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

	bsv1alpha1 "janus-idp.io/backstage-operator/api/v1alpha1"
	"janus-idp.io/backstage-operator/pkg/utils"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type BackstageServiceFactory struct{}

func (f BackstageServiceFactory) newBackstageObject() BackstageObject {
	return &BackstageService{service: &corev1.Service{}}
}

type BackstageService struct {
	service *corev1.Service
}

func init() {
	registerConfig("service.yaml", BackstageServiceFactory{}, Mandatory)
}

// implementation of BackstageObject interface
func (b *BackstageService) Object() client.Object {
	return b.service
}

// implementation of BackstageObject interface
func (b *BackstageService) addToModel(model *RuntimeModel, backstageMeta bsv1alpha1.Backstage, ownsRuntime bool) {
	model.backstageService = b
	model.setObject(b)

	b.service.SetName(utils.GenerateRuntimeObjectName(backstageMeta.Name, "service"))
	utils.GenerateLabel(&b.service.Spec.Selector, backstageAppLabel, fmt.Sprintf("backstage-%s", backstageMeta.Name))

}

// implementation of BackstageObject interface
func (b *BackstageService) EmptyObject() client.Object {
	return &corev1.Service{}
}

// implementation of BackstageObject interface
func (b *BackstageService) validate(model *RuntimeModel) error {
	return nil
}
