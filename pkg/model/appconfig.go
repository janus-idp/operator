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

	bsv1alpha1 "janus-idp.io/backstage-operator/api/v1alpha1"
	"janus-idp.io/backstage-operator/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const defaultDir = "/test/dir"

type AppConfig struct {
	//path      string
	configMap *corev1.ConfigMap
}

func newAppConfig() *AppConfig {
	return &AppConfig{configMap: &corev1.ConfigMap{}}
}

func (b *AppConfig) Object() client.Object {
	return b.configMap
}

func (b *AppConfig) initMetainfo(backstageMeta bsv1alpha1.Backstage, ownsRuntime bool) {
	initMetainfo(b, backstageMeta, ownsRuntime)
	b.configMap.SetName(utils.GenerateRuntimeObjectName(backstageMeta.Name, "default-appconfig"))
}

func (b *AppConfig) updateBackstagePod(pod *backstagePod) {
	path := defaultDir
	for k := range b.configMap.Data {
		path = filepath.Join(path, k)
	}
	pod.addAppConfig(b.configMap.Name, path)
}

func (b *AppConfig) addToModel(model *runtimeModel) {
	// nothing to add
}
