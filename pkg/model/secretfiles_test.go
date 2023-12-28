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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	//corev1 "k8s.io/api/core/v1"

	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultSecretFiles(t *testing.T) {

	bs := simpleTestBackstage

	testObj := createBackstageTest(bs).withDefaultConfig(true).addToDefaultConfig("secret-files.yaml", "s-files.yaml")

	model, err := InitObjects(context.TODO(), bs, testObj.detailedSpec, true, false)

	assert.NoError(t, err)

	deployment := model[0].(*BackstageDeployment)
	assert.NotNil(t, deployment)

	assert.Equal(t, 1, len(deployment.deployment.Spec.Template.Spec.Containers[0].VolumeMounts))
	assert.Equal(t, 1, len(deployment.deployment.Spec.Template.Spec.Volumes))

}

func TestSpecifiedSecretFiles(t *testing.T) {

	bs := simpleTestBackstage

	sec1 := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "secret1",
			Namespace: "ns123",
		},
		StringData: map[string]string{"conf.yaml": ""},
	}

	sec2 := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "secret2",
			Namespace: "ns123",
		},
		Data: map[string][]byte{"conf2.yaml": []byte{}},
	}

	testObj := createBackstageTest(bs).withDefaultConfig(true)

	testObj.detailedSpec.AddConfigObject(&SecretFiles{Secret: &sec1, MountPath: "/my/path"})
	testObj.detailedSpec.AddConfigObject(&SecretFiles{Secret: &sec2, MountPath: "/my/path"})

	model, err := InitObjects(context.TODO(), bs, testObj.detailedSpec, true, false)

	assert.NoError(t, err)
	assert.True(t, len(model) > 0)

	deployment := model[0].(*BackstageDeployment)
	assert.NotNil(t, deployment)

	assert.Equal(t, 2, len(deployment.deployment.Spec.Template.Spec.Containers[0].VolumeMounts))
	assert.Equal(t, 0, len(deployment.deployment.Spec.Template.Spec.Containers[0].Args))
	assert.Equal(t, 2, len(deployment.deployment.Spec.Template.Spec.Volumes))

}

func TestDefaultAndSpecifiedSecretFiles(t *testing.T) {

	bs := simpleTestBackstage

	testObj := createBackstageTest(bs).withDefaultConfig(true).addToDefaultConfig("secret-files.yaml", "s-files.yaml")

	sec := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "secret1",
			Namespace: "ns123",
		},
		StringData: map[string]string{"conf.yaml": ""},
	}

	//testObj.detailedSpec.Details.AddAppConfig(cm, "/my/path")
	testObj.detailedSpec.AddConfigObject(&SecretFiles{Secret: &sec, MountPath: "/my/path"})

	model, err := InitObjects(context.TODO(), bs, testObj.detailedSpec, true, false)

	assert.NoError(t, err)
	assert.True(t, len(model) > 0)

	deployment := model[0].(*BackstageDeployment)
	assert.NotNil(t, deployment)

	assert.Equal(t, 2, len(deployment.deployment.Spec.Template.Spec.Containers[0].VolumeMounts))
	assert.Equal(t, 0, len(deployment.deployment.Spec.Template.Spec.Containers[0].Args))
	assert.Equal(t, 2, len(deployment.deployment.Spec.Template.Spec.Volumes))

}
