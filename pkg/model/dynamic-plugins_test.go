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

func TestDynamicPluginsValidationFailed(t *testing.T) {

	bs := simpleTestBackstage

	testObj := createBackstageTest(bs).withDefaultConfig(true).addToDefaultConfig("dynamic-plugins.yaml", "dynamic-plugins1.yaml")

	_, err := InitObjects(context.TODO(), bs, testObj.detailedSpec, true, false)

	//"failed object validation, reason: failed to apply dynamic plugins, no deployment.spec.template.spec.volumes.configMap.name = 'default-dynamic-plugins' configured\n")
	assert.Error(t, err)
}

func TestDynamicPluginsConfigured(t *testing.T) {

	bs := simpleTestBackstage

	testObj := createBackstageTest(bs).withDefaultConfig(true).
		addToDefaultConfig("dynamic-plugins.yaml", "dynamic-plugins1.yaml").
		addToDefaultConfig("deployment.yaml", "janus-deployment.yaml")

	model, err := InitObjects(context.TODO(), bs, testObj.detailedSpec, true, false)

	assert.NoError(t, err)
	assert.NotNil(t, model.backstageDeployment)

}
