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
	"fmt"
	"testing"

	"k8s.io/utils/pointer"

	"janus-idp.io/backstage-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/assert"
)

// NOTE: to make it work locally env var LOCALBIN should point to the directory where default-config folder located
func TestInitDefaultDeploy(t *testing.T) {

	//setTestEnv()

	bs := v1alpha1.Backstage{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bs",
			Namespace: "ns123",
		},
		Spec: v1alpha1.BackstageSpec{
			Database: v1alpha1.Database{
				EnableLocalDb: pointer.Bool(false),
			},
		},
	}

	testObj := createBackstageTest(bs).withDefaultConfig(true)

	model, err := InitObjects(context.TODO(), bs, testObj.detailedSpec, true, false)

	assert.NoError(t, err)
	assert.True(t, len(model.Objects) > 0)
	assert.Equal(t, "bs-deployment", model.backstageDeployment.Object().GetName())
	assert.Equal(t, "ns123", model.backstageDeployment.Object().GetNamespace())
	assert.Equal(t, 2, len(model.backstageDeployment.Object().GetLabels()))
	//	assert.Equal(t, 1, len(model[0].Object().GetOwnerReferences()))

	bsDeployment := model.backstageDeployment
	assert.NotNil(t, bsDeployment.pod.container)
	assert.Equal(t, backstageContainerName, bsDeployment.pod.container.Name)
	assert.NotNil(t, bsDeployment.pod.volumes)

	//	assert.Equal(t, "Backstage", bsDeployment.deployment.OwnerReferences[0].Kind)

	bsService := model.backstageService
	assert.Equal(t, "bs-service", bsService.service.Name)
	assert.True(t, len(bsService.service.Spec.Ports) > 0)

	assert.Equal(t, fmt.Sprintf("backstage-%s", "bs"), bsDeployment.deployment.Spec.Template.ObjectMeta.Labels[backstageAppLabel])
	assert.Equal(t, fmt.Sprintf("backstage-%s", "bs"), bsService.service.Spec.Selector[backstageAppLabel])

}

func TestIfEmptyObjectIsValid(t *testing.T) {

	bs := simpleTestBackstage
	testObj := createBackstageTest(bs).withDefaultConfig(true)

	model, err := InitObjects(context.TODO(), bs, testObj.detailedSpec, true, false)
	assert.NoError(t, err)

	assert.Equal(t, len(model.Objects), 2)

}

// [GA]Can be helpful to explore new model features (for example for Db and Route removing
// Do not remove it.

//func TestIfModelObjectAndArrayElementIsTheSame(t *testing.T) {
//
//	bs := simpleTestBackstage
//	testObj := createBackstageTest(bs).withDefaultConfig(true)
//
//	model, err := InitObjects(context.TODO(), bs, testObj.detailedSpec, true, false)
//	assert.NoError(t, err)
//
//	bs.Spec.Application = &v1alpha1.Application{
//		Route: &v1alpha1.Route{
//			Enabled: pointer.Bool(false),
//		},
//	}
//
//	route := BackstageRouteFactory{}.newBackstageObject()
//	model.Objects = append(model.Objects, route)
//	for _, o := range model.Objects {
//
//		if _, ok := o.(*BackstageRoute); ok {
//			t.Log(">>>>>>>>>>>OB>>>>>>>>>>>>>> ", o.Object())
//			assert.Nil(t, o.Object())
//			break
//		}
//	}
//	//t.Error("Model does not contain BackstageRoute")
//
//}
