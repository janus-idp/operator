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

func TestRouteSpec(t *testing.T) {
	bs := simpleTestBackstage

	testObj := createBackstageTest(bs).withDefaultConfig(true).addToDefaultConfig("route.yaml", "route.yaml")

	model, err := InitObjects(context.TODO(), bs, testObj.detailedSpec, true, true)

	assert.NoError(t, err)

	assert.NotNil(t, model.route)
	assert.Equal(t, model.backstageService.service.Name, model.route.route.Spec.To.Name)

}