package model

import (
	"context"
	"fmt"

	bsv1alpha1 "janus-idp.io/backstage-operator/api/v1alpha1"

	"janus-idp.io/backstage-operator/pkg/utils"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const deploymentKey = "deployment.yaml"
const backstageAppLabel = "backstage.io/app"

const (
	Mandatory        needType = "Mandatory"
	NotMandatory     needType = "Optional"
	ForLocalDatabase needType = "ForLocalDatabase"
	ForOpenshift     needType = "ForOpenshift"
)

var runtimeConfig = []ObjectConfig{
	{Key: deploymentKey, BackstageObject: newBackstageDeployment(), need: Mandatory},
	{Key: "service.yaml", BackstageObject: newBackstageService(), need: Mandatory},
	{Key: "db-statefulset.yaml", BackstageObject: newDbStatefulSet(), need: ForLocalDatabase},
	{Key: "db-service.yaml", BackstageObject: newDbService(), need: ForLocalDatabase},
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
}

type BackstageConfObject interface {
	BackstageObject
	updateBackstagePod(pod *backstagePod)
}

func (c *ObjectConfig) isEmpty() bool {
	return c.BackstageObject == nil
}

//type RuntimeModel struct {
//	BackstageDeployment      appsv1.Deployment
//	BackstageService         corev1.Service
//	AppConfig               corev1.ConfigMap
//	ExtraConfigMapToFiles   corev1.ConfigMap
//	ExtraConfigMapToEnvVars corev1.ConfigMap
//	ExtraSecretToFiles      corev1.Secret
//	ExtraSecretToEnvVars    corev1.Secret
//	ExtraEnvVars             map[string]string
//
//	LocalDbStatefulSet appsv1.StatefulSet
//	LocalDbService     corev1.Service
//
//	NetworkingRoute   openshift.Route
//	NetworkingIngress networkingv1.Ingress
//}

func InitObjects(ctx context.Context, backstageMeta bsv1alpha1.Backstage, backstageSpec *DetailedBackstageSpec, ownsRuntime bool, isOpenshift bool) ([]BackstageObject, error) {

	// 3 phases of Backstage configuration:
	// 1- load from Operator defaults, modify metadata (labels, selectors..) and namespace as needed
	// 2- overlay some/all objects with Backstage.spec.rawRuntimeConfig CM
	// 3- override some parameters defined in Backstage.spec.application
	// At the end there should be an array of runtime Objects to apply (order optimized)

	lg := log.FromContext(ctx)

	runtimeModel := make([]BackstageObject, 0)
	var backstageDeployment *BackstageDeployment
	var backstagePod *backstagePod
	// Phase 1:
	for _, conf := range runtimeConfig {
		backstageObject := conf.BackstageObject
		if err := utils.ReadYamlFile(utils.DefFile(conf.Key), backstageObject.Object()); err != nil {
			if conf.need == Mandatory || (conf.need == ForLocalDatabase && *backstageSpec.EnableLocalDb) {
				return nil, err
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

		// do not add if not openshift
		if !isOpenshift && conf.need == ForOpenshift {
			continue
		}

		backstageObject.initMetainfo(backstageMeta, ownsRuntime)

		if conf.Key == deploymentKey {
			backstageDeployment = backstageObject.(*BackstageDeployment)
			//(backstageObject.Object()).(*appsv1.Deployment)
		}
		runtimeModel = append(runtimeModel, backstageObject)
	}

	// initialize Backstage Pod object
	if backstageDeployment == nil {
		return nil, fmt.Errorf("failed to identify Backstage Deployment by %s, it should not happen normally", deploymentKey)
	} else {
		backstagePod = newBackstagePod(backstageDeployment.deployment)
		backstageDeployment.pod = backstagePod
	}

	// update Backstage Pod with parts (volume, container, volumeMounts)
	for _, bso := range runtimeModel {
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

	return runtimeModel, nil
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
