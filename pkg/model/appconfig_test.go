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
	"janus-idp.io/backstage-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"

	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpecifiedAppConfig(t *testing.T) {
	setTestEnv()

	meta := v1alpha1.Backstage{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bs",
			Namespace: "ns123",
		},
	}

	bs := DetailedBackstageSpec{BackstageSpec: meta.Spec}
	yaml, err := readTestYamlFile("app-config1.yaml")
	bs.Details.RawConfig = map[string]string{}
	bs.Details.RawConfig["app-config.yaml"] = string(yaml)
	bs.EnableLocalDb = pointer.Bool(false)

	model, err := InitObjects(context.TODO(), meta, &bs, true, false)

	assert.NoError(t, err)
	assert.True(t, len(model) > 0)

	//deployment := getBackstageDeployment(model)
	deployment := model[0].(*BackstageDeployment)
	assert.NotNil(t, deployment)

	assert.Equal(t, 1, len(deployment.deployment.Spec.Template.Spec.Containers[0].VolumeMounts))
	assert.Equal(t, 2, len(deployment.deployment.Spec.Template.Spec.Containers[0].Args))
	assert.Equal(t, 1, len(deployment.deployment.Spec.Template.Spec.Volumes))

}
