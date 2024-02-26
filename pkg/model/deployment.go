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
	"fmt"
	"os"

	corev1 "k8s.io/api/core/v1"

	bsv1alpha1 "janus-idp.io/backstage-operator/api/v1alpha1"

	"janus-idp.io/backstage-operator/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const BackstageImageEnvVar = "RELATED_IMAGE_backstage"
const defaultMountDir = "/opt/app-root/src"

type BackstageDeploymentFactory struct{}

func (f BackstageDeploymentFactory) newBackstageObject() RuntimeObject {
	return &BackstageDeployment{}
}

type BackstageDeployment struct {
	deployment *appsv1.Deployment
	//	pod        *backstagePod
}

func init() {
	registerConfig("deployment.yaml", BackstageDeploymentFactory{})
}

func DeploymentName(backstageName string) string {
	return utils.GenerateRuntimeObjectName(backstageName, "deployment")
}

// implementation of RuntimeObject interface
func (b *BackstageDeployment) Object() client.Object {
	return b.deployment
}

func (b *BackstageDeployment) setObject(obj client.Object, backstageName string) {
	b.deployment = nil
	if obj != nil {
		b.deployment = obj.(*appsv1.Deployment)
	}
}

// implementation of RuntimeObject interface
func (b *BackstageDeployment) EmptyObject() client.Object {
	return &appsv1.Deployment{}
}

// implementation of RuntimeObject interface
func (b *BackstageDeployment) addToModel(model *BackstageModel, backstage bsv1alpha1.Backstage, ownsRuntime bool) (bool, error) {
	if b.deployment == nil {
		return false, fmt.Errorf("Backstage Deployment is not initialized, make sure there is deployment.yaml in default or raw configuration")
	}
	model.backstageDeployment = b
	model.setRuntimeObject(b)

	if backstage.Spec.Application != nil {
		b.setReplicas(backstage.Spec.Application.Replicas)
		b.setImagePullSecrets(backstage.Spec.Application.ImagePullSecrets)
		b.setImage(backstage.Spec.Application.Image)
		b.addExtraEnvs(backstage.Spec.Application.ExtraEnvs)
	}

	// override image with env var
	// [GA] TODO Do we need this feature?
	if os.Getenv(BackstageImageEnvVar) != "" {
		b.deployment.Spec.Template.Spec.Containers[0].Image = os.Getenv(BackstageImageEnvVar)
		// TODO workaround for the (janus-idp, rhdh) case where we have
		// exactly the same image for initContainer and want it to be overriden
		// the same way as Backstage's one
		for i := range b.deployment.Spec.Template.Spec.InitContainers {
			b.deployment.Spec.Template.Spec.InitContainers[i].Image = os.Getenv(BackstageImageEnvVar)
		}
	}

	return true, nil
}

// implementation of RuntimeObject interface
func (b *BackstageDeployment) validate(model *BackstageModel, backstage bsv1alpha1.Backstage) error {
	for _, bso := range model.RuntimeObjects {
		if bs, ok := bso.(BackstagePodContributor); ok {
			bs.updatePod(b.deployment)
		}
	}

	if backstage.Spec.Application != nil {
		application := backstage.Spec.Application
		// AppConfig
		if application.AppConfig != nil {
			mountPath := application.AppConfig.MountPath
			for _, cm := range model.appConfigs {
				newAppConfig(mountPath, &cm.ConfigMap, cm.Key).updatePod(b.deployment)
			}

			//for _, spec := range application.AppConfig.ConfigMaps {
			//	configMap, err := getAppConfigMap(spec.Name, spec.Key, model.appConfigs)
			//	if err != nil {
			//		return fmt.Errorf("app-config configuration failed %w", err)
			//	}
			//	newAppConfig(mountPath, configMap, spec.Key).updatePod(b.deployment)
			//}
		}

		//DynaPlugins
		if application.DynamicPluginsConfigMapName != "" {
			if dynamicPluginsInitContainer(b.deployment.Spec.Template.Spec.InitContainers) == nil {
				return fmt.Errorf("deployment validation failed, dynamic plugin name configured but no InitContainer %s defined", dynamicPluginInitContainerName)
			}
			newDynamicPlugins(application.DynamicPluginsConfigMapName).updatePod(b.deployment)
		}
		//Ext (4)
		if application.ExtraFiles != nil {
			mountPath := application.ExtraFiles.MountPath
			for _, spec := range application.ExtraFiles.ConfigMaps {
				newConfigMapFiles(mountPath, spec.Name, spec.Key).updatePod(b.deployment)
			}
			for _, spec := range application.ExtraFiles.Secrets {
				newSecretFiles(mountPath, spec.Name, spec.Key).updatePod(b.deployment)
			}
		}
		if application.ExtraEnvs != nil {
			for _, spec := range application.ExtraEnvs.ConfigMaps {
				newConfigMapEnvs(spec.Name, spec.Key).updatePod(b.deployment)
			}
			for _, spec := range application.ExtraEnvs.Secrets {
				newSecretEnvs(spec.Name, spec.Key).updatePod(b.deployment)
			}
		}
	}

	//DbSecret
	if model.LocalDbSecret != nil {
		utils.AddEnvVarsFrom(&b.deployment.Spec.Template.Spec.Containers[0], utils.SecretObjectKind,
			model.LocalDbSecret.secret.Name, "")
		//b.pod.setEnvsFromSecret(model.LocalDbSecret.secret.Name)
	}

	return nil
}

func (b *BackstageDeployment) setMetaInfo(backstageName string) {
	b.deployment.SetName(DeploymentName(backstageName))
	utils.GenerateLabel(&b.deployment.Spec.Template.ObjectMeta.Labels, backstageAppLabel, fmt.Sprintf("backstage-%s", backstageName))
	utils.GenerateLabel(&b.deployment.Spec.Selector.MatchLabels, backstageAppLabel, fmt.Sprintf("backstage-%s", backstageName))
}

func (b *BackstageDeployment) container() *corev1.Container {
	return &b.deployment.Spec.Template.Spec.Containers[0]
}

// sets the amount of replicas (used by CR config)
func (b *BackstageDeployment) setReplicas(replicas *int32) {
	if replicas != nil {
		b.deployment.Spec.Replicas = replicas
	}
}

// sets container image name of Backstage Container
func (b *BackstageDeployment) setImage(image *string) {
	if image != nil {
		b.container().Image = *image
		//b.deployment.Spec.Template.Spec.Containers[0].Image = *image
	}
}

// adds environment variables to the Backstage Container
func (b *BackstageDeployment) addContainerEnvVar(env bsv1alpha1.Env) {
	b.deployment.Spec.Template.Spec.Containers[0].Env =
		append(b.deployment.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
			Name:  env.Name,
			Value: env.Value,
		})
}

// adds environment from source to the Backstage Container
func (b *BackstageDeployment) addExtraEnvs(extraEnvs *bsv1alpha1.ExtraEnvs) {
	if extraEnvs != nil {
		for _, e := range extraEnvs.Envs {
			b.addContainerEnvVar(e)
		}
	}
}

// sets pullSecret for Backstage Pod
func (b *BackstageDeployment) setImagePullSecrets(pullSecrets []string) {
	for _, ps := range pullSecrets {
		b.deployment.Spec.Template.Spec.ImagePullSecrets = append(b.deployment.Spec.Template.Spec.ImagePullSecrets,
			corev1.LocalObjectReference{Name: ps})
	}
}

// find, validate and return app-config's configMap
//func getAppConfigMap(name, key string, configs []corev1.ConfigMap) (*corev1.ConfigMap, error) {
//	for _, cm := range configs {
//		if cm.Name == name {
//			if key != "" {
//				if _, ok := cm.Data[key]; ok {
//					return &cm, nil
//				} else {
//					return nil, fmt.Errorf("key %s not found", key)
//				}
//			}
//			return &cm, nil
//		}
//	}
//	return nil, fmt.Errorf("configMap %s not found", name)
//}
