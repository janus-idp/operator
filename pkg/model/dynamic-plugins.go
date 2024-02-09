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
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"path/filepath"

	"janus-idp.io/backstage-operator/pkg/utils"
	"k8s.io/utils/pointer"

	"janus-idp.io/backstage-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const dynamicPluginInitContainerName = "install-dynamic-plugins"

type DynamicPluginsFactory struct{}

func (f DynamicPluginsFactory) newBackstageObject() RuntimeObject {
	return &DynamicPlugins{ /*ConfigMap: &corev1.ConfigMap{}*/ }
}

type DynamicPlugins struct {
	ConfigMap *corev1.ConfigMap
}

func init() {
	registerConfig("dynamic-plugins.yaml", DynamicPluginsFactory{})
}

func newDynamicPlugins(configMapName string) *DynamicPlugins {
	return &DynamicPlugins{ConfigMap: &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: configMapName},
	}}
}

// implementation of RuntimeObject interface
func (p *DynamicPlugins) Object() client.Object {
	return p.ConfigMap
}

func (p *DynamicPlugins) setObject(obj client.Object, backstageName string) {
	p.ConfigMap = nil
	if obj != nil {
		p.ConfigMap = obj.(*corev1.ConfigMap)
		p.ConfigMap.SetName(utils.GenerateRuntimeObjectName(backstageName, "default-dynamic-plugins"))
	}

}

// implementation of RuntimeObject interface
func (p *DynamicPlugins) EmptyObject() client.Object {
	return &corev1.ConfigMap{}
}

// implementation of RuntimeObject interface
func (p *DynamicPlugins) addToModel(model *BackstageModel, backstageMeta v1alpha1.Backstage, ownsRuntime bool) error {
	if p.ConfigMap != nil {
		model.setRuntimeObject(p)
	}
	return nil
}

// implementation of PodContributor interface
func (p *DynamicPlugins) updatePod(pod *backstagePod) {

	//it relies on implementation where dynamic-plugin initContainer
	//uses specified ConfigMap for producing app-config with dynamic-plugins
	//For this implementation:
	//- backstage contaier and dynamic-plugin initContainer must share a volume
	//  where initContainer writes and backstage container reads produced app-config
	//- app-config path should be set as a --config parameter of backstage container
	//in the deployment manifest

	//it creates a volume with dynamic-plugins ConfigMap (there should be a key named "dynamic-plugins.yaml")
	//and mount it to the dynamic-plugin initContainer's WorkingDir (what if not specified?)
	initContainer := dynamicPluginsInitContainer(pod.parent.Spec.Template.Spec.InitContainers)
	if initContainer == nil {
		// it will fail on validate
		return
	}

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
		pod.appendOrReplaceInitContainerVolumeMount(corev1.VolumeMount{
			Name:      volName,
			MountPath: filepath.Join(initContainer.WorkingDir, file),
			SubPath:   file,
			ReadOnly:  true,
		}, dynamicPluginInitContainerName)
	}
}

// implementation of RuntimeObject interface
// ConfigMap name must be the same as (deployment.yaml).spec.template.spec.volumes.name.dynamic-plugins-conf.ConfigMap.name
func (p *DynamicPlugins) validate(model *BackstageModel, backstage v1alpha1.Backstage) error {

	initContainer := dynamicPluginsInitContainer(model.backstageDeployment.deployment.Spec.Template.Spec.InitContainers)
	if initContainer == nil {
		return fmt.Errorf("failed to find initContainer named %s", dynamicPluginInitContainerName)
	}
	// override image with env var
	// [GA] Do we need this feature?
	if os.Getenv(BackstageImageEnvVar) != "" {
		// TODO workaround for the (janus-idp, rhdh) case where we have
		// exactly the same image for initContainer and want it to be overriden
		// the same way as Backstage's one
		initContainer.Image = os.Getenv(BackstageImageEnvVar)
	}
	return nil
}

// returns initContainer supposed to initialize DynamicPlugins
// TODO consider to use a label to identify instead
func dynamicPluginsInitContainer(initContainers []corev1.Container) *corev1.Container {
	for _, ic := range initContainers {
		if ic.Name == dynamicPluginInitContainerName {
			return &ic
		}
	}
	return nil
}
