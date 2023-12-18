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

	"sigs.k8s.io/controller-runtime/pkg/log"

	bsv1alpha1 "janus-idp.io/backstage-operator/api/v1alpha1"

	"janus-idp.io/backstage-operator/pkg/utils"
)

const backstageAppLabel = "backstage.io/app"

// Backstage configuration scaffolding with empty BackstageObjects.
// Here're all possible objects for configuration, can be:
// Mandatory - Backstage Deployment (Pod), Service
// Optional - mostly (but not only) Bckstage Pod configuration objects (AppConfig, ExtraConfig)
// ForLocalDatabase - mandatory if EnabledLocalDb, ignored otherwise
// ForOpenshift - if configured, used for Openshift deployment, ignored otherwise
var runtimeConfig = []ObjectConfig{
	{Key: "deployment.yaml", ObjectFactory: BackstageDeploymentFactory{}, need: Mandatory},
	{Key: "service.yaml", ObjectFactory: BackstageServiceFactory{}, need: Mandatory},
	{Key: "db-statefulset.yaml", ObjectFactory: DbStatefulSetFactory{}, need: ForLocalDatabase},
	{Key: "db-service.yaml", ObjectFactory: DbServiceFactory{}, need: ForLocalDatabase},
	{Key: "db-secret.yaml", ObjectFactory: DbSecretFactory{}, need: ForLocalDatabase},
	{Key: "app-config.yaml", ObjectFactory: AppConfigFactory{}, need: Optional},
	//{Key: "configmap-files.yaml", ObjectFactory: newBackstageDeployment(), need: Optional},
	//{Key: "secret-files.yaml", BackstageObject: newBackstageDeployment(), need: Optional},
	//{Key: "configmap-envs.yaml", BackstageObject: newBackstageDeployment(), need: Optional},
	//{Key: "secret-envs.yaml", BackstageObject: newBackstageDeployment(), need: Optional},
	{Key: "route.yaml", ObjectFactory: BackstageRouteFactory{}, need: ForOpenshift},
}

// internal object model to simplify management dealing with structured objects
type runtimeModel struct {
	backstageDeployment *BackstageDeployment
	backstageService    *BackstageService

	localDbStatefulSet *DbStatefulSet
	localDbService     *DbService
	localDbSecret      *DbSecret
}

// Main loop for configuring and making the array of objects to reconsile
func InitObjects(ctx context.Context, backstageMeta bsv1alpha1.Backstage, backstageSpec *DetailedBackstageSpec, ownsRuntime bool, isOpenshift bool) ([]BackstageObject, error) {

	// 3 phases of Backstage configuration:
	// 1- load from Operator defaults, modify metadata (labels, selectors..) and namespace as needed
	// 2- overlay some/all objects with Backstage.spec.rawRuntimeConfig CM
	// 3- override some parameters defined in Backstage.spec.application
	// At the end there should be an array of runtime Objects to apply (order optimized)

	lg := log.FromContext(ctx)

	objectList := make([]BackstageObject, 0)
	runtimeModel := &runtimeModel{}

	for _, conf := range runtimeConfig {

		backstageObject := conf.ObjectFactory.newBackstageObject()
		var defaultErr error
		var overlayErr error

		// read default configuration
		if err := utils.ReadYamlFile(utils.DefFile(conf.Key), backstageObject.Object()); err != nil {
			defaultErr = fmt.Errorf("failed to read default value for the key %s, reason: %s", conf.Key, err)
			//lg.V(1).Info("failed reading default config", "error", err.Error())
		}

		// overlay with or add rawConfig
		overlay, overlayExist := backstageSpec.Details.RawConfig[conf.Key]
		if overlayExist {
			if err := utils.ReadYaml([]byte(overlay), backstageObject.Object()); err != nil {
				overlayErr = fmt.Errorf("failed to read overlay value for the key %s, reason: %s", conf.Key, err)
			}
		}

		if overlayErr != nil || (!overlayExist && defaultErr != nil) {
			if conf.need == Mandatory || (conf.need == ForLocalDatabase && *backstageSpec.EnableLocalDb) {
				return nil, errors.Join(defaultErr, overlayErr)
			} else {
				lg.V(1).Info("failed to read default value for optional key. Ignored \n", conf.Key, errors.Join(defaultErr, overlayErr))
				continue
			}
		}

		// do not add if local db is disabled
		if !backstageSpec.LocalDbEnabled() && conf.need == ForLocalDatabase {
			continue
		}

		// do not add if not openshift
		if !isOpenshift && conf.need == ForOpenshift {
			continue
		}

		// populate BackstageObject metainfo (names, labels, selsctors etc) for consistency
		backstageObject.initMetainfo(backstageMeta, ownsRuntime)

		// finally add the object to the model and list
		backstageObject.addToModel(runtimeModel)
		objectList = append(objectList, backstageObject)
	}

	// update local-db conf objects
	if backstageSpec.LocalDbEnabled() {
		for _, bso := range objectList {
			if ldco, ok := bso.(LocalDbConfObject); ok {
				ldco.updateLocalDbPod(runtimeModel)
			}
		}
	}

	// create Backstage Pod object
	if runtimeModel.backstageDeployment == nil {
		return nil, fmt.Errorf("failed to identify Backstage Deployment by %s, it should not happen normally", "deployment.xml")
	}
	backstagePod, err := newBackstagePod(runtimeModel.backstageDeployment)
	if err != nil {
		return nil, fmt.Errorf("failed to create Backstage Pod: %s", err)
	}

	// update Backstage Pod with parts (volumes, container)
	// according to default configuration
	for _, bso := range objectList {
		if bs, ok := bso.(BackstageConfObject); ok {
			bs.updateBackstagePod(backstagePod)
		}
	}

	// Phase 3: process Backstage.spec
	if backstageSpec.Application != nil {
		runtimeModel.backstageDeployment.setReplicas(backstageSpec.Application.Replicas)
		backstagePod.appendImagePullSecrets(backstageSpec.Application.ImagePullSecrets)
		backstagePod.setImage(backstageSpec.Application.Image)
	}
	// TODO API
	if backstageSpec.Details.AppConfigs != nil {
		for _, ac := range backstageSpec.Details.AppConfigs {
			backstagePod.addAppConfig(ac.ConfigMapName, ac.FilePath)
		}
	}

	return objectList, nil
}

// Every BackstageObject.initMetainfo should as minimum call this
func initMetainfo(modelObject BackstageObject, backstageMeta bsv1alpha1.Backstage, ownsRuntime bool) {
	modelObject.Object().SetNamespace(backstageMeta.Namespace)
	modelObject.Object().SetLabels(utils.SetKubeLabels(modelObject.Object().GetLabels(), backstageMeta.Name))
	//if ownsRuntime {
	//if err = controllerutil.SetControllerReference(&backstageMeta, modelObject.Object(), r.Scheme); err != nil {
	//	//return fmt.Errorf("failed to set owner reference: %s", err)
	//}
}
