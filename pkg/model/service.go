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

type BackstageService struct {
	service *corev1.Service
}

func newBackstageService() *BackstageService {
	return &BackstageService{service: &corev1.Service{}}
}

func (s *BackstageService) Object() client.Object {
	return s.service
}

func (s *BackstageService) initMetainfo(backstageMeta bsv1alpha1.Backstage, ownsRuntime bool) {
	initMetainfo(s, backstageMeta, ownsRuntime)
	s.service.SetName(utils.GenerateRuntimeObjectName(backstageMeta.Name, "service"))
	utils.GenerateLabel(&s.service.Spec.Selector, backstageAppLabel, fmt.Sprintf("backstage-%s", backstageMeta.Name))
}

func (b *BackstageService) addToModel(model *runtimeModel) {
	model.backstageService = b
}
