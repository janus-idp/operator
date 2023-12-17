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
	"strconv"

	"k8s.io/apimachinery/pkg/util/rand"

	bsv1alpha1 "janus-idp.io/backstage-operator/api/v1alpha1"
	"janus-idp.io/backstage-operator/pkg/utils"
	corev1 "k8s.io/api/core/v1"

	//	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DbSecretFactory struct{}

func (f DbSecretFactory) newBackstageObject() BackstageObject {
	return &DbSecret{secret: &corev1.Secret{}}
}

type DbSecret struct {
	secret *corev1.Secret
}

//func newDbSecret() *DbSecret {
//	return &DbSecret{secret: &corev1.Secret{}}
//}

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

func (b *DbSecret) EmptyObject() client.Object {
	return &corev1.Secret{}
}

func (b *DbSecret) updateLocalDbPod(model *runtimeModel) {
	dbservice := model.localDbService.service

	// fill the host with localDb service name
	b.secret.StringData["POSTGRES_HOST"] = dbservice.Name
	b.secret.StringData["POSTGRES_PORT"] = strconv.FormatInt(int64(dbservice.Spec.Ports[0].Port), 10)

	// populate db statefulset
	model.localDbStatefulSet.appendContainerEnvFrom(corev1.EnvFromSource{
		SecretRef: &corev1.SecretEnvSource{
			LocalObjectReference: corev1.LocalObjectReference{Name: b.secret.Name},
		},
	})

}

func (b *DbSecret) updateBackstagePod(pod *backstagePod) {
	// populate backstage deployment
	pod.appendContainerEnvFrom(corev1.EnvFromSource{
		SecretRef: &corev1.SecretEnvSource{
			LocalObjectReference: corev1.LocalObjectReference{Name: b.secret.Name},
		},
	})
}

func (b *DbSecret) OnCreate() error {

	if b.secret.StringData["POSTGRES_PASSWORD"] == "" {
		pswd := rand.String(8)
		b.secret.StringData["POSTGRES_PASSWORD"] = pswd
		b.secret.StringData["POSTGRESQL_ADMIN_PASSWORD"] = pswd
	}

	return nil
}
