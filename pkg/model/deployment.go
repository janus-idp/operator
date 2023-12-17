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

	bsv1alpha1 "janus-idp.io/backstage-operator/api/v1alpha1"

	"janus-idp.io/backstage-operator/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type BackstageDeploymentFactory struct{}

func (f BackstageDeploymentFactory) newBackstageObject() BackstageObject {
	return &BackstageDeployment{deployment: &appsv1.Deployment{}}
}

type BackstageDeployment struct {
	deployment *appsv1.Deployment
	pod        *backstagePod
}

func getBackstageDeployment(bsobjects []BackstageObject) *BackstageDeployment {
	for _, obj := range bsobjects {
		if bs, ok := obj.(*BackstageDeployment); ok {
			return bs
		}
	}
	return nil
}

func (b *BackstageDeployment) Object() client.Object {
	return b.deployment
}

func (b *BackstageDeployment) EmptyObject() client.Object {

	return &appsv1.Deployment{}
}

func (b *BackstageDeployment) initMetainfo(backstageMeta bsv1alpha1.Backstage, ownsRuntime bool) {
	initMetainfo(b, backstageMeta, ownsRuntime)
	b.deployment.SetName(utils.GenerateRuntimeObjectName(backstageMeta.Name, "deployment"))
	utils.GenerateLabel(&b.deployment.Spec.Template.ObjectMeta.Labels, backstageAppLabel, fmt.Sprintf("backstage-%s", backstageMeta.Name))
	utils.GenerateLabel(&b.deployment.Spec.Selector.MatchLabels, backstageAppLabel, fmt.Sprintf("backstage-%s", backstageMeta.Name))
}

func (b *BackstageDeployment) addToModel(model *runtimeModel) {
	model.backstageDeployment = b
}

func (b *BackstageDeployment) setReplicas(replicas *int32) {
	if replicas != nil {
		b.deployment.Spec.Replicas = replicas
	}
}
