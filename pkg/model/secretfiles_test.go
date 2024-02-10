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

	bsv1alpha1 "janus-idp.io/backstage-operator/api/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	//corev1 "k8s.io/api/core/v1"

	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	secretFilesTestSecret = corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "secret1",
			Namespace: "ns123",
		},
		StringData: map[string]string{"conf.yaml": ""},
	}

	secretFilesTestSecret2 = corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "secret2",
			Namespace: "ns123",
		},
		StringData: map[string]string{"conf2.yaml": ""},
	}

	secretFilesTestBackstage = bsv1alpha1.Backstage{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bs",
			Namespace: "ns123",
		},
		Spec: bsv1alpha1.BackstageSpec{
			Application: &bsv1alpha1.Application{
				ExtraFiles: &bsv1alpha1.ExtraFiles{
					MountPath: "/my/path",
					Secrets:   []bsv1alpha1.ObjectKeyRef{},
				},
			},
		},
	}
)

func TestDefaultSecretFiles(t *testing.T) {

	bs := simpleTestBackstage()

	testObj := createBackstageTest(bs).withDefaultConfig(true).addToDefaultConfig("secret-files.yaml", "raw-secret-files.yaml")

	model, err := InitObjects(context.TODO(), bs, testObj.rawConfig, true, false, testObj.scheme)

	assert.NoError(t, err)

	deployment := model.backstageDeployment
	assert.NotNil(t, deployment)

	assert.Equal(t, 1, len(deployment.deployment.Spec.Template.Spec.Containers[0].VolumeMounts))
	assert.Equal(t, 1, len(deployment.deployment.Spec.Template.Spec.Volumes))

}

func TestSpecifiedSecretFiles(t *testing.T) {

	bs := *secretFilesTestBackstage.DeepCopy()
	sf := &bs.Spec.Application.ExtraFiles.Secrets
	*sf = append(*sf, bsv1alpha1.ObjectKeyRef{Name: secretFilesTestSecret.Name})
	*sf = append(*sf, bsv1alpha1.ObjectKeyRef{Name: secretFilesTestSecret2.Name})

	testObj := createBackstageTest(bs).withDefaultConfig(true)

	model, err := InitObjects(context.TODO(), bs, testObj.rawConfig, true, false, testObj.scheme)

	assert.NoError(t, err)
	assert.True(t, len(model.RuntimeObjects) > 0)

	deployment := model.backstageDeployment
	assert.NotNil(t, deployment)

	assert.Equal(t, 2, len(deployment.deployment.Spec.Template.Spec.Containers[0].VolumeMounts))
	assert.Equal(t, 0, len(deployment.deployment.Spec.Template.Spec.Containers[0].Args))
	assert.Equal(t, 2, len(deployment.deployment.Spec.Template.Spec.Volumes))

	//t.Log(">>>>", deployment.deployment.Spec.Template.Spec.Containers[0].VolumeMounts)

}

func TestDefaultAndSpecifiedSecretFiles(t *testing.T) {

	bs := *secretFilesTestBackstage.DeepCopy()
	sf := &bs.Spec.Application.ExtraFiles.Secrets
	*sf = append(*sf, bsv1alpha1.ObjectKeyRef{Name: secretFilesTestSecret.Name})
	testObj := createBackstageTest(bs).withDefaultConfig(true).addToDefaultConfig("secret-files.yaml", "raw-secret-files.yaml")

	model, err := InitObjects(context.TODO(), bs, testObj.rawConfig, true, false, testObj.scheme)

	assert.NoError(t, err)
	assert.True(t, len(model.RuntimeObjects) > 0)

	deployment := model.backstageDeployment
	assert.NotNil(t, deployment)

	assert.Equal(t, 2, len(deployment.deployment.Spec.Template.Spec.Containers[0].VolumeMounts))
	assert.Equal(t, 0, len(deployment.deployment.Spec.Template.Spec.Containers[0].Args))
	assert.Equal(t, 2, len(deployment.deployment.Spec.Template.Spec.Volumes))

}
