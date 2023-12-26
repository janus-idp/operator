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
	"janus-idp.io/backstage-operator/api/v1alpha1"
	"janus-idp.io/backstage-operator/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ConfigMapEnvsFactory struct{}

func (f ConfigMapEnvsFactory) newBackstageObject() BackstageObject {
	return &ConfigMapEnvs{ConfigMap: &corev1.ConfigMap{}}
}

type ConfigMapEnvs struct {
	ConfigMap *corev1.ConfigMap
}

func (p *ConfigMapEnvs) Object() client.Object {
	return p.ConfigMap
}

func (p *ConfigMapEnvs) initMetainfo(backstageMeta v1alpha1.Backstage, ownsRuntime bool) {
	initMetainfo(p, backstageMeta, ownsRuntime)
	p.ConfigMap.SetName(utils.GenerateRuntimeObjectName(backstageMeta.Name, "default-configmapenvs"))
}

func (p *ConfigMapEnvs) EmptyObject() client.Object {
	return &corev1.ConfigMap{}
}

func (p *ConfigMapEnvs) addToModel(model *runtimeModel) {
	// nothing
}

func (p *ConfigMapEnvs) updateBackstagePod(pod *backstagePod) {

	pod.appendContainerEnvFrom(corev1.EnvFromSource{
		ConfigMapRef: &corev1.ConfigMapEnvSource{
			LocalObjectReference: corev1.LocalObjectReference{Name: p.ConfigMap.Name}}})

}
