package appconfig

import (
	"path/filepath"

	"janus-idp.io/backstage-operator/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/pointer"
)

const (
	secretObjectKind    = "Secret"
	configMapObjectKind = "ConfigMap"
)

type objectKind string

type podMutator struct {
	podSpec   *corev1.PodSpec
	container *corev1.Container
}

func (p *podMutator) mountFilesFrom(kind objectKind, objectName string, mountPath string, singleFileName string) {

	volName := utils.GenerateVolumeNameFromCmOrSecret(objectName)
	volSrc := corev1.VolumeSource{}
	if kind == configMapObjectKind {
		volSrc.ConfigMap = &corev1.ConfigMapVolumeSource{
			LocalObjectReference: corev1.LocalObjectReference{Name: objectName},
			DefaultMode:          pointer.Int32(420),
		}
	} else if kind == secretObjectKind {
		volSrc.Secret = &corev1.SecretVolumeSource{
			SecretName:  objectName,
			DefaultMode: pointer.Int32(420),
		}
	}
	p.podSpec.Volumes = append(p.podSpec.Volumes, corev1.Volume{Name: volName, VolumeSource: volSrc})

	vm := corev1.VolumeMount{Name: volName, MountPath: filepath.Join(mountPath, objectName, singleFileName), SubPath: singleFileName}
	p.container.VolumeMounts = append(p.container.VolumeMounts, vm)
}

func (p *podMutator) addEnvVarsFrom(kind objectKind, objectName string, singleVarName string) {
	if singleVarName == "" {
		envFromSrc := corev1.EnvFromSource{}
		if kind == configMapObjectKind {
			envFromSrc.ConfigMapRef = &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: objectName}}
		} else if kind == secretObjectKind {
			envFromSrc.SecretRef = &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: objectName}}
		}
		p.container.EnvFrom = append(p.container.EnvFrom, envFromSrc)
	} else {
		envVarSrc := &corev1.EnvVarSource{}
		if kind == configMapObjectKind {
			envVarSrc.ConfigMapKeyRef = &corev1.ConfigMapKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: objectName,
				},
				Key: singleVarName,
			}
		} else if kind == secretObjectKind {
			envVarSrc.SecretKeyRef = &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: objectName,
				},
				Key: singleVarName,
			}
		}
		p.container.Env = append(p.container.Env, corev1.EnvVar{
			Name:      singleVarName,
			ValueFrom: envVarSrc,
		})
	}
}
