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
	bsv1alpha1 "janus-idp.io/backstage-operator/api/v1alpha1"
	"janus-idp.io/backstage-operator/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DbSecret struct {
	secret *corev1.Secret
}

func newDbSecret() *DbSecret {
	return &DbSecret{secret: &corev1.Secret{}}
}

func (b *DbSecret) Object() client.Object {
	return b.secret
}

func (b *DbSecret) initMetainfo(backstageMeta bsv1alpha1.Backstage, ownsRuntime bool) {
	initMetainfo(b, backstageMeta, ownsRuntime)
	b.secret.SetName(utils.GenerateRuntimeObjectName(backstageMeta.Name, "default-dbsecret"))
}

func (b *DbSecret) addToModel(model *runtimeModel) {
	model.localDbSecret = b
}

func (b *DbSecret) updateSecret(backstageDeployment *BackstageDeployment, localDbDeployment *DbStatefulSet, localDbService *DbService) error {
	b.secret.StringData["POSTGRES_HOST"] = localDbService.service.Name
	//TODO
	return nil
}
