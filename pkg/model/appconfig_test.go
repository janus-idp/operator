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

	corev1 "k8s.io/api/core/v1"

	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDefaultAppConfig(t *testing.T) {

	bs := simpleTestBackstage()

	testObj := createBackstageTest(bs).withDefaultConfig(true).addToDefaultConfig("app-config.yaml", "raw-app-config.yaml")

	model, err := InitObjects(context.TODO(), bs, testObj.detailedSpec, true, false, testObj.scheme)

	assert.NoError(t, err)
	assert.True(t, len(model.Objects) > 0)

	deployment := model.backstageDeployment
	assert.NotNil(t, deployment)

	assert.Equal(t, 1, len(deployment.deployment.Spec.Template.Spec.Containers[0].VolumeMounts))
	assert.Equal(t, 2, len(deployment.deployment.Spec.Template.Spec.Containers[0].Args))
	assert.Equal(t, 1, len(deployment.deployment.Spec.Template.Spec.Volumes))

}

func TestSpecifiedAppConfig(t *testing.T) {

	bs := simpleTestBackstage()

	cm := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "app-config1",
			Namespace: "ns123",
		},
		Data: map[string]string{"conf.yaml": ""},
	}

	cm2 := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "app-config2",
			Namespace: "ns123",
		},
		Data: map[string]string{"conf2.yaml": ""},
	}

	testObj := createBackstageTest(bs).withDefaultConfig(true)

	testObj.detailedSpec.AddConfigObject(&AppConfig{ConfigMap: &cm, MountPath: "/my/path"})
	testObj.detailedSpec.AddConfigObject(&AppConfig{ConfigMap: &cm2, MountPath: "/my/path"})

	model, err := InitObjects(context.TODO(), bs, testObj.detailedSpec, true, false, testObj.scheme)

	assert.NoError(t, err)
	assert.True(t, len(model.Objects) > 0)

	deployment := model.backstageDeployment
	assert.NotNil(t, deployment)

	assert.Equal(t, 2, len(deployment.deployment.Spec.Template.Spec.Containers[0].VolumeMounts))
	assert.Equal(t, 4, len(deployment.deployment.Spec.Template.Spec.Containers[0].Args))
	assert.Equal(t, 2, len(deployment.deployment.Spec.Template.Spec.Volumes))

}

func TestDefaultAndSpecifiedAppConfig(t *testing.T) {

	bs := simpleTestBackstage()

	testObj := createBackstageTest(bs).withDefaultConfig(true).addToDefaultConfig("app-config.yaml", "raw-app-config.yaml")

	cm := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "app-config1",
			Namespace: "ns123",
		},
		Data: map[string]string{"conf.yaml": ""},
	}

	testObj.detailedSpec.AddConfigObject(&AppConfig{ConfigMap: &cm, MountPath: "/my/path"})

	model, err := InitObjects(context.TODO(), bs, testObj.detailedSpec, true, false, testObj.scheme)

	assert.NoError(t, err)
	assert.True(t, len(model.Objects) > 0)

	deployment := model.backstageDeployment
	assert.NotNil(t, deployment)

	assert.Equal(t, 2, len(deployment.deployment.Spec.Template.Spec.Containers[0].VolumeMounts))
	assert.Equal(t, 4, len(deployment.deployment.Spec.Template.Spec.Containers[0].Args))
	assert.Equal(t, 2, len(deployment.deployment.Spec.Template.Spec.Volumes))

}
