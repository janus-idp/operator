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
	"fmt"

	bs "janus-idp.io/backstage-operator/api/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

const backstageContainerName = "backstage-backend"
const defaultDir = "/opt/app-root/src"

// Pod containing Backstage business logic runtime objects (container, volumes)
type backstagePod struct {
	container *corev1.Container
	volumes   *[]corev1.Volume
	parent    *appsv1.Deployment
}

// Constructor for Backstage Pod type.
// Always use it and do not create backstagePod type manually
// Current implementation relies on the fact that Pod contains single Backstage Container
func newBackstagePod(bsdeployment *BackstageDeployment) (*backstagePod, error) {

	if bsdeployment.deployment == nil {
		return nil, fmt.Errorf("deployment not defined")
	}

	podSpec := bsdeployment.deployment.Spec.Template.Spec
	if len(podSpec.Containers) != 1 {
		return nil, fmt.Errorf("failed to create Backstage Pod. Only one Container, "+
			"treated as Backstage Container expected, but found %v", len(podSpec.Containers))
	}

	bspod := &backstagePod{
		parent:    bsdeployment.deployment,
		container: &podSpec.Containers[0],
		volumes:   &podSpec.Volumes,
	}

	bsdeployment.pod = bspod

	return bspod, nil
}

// appends Volume to the Backstage Pod
func (p backstagePod) appendVolume(volume corev1.Volume) {
	*p.volumes = append(*p.volumes, volume)
	p.parent.Spec.Template.Spec.Volumes = *p.volumes
}

// appends --config argument to the Backstage Container command line
func (p backstagePod) appendConfigArg(appConfigPath string) {
	p.container.Args = append(p.container.Args, []string{"--config", appConfigPath}...)
}

// appends VolumeMount to the Backstage Container
func (p backstagePod) appendContainerVolumeMount(mount corev1.VolumeMount) {
	p.container.VolumeMounts = append(p.container.VolumeMounts, mount)
}

// adds environment variable to the Backstage Container using ConfigMap or Secret source
func (p backstagePod) addContainerEnvFrom(envFrom corev1.EnvFromSource) {
	p.container.EnvFrom = append(p.container.EnvFrom, envFrom)
}

// adds environment variables to the Backstage Container
func (p backstagePod) addContainerEnvVar(env bs.Env) {
	p.container.Env = append(p.container.Env, corev1.EnvVar{
		Name:  env.Name,
		Value: env.Value,
	})
}

// adds environment from source to the Backstage Container
func (p backstagePod) addContainerEnvVarSource(name string, envVarSource *corev1.EnvVarSource) {
	p.container.Env = append(p.container.Env, corev1.EnvVar{
		Name:      name,
		ValueFrom: envVarSource,
	})
}

// sets pullSecret for Backstage Pod
func (p backstagePod) setImagePullSecrets(pullSecrets []string) {
	for _, ps := range pullSecrets {
		p.parent.Spec.Template.Spec.ImagePullSecrets = append(p.parent.Spec.Template.Spec.ImagePullSecrets,
			corev1.LocalObjectReference{Name: ps})
	}
}

// sets container image name of Backstage Container
func (p backstagePod) setImage(image *string) {
	if image != nil {
		p.container.Image = *image
	}
}
