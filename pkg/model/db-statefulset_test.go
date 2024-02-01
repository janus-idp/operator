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
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// It tests the overriding image feature
// [GA] if we need this (and like this) feature
// we need to think about simple template engine
// for substitution env vars instead.
// Current implementation is not good
func TestOverrideDbImage(t *testing.T) {
	bs := simpleTestBackstage()

	testObj := createBackstageTest(bs).withDefaultConfig(true).
		addToDefaultConfig("db-statefulset.yaml", "janus-db-statefulset.yaml").withLocalDb("")

	_ = os.Setenv(LocalDbImageEnvVar, "dummy")

	model, err := InitObjects(context.TODO(), bs, testObj.detailedSpec, true, false, testObj.scheme)
	assert.NoError(t, err)

	assert.Equal(t, "dummy", model.localDbStatefulSet.statefulSet.Spec.Template.Spec.Containers[0].Image)
}
