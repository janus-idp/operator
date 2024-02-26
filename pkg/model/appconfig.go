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

	appsv1 "k8s.io/api/apps/v1"

	bsv1alpha1 "janus-idp.io/backstage-operator/api/v1alpha1"
	"janus-idp.io/backstage-operator/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type AppConfigFactory struct{}

// factory method to create App Config object
func (f AppConfigFactory) newBackstageObject() RuntimeObject {
	return &AppConfig{MountPath: defaultMountDir}
}

// structure containing ConfigMap where keys are Backstage ConfigApp file names and vaues are contents of the files
// Mount path is a patch to the follder to place the files to
type AppConfig struct {
	ConfigMap *corev1.ConfigMap
	MountPath string
	Key       string
}

func init() {
	registerConfig("app-config.yaml", AppConfigFactory{})
}

func AppConfigDefaultName(backstageName string) string {
	return utils.GenerateRuntimeObjectName(backstageName, "default-appconfig")
}

func newAppConfig(mountPath string, cm *corev1.ConfigMap, key string) *AppConfig {
	return &AppConfig{
		ConfigMap: cm,
		MountPath: mountPath,
		Key:       key,
	}
}

// implementation of RuntimeObject interface
func (b *AppConfig) Object() client.Object {
	return b.ConfigMap
}

// implementation of RuntimeObject interface
func (b *AppConfig) setObject(obj client.Object, backstageName string) {
	b.ConfigMap = nil
	if obj != nil {
		b.ConfigMap = obj.(*corev1.ConfigMap)
	}
}

// implementation of RuntimeObject interface
func (b *AppConfig) EmptyObject() client.Object {
	return &corev1.ConfigMap{}
}

// implementation of RuntimeObject interface
func (b *AppConfig) addToModel(model *BackstageModel, backstageMeta bsv1alpha1.Backstage, ownsRuntime bool) (bool, error) {
	if b.ConfigMap != nil {
		model.setRuntimeObject(b)
		return true, nil
	}
	return false, nil
}

// implementation of RuntimeObject interface
func (b *AppConfig) validate(model *BackstageModel, backstage bsv1alpha1.Backstage) error {
	return nil
}

func (b *AppConfig) setMetaInfo(backstageName string) {
	b.ConfigMap.SetName(AppConfigDefaultName(backstageName))
}

// implementation of BackstagePodContributor interface
// it contrubutes to Volumes, container.VolumeMounts and contaiter.Args
func (b *AppConfig) updatePod(deployment *appsv1.Deployment) {

	utils.MountFilesFrom(&deployment.Spec.Template.Spec, &deployment.Spec.Template.Spec.Containers[0], utils.ConfigMapObjectKind,
		b.ConfigMap.Name, b.MountPath, b.Key, b.ConfigMap.Data)

	fileDir := b.MountPath
	for file := range b.ConfigMap.Data {
		if b.Key == "" || b.Key == file {
			deployment.Spec.Template.Spec.Containers[0].Args =
				append(deployment.Spec.Template.Spec.Containers[0].Args, []string{"--config", filepath.Join(fileDir, file)}...)
		}
	}
}
