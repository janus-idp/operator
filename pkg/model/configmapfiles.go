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

type ConfigMapFilesFactory struct{}

func (f ConfigMapFilesFactory) newBackstageObject() BackstageObject {
	return &ConfigMapFiles{ConfigMap: &corev1.ConfigMap{}, MountPath: defaultDir}
}

type ConfigMapFiles struct {
	ConfigMap *corev1.ConfigMap
	MountPath string
}

func init() {
	registerConfig("configmap-files.yaml", ConfigMapFilesFactory{}, Optional)
}

// implementation of BackstageObject interface
func (p *ConfigMapFiles) Object() client.Object {
	return p.ConfigMap
}

// implementation of BackstageObject interface
func (p *ConfigMapFiles) initMetainfo(backstageMeta v1alpha1.Backstage, ownsRuntime bool) {
	initMetainfo(p, backstageMeta, ownsRuntime)
	p.ConfigMap.SetName(utils.GenerateRuntimeObjectName(backstageMeta.Name, "default-configmapfiles"))
}

// implementation of BackstageObject interface
func (p *ConfigMapFiles) EmptyObject() client.Object {
	return &corev1.ConfigMap{}
}

// implementation of BackstageObject interface
func (p *ConfigMapFiles) addToModel(model *RuntimeModel) {
	// nothing
}

// implementation of BackstageObject interface
func (p *ConfigMapFiles) validate(model *RuntimeModel) error {
	return nil
}

// implementation of BackstagePodContributor interface
func (p *ConfigMapFiles) updateBackstagePod(pod *backstagePod) {

	volName := utils.GenerateVolumeNameFromCmOrSecret(p.ConfigMap.Name)

	volSource := corev1.VolumeSource{
		ConfigMap: &corev1.ConfigMapVolumeSource{
			DefaultMode:          pointer.Int32(420),
			LocalObjectReference: corev1.LocalObjectReference{Name: p.ConfigMap.Name},
		},
	}
	pod.appendVolume(corev1.Volume{
		Name:         volName,
		VolumeSource: volSource,
	})

	for file := range p.ConfigMap.Data {

		pod.appendContainerVolumeMount(corev1.VolumeMount{
			Name:      volName,
			MountPath: filepath.Join(p.MountPath, file),
			SubPath:   file,
		})

	}

}
