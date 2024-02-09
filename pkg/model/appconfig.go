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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"path/filepath"

	bsv1alpha1 "janus-idp.io/backstage-operator/api/v1alpha1"
	"janus-idp.io/backstage-operator/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type AppConfigFactory struct{}

// factory method to create App Config object
func (f AppConfigFactory) newBackstageObject() RuntimeObject {
	return &AppConfig{MountPath: defaultDir}
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

func newAppConfig(mountPath string, name string, key string) *AppConfig {
	return &AppConfig{
		ConfigMap: &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: name},
		},
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
		b.ConfigMap.SetName(utils.GenerateRuntimeObjectName(backstageName, "default-appconfig"))
	}
}

// implementation of RuntimeObject interface
func (b *AppConfig) EmptyObject() client.Object {
	return &corev1.ConfigMap{}
}

// implementation of RuntimeObject interface
func (b *AppConfig) addToModel(model *BackstageModel, backstageMeta bsv1alpha1.Backstage, ownsRuntime bool) error {
	if b.ConfigMap != nil {
		model.setRuntimeObject(b)
	}
	return nil
}

// implementation of RuntimeObject interface
func (b *AppConfig) validate(model *BackstageModel, backstage bsv1alpha1.Backstage) error {
	return nil
}

// implementation of PodContributor interface
// it contrubutes to Volumes, container.VolumeMounts and contaiter.Args
func (b *AppConfig) updatePod(pod *backstagePod) {

	volName := utils.GenerateVolumeNameFromCmOrSecret(b.ConfigMap.Name)

	volSource := corev1.VolumeSource{
		ConfigMap: &corev1.ConfigMapVolumeSource{
			DefaultMode:          pointer.Int32(420),
			LocalObjectReference: corev1.LocalObjectReference{Name: b.ConfigMap.Name},
		},
	}
	pod.appendVolume(corev1.Volume{
		Name:         volName,
		VolumeSource: volSource,
	})

	//for file := range b.ConfigMap.Data {
	//	if b.Key == "" || (b.Key == file) {
	//		pod.appendContainerVolumeMount(corev1.VolumeMount{
	//			Name:      volName,
	//			MountPath: filepath.Join(b.MountPath, file),
	//			SubPath:   file,
	//		})
	//
	//		pod.appendConfigArg(filepath.Join(b.MountPath, file))
	//	}
	//}

	// One configMap - one appConfig
	// Problem: we need to know file path to form --config CL args
	// If we want not to read CM - need to point file name (key) which should fit CM data.key
	// Otherwise - we can read it and not specify
	// Path to appConfig: /<mountPath>/<configMapName>/<file(key) name>
	// Preferences:
	// - not to read CM.Data on external files (Less permissive operator, not needed CM read/list)
	// - not to use SubPath mounting CM to make Kubernetes refresh data if CM changed
	vm := corev1.VolumeMount{Name: volName, MountPath: filepath.Join(b.MountPath, b.ConfigMap.Name)}
	pod.container.VolumeMounts = append(pod.container.VolumeMounts, vm)

	appConfigPath := filepath.Join(b.MountPath, b.ConfigMap.Name, b.Key)
	pod.container.Args = append(pod.container.Args, []string{"--config", appConfigPath}...)
	//pod.appendConfigArg(filepath.Join(b.MountPath, b.ConfigMap.Name, b.Key))

}
