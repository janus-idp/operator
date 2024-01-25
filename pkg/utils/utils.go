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

package utils

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/util/yaml"
)

func SetKubeLabels(labels map[string]string, backstageName string) map[string]string {
	if labels == nil {
		labels = map[string]string{}
	}
	labels["app.kubernetes.io/name"] = "backstage"
	labels["app.kubernetes.io/instance"] = backstageName

	return labels
}

// GenerateLabel generates backstage-{Id} for labels or selectors
func GenerateLabel(labels *map[string]string, name string, value string) {
	if *labels == nil {
		*labels = map[string]string{}
	}
	(*labels)[name] = value
}

// GenerateRuntimeObjectName generates name using BackstageCR name and objectType which is ConfigObject Key without '.yaml' (like 'deployment')
func GenerateRuntimeObjectName(backstageCRName string, objectType string) string {
	return fmt.Sprintf("%s-%s", backstageCRName, objectType)
}

// GenerateVolumeNameFromCmOrSecret generates volume name for mounting ConfigMap or Secret
func GenerateVolumeNameFromCmOrSecret(cmOrSecretName string) string {
	return fmt.Sprintf("vol-%s", cmOrSecretName)
}

func ReadYaml(manifest []byte, object interface{}) error {
	dec := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(manifest), 1000)
	if err := dec.Decode(object); err != nil {
		return fmt.Errorf("failed to decode YAML: %w", err)
	}
	return nil
}

func ReadYamlFile(path string, object metav1.Object) error {

	b, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return fmt.Errorf("failed to read YAML file: %w", err)
	}
	return ReadYaml(b, object)
}

func DefFile(key string) string {
	return filepath.Join(os.Getenv("LOCALBIN"), "default-config", key)
}
