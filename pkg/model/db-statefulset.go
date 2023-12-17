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

	corev1 "k8s.io/api/core/v1"

	bsv1alpha1 "janus-idp.io/backstage-operator/api/v1alpha1"
	"janus-idp.io/backstage-operator/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DbStatefulSetFactory struct{}

func (f DbStatefulSetFactory) newBackstageObject() BackstageObject {
	return &DbStatefulSet{statefulSet: &appsv1.StatefulSet{}}
}

type DbStatefulSet struct {
	statefulSet *appsv1.StatefulSet
}

//func newDbStatefulSet() *DbStatefulSet {
//	return &DbStatefulSet{statefulSet: &appsv1.StatefulSet{}}
//}

func (b *DbStatefulSet) Object() client.Object {
	return b.statefulSet
}

func (b *DbStatefulSet) initMetainfo(backstageMeta bsv1alpha1.Backstage, ownsRuntime bool) {
	initMetainfo(b, backstageMeta, ownsRuntime)
	b.statefulSet.SetName(utils.GenerateRuntimeObjectName(backstageMeta.Name, "db-statefulset"))
	utils.GenerateLabel(&b.statefulSet.Spec.Template.ObjectMeta.Labels, backstageAppLabel, fmt.Sprintf("backstage-db-%s", backstageMeta.Name))
	utils.GenerateLabel(&b.statefulSet.Spec.Selector.MatchLabels, backstageAppLabel, fmt.Sprintf("backstage-db-%s", backstageMeta.Name))
}

func (b *DbStatefulSet) addToModel(model *runtimeModel) {
	model.localDbStatefulSet = b
}

func (b *DbStatefulSet) EmptyObject() client.Object {
	return &appsv1.StatefulSet{}
}

// NOTE we consider single container here
func (b *DbStatefulSet) appendContainerEnvFrom(envFrom corev1.EnvFromSource) {
	b.statefulSet.Spec.Template.Spec.Containers[0].EnvFrom = append(b.statefulSet.Spec.Template.Spec.Containers[0].EnvFrom, envFrom)
}
