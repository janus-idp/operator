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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	bsv1alpha1 "janus-idp.io/backstage-operator/api/v1alpha1"
	"k8s.io/utils/pointer"

	"janus-idp.io/backstage-operator/pkg/utils"

	"github.com/stretchr/testify/assert"
)

func TestDefaultRoute(t *testing.T) {
	bs := bsv1alpha1.Backstage{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "TestSpecifiedRoute",
			Namespace: "ns123",
		},
		Spec: bsv1alpha1.BackstageSpec{
			Application: &bsv1alpha1.Application{
				Route: &bsv1alpha1.Route{
					Enabled: pointer.Bool(true),
					Host:    "TestSpecifiedRoute",
					TLS:     nil,
				},
			},
		},
	}
	assert.True(t, bs.Spec.IsRouteEnabled())

	testObj := createBackstageTest(bs).withDefaultConfig(true).addToDefaultConfig("route.yaml", "raw-route.yaml")

	model, err := InitObjects(context.TODO(), bs, testObj.rawConfig, true, true, testObj.scheme)

	assert.NoError(t, err)

	assert.NotNil(t, model.route)

	assert.Equal(t, utils.GenerateRuntimeObjectName(bs.Name, "route"), model.route.route.Name)
	assert.Equal(t, model.backstageService.service.Name, model.route.route.Spec.To.Name)

	//	assert.Empty(t, model.route.route.Spec.Host)
}

func TestSpecifiedRoute(t *testing.T) {
	bs := bsv1alpha1.Backstage{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "TestSpecifiedRoute",
			Namespace: "ns123",
		},
		Spec: bsv1alpha1.BackstageSpec{
			Application: &bsv1alpha1.Application{
				Route: &bsv1alpha1.Route{
					Enabled: pointer.Bool(true),
					Host:    "TestSpecifiedRoute",
					TLS:     nil,
				},
			},
		},
	}

	assert.True(t, bs.Spec.IsRouteEnabled())

	// Test w/o default route configured
	testObjNoDef := createBackstageTest(bs).withDefaultConfig(true)
	model, err := InitObjects(context.TODO(), bs, testObjNoDef.rawConfig, true, true, testObjNoDef.scheme)

	assert.NoError(t, err)
	assert.NotNil(t, model.route)

	// check if what we have is what we specified in bs
	assert.Equal(t, utils.GenerateRuntimeObjectName(bs.Name, "route"), model.route.route.Name)
	assert.Equal(t, bs.Spec.Application.Route.Host, model.route.route.Spec.Host)

	// Test with default route configured
	testObjWithDef := testObjNoDef.addToDefaultConfig("route.yaml", "raw-route.yaml")
	model, err = InitObjects(context.TODO(), bs, testObjWithDef.rawConfig, true, true, testObjWithDef.scheme)

	assert.NoError(t, err)
	assert.NotNil(t, model.route)

	// check if what we have is what we specified in bs
	assert.Equal(t, utils.GenerateRuntimeObjectName(bs.Name, "route"), model.route.route.Name)
	assert.Equal(t, bs.Spec.Application.Route.Host, model.route.route.Spec.Host)
}
