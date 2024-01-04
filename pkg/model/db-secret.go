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

func init() {
	registerConfig("db-secret.yaml", DbSecretFactory{}, ForLocalDatabase)
}

// implementation of BackstageObject interface
func (b *DbSecret) Object() client.Object {
	return b.secret
}

// implementation of BackstageObject interface
func (b *DbSecret) initMetainfo(backstageMeta bsv1alpha1.Backstage, ownsRuntime bool) {
	initMetainfo(b, backstageMeta, ownsRuntime)
	b.secret.SetName(utils.GenerateRuntimeObjectName(backstageMeta.Name, "default-dbsecret"))
}

// implementation of BackstageObject interface
func (b *DbSecret) addToModel(model *RuntimeModel) {
	model.localDbSecret = b
}

// implementation of BackstageObject interface
func (b *DbSecret) EmptyObject() client.Object {
	return &corev1.Secret{}
}

// implementation of BackstageObject interface
func (b *DbSecret) validate(model *RuntimeModel) error {
	return nil
}

// implementation of LocalDbPodContributor interface
// contributes username, password, host and port to PostgreSQL container from the Secret EnvVars source
// if "template" Secret does not contain password/username (or empty) random one will be generated
func (b *DbSecret) updateLocalDbPod(model *RuntimeModel) {
	dbservice := model.localDbService.service

	// check POSTGRES_PASSWORD and generate random one if not found
	if b.secret.StringData["POSTGRES_PASSWORD"] == "" {
		pswd := rand.String(8)
		b.secret.StringData["POSTGRES_PASSWORD"] = pswd
		b.secret.StringData["POSTGRESQL_ADMIN_PASSWORD"] = pswd
	}

	// check POSTGRES_USER and generate random one if not found
	if b.secret.StringData["POSTGRES_USER"] == "" {
		b.secret.StringData["POSTGRES_USER"] = rand.String(8)
	}

	// fill the host with localDb service name
	b.secret.StringData["POSTGRES_HOST"] = dbservice.Name
	b.secret.StringData["POSTGRES_PORT"] = strconv.FormatInt(int64(dbservice.Spec.Ports[0].Port), 10)

	// populate db statefulset
	model.localDbStatefulSet.setSecretNameEnvFrom(corev1.EnvFromSource{
		SecretRef: &corev1.SecretEnvSource{
			LocalObjectReference: corev1.LocalObjectReference{Name: b.secret.Name},
		},
	})

	model.localDbSecret.secret = b.secret

}

// implementation of BackstagePodContributor interface
func (b *DbSecret) updateBackstagePod(pod *backstagePod) {
	// populate backstage deployment
	pod.addContainerEnvFrom(corev1.EnvFromSource{
		SecretRef: &corev1.SecretEnvSource{
			LocalObjectReference: corev1.LocalObjectReference{Name: b.secret.Name},
		},
	})
}
