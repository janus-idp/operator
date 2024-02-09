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

type DbSecretFactory struct{}

func (f DbSecretFactory) newBackstageObject() RuntimeObject {
	return &DbSecret{}
}

type DbSecret struct {
	secret *corev1.Secret
}

func init() {
	registerConfig("db-secret.yaml", DbSecretFactory{})
}

func DbSecretDefaultName(backstageName string) string {
	return utils.GenerateRuntimeObjectName(backstageName, "default-dbsecret")
}

// implementation of RuntimeObject interface
func (b *DbSecret) Object() client.Object {
	return b.secret
}

func (b *DbSecret) setObject(obj client.Object, name string) {
	b.secret = nil
	if obj != nil {
		b.secret = obj.(*corev1.Secret)
	}
}

// implementation of RuntimeObject interface
func (b *DbSecret) addToModel(model *BackstageModel, backstage bsv1alpha1.Backstage, ownsRuntime bool) error {
	if b.secret == nil && !backstage.Spec.IsAuthSecretSpecified() {
		return nil
	}

	if backstage.Spec.IsAuthSecretSpecified() {
		b.secret = &corev1.Secret{}
		b.secret.SetName(backstage.Spec.Database.AuthSecretName)
	} else {
		b.secret.SetName(DbSecretDefaultName(backstage.Name))
	}

	model.LocalDbSecret = b
	//model.setRuntimeObject(b)

	return nil
}

// implementation of RuntimeObject interface
func (b *DbSecret) EmptyObject() client.Object {
	return &corev1.Secret{}
}

// implementation of RuntimeObject interface
func (b *DbSecret) validate(model *BackstageModel, backstage bsv1alpha1.Backstage) error {
	return nil
}

//func (b *DbSecret) updateSecret(model *BackstageModel) {
//
//	dbservice := model.LocalDbService.service
//	if b.secret.StringData == nil {
//		b.secret.StringData = map[string]string{}
//	}
//	// fill the host with localDb service name
//	b.secret.StringData["POSTGRES_HOST"] = dbservice.Name
//
//	//// fill the port with localDb service port
//	b.secret.StringData["POSTGRES_PORT"] = strconv.FormatInt(int64(dbservice.Spec.Ports[0].Port), 10)
//
//	// populate db statefulset
//	model.localDbStatefulSet.setSecretNameEnvFrom(corev1.EnvFromSource{
//		SecretRef: &corev1.SecretEnvSource{
//			LocalObjectReference: corev1.LocalObjectReference{Name: b.secret.Name},
//		},
//	})
//
//	// populate backstage deployment
//	model.backstageDeployment.pod.addContainerEnvFrom(corev1.EnvFromSource{
//		SecretRef: &corev1.SecretEnvSource{
//			LocalObjectReference: corev1.LocalObjectReference{Name: b.secret.Name},
//		},
//	})
//}
//
//func generatePassword(length int) (string, error) {
//	bytes := make([]byte, length)
//	if _, err := rand.Read(bytes); err != nil {
//		return "", err
//	}
//	// Encode the password to prevent special characters
//	return base64.StdEncoding.EncodeToString(bytes), nil
//}
