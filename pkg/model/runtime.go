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

	openshift "github.com/openshift/api/route/v1"

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

// BackstageModel represents internal object model
type BackstageModel struct {
	backstageDeployment *BackstageDeployment
	backstageService    *BackstageService

	localDbStatefulSet *DbStatefulSet
	LocalDbService     *DbService
	//LocalDbSecret      *DbSecret

	route *BackstageRoute

	RuntimeObjects []RuntimeObject
}

func (model *BackstageModel) setRuntimeObject(object RuntimeObject) {
	for i, obj := range model.RuntimeObjects {
		if reflect.TypeOf(obj) == reflect.TypeOf(object) {
			model.RuntimeObjects[i] = object
			return
		}
	}
	model.RuntimeObjects = append(model.RuntimeObjects, object)
}

// Registers config object
func registerConfig(key string, factory ObjectFactory, need needType) {
	runtimeConfig = append(runtimeConfig, ObjectConfig{Key: key, ObjectFactory: factory, need: need})
}

// InitObjects performs a main loop for configuring and making the array of objects to reconcile
func InitObjects(ctx context.Context, backstageMeta bsv1alpha1.Backstage, backstageSpec *DetailedBackstageSpec, ownsRuntime bool, isOpenshift bool, scheme *runtime.Scheme) (*BackstageModel, error) {

	// 3 phases of Backstage configuration:
	// 1- load from Operator defaults, modify metadata (labels, selectors..) and namespace as needed
	// 2- overlay some/all objects with Backstage.spec.rawRuntimeConfig CM
	// 3- override some parameters defined in Backstage.spec.application
	// At the end there should be an array of runtime RuntimeObjects to apply (order optimized)

	lg := log.FromContext(ctx)
	lg.V(1)

	model := &BackstageModel{RuntimeObjects: make([]RuntimeObject, 0) /*, generateDbPassword: backstageSpec.GenerateDbPassword*/}

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
	for _, bso := range model.RuntimeObjects {
		if bs, ok := bso.(PodContributor); ok {
			bs.updatePod(backstagePod)
		}
	}

	// Phase 3: process Backstage.spec, getting final desired state
	if backstageSpec.Application != nil {
		model.backstageDeployment.setReplicas(backstageSpec.Application.Replicas)
		backstagePod.setImagePullSecrets(backstageSpec.Application.ImagePullSecrets)
		backstagePod.setImage(backstageSpec.Application.Image)

		backstagePod.addExtraEnvs(backstageSpec.Application.ExtraEnvs)

	}

	// Route...
	// TODO: nicer proccessing
	if isOpenshift && backstageSpec.IsRouteEnabled() && !backstageSpec.IsRouteEmpty() {
		if model.route == nil {
			br := BackstageRoute{route: &openshift.Route{}}
			br.addToModel(model, backstageMeta, ownsRuntime)
		}
		model.route.patchRoute(*backstageSpec.Application.Route)
	}

	// Local DB Secret...
	// if exists - initiated from existed, otherwise:
	//  if specified - get from spec
	//  if not specified - generate
	// TODO
	var dbSecretName string
	if !backstageSpec.IsAuthSecretSpecified() {
		dbSecretName = DbSecretDefaultName(backstageMeta.Name)
	} else {
		dbSecretName = backstageSpec.Database.AuthSecretName
	}
	if backstageSpec.IsLocalDbEnabled() {
		model.localDbStatefulSet.setDbEnvsFromSecret(dbSecretName)
	}
	backstagePod.setEnvsFromSecret(dbSecretName)

	// contribute to Backstage config
	for _, v := range backstageSpec.ConfigObjects {
		v.updatePod(backstagePod)
	}

	// set generic metainfo and validate all
	for _, v := range model.RuntimeObjects {
		setMetaInfo(v, backstageMeta, ownsRuntime, scheme)
		err := v.validate(model)
		if err != nil {
			return nil, fmt.Errorf("failed object validation, reason: %s", err)
		}
	}

	return model, nil
}

func (model *BackstageModel) addDefaultsAndRaw(backstageMeta bsv1alpha1.Backstage, backstageSpec *DetailedBackstageSpec, ownsRuntime bool, isOpenshift bool) error {
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

		// do not add if ForOpenshift and (cluster is not Openshift OR route is not enabled in CR)
		if conf.need == ForOpenshift && (!isOpenshift || !backstageSpec.IsRouteEnabled()) {
			continue
		}

		// finally add the object to the model and list
		backstageObject.addToModel(model, backstageMeta, ownsRuntime)
	}

	return nil
}

// Every RuntimeObject.setMetaInfo should as minimum call this
func setMetaInfo(modelObject RuntimeObject, backstageMeta bsv1alpha1.Backstage, ownsRuntime bool, scheme *runtime.Scheme) {
	modelObject.Object().SetNamespace(backstageMeta.Namespace)
	modelObject.Object().SetLabels(utils.SetKubeLabels(modelObject.Object().GetLabels(), backstageMeta.Name))

	if ownsRuntime {
		if err := controllerutil.SetControllerReference(&backstageMeta, modelObject.Object(), scheme); err != nil {
			//error should never have happened,
			//otherwise the Operator has invalid (not a runtime.Object) or non-registered type.
			//In both cases it will fail before this place
			panic(err)
		}
	}

}
