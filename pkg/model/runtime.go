//
// Copyright (c) 2023 Red Hat, Inc.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package model

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"sigs.k8s.io/controller-runtime/pkg/log"

	bsv1alpha1 "janus-idp.io/backstage-operator/api/v1alpha1"

	"janus-idp.io/backstage-operator/pkg/utils"
)

const backstageAppLabel = "backstage.io/app"

// Backstage configuration scaffolding with empty BackstageObjects.
// There are all possible objects for configuration, can be:
// Mandatory - Backstage Deployment (Pod), Service
// Optional - mostly (but not only) Backstage Pod configuration objects (AppConfig, ExtraConfig)
// ForLocalDatabase - mandatory if EnabledLocalDb, ignored otherwise
// ForOpenshift - if configured, used for Openshift deployment, ignored otherwise
var runtimeConfig []ObjectConfig

// RuntimeModel represents internal object model
type RuntimeModel struct {
	backstageDeployment *BackstageDeployment
	backstageService    *BackstageService

	localDbStatefulSet *DbStatefulSet
	localDbService     *DbService
	localDbSecret      *DbSecret

	route *BackstageRoute

	Objects []BackstageObject
}

func (model *RuntimeModel) setObject(object BackstageObject) {
	for i, obj := range model.Objects {
		if reflect.TypeOf(obj) == reflect.TypeOf(object) {
			model.Objects[i] = object
			return
		}
	}
	model.Objects = append(model.Objects, object)
}

// Registers config object
func registerConfig(key string, factory ObjectFactory, need needType) {
	runtimeConfig = append(runtimeConfig, ObjectConfig{Key: key, ObjectFactory: factory, need: need})
}

// InitObjects performs a main loop for configuring and making the array of objects to reconcile
func InitObjects(ctx context.Context, backstageMeta bsv1alpha1.Backstage, backstageSpec *DetailedBackstageSpec, ownsRuntime bool, isOpenshift bool, scheme *runtime.Scheme) (*RuntimeModel, error) {

	// 3 phases of Backstage configuration:
	// 1- load from Operator defaults, modify metadata (labels, selectors..) and namespace as needed
	// 2- overlay some/all objects with Backstage.spec.rawRuntimeConfig CM
	// 3- override some parameters defined in Backstage.spec.application
	// At the end there should be an array of runtime Objects to apply (order optimized)

	lg := log.FromContext(ctx)
	lg.V(1)

	model := &RuntimeModel{Objects: make([]BackstageObject, 0) /*, generateDbPassword: backstageSpec.GenerateDbPassword*/}

	if err := model.addDefaultsAndRaw(backstageMeta, backstageSpec, ownsRuntime, isOpenshift); err != nil {
		return nil, fmt.Errorf("failed to initialize objects %w", err)
	}

	if model.backstageDeployment == nil {
		return nil, fmt.Errorf("failed to identify Backstage Deployment by %s, it should not happen normally", "deployment.yaml")
	}
	if backstageSpec.IsLocalDbEnabled() && model.localDbStatefulSet == nil {
		return nil, fmt.Errorf("failed to identify Local DB StatefulSet by %s, it should not happen normally", "db-statefulset.yaml")
	}

	// create Backstage Pod object
	backstagePod, err := newBackstagePod(model.backstageDeployment)
	if err != nil {
		return nil, fmt.Errorf("failed to create Backstage Pod: %s", err)
	}

	// init default meta info (name, namespace, owner) and update Backstage Pod with contributions (volumes, container)
	for _, bso := range model.Objects {
		if bs, ok := bso.(BackstagePodContributor); ok {
			bs.updateBackstagePod(backstagePod)
		}
	}

	// Phase 3: process Backstage.spec, getting final desired state
	if backstageSpec.Application != nil {
		model.backstageDeployment.setReplicas(backstageSpec.Application.Replicas)
		backstagePod.setImagePullSecrets(backstageSpec.Application.ImagePullSecrets)
		backstagePod.setImage(backstageSpec.Application.Image)

		backstagePod.addExtraEnvs(backstageSpec.Application.ExtraEnvs)

		//if backstageSpec.Application.ExtraEnvs != nil {
		//	for _, e := range backstageSpec.Application.ExtraEnvs.Envs {
		//		backstagePod.addContainerEnvVar(e)
		//	}
		//}
	}
	// Route...
	if isOpenshift && backstageSpec.IsRouteEnabled() {
		newBackstageRoute(*backstageSpec.Application.Route).addToModel(model, backstageMeta, ownsRuntime)
	}

	// Local DB Secret...
	// if exists - initiated from existed, otherwise:
	//  if specified - get from spec
	//  if not specified - generate
	if backstageSpec.IsLocalDbEnabled() {
		backstageSpec.LocalDbSecret.addToModel(model, backstageMeta, ownsRuntime)
		backstageSpec.LocalDbSecret.updateSecret(model)
	}

	// contribute to Backstage config
	for _, v := range backstageSpec.ConfigObjects {
		v.updateBackstagePod(backstagePod)
	}

	// set generic metainfo and validate all
	for _, v := range model.Objects {
		setMetaInfo(v, backstageMeta, ownsRuntime, scheme)
		err := v.validate(model)
		if err != nil {
			return nil, fmt.Errorf("failed object validation, reason: %s", err)
		}
	}

	return model, nil
}

func (model *RuntimeModel) addDefaultsAndRaw(backstageMeta bsv1alpha1.Backstage, backstageSpec *DetailedBackstageSpec, ownsRuntime bool, isOpenshift bool) error {
	// looping through the registered runtimeConfig objects initializing the model
	for _, conf := range runtimeConfig {

		// creating the instance of backstageObject
		backstageObject := conf.ObjectFactory.newBackstageObject()
		var defaultErr error
		var overlayErr error

		// reading default configuration defined in the default-config/[key] file
		// mounted from the 'default-config' ConfigMap
		// this is a cluster scope configuration applying to every Backstage CR by default
		if err := utils.ReadYamlFile(utils.DefFile(conf.Key), backstageObject.Object()); err != nil {
			defaultErr = fmt.Errorf("failed to read default value for the key %s, reason: %s", conf.Key, err)
			//lg.V(1).Info("failed reading default config", "error", err.Error())
		}

		// reading configuration defined in BackstageCR.Spec.RawConfigContent ConfigMap
		// if present, backstageObject's default configuration will be overridden
		overlay, overlayExist := backstageSpec.RawConfigContent[conf.Key]
		if overlayExist {
			if err := utils.ReadYaml([]byte(overlay), backstageObject.Object()); err != nil {
				overlayErr = fmt.Errorf("failed to read overlay value for the key %s, reason: %s", conf.Key, err)
			}
		}

		// throw the error if raw configuration exists and is invalid
		// throw the error if there is invalid or no configuration (default|raw) for Mandatory object
		// continue if there is invalid or no configuration (default|raw) for Optional object
		// TODO separate the case when configuration does not exist (intentionally) from invalid configuration
		if overlayErr != nil || (!overlayExist && defaultErr != nil) {
			if conf.need == Mandatory || (conf.need == ForLocalDatabase && backstageSpec.IsLocalDbEnabled()) {
				return errors.Join(defaultErr, overlayErr)
			} else {
				//lg.V(1).Info("failed to read default value for optional key. Ignored \n", conf.Key, errors.Join(defaultErr, overlayErr))
				continue
			}
		}

		// do not add if ForLocalDatabase and LocalDb is disabled
		if !backstageSpec.IsLocalDbEnabled() && conf.need == ForLocalDatabase {
			continue
		}

		// do not add if ForOpenshift and cluster is not Openshift
		if !isOpenshift && conf.need == ForOpenshift {
			continue
		}

		// finally add the object to the model and list
		backstageObject.addToModel(model, backstageMeta, ownsRuntime)
	}

	return nil
}

// Every BackstageObject.setMetaInfo should as minimum call this
func setMetaInfo(modelObject BackstageObject, backstageMeta bsv1alpha1.Backstage, ownsRuntime bool, scheme *runtime.Scheme) {
	modelObject.Object().SetNamespace(backstageMeta.Namespace)
	modelObject.Object().SetLabels(utils.SetKubeLabels(modelObject.Object().GetLabels(), backstageMeta.Name))

	if ownsRuntime {
		err := controllerutil.SetControllerReference(&backstageMeta, modelObject.Object(), scheme)
		if err != nil {
			//error should never have happened,
			//otherwise the Operator has invalid (not a runtime.Object) or non-registered type.
			//In both cases it will fail before this place
			panic(err)
		}
	}

}
