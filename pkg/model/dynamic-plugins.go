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

	"janus-idp.io/backstage-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DynamicPluginsFactory struct{}

func (f DynamicPluginsFactory) newBackstageObject() BackstageObject {
	return &DynamicPlugins{configMap: &corev1.ConfigMap{}}
}

type DynamicPlugins struct {
	configMap *corev1.ConfigMap
}

func init() {
	registerConfig("dynamic-plugins.yaml", DynamicPluginsFactory{}, Optional)
}

// implementation of BackstageObject interface
func (p *DynamicPlugins) Object() client.Object {
	return p.configMap
}

// implementation of BackstageObject interface
func (p *DynamicPlugins) initMetainfo(backstageMeta v1alpha1.Backstage, ownsRuntime bool) {
	initMetainfo(p, backstageMeta, ownsRuntime)
}

// implementation of BackstageObject interface
func (p *DynamicPlugins) EmptyObject() client.Object {
	return &corev1.ConfigMap{}
}

// implementation of BackstageObject interface
func (p *DynamicPlugins) addToModel(model *RuntimeModel) {
	// nothing
}

// implementation of BackstageObject interface
// configMap name must be the same as (deployment.yaml).spec.template.spec.volumes.name.dynamic-plugins-conf.configMap.name
func (p *DynamicPlugins) validate(model *RuntimeModel) error {

	for _, v := range *model.backstageDeployment.pod.volumes {
		if v.ConfigMap != nil && v.ConfigMap.Name == p.configMap.Name {
			return nil
		}
	}

	return fmt.Errorf("failed to apply dynamic plugins, no deployment.spec.template.spec.volumes.configMap.name = '%s' configured", p.configMap.Name)
}
