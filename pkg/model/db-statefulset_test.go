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

	"k8s.io/utils/ptr"

	bsv1alpha1 "redhat-developer/red-hat-developer-hub-operator/api/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/assert"
)

var dbStatefulSetBackstage = &bsv1alpha1.Backstage{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "bs",
		Namespace: "ns123",
	},
	Spec: bsv1alpha1.BackstageSpec{
		Database:    &bsv1alpha1.Database{},
		Application: &bsv1alpha1.Application{},
	},
}

// It tests the overriding image feature
func TestOverrideDbImage(t *testing.T) {
	bs := *dbStatefulSetBackstage.DeepCopy()

	bs.Spec.Database.EnableLocalDb = ptr.To(false)

	testObj := createBackstageTest(bs).withDefaultConfig(true).
		addToDefaultConfig("db-statefulset.yaml", "janus-db-statefulset.yaml").withLocalDb()

	_ = os.Setenv(LocalDbImageEnvVar, "dummy")

	model, err := InitObjects(context.TODO(), bs, testObj.externalConfig, true, false, testObj.scheme)
	assert.NoError(t, err)

	assert.Equal(t, "dummy", model.localDbStatefulSet.statefulSet.Spec.Template.Spec.Containers[0].Image)
}

// test bs.Spec.Application.ImagePullSecrets shared with StatefulSet
func TestImagePullSecretSpec(t *testing.T) {
	bs := *dbStatefulSetBackstage.DeepCopy()
	bs.Spec.Application.ImagePullSecrets = []string{"my-secret1", "my-secret2"}

	testObj := createBackstageTest(bs).withDefaultConfig(true)
	model, err := InitObjects(context.TODO(), bs, testObj.externalConfig, true, false, testObj.scheme)
	assert.NoError(t, err)

	assert.Equal(t, 2, len(model.localDbStatefulSet.statefulSet.Spec.Template.Spec.ImagePullSecrets))
	assert.Equal(t, "my-secret1", model.localDbStatefulSet.statefulSet.Spec.Template.Spec.ImagePullSecrets[0].Name)
	assert.Equal(t, "my-secret2", model.localDbStatefulSet.statefulSet.Spec.Template.Spec.ImagePullSecrets[1].Name)

	// no image pull secrets specified
	bs = *dbStatefulSetBackstage.DeepCopy()
	testObj = createBackstageTest(bs).withDefaultConfig(true).
		addToDefaultConfig("db-statefulset.yaml", "ips-deployment.yaml")

	model, err = InitObjects(context.TODO(), bs, testObj.externalConfig, true, true, testObj.scheme)
	assert.NoError(t, err)

	// if imagepullsecrets not defined - default used
	assert.Equal(t, 2, len(model.localDbStatefulSet.statefulSet.Spec.Template.Spec.ImagePullSecrets))
	assert.Equal(t, "ips1", model.localDbStatefulSet.statefulSet.Spec.Template.Spec.ImagePullSecrets[0].Name)
	assert.Equal(t, "ips2", model.localDbStatefulSet.statefulSet.Spec.Template.Spec.ImagePullSecrets[1].Name)

	// empty list of image pull secrets
	bs = *dbStatefulSetBackstage.DeepCopy()
	bs.Spec.Application.ImagePullSecrets = []string{}

	testObj = createBackstageTest(bs).withDefaultConfig(true).
		addToDefaultConfig("db-statefulset.yaml", "ips-deployment.yaml")

	model, err = InitObjects(context.TODO(), bs, testObj.externalConfig, true, true, testObj.scheme)
	assert.NoError(t, err)

	assert.Equal(t, 0, len(model.localDbStatefulSet.statefulSet.Spec.Template.Spec.ImagePullSecrets))
}
