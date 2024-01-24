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
	bs "janus-idp.io/backstage-operator/api/v1alpha1"
)

// extension of Backstage.Spec to make it possible to work on model package level
type DetailedBackstageSpec struct {
	bs.BackstageSpec
	RawConfigContent map[string]string
	ConfigObjects    backstageConfigs
	LocalDbSecret    DbSecret
}

// array of BackstagePodContributor interfaces
type backstageConfigs []BackstagePodContributor

func (a *DetailedBackstageSpec) AddConfigObject(obj BackstagePodContributor) {
	a.ConfigObjects = append(a.ConfigObjects, obj)
}

//func (a *DetailedBackstageSpec) SetDbSecret(secret *corev1.Secret) {
//	a.LocalDbSecret = DbSecret{secret: secret}
//}
