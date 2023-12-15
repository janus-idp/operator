package model

import (
	"context"
	"fmt"

	bsv1alpha1 "janus-idp.io/backstage-operator/api/v1alpha1"

	"janus-idp.io/backstage-operator/pkg/utils"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

//const (
//	deploymentKey = "deployment.yaml"
//	dbDeploymentKey = "db-statefulset.yaml"
//)

const backstageAppLabel = "backstage.io/app"

const (
	Mandatory        needType = "Mandatory"
	NotMandatory     needType = "Optional"
	ForLocalDatabase needType = "ForLocalDatabase"
	ForOpenshift     needType = "ForOpenshift"
)

var runtimeConfig = []ObjectConfig{
	{Key: "deployment.yaml", BackstageObject: newBackstageDeployment(), need: Mandatory},
	{Key: "service.yaml", BackstageObject: newBackstageService(), need: Mandatory},
	{Key: "db-statefulset.yaml", BackstageObject: newDbStatefulSet(), need: ForLocalDatabase},
	{Key: "db-service.yaml", BackstageObject: newDbService(), need: ForLocalDatabase},
	{Key: "db-secret.yaml", BackstageObject: newDbSecret(), need: ForLocalDatabase},
	{Key: "app-config.yaml", BackstageObject: newAppConfig(), need: NotMandatory},
	{Key: "configmap-files.yaml", BackstageObject: newBackstageDeployment(), need: NotMandatory},
	{Key: "secret-files.yaml", BackstageObject: newBackstageDeployment(), need: NotMandatory},
	{Key: "configmap-envs.yaml", BackstageObject: newBackstageDeployment(), need: NotMandatory},
	{Key: "secret-envs.yaml", BackstageObject: newBackstageDeployment(), need: NotMandatory},
	{Key: "route.yaml", BackstageObject: newRoute(), need: ForOpenshift},
}

type needType string

type ObjectConfig struct {
	BackstageObject BackstageObject
	Key             string
	need            needType
}

type BackstageObject interface {
	Object() client.Object
	initMetainfo(backstageMeta bsv1alpha1.Backstage, ownsRuntime bool)
	addToModel(model *runtimeModel)
}

type BackstageConfObject interface {
	BackstageObject
	updateBackstagePod(pod *backstagePod)
}

// internal object model to simplify management dealing with structured objects
type runtimeModel struct {
	backstageDeployment *BackstageDeployment
	backstageService    *BackstageService

	localDbStatefulSet *DbStatefulSet
	localDbService     *DbService
	localDbSecret      *DbSecret
}

func InitObjects(ctx context.Context, backstageMeta bsv1alpha1.Backstage, backstageSpec *DetailedBackstageSpec, ownsRuntime bool, isOpenshift bool) ([]BackstageObject, error) {

	// 3 phases of Backstage configuration:
	// 1- load from Operator defaults, modify metadata (labels, selectors..) and namespace as needed
	// 2- overlay some/all objects with Backstage.spec.rawRuntimeConfig CM
	// 3- override some parameters defined in Backstage.spec.application
	// At the end there should be an array of runtime Objects to apply (order optimized)

	lg := log.FromContext(ctx)

	objectList := make([]BackstageObject, 0)
	runtimeModel := &runtimeModel{}

	//var backstageDeployment *BackstageDeployment
	//var localDbDeployment *DbStatefulSet
	// Phase 1:
	for _, conf := range runtimeConfig {
		backstageObject := conf.BackstageObject
		if err := utils.ReadYamlFile(utils.DefFile(conf.Key), backstageObject.Object()); err != nil {
			if conf.need == Mandatory || (conf.need == ForLocalDatabase && *backstageSpec.EnableLocalDb) {
				return nil, fmt.Errorf("failed to read default value for the key %s, reason: %s", conf.Key, err)
			} else {
				lg.Info("failed to read default value for optional key. Ignored \n", conf.Key, err)
				continue
			}
		}

		// Phase 2: overlay with rawConfig if any
		overlay, ok := backstageSpec.Details.RawConfig[conf.Key]
		if ok {
			if err := utils.ReadYaml([]byte(overlay), backstageObject.Object()); err != nil {
				// consider all values set intentionally, "need" ignored, always throw error
				return nil, fmt.Errorf("failed to read overlay value for the key %s, reason: %s", conf.Key, err)
			}
		}

		// do not add if not for local db
		if !*backstageSpec.EnableLocalDb && conf.need == ForLocalDatabase {
			continue
		}

		// do not add if not openshift
		if !isOpenshift && conf.need == ForOpenshift {
			continue
		}

		backstageObject.initMetainfo(backstageMeta, ownsRuntime)

		// finally add the object to the model and list
		backstageObject.addToModel(runtimeModel)
		objectList = append(objectList, backstageObject)
	}

	// create Backstage Pod object
	if runtimeModel.backstageDeployment == nil {
		return nil, fmt.Errorf("failed to identify Backstage Deployment by %s, it should not happen normally", "deployment.xml")
	}
	backstagePod, err := newBackstagePod(runtimeModel.backstageDeployment)
	if err != nil {
		return nil, fmt.Errorf("failed to create Backstage Pod: %s", err)
	}

	// update local-db-secret
	if *backstageSpec.EnableLocalDb {
		err := runtimeModel.localDbSecret.updateSecret(runtimeModel.backstageDeployment, runtimeModel.localDbStatefulSet,
			runtimeModel.localDbService)
		if err != nil {
			return nil, fmt.Errorf("failed to update LocalDb Secret: %s", err)
		}
	}

	// update Backstage Pod with parts (volumes, container)
	// according to default configuration
	for _, bso := range objectList {
		if bs, ok := bso.(BackstageConfObject); ok {
			bs.updateBackstagePod(backstagePod)
		}
	}

	// Phase 3: process Backstage.spec
	// TODO API
	//backstageDeployment.setReplicas(backstageSpec.replicas)
	//backstagePod.addImagePullSecrets(backstageSpec.imagePullSecrets)
	//backstagePod.container.setImage(backstageSpec.image)

	// TODO API
	//if backstageSpec.AppConfigs != nil {
	//	for _, ac := range backstageSpec.AppConfigs {
	//		backstagePod.addAppConfig(ac.Name, ac.FilePath)
	//	}
	//}

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
