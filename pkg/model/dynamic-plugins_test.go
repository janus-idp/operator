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

func TestDynamicPluginsValidationFailed(t *testing.T) {

	bs := simpleTestBackstage

	testObj := createBackstageTest(bs).withDefaultConfig(true).
		addToDefaultConfig("dynamic-plugins.yaml", "dynamic-plugins1.yaml")

	_, err := InitObjects(context.TODO(), bs, testObj.detailedSpec, true, false)

	//"failed object validation, reason: failed to find initContainer named install-dynamic-plugins")
	assert.Error(t, err)

}

func TestDefaultDynamicPlugins(t *testing.T) {

	bs := simpleTestBackstage

	testObj := createBackstageTest(bs).withDefaultConfig(true).
		addToDefaultConfig("dynamic-plugins.yaml", "dynamic-plugins1.yaml").
		addToDefaultConfig("deployment.yaml", "janus-deployment.yaml")

	model, err := InitObjects(context.TODO(), bs, testObj.detailedSpec, true, false)

	assert.NoError(t, err)
	assert.NotNil(t, model.backstageDeployment)
	//dynamic-plugins-root
	//dynamic-plugins-npmrc
	//vol-default-dynamic-plugins
	assert.Equal(t, 3, len(model.backstageDeployment.deployment.Spec.Template.Spec.Volumes))
	//for _, v := range model.backstageDeployment.deployment.Spec.Template.Spec.Volumes {
	//	t.Log(">>>>>>>>>>>>>>>>>>>> ", v.Name, v.ConfigMap)
	//
	//}

}

func TestSpecifiedDynamicPlugins(t *testing.T) {

	bs := simpleTestBackstage

	testObj := createBackstageTest(bs).withDefaultConfig(true).
		addToDefaultConfig("dynamic-plugins.yaml", "dynamic-plugins1.yaml").
		addToDefaultConfig("deployment.yaml", "janus-deployment.yaml")

	cm := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dplugin",
			Namespace: "ns123",
		},
		Data: map[string]string{"dynamic-plugins.yaml": ""},
	}

	testObj.detailedSpec.AddConfigObject(&DynamicPlugins{ConfigMap: &cm})

	model, err := InitObjects(context.TODO(), bs, testObj.detailedSpec, true, false)

	assert.NoError(t, err)
	assert.NotNil(t, model)

	//for _, v := range model.backstageDeployment.deployment.Spec.Template.Spec.Volumes {
	//	t.Log(">>>>>>>>>>>>>>>>>>>> ", v.Name, v.ConfigMap)
	//
	//}
	//
	//for _, v := range model.backstageDeployment.deployment.Spec.Template.Spec.InitContainers {
	//	t.Log(">>>>>>>MOUNT>>>>>>>>>>>>> ", v.Name, v.VolumeMounts)
	//
	//}

	//"failed object validation, reason: failed to apply dynamic plugins, no deployment.spec.template.spec.volumes.ConfigMap.name = 'default-dynamic-plugins' configured\n")
	//assert.Error(t, err)
}
