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

	"github.com/stretchr/testify/assert"
)

func TestDefaultWithDefinedSecrets(t *testing.T) {

	bs := simpleTestBackstage

	// expected generatePassword = false (default db-secret defined) will come from preprocess
	testObj := createBackstageTest(bs).withDefaultConfig(true).withLocalDb(false).addToDefaultConfig("db-secret.yaml", "db-defined-secret.yaml")

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

func TestEmptyDbSecret(t *testing.T) {

	bs := simpleTestBackstage

	// expected generatePassword = false (default db-secret defined) will come from preprocess
	testObj := createBackstageTest(bs).withDefaultConfig(true).withLocalDb(false).addToDefaultConfig("db-secret.yaml", "db-empty-secret.yaml")

	model, err := InitObjects(context.TODO(), bs, testObj.detailedSpec, true, false)

	assert.NoError(t, err)
	assert.NotNil(t, model.localDbSecret)
	assert.Equal(t, "bs-default-dbsecret", model.localDbSecret.secret.Name)
	assert.NotEmpty(t, model.localDbSecret.secret.StringData["POSTGRES_USER"])
	_, ok := model.localDbSecret.secret.StringData["POSTGRES_PASSWORD"]
	assert.True(t, ok)
	//	assert.NotEmpty(t, model.localDbSecret.secret.StringData["POSTGRES_PASSWORD"])

	assert.Equal(t, "postgres", model.localDbSecret.secret.StringData["POSTGRES_USER"])

	dbss := model.localDbStatefulSet
	assert.NotNil(t, dbss)
	assert.Equal(t, 1, len(dbss.container().EnvFrom))

	assert.Equal(t, model.localDbSecret.secret.Name, dbss.container().EnvFrom[0].SecretRef.Name)
}

func TestDefaultWithGeneratedSecrets(t *testing.T) {
	bs := simpleTestBackstage

	// expected generatePassword = true (no db-secret defined) will come from preprocess
	testObj := createBackstageTest(bs).withDefaultConfig(true).withLocalDb(true).addToDefaultConfig("db-secret.yaml", "db-generated-secret.yaml")

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
	bs.Spec.Database.AuthSecretName = "custom-db-secret"

	// expected generatePassword = false (db-secret defined in the spec) will come from preprocess
	testObj := createBackstageTest(bs).withDefaultConfig(true).withLocalDb(false).addToDefaultConfig("db-secret.yaml", "db-generated-secret.yaml")

	model, err := InitObjects(context.TODO(), bs, testObj.detailedSpec, true, false)

	assert.NoError(t, err)
	assert.Equal(t, "custom-db-secret", model.localDbSecret.secret.Name)

	//assert.Equal(t, sec1["POSTGRES_USER"], model.localDbSecret.secret.StringData["POSTGRES_USER"])
	//assert.NotEmpty(t, model.localDbSecret.secret.StringData["POSTGRES_USER"])
	//assert.Equal(t, sec1.StringData["POSTGRES_PASSWORD"], model.localDbSecret.secret.StringData["POSTGRES_PASSWORD"])
	//assert.NotEmpty(t, model.localDbSecret.secret.StringData["POSTGRES_PASSWORD"])
	//assert.Equal(t, model.localDbSecret.secret.Name, model.localDbStatefulSet.container().EnvFrom[0].SecretRef.Name)

}
