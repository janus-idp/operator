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

func (f ConfigMapFilesFactory) newBackstageObject() RuntimeObject {
	return &ConfigMapFiles{ /*ConfigMap: &corev1.ConfigMap{},*/ MountPath: defaultDir}
}

type ConfigMapFiles struct {
	ConfigMap *corev1.ConfigMap
	MountPath string
	Key       string
}

func init() {
	registerConfig("configmap-files.yaml", ConfigMapFilesFactory{})
}

// implementation of RuntimeObject interface
func (p *ConfigMapFiles) Object() client.Object {
	return p.ConfigMap
}

func (p *ConfigMapFiles) setObject(object client.Object) {
	p.ConfigMap = nil
	if object != nil {
		p.ConfigMap = object.(*corev1.ConfigMap)
	}

}

// implementation of RuntimeObject interface
func (p *ConfigMapFiles) EmptyObject() client.Object {
	return &corev1.ConfigMap{}
}

// implementation of RuntimeObject interface
func (p *ConfigMapFiles) addToModel(model *BackstageModel, backstageMeta v1alpha1.Backstage, ownsRuntime bool) error {
	if p.ConfigMap != nil {
		model.setRuntimeObject(p)
		p.ConfigMap.SetName(utils.GenerateRuntimeObjectName(backstageMeta.Name, "default-configmapfiles"))
	}
	return nil
}

// implementation of RuntimeObject interface
func (p *ConfigMapFiles) validate(model *BackstageModel) error {
	return nil
}

// implementation of PodContributor interface
func (p *ConfigMapFiles) updatePod(pod *backstagePod) {

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

	vm := corev1.VolumeMount{Name: volName, MountPath: filepath.Join(p.MountPath, p.ConfigMap.Name, p.Key), SubPath: p.Key}
	pod.container.VolumeMounts = append(pod.container.VolumeMounts, vm)

}
