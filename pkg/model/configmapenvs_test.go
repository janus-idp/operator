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

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfigMapEnvFrom(t *testing.T) {

	bs := simpleTestBackstage()

	testObj := createBackstageTest(bs).withDefaultConfig(true).addToDefaultConfig("configmap-envs.yaml", "raw-cm-envs.yaml")

	model, err := InitObjects(context.TODO(), bs, testObj.detailedSpec, true, false, testObj.scheme)

	assert.NoError(t, err)
	assert.NotNil(t, model)

	bscontainer := model.backstageDeployment.pod.container
	assert.NotNil(t, bscontainer)

	assert.Equal(t, 1, len(bscontainer.EnvFrom))
	assert.Equal(t, 0, len(bscontainer.Env))

}

func TestSpecifiedConfigMapEnvs(t *testing.T) {

	bs := simpleTestBackstage()

	testObj := createBackstageTest(bs).withDefaultConfig(true)

	cm := corev1.ConfigMap{Data: map[string]string{
		"ENV1": "Val",
	}}

	testObj.detailedSpec.AddConfigObject(&ConfigMapEnvs{ConfigMap: &cm, Key: "ENV1"})

	model, err := InitObjects(context.TODO(), bs, testObj.detailedSpec, true, false, testObj.scheme)

	assert.NoError(t, err)
	assert.NotNil(t, model)

	bscontainer := model.backstageDeployment.pod.container
	assert.NotNil(t, bscontainer)
	assert.Equal(t, 1, len(bscontainer.Env))
	assert.NotNil(t, bscontainer.Env[0])
	assert.Equal(t, "ENV1", bscontainer.Env[0].ValueFrom.ConfigMapKeyRef.Key)

}
