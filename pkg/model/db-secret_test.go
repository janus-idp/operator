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
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/assert"
)

func TestDefaultWithDefinedSecrets(t *testing.T) {

	bs := simpleTestBackstage

	testObj := createBackstageTest(bs).withDefaultConfig(true).withLocalDb().addToDefaultConfig("db-secret.yaml", "db-defined-secret.yaml")

	model, err := InitObjects(context.TODO(), bs, testObj.detailedSpec, true, false)

	assert.NoError(t, err)
	assert.NotNil(t, model.localDbSecret)
	assert.Equal(t, "bs-default-dbsecret", model.localDbSecret.secret.Name)
	assert.Equal(t, "postgres", model.localDbSecret.secret.StringData["POSTGRES_USER"])

	dbss := model.localDbStatefulSet
	assert.NotNil(t, dbss)
	assert.Equal(t, 1, len(dbss.container().EnvFrom))

	assert.Equal(t, model.localDbSecret.secret.Name, dbss.container().EnvFrom[0].SecretRef.Name)
}

func TestDefaultWithGeneratedSecrets(t *testing.T) {
	bs := simpleTestBackstage

	testObj := createBackstageTest(bs).withDefaultConfig(true).withLocalDb().addToDefaultConfig("db-secret.yaml", "db-generated-secret.yaml")

	model, err := InitObjects(context.TODO(), bs, testObj.detailedSpec, true, false)

	assert.NoError(t, err)
	assert.Equal(t, "bs-default-dbsecret", model.localDbSecret.secret.Name)
	assert.NotEmpty(t, model.localDbSecret.secret.StringData["POSTGRES_USER"])
	assert.NotEmpty(t, model.localDbSecret.secret.StringData["POSTGRES_PASSWORD"])

	dbss := model.localDbStatefulSet
	assert.NotNil(t, dbss)
	assert.Equal(t, 1, len(dbss.container().EnvFrom))
	assert.Equal(t, model.localDbSecret.secret.Name, dbss.container().EnvFrom[0].SecretRef.Name)
}

func TestSpecifiedSecret(t *testing.T) {
	bs := simpleTestBackstage

	sec1 := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "custom-db-secret",
			Namespace: "ns123",
		},
		StringData: map[string]string{
			"POSTGRES_USER":     "user",
			"POSTGRES_PASSWORD": "password",
		},
	}

	testObj := createBackstageTest(bs).withDefaultConfig(true).withLocalDb().addToDefaultConfig("db-secret.yaml", "db-generated-secret.yaml")

	testObj.detailedSpec.AddConfigObject(&DbSecret{secret: &sec1})

	model, err := InitObjects(context.TODO(), bs, testObj.detailedSpec, true, false)

	assert.NoError(t, err)
	assert.Equal(t, "custom-db-secret", model.localDbSecret.secret.Name)

	assert.NotEmpty(t, model.localDbSecret.secret.StringData["POSTGRES_USER"])
	assert.NotEmpty(t, model.localDbSecret.secret.StringData["POSTGRES_PASSWORD"])
	assert.Equal(t, model.localDbSecret.secret.Name, model.localDbStatefulSet.container().EnvFrom[0].SecretRef.Name)

}
