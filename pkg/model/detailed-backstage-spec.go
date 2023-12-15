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

import bs "janus-idp.io/backstage-operator/api/v1alpha1"

type DetailedBackstageSpec struct {
	bs.BackstageSpec
	Details SpecDetails
}

type SpecDetails struct {
	RawConfig              map[string]string
	AppConfigs             []AppConfigDetails
	ExtraSecretsToFiles    []ExtraSecretToFilesDetails
	ExtraSecretsToEnvs     []ExtraSecretToEnvsDetails
	ExtraConfigMapsToFiles []ExtraConfigMapToFilesDetails
	ExtraConfigMapsToEnvs  []ExtraConfigMapToEnvsDetails
}

type AppConfigDetails struct {
	ConfigMapName string
	FilePath      string
}

type ExtraSecretToFilesDetails struct {
	SecretName string
	FilePaths  []string
}

type ExtraSecretToEnvsDetails struct {
	SecretName string
	Envs       []string
}

type ExtraConfigMapToFilesDetails struct {
	ConfigMapName string
	FilePaths     []string
}

type ExtraConfigMapToEnvsDetails struct {
	ConfigMapName string
	Envs          []string
}