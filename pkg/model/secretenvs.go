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

type SecretEnvsFactory struct{}

func (f SecretEnvsFactory) newBackstageObject() BackstageObject {
	return &SecretEnvs{Secret: &corev1.Secret{}}
}

type SecretEnvs struct {
	Secret *corev1.Secret
	Key    string
}

func init() {
	registerConfig("secret-envs.yaml", SecretEnvsFactory{}, Optional)
}

// implementation of BackstageObject interface
func (p *SecretEnvs) Object() client.Object {
	return p.Secret
}

// implementation of BackstageObject interface
//func (p *SecretEnvs) setMetaInfo(backstageMeta v1alpha1.Backstage, ownsRuntime bool) {
//	setMetaInfo(p, backstageMeta, ownsRuntime)
//	p.Secret.SetName(utils.GenerateRuntimeObjectName(backstageMeta.Name, "default-secretenvs"))
//}

// implementation of BackstageObject interface
func (p *SecretEnvs) EmptyObject() client.Object {
	return &corev1.Secret{}
}

// implementation of BackstageObject interface
func (p *SecretEnvs) addToModel(model *RuntimeModel, backstageMeta v1alpha1.Backstage, ownsRuntime bool) {
	model.setObject(p)

	p.Secret.SetName(utils.GenerateRuntimeObjectName(backstageMeta.Name, "default-secretenvs"))
}

// implementation of BackstageObject interface
func (p *SecretEnvs) validate(model *RuntimeModel) error {
	return nil
}

// implementation of BackstagePodContributor interface
func (p *SecretEnvs) updateBackstagePod(pod *backstagePod) {
	if p.Key == "" {
		pod.addContainerEnvFrom(corev1.EnvFromSource{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: p.Secret.Name}}})
	} else if _, ok := p.Secret.Data[p.Key]; ok {
		pod.addContainerEnvVarSource(p.Key, &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: p.Secret.Name,
				},
				Key: p.Key,
			},
		})
	}
}
