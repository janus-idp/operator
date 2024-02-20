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

	bsv1alpha1 "janus-idp.io/backstage-operator/api/v1alpha1"
	"k8s.io/utils/pointer"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/assert"
)

func TestDynamicPluginsValidationFailed(t *testing.T) {

	bs := bsv1alpha1.Backstage{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bs",
			Namespace: "ns123",
		},
		Spec: bsv1alpha1.BackstageSpec{
			Database: &bsv1alpha1.Database{
				EnableLocalDb: pointer.Bool(false),
			},
		},
	}

	testObj := createBackstageTest(bs).withDefaultConfig(true).
		addToDefaultConfig("dynamic-plugins.yaml", "raw-dynamic-plugins.yaml")

	_, err := InitObjects(context.TODO(), bs, testObj.rawConfig, []corev1.ConfigMap{}, true, false, testObj.scheme)

	//"failed object validation, reason: failed to find initContainer named install-dynamic-plugins")
	assert.Error(t, err)

}

func TestDefaultDynamicPlugins(t *testing.T) {

	bs := simpleTestBackstage()

	testObj := createBackstageTest(bs).withDefaultConfig(true).
		addToDefaultConfig("dynamic-plugins.yaml", "raw-dynamic-plugins.yaml").
		addToDefaultConfig("deployment.yaml", "janus-deployment.yaml")

	model, err := InitObjects(context.TODO(), bs, testObj.rawConfig, []corev1.ConfigMap{}, true, false, testObj.scheme)

	assert.NoError(t, err)
	assert.NotNil(t, model.backstageDeployment)
	//dynamic-plugins-root
	//dynamic-plugins-npmrc
	//vol-default-dynamic-plugins
	assert.Equal(t, 3, len(model.backstageDeployment.deployment.Spec.Template.Spec.Volumes))

	ic := initContainer(model)
	assert.NotNil(t, ic)
	//dynamic-plugins-root
	//dynamic-plugins-npmrc
	//vol-default-dynamic-plugins
	assert.Equal(t, 3, len(ic.VolumeMounts))
}

func TestSpecifiedDynamicPlugins(t *testing.T) {

	bs := simpleTestBackstage()

	testObj := createBackstageTest(bs).withDefaultConfig(true).
		addToDefaultConfig("dynamic-plugins.yaml", "raw-dynamic-plugins.yaml").
		addToDefaultConfig("deployment.yaml", "janus-deployment.yaml")

	_ = corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dplugin",
			Namespace: "ns123",
		},
		Data: map[string]string{"dynamic-plugins.yaml": ""},
	}

	//testObj.detailedSpec.AddConfigObject(&DynamicPlugins{ConfigMap: &cm})

	model, err := InitObjects(context.TODO(), bs, testObj.rawConfig, []corev1.ConfigMap{}, true, false, testObj.scheme)

	assert.NoError(t, err)
	assert.NotNil(t, model)

	ic := initContainer(model)
	assert.NotNil(t, ic)
	//dynamic-plugins-root
	//dynamic-plugins-npmrc
	//vol-dplugin
	assert.Equal(t, 3, len(ic.VolumeMounts))
}

func initContainer(model *BackstageModel) *corev1.Container {
	for _, v := range model.backstageDeployment.deployment.Spec.Template.Spec.InitContainers {
		if v.Name == dynamicPluginInitContainerName {
			return &v
		}
	}
	return nil
}
