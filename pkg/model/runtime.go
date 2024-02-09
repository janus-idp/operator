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
	"os"
	"reflect"

	"sigs.k8s.io/controller-runtime/pkg/client"

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
	localDbEnabled bool
	isOpenshift    bool

	backstageDeployment *BackstageDeployment
	backstageService    *BackstageService

	localDbStatefulSet *DbStatefulSet
	LocalDbService     *DbService
	LocalDbSecret      *DbSecret

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
func registerConfig(key string, factory ObjectFactory) {
	runtimeConfig = append(runtimeConfig, ObjectConfig{Key: key, ObjectFactory: factory /*, need: need*/})
}

// InitObjects performs a main loop for configuring and making the array of objects to reconcile
func InitObjects(ctx context.Context, backstage bsv1alpha1.Backstage, backstageSpec *DetailedBackstageSpec, ownsRuntime bool, isOpenshift bool, scheme *runtime.Scheme) (*BackstageModel, error) {

	// 3 phases of Backstage configuration:
	// 1- load from Operator defaults, modify metadata (labels, selectors..) and namespace as needed
	// 2- overlay some/all objects with Backstage.spec.rawRuntimeConfig CM
	// 3- override some parameters defined in Backstage.spec.application
	// At the end there should be an array of runtime RuntimeObjects to apply (order optimized)

	lg := log.FromContext(ctx)
	lg.V(1)

	model := &BackstageModel{RuntimeObjects: make([]RuntimeObject, 0), localDbEnabled: backstageSpec.IsLocalDbEnabled(), isOpenshift: isOpenshift}

	// looping through the registered runtimeConfig objects initializing the model
	for _, conf := range runtimeConfig {

		// creating the instance of backstageObject
		backstageObject := conf.ObjectFactory.newBackstageObject()

		var obj client.Object = backstageObject.EmptyObject()
		if err := utils.ReadYamlFile(utils.DefFile(conf.Key), obj); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return nil, fmt.Errorf("failed to read default value for the key %s, reason: %s", conf.Key, err)
			}
		} else {
			backstageObject.setObject(obj, backstage.Name)
		}

		// reading configuration defined in BackstageCR.Spec.RawConfigContent ConfigMap
		// if present, backstageObject's default configuration will be overridden
		overlay, overlayExist := backstageSpec.RawConfigContent[conf.Key]
		if overlayExist {
			if err := utils.ReadYaml([]byte(overlay), obj); err != nil {
				return nil, fmt.Errorf("failed to read overlay value for the key %s, reason: %s", conf.Key, err)
			} else {
				backstageObject.setObject(obj, backstage.Name)
			}
		}

		// apply spec and add the object to the model and list
		if err := backstageObject.addToModel(model, backstage, ownsRuntime); err != nil {
			return nil, fmt.Errorf("failed to initialize %s reason: %s", backstageObject, err)
		}
	}
	//////////////////////
	// init default meta info (name, namespace, owner) and update Backstage Pod with contributions (volumes, container)
	for _, bso := range model.RuntimeObjects {
		if bs, ok := bso.(PodContributor); ok {
			bs.updatePod(model.backstageDeployment.pod)
		}
	}

	if backstageSpec.IsLocalDbEnabled() {
		model.localDbStatefulSet.setDbEnvsFromSecret(model.LocalDbSecret.secret.Name)
		//model.backstageDeployment.pod.setEnvsFromSecret(model.LocalDbSecret.secret.Name)
	}

	// contribute to Backstage config
	for _, v := range backstageSpec.ConfigObjects {
		v.updatePod(model.backstageDeployment.pod)
	}
	/////////////////

	// set generic metainfo and validate all
	for _, v := range model.RuntimeObjects {
		setMetaInfo(v, backstage, ownsRuntime, scheme)
		err := v.validate(model, backstage)
		if err != nil {
			return nil, fmt.Errorf("failed object validation, reason: %s", err)
		}
	}

	return model, nil
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
