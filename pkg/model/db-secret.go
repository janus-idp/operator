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
	"encoding/base64"
	"strconv"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"crypto/rand"

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
	secret        *corev1.Secret
	nameSpecified bool
}

// TODO: consider to get it back
//func init() {
//	registerConfig("db-secret.yaml", DbSecretFactory{}, ForLocalDatabase)
//}

func DbSecretDefaultName(backstageName string) string {
	return utils.GenerateRuntimeObjectName(backstageName, "default-dbsecret")
}

func NewDbSecretFromSpec(name string) DbSecret {
	return DbSecret{
		secret: &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		},
		nameSpecified: true,
	}
}

func ExistedDbSecret(sec corev1.Secret) DbSecret {
	return DbSecret{
		secret:        &sec,
		nameSpecified: true,
	}
}

func GenerateDbSecret() DbSecret {
	// generate password
	pswd, _ := generatePassword(24)
	return DbSecret{
		secret: &corev1.Secret{
			StringData: map[string]string{
				"POSTGRES_PASSWORD":         pswd,
				"POSTGRESQL_ADMIN_PASSWORD": pswd,
				"POSTGRES_USER":             "postgres",
			},
		},
		nameSpecified: false,
	}
}

// implementation of BackstageObject interface
func (b *DbSecret) Object() client.Object {
	return b.secret
}

// implementation of BackstageObject interface
func (b *DbSecret) addToModel(model *RuntimeModel, backstageMeta bsv1alpha1.Backstage, ownsRuntime bool) {
	model.localDbSecret = b
	model.setObject(b)

	// TODO refactor it: b.secret should not be nil at this stage
	if b.secret == nil {
		b.secret = GenerateDbSecret().secret
	}

	if !b.nameSpecified {
		b.secret.SetName(DbSecretDefaultName(backstageMeta.Name))
	}
}

// implementation of BackstageObject interface
func (b *DbSecret) EmptyObject() client.Object {
	return &corev1.Secret{}
}

// implementation of BackstageObject interface
func (b *DbSecret) validate(model *RuntimeModel) error {
	return nil
}

// implementation of BackstagePodContributor interface
//func (b *DbSecret) updateBackstagePod(pod *backstagePod) {
//	// populate backstage deployment
//	pod.addContainerEnvFrom(corev1.EnvFromSource{
//		SecretRef: &corev1.SecretEnvSource{
//			LocalObjectReference: corev1.LocalObjectReference{Name: b.secret.Name},
//		},
//	})
//}

func (b *DbSecret) updateSecret(model *RuntimeModel) {

	dbservice := model.localDbService.service
	if b.secret.StringData == nil {
		b.secret.StringData = map[string]string{}
	}
	// fill the host with localDb service name
	b.secret.StringData["POSTGRES_HOST"] = dbservice.Name

	//// fill the port with localDb service port
	b.secret.StringData["POSTGRES_PORT"] = strconv.FormatInt(int64(dbservice.Spec.Ports[0].Port), 10)

	// populate db statefulset
	model.localDbStatefulSet.setSecretNameEnvFrom(corev1.EnvFromSource{
		SecretRef: &corev1.SecretEnvSource{
			LocalObjectReference: corev1.LocalObjectReference{Name: b.secret.Name},
		},
	})

	// populate backstage deployment
	model.backstageDeployment.pod.addContainerEnvFrom(corev1.EnvFromSource{
		SecretRef: &corev1.SecretEnvSource{
			LocalObjectReference: corev1.LocalObjectReference{Name: b.secret.Name},
		},
	})
}

func generatePassword(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	// Encode the password to prevent special characters
	return base64.StdEncoding.EncodeToString(bytes), nil
}
