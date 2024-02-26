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
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"janus-idp.io/backstage-operator/api/v1alpha1"
	"janus-idp.io/backstage-operator/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SecretFilesFactory struct{}

func (f SecretFilesFactory) newBackstageObject() RuntimeObject {
	return &SecretFiles{ /*Secret: &corev1.Secret{},*/ MountPath: defaultMountDir}
}

type SecretFiles struct {
	Secret    *corev1.Secret
	MountPath string
	Key       string
}

func init() {
	registerConfig("secret-files.yaml", SecretFilesFactory{})
}

func newSecretFiles(mountPath string, name string, key string) *SecretFiles {
	return &SecretFiles{
		Secret: &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: name},
		},
		MountPath: mountPath,
		Key:       key,
	}
}

// implementation of RuntimeObject interface
func (p *SecretFiles) Object() client.Object {
	return p.Secret
}

func (p *SecretFiles) setObject(obj client.Object, backstageName string) {
	p.Secret = nil
	if obj != nil {
		p.Secret = obj.(*corev1.Secret)
	}
}

// implementation of RuntimeObject interface
func (p *SecretFiles) EmptyObject() client.Object {
	return &corev1.Secret{}
}

// implementation of RuntimeObject interface
func (p *SecretFiles) addToModel(model *BackstageModel, backstageMeta v1alpha1.Backstage, ownsRuntime bool) (bool, error) {
	if p.Secret != nil {
		model.setRuntimeObject(p)
		return true, nil
	}
	return false, nil
}

// implementation of RuntimeObject interface
func (p *SecretFiles) validate(model *BackstageModel, backstage v1alpha1.Backstage) error {
	return nil
}

func (p *SecretFiles) setMetaInfo(backstageName string) {
	p.Secret.SetName(utils.GenerateRuntimeObjectName(backstageName, "default-secretfiles"))
}

// implementation of BackstagePodContributor interface
func (p *SecretFiles) updatePod(depoyment *appsv1.Deployment) {

	utils.MountFilesFrom(&depoyment.Spec.Template.Spec, &depoyment.Spec.Template.Spec.Containers[0], utils.SecretObjectKind,
		p.Secret.Name, p.MountPath, p.Key, p.Secret.StringData)
}
