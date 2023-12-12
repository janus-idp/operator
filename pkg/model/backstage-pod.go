package model

import (
	"fmt"
	"path/filepath"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/pointer"
)

const backstageContainerName = "backstage-backend"

type backstagePod struct {
	container *corev1.Container
	volumes   []corev1.Volume
	podSpec   corev1.PodSpec
}

func newBackstagePod(deployment *appsv1.Deployment) *backstagePod {

	result := &backstagePod{}
	result.podSpec = deployment.Spec.Template.Spec
	// interested in Backstage container only and expected it to be the only one
	for _, c := range result.podSpec.Containers {
		result.container = &c
		result.container.Name = backstageContainerName
		break
	}
	if result.podSpec.Volumes == nil {
		result.volumes = []corev1.Volume{}
	} else {
		result.volumes = result.podSpec.Volumes
	}

	return result
}

func (p backstagePod) addExtraFile(configMaps []string, secrets []string) {

	panic("TODO")
}

func (p backstagePod) extraEnvVars(configMaps []corev1.ConfigMap, secrets []corev1.Secret, envs map[string]string) {

	panic("TODO")
}

func (p backstagePod) addAppConfig(configMapName string, filePath string) {

	volName := fmt.Sprintf("app-config-%s", configMapName)
	volSource := corev1.VolumeSource{
		ConfigMap: &corev1.ConfigMapVolumeSource{
			DefaultMode:          pointer.Int32(420),
			LocalObjectReference: corev1.LocalObjectReference{Name: configMapName},
		},
	}
	p.volumes = append(p.volumes, corev1.Volume{
		Name:         volName,
		VolumeSource: volSource,
	})

	p.container.VolumeMounts = append(p.container.VolumeMounts, corev1.VolumeMount{
		Name:      volName,
		MountPath: filePath,
		SubPath:   filepath.Base(filePath),
	})
	p.container.Args = append(p.container.Args, fmt.Sprintf("--config='%s'", filePath))
}

func (p backstagePod) addImagePullSecrets(pullSecrets []corev1.LocalObjectReference) {
	p.podSpec.ImagePullSecrets = append(p.podSpec.ImagePullSecrets, pullSecrets...)
}

func (p backstagePod) setImage(image string) {
	if image != "" {
		p.container.Image = image
	}
}
