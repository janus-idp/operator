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
	"redhat-developer/red-hat-developer-hub-operator/api/v1alpha1"
	"redhat-developer/red-hat-developer-hub-operator/pkg/utils"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SecretEnvsFactory struct{}

func (f SecretEnvsFactory) newBackstageObject() RuntimeObject {
	return &SecretEnvs{}
}

type SecretEnvs struct {
	Secret *corev1.Secret
	Key    string
}

func init() {
	registerConfig("secret-envs.yaml", SecretEnvsFactory{})
}

// implementation of RuntimeObject interface
func (p *SecretEnvs) Object() client.Object {
	return p.Secret
}

func addSecretEnvs(spec v1alpha1.BackstageSpec, deployment *appsv1.Deployment) error {

	if spec.Application == nil || spec.Application.ExtraEnvs == nil || spec.Application.ExtraEnvs.Secrets == nil {
		return nil
	}

	for _, sec := range spec.Application.ExtraEnvs.Secrets {
		se := SecretEnvs{
			Secret: &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: sec.Name}},
			Key:    sec.Key,
		}
		se.updatePod(deployment)
	}
	return nil
}

func (p *SecretEnvs) setObject(obj client.Object) {
	p.Secret = nil
	if obj != nil {
		p.Secret = obj.(*corev1.Secret)
	}
}

// implementation of RuntimeObject interface
func (p *SecretEnvs) EmptyObject() client.Object {
	return &corev1.Secret{}
}

// implementation of RuntimeObject interface
func (p *SecretEnvs) addToModel(model *BackstageModel, _ v1alpha1.Backstage) (bool, error) {
	if p.Secret != nil {
		model.setRuntimeObject(p)
		return true, nil
	}
	return false, nil
}

// implementation of RuntimeObject interface
func (p *SecretEnvs) validate(_ *BackstageModel, _ v1alpha1.Backstage) error {
	return nil
}

func (p *SecretEnvs) setMetaInfo(backstageName string) {
	p.Secret.SetName(utils.GenerateRuntimeObjectName(backstageName, "backstage-envs"))
}

// implementation of BackstagePodContributor interface
func (p *SecretEnvs) updatePod(deployment *appsv1.Deployment) {

	utils.AddEnvVarsFrom(&deployment.Spec.Template.Spec.Containers[0], utils.SecretObjectKind,
		p.Secret.Name, p.Key)
}
