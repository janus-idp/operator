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
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/utils/pointer"

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

func (p *SecretFiles) setObject(obj client.Object, name string) {
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
func (p *SecretFiles) addToModel(model *BackstageModel, backstageMeta v1alpha1.Backstage, ownsRuntime bool) error {
	if p.Secret != nil {
		model.setRuntimeObject(p)
		p.Secret.SetName(utils.GenerateRuntimeObjectName(backstageMeta.Name, "default-secretfiles"))
	}
	return nil
}

// implementation of RuntimeObject interface
func (p *SecretFiles) validate(model *BackstageModel, backstage v1alpha1.Backstage) error {
	return nil
}

// implementation of PodContributor interface
func (p *SecretFiles) updatePod(pod *backstagePod) {

	volName := utils.GenerateVolumeNameFromCmOrSecret(p.Secret.Name)

	volSource := corev1.VolumeSource{
		Secret: &corev1.SecretVolumeSource{
			DefaultMode: pointer.Int32(420),
			SecretName:  p.Secret.Name,
		},
	}

	pod.appendVolume(corev1.Volume{
		Name:         volName,
		VolumeSource: volSource,
	})

	vm := corev1.VolumeMount{Name: volName, MountPath: filepath.Join(p.MountPath, p.Secret.Name, p.Key), SubPath: p.Key}
	pod.container.VolumeMounts = append(pod.container.VolumeMounts, vm)
}
