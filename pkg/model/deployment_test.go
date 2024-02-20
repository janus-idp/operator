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

	corev1 "k8s.io/api/core/v1"

	bsv1alpha1 "janus-idp.io/backstage-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"

	"github.com/stretchr/testify/assert"
)

func TestImagePullSecrets(t *testing.T) {

}

// It tests the overriding image feature
// [GA] if we need this (and like this) feature
// we need to think about simple template engine
// for substitution env vars instead.
// Janus image specific
func TestOverrideBackstageImage(t *testing.T) {
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
		addToDefaultConfig("deployment.yaml", "janus-deployment.yaml")

	_ = os.Setenv(BackstageImageEnvVar, "dummy")

	model, err := InitObjects(context.TODO(), bs, testObj.rawConfig, []corev1.ConfigMap{}, true, false, testObj.scheme)
	assert.NoError(t, err)

	assert.Equal(t, "dummy", model.backstageDeployment.pod.container.Image)
	assert.Equal(t, "dummy", model.backstageDeployment.deployment.Spec.Template.Spec.InitContainers[0].Image)

	//t.Log(">>>>>>>>>>>>>>>>", model.backstageDeployment.Object().GetOwnerReferences()[0].Kind)

	//t.Log(">>>>>>>>>>>>>>>>", testObj.scheme.AllKnownTypes())

}
