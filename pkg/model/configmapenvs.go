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

func (f ConfigMapEnvsFactory) newBackstageObject() RuntimeObject {
	return &ConfigMapEnvs{ConfigMap: &corev1.ConfigMap{}}
}

type ConfigMapEnvs struct {
	ConfigMap *corev1.ConfigMap
	Key       string
}

func init() {
	registerConfig("configmap-envs.yaml", ConfigMapEnvsFactory{}, Optional)
}

// Object implements RuntimeObject interface
func (p *ConfigMapEnvs) Object() client.Object {
	return p.ConfigMap
}

// EmptyObject implements RuntimeObject interface
func (p *ConfigMapEnvs) EmptyObject() client.Object {
	return &corev1.ConfigMap{}
}

// implementation of RuntimeObject interface
func (p *ConfigMapEnvs) addToModel(model *BackstageModel, backstageMeta v1alpha1.Backstage, ownsRuntime bool) {
	model.setRuntimeObject(p)
	p.ConfigMap.SetName(utils.GenerateRuntimeObjectName(backstageMeta.Name, "default-configmapenvs"))
}

// implementation of RuntimeObject interface
func (p *ConfigMapEnvs) validate(model *BackstageModel) error {
	return nil
}

// implementation of PodContributor interface
func (p *ConfigMapEnvs) updatePod(pod *backstagePod) {

	if p.Key == "" {
		pod.addContainerEnvFrom(corev1.EnvFromSource{
			ConfigMapRef: &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: p.ConfigMap.Name}}})
	} else {
		envVarSource := &corev1.EnvVarSource{
			ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: p.ConfigMap.Name,
				},
				Key: p.Key,
			},
		}
		pod.container.Env = append(pod.container.Env, corev1.EnvVar{
			Name:      p.Key,
			ValueFrom: envVarSource,
		})
	}

	//if p.Key == "" || (p.Key == p.ConfigMap.Name) {
	//	pod.addContainerEnvFrom(corev1.EnvFromSource{
	//		ConfigMapRef: &corev1.ConfigMapEnvSource{
	//			LocalObjectReference: corev1.LocalObjectReference{Name: p.ConfigMap.Name}}})
	//}

	//if p.Key == "" {
	//	pod.addContainerEnvFrom(corev1.EnvFromSource{
	//		ConfigMapRef: &corev1.ConfigMapEnvSource{
	//			LocalObjectReference: corev1.LocalObjectReference{Name: p.ConfigMap.Name}}})
	//} else if _, ok := p.ConfigMap.Data[p.Key]; ok {
	//	pod.addContainerEnvVarSource(p.Key, &corev1.EnvVarSource{
	//		ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
	//			LocalObjectReference: corev1.LocalObjectReference{
	//				Name: p.ConfigMap.Name,
	//			},
	//			Key: p.Key,
	//		},
	//	})
	//}
}
