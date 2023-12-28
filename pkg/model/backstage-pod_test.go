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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func TestSingleBackstageContainer(t *testing.T) {
	depl := &appsv1.Deployment{}
	_, err := newBackstagePod(&BackstageDeployment{deployment: depl})
	require.EqualErrorf(t, err, "failed to create Backstage Pod. Only one Container, "+
		"treated as Backstage Container expected, but found 0", "Must fail as no containers specified")

	depl.Spec.Template.Spec.Containers = append(depl.Spec.Template.Spec.Containers, corev1.Container{Name: "backstage-backend"})
	p, err := newBackstagePod(&BackstageDeployment{deployment: depl})
	require.NoError(t, err)
	assert.Equal(t, &depl.Spec.Template.Spec.Containers[0], p.container)

	depl.Spec.Template.Spec.Containers = append(depl.Spec.Template.Spec.Containers, corev1.Container{Name: "backstage-backend2"})
	_, err = newBackstagePod(&BackstageDeployment{deployment: depl})
	require.EqualErrorf(t, err, "failed to create Backstage Pod. Only one Container, "+
		"treated as Backstage Container expected, but found 2", "Must fail as 2 containers specified")
}

func TestIfBasckstagePodPointsToDeployment(t *testing.T) {
	depl := &appsv1.Deployment{}
	depl.Spec.Template.Spec.Containers = append(depl.Spec.Template.Spec.Containers, corev1.Container{Name: "backstage-backend"})

	testPod, err := newBackstagePod(&BackstageDeployment{deployment: depl})
	assert.NoError(t, err)

	bc := testPod.container

	assert.Equal(t, bc, &testPod.parent.Spec.Template.Spec.Containers[0])
	assert.Equal(t, testPod.parent.Spec.Template.Spec.Containers[0].Name, bc.Name)

	assert.Equal(t, 0, len(testPod.parent.Spec.Template.Spec.Containers[0].Env))
	assert.Equal(t, 0, len(bc.Env))
	testPod.addContainerEnvVar(corev1.EnvVar{Name: "myKey", Value: "myValue"})
	assert.Equal(t, 1, len(bc.Env))
	assert.Equal(t, 1, len(testPod.parent.Spec.Template.Spec.Containers[0].Env))

	assert.Equal(t, 0, len(testPod.parent.Spec.Template.Spec.Containers[0].VolumeMounts))
	assert.Equal(t, 0, len(bc.VolumeMounts))
	testPod.appendContainerVolumeMount(corev1.VolumeMount{
		Name: "mount",
	})
	assert.Equal(t, 1, len(testPod.parent.Spec.Template.Spec.Containers[0].VolumeMounts))
	assert.Equal(t, 1, len(bc.VolumeMounts))

	assert.Equal(t, 0, len(testPod.parent.Spec.Template.Spec.Volumes))
	assert.Equal(t, 0, len(*testPod.volumes))
	testPod.appendVolume(corev1.Volume{Name: "vol"})
	assert.Equal(t, 1, len(testPod.parent.Spec.Template.Spec.Volumes))
	assert.Equal(t, 1, len(*testPod.volumes))

	assert.Equal(t, 0, len(testPod.parent.Spec.Template.Spec.Containers[0].Args))
	assert.Equal(t, 0, len(testPod.container.Args))
	testPod.appendConfigArg("/test.yaml")
	assert.Equal(t, 2, len(testPod.parent.Spec.Template.Spec.Containers[0].Args))
	assert.Equal(t, 2, len(testPod.container.Args))

	assert.Equal(t, 0, len(testPod.parent.Spec.Template.Spec.Containers[0].EnvFrom))
	assert.Equal(t, 0, len(testPod.container.EnvFrom))
	testPod.addContainerEnvFrom(
		corev1.EnvFromSource{ConfigMapRef: &corev1.ConfigMapEnvSource{
			LocalObjectReference: corev1.LocalObjectReference{Name: "cm1"},
		}})
	testPod.addContainerEnvFrom(
		corev1.EnvFromSource{SecretRef: &corev1.SecretEnvSource{
			LocalObjectReference: corev1.LocalObjectReference{Name: "sec"},
		}})
	assert.Equal(t, 2, len(testPod.parent.Spec.Template.Spec.Containers[0].EnvFrom))
	assert.Equal(t, 2, len(testPod.container.EnvFrom))

}
