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

	bs := v1alpha1.Backstage{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bs",
			Namespace: "ns123",
		},
		Spec: v1alpha1.BackstageSpec{
			EnableLocalDb: pointer.Bool(true),
		},
	}

	model, err := InitObjects(context.TODO(), bs, &DetailedBackstageSpec{BackstageSpec: bs.Spec}, true, false)

	assert.NoError(t, err)
	assert.True(t, len(model) > 0)
	assert.Equal(t, "bs-deployment", model[0].Object().GetName())
	assert.Equal(t, "ns123", model[0].Object().GetNamespace())
	assert.Equal(t, 2, len(model[0].Object().GetLabels()))
	//	assert.Equal(t, 1, len(model[0].Object().GetOwnerReferences()))

	bsDeployment := model[0].(*BackstageDeployment)
	assert.NotNil(t, bsDeployment.pod.container)
	assert.Equal(t, backstageContainerName, bsDeployment.pod.container.Name)
	assert.NotNil(t, bsDeployment.pod.volumes)

	for _, vol := range bsDeployment.deployment.Spec.Template.Spec.Volumes {
		fmt.Printf("vol %v \n", vol)
	}

	for _, vm := range bsDeployment.deployment.Spec.Template.Spec.Containers[0].VolumeMounts {
		fmt.Printf("vol Mount %v \n", vm)
	}

	for _, vol1 := range *bsDeployment.pod.volumes {
		fmt.Printf("vol %v \n", vol1)
	}

	for _, vm1 := range bsDeployment.pod.container.VolumeMounts {
		fmt.Printf("vol Mount %v \n", vm1)
	}

	//	assert.Equal(t, "Backstage", bsDeployment.deployment.OwnerReferences[0].Kind)

	bsService := model[1].(*BackstageService)
	assert.Equal(t, "bs-service", bsService.service.Name)
	assert.True(t, len(bsService.service.Spec.Ports) > 0)

	assert.Equal(t, fmt.Sprintf("backstage-%s", "bs"), bsDeployment.deployment.Spec.Template.ObjectMeta.Labels[backstageAppLabel])
	assert.Equal(t, fmt.Sprintf("backstage-%s", "bs"), bsService.service.Spec.Selector[backstageAppLabel])

}
