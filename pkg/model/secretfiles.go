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

	"k8s.io/utils/pointer"

	"janus-idp.io/backstage-operator/api/v1alpha1"
	"janus-idp.io/backstage-operator/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SecretFilesFactory struct{}

func (f SecretFilesFactory) newBackstageObject() BackstageObject {
	return &SecretFiles{Secret: &corev1.Secret{}, MountPath: defaultDir}
}

type SecretFiles struct {
	Secret    *corev1.Secret
	MountPath string
}

func init() {
	registerConfig("secret-files.yaml", SecretFilesFactory{}, Optional)
}

// implementation of BackstageObject interface
func (p *SecretFiles) Object() client.Object {
	return p.Secret
}

// implementation of BackstageObject interface
func (p *SecretFiles) initMetainfo(backstageMeta v1alpha1.Backstage, ownsRuntime bool) {
	initMetainfo(p, backstageMeta, ownsRuntime)
	p.Secret.SetName(utils.GenerateRuntimeObjectName(backstageMeta.Name, "default-secretfiles"))
}

// implementation of BackstageObject interface
func (p *SecretFiles) EmptyObject() client.Object {
	return &corev1.Secret{}
}

// implementation of BackstageObject interface
func (p *SecretFiles) addToModel(model *RuntimeModel) {
	// nothing
}

// implementation of BackstageObject interface
func (p *SecretFiles) validate(model *RuntimeModel) error {
	return nil
}

// implementation of BackstagePodContributor interface
func (p *SecretFiles) updateBackstagePod(pod *backstagePod) {

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

	for file := range p.Secret.Data {

		pod.appendContainerVolumeMount(corev1.VolumeMount{
			Name:      volName,
			MountPath: filepath.Join(p.MountPath, file),
			SubPath:   file,
		})

	}

	for file := range p.Secret.StringData {

		pod.appendContainerVolumeMount(corev1.VolumeMount{
			Name:      volName,
			MountPath: filepath.Join(p.MountPath, file),
			SubPath:   file,
		})

	}

}
