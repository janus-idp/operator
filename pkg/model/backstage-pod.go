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
	"path/filepath"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/pointer"
)

const backstageContainerName = "backstage-backend"

// Pod containing Backstage business logic runtime objects (container, volumes)
type backstagePod struct {
	container *corev1.Container
	volumes   []corev1.Volume
	parent    *appsv1.Deployment
}

// Constructor for Backstage Pod type.
// Always use it and do not create backstagePod manually
// Current implementation relies on the fact that Pod contains single container
// (a Backstage Container)
// In the future, if needed, other logic can be implemented, (for example:
// a name of Backstage Container can be writen as predefined Pod's annotation, etc)
func newBackstagePod(bsdeployment *BackstageDeployment) (*backstagePod, error) {

	podSpec := &bsdeployment.deployment.Spec.Template.Spec
	if len(podSpec.Containers) != 1 {
		return nil, fmt.Errorf("failed to create Backstage Pod. For the time only one Container,"+
			"treated as Backstage Container expected, but found %v", len(podSpec.Containers))
	}

	bspod := &backstagePod{
		parent:    bsdeployment.deployment,
		container: &podSpec.Containers[0],
		volumes:   podSpec.Volumes,
	}

	bsdeployment.pod = bspod

	return bspod, nil
}

func (p backstagePod) addExtraFileFromSecrets(secrets []string) {

	panic("TODO")
}

func (p backstagePod) addExtraFileFromConfigMaps(configMaps []string) {

	panic("TODO")
}

func (p backstagePod) addExtraEnvVarFromSecrets(secretNames []string) {
	for _, secretName := range secretNames {
		envSource := &corev1.SecretEnvSource{
			LocalObjectReference: corev1.LocalObjectReference{Name: secretName},
		}

		p.appendContainerEnvFrom(corev1.EnvFromSource{
			Prefix:    "secret-",
			SecretRef: envSource,
		})
	}
}

func (p backstagePod) addExtraEnvVarFromConfigMaps(configMapNames []string) {
	for _, cmName := range configMapNames {
		envSource := &corev1.ConfigMapEnvSource{
			LocalObjectReference: corev1.LocalObjectReference{Name: cmName},
		}

		p.appendContainerEnvFrom(corev1.EnvFromSource{
			Prefix:       "cm-",
			ConfigMapRef: envSource,
		})
	}
}

func (p backstagePod) addExtraEnvVars(envVars map[string]string) {
	for name, value := range envVars {

		p.appendContainerEnvVar(corev1.EnvVar{
			Name:  name,
			Value: value,
		})
	}
}

// Add x.y.z.app-config.yaml file to the Backstage configuration
func (p backstagePod) addAppConfig(configMapName string, filePath string) {

	volName := fmt.Sprintf("app-config-%s", configMapName)
	volSource := corev1.VolumeSource{
		ConfigMap: &corev1.ConfigMapVolumeSource{
			DefaultMode:          pointer.Int32(420),
			LocalObjectReference: corev1.LocalObjectReference{Name: configMapName},
		},
	}
	p.appendVolume(corev1.Volume{
		Name:         volName,
		VolumeSource: volSource,
	})

	p.appendContainerVolumeMount(corev1.VolumeMount{
		Name:      volName,
		MountPath: filePath,
		SubPath:   filepath.Base(filePath),
	})
	p.appendContainerArgs([]string{"--config", filePath})

}

func (p backstagePod) appendVolume(volume corev1.Volume) {
	p.volumes = append(p.volumes, volume)
	p.parent.Spec.Template.Spec.Volumes = p.volumes
}

func (p backstagePod) appendContainerArgs(args []string) {
	p.container.Args = append(p.container.Args, args...)
	p.parent.Spec.Template.Spec.Containers[0].Args = p.container.Args
}

func (p backstagePod) appendContainerVolumeMount(mount corev1.VolumeMount) {
	p.container.VolumeMounts = append(p.container.VolumeMounts, mount)
	p.parent.Spec.Template.Spec.Containers[0].VolumeMounts = p.container.VolumeMounts
}

func (p backstagePod) appendContainerEnvFrom(envFrom corev1.EnvFromSource) {
	p.container.EnvFrom = append(p.container.EnvFrom, envFrom)
	p.parent.Spec.Template.Spec.Containers[0].EnvFrom = p.container.EnvFrom
}

func (p backstagePod) appendContainerEnvVar(env corev1.EnvVar) {
	p.container.Env = append(p.container.Env, env)
	p.parent.Spec.Template.Spec.Containers[0].Env = p.container.Env
}

func (p backstagePod) appendImagePullSecrets(pullSecrets []corev1.LocalObjectReference) {
	p.parent.Spec.Template.Spec.ImagePullSecrets = append(p.parent.Spec.Template.Spec.ImagePullSecrets, pullSecrets...)
}

func (p backstagePod) setImage(image string) {
	if image != "" {
		p.container.Image = image
	}
}
