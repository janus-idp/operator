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

	"k8s.io/utils/ptr"

	corev1 "k8s.io/api/core/v1"

	bsv1alpha1 "redhat-developer/red-hat-developer-hub-operator/api/v1alpha1"

	"redhat-developer/red-hat-developer-hub-operator/pkg/utils"

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
}

func init() {
	registerConfig("deployment.yaml", BackstageDeploymentFactory{})
}

func DeploymentName(backstageName string) string {
	return utils.GenerateRuntimeObjectName(backstageName, "backstage")
}

// implementation of RuntimeObject interface
func (b *BackstageDeployment) Object() client.Object {
	return b.deployment
}

func (b *BackstageDeployment) setObject(obj client.Object) {
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
func (b *BackstageDeployment) addToModel(model *BackstageModel, _ bsv1alpha1.Backstage) (bool, error) {
	if b.deployment == nil {
		return false, fmt.Errorf("Backstage Deployment is not initialized, make sure there is deployment.yaml in default or raw configuration")
	}
	model.backstageDeployment = b
	model.setRuntimeObject(b)

	// override image with env var
	// [GA] TODO Do we need this feature?
	if os.Getenv(BackstageImageEnvVar) != "" {
		b.setImage(ptr.To(os.Getenv(BackstageImageEnvVar)))
	}

	return true, nil
}

// implementation of RuntimeObject interface
func (b *BackstageDeployment) validate(model *BackstageModel, backstage bsv1alpha1.Backstage) error {

	if backstage.Spec.Application != nil {
		b.setReplicas(backstage.Spec.Application.Replicas)
		b.setImagePullSecrets(backstage.Spec.Application.ImagePullSecrets)
		b.setImage(backstage.Spec.Application.Image)
		b.addExtraEnvs(backstage.Spec.Application.ExtraEnvs)
	}

	for _, bso := range model.RuntimeObjects {
		if bs, ok := bso.(BackstagePodContributor); ok {
			bs.updatePod(b.deployment)
		}
	}

	addAppConfigs(backstage.Spec, b.deployment, model)

	addConfigMapFiles(backstage.Spec, b.deployment, model)

	addConfigMapEnvs(backstage.Spec, b.deployment, model)

	if err := addSecretFiles(backstage.Spec, b.deployment); err != nil {
		return err
	}

	if err := addSecretEnvs(backstage.Spec, b.deployment); err != nil {
		return err
	}
	if err := addDynamicPlugins(backstage.Spec, b.deployment, model); err != nil {
		return err
	}

	//DbSecret
	if backstage.Spec.IsAuthSecretSpecified() {
		utils.SetDbSecretEnvVar(b.container(), backstage.Spec.Database.AuthSecretName)
	} else if model.LocalDbSecret != nil {
		utils.SetDbSecretEnvVar(b.container(), model.LocalDbSecret.secret.Name)
	}

	return nil
}

func (b *BackstageDeployment) setMetaInfo(backstageName string) {
	b.deployment.SetName(DeploymentName(backstageName))
	utils.GenerateLabel(&b.deployment.Spec.Template.ObjectMeta.Labels, BackstageAppLabel, fmt.Sprintf("backstage-%s", backstageName))
	utils.GenerateLabel(&b.deployment.Spec.Selector.MatchLabels, BackstageAppLabel, fmt.Sprintf("backstage-%s", backstageName))
}

func (b *BackstageDeployment) container() *corev1.Container {
	return &b.deployment.Spec.Template.Spec.Containers[0]
}

func (b *BackstageDeployment) podSpec() *corev1.PodSpec {
	return &b.deployment.Spec.Template.Spec
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
		// this is a workaround for RHDH/Janus configuration
		// it is not a fact that all the containers should be updated
		// in general case need something smarter
		// to mark/recognize containers for update
		if len(b.podSpec().InitContainers) > 0 {
			i, ic := dynamicPluginsInitContainer(b.podSpec().InitContainers)
			if ic != nil {
				b.podSpec().InitContainers[i].Image = *image
			}
		}
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
