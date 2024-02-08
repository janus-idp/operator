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

	bsv1alpha1 "janus-idp.io/backstage-operator/api/v1alpha1"

	"janus-idp.io/backstage-operator/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const BackstageImageEnvVar = "RELATED_IMAGE_backstage"

type BackstageDeploymentFactory struct{}

func (f BackstageDeploymentFactory) newBackstageObject() RuntimeObject {
	return &BackstageDeployment{ /*deployment: &appsv1.Deployment{}*/ }
}

type BackstageDeployment struct {
	deployment *appsv1.Deployment
	pod        *backstagePod
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

func (b *BackstageDeployment) setObject(object client.Object) {
	b.deployment = nil
	if object != nil {
		b.deployment = object.(*appsv1.Deployment)
	}
}

// implementation of RuntimeObject interface
func (b *BackstageDeployment) EmptyObject() client.Object {
	return &appsv1.Deployment{}
}

// implementation of RuntimeObject interface
func (b *BackstageDeployment) addToModel(model *BackstageModel, backstage bsv1alpha1.Backstage, ownsRuntime bool) error {
	if b.deployment == nil {
		return fmt.Errorf("Backstage Deployment is not initialized, make sure there is deployment.yaml in default or raw configuration")
	}
	model.backstageDeployment = b
	model.setRuntimeObject(b)

	b.deployment.SetName(utils.GenerateRuntimeObjectName(backstage.Name, "deployment"))
	utils.GenerateLabel(&b.deployment.Spec.Template.ObjectMeta.Labels, backstageAppLabel, fmt.Sprintf("backstage-%s", backstage.Name))
	utils.GenerateLabel(&b.deployment.Spec.Selector.MatchLabels, backstageAppLabel, fmt.Sprintf("backstage-%s", backstage.Name))

	// fill the Pod
	// create Backstage Pod object
	var err error
	b.pod, err = newBackstagePod(model.backstageDeployment)
	if err != nil {
		return fmt.Errorf("failed to create Backstage Pod: %s", err)
	}

	if backstage.Spec.Application != nil {
		b.setReplicas(backstage.Spec.Application.Replicas)
		b.pod.setImagePullSecrets(backstage.Spec.Application.ImagePullSecrets)
		b.pod.setImage(backstage.Spec.Application.Image)
		b.pod.addExtraEnvs(backstage.Spec.Application.ExtraEnvs)
	}

	// override image with env var
	// [GA] TODO Do we need this feature?
	if os.Getenv(BackstageImageEnvVar) != "" {
		b.pod.container.Image = os.Getenv(BackstageImageEnvVar)
		// TODO workaround for the (janus-idp, rhdh) case where we have
		// exactly the same image for initContainer and want it to be overriden
		// the same way as Backstage's one
		for i := range b.deployment.Spec.Template.Spec.InitContainers {
			b.deployment.Spec.Template.Spec.InitContainers[i].Image = os.Getenv(BackstageImageEnvVar)
		}
	}

	return nil
}

// implementation of RuntimeObject interface
func (b *BackstageDeployment) validate(model *BackstageModel) error {
	//for _, bso := range model.RuntimeObjects {
	//	if bs, ok := bso.(PodContributor); ok {
	//		bs.updatePod(b.pod)
	//	}
	//}
	//if backstage.Spec.Application != nil {
	//	// AppConfig
	//	// DynaPlugins
	//	// Ext (4)
	//	// DbSecret
	//}

	//for _, v := range backstage.Spec.ConfigObjects {
	//	v.updatePod(b.pod)
	//}
	return nil
}

// sets the amount of replicas (used by CR config)
func (b *BackstageDeployment) setReplicas(replicas *int32) {
	if replicas != nil {
		b.deployment.Spec.Replicas = replicas
	}
}
