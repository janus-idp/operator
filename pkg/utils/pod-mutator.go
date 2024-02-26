package utils

import (
	"path/filepath"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/pointer"
)

const (
	SecretObjectKind    = "Secret"
	ConfigMapObjectKind = "ConfigMap"
)

type ObjectKind string

type PodMutator struct {
	PodSpec   *corev1.PodSpec
	Container *corev1.Container
}

func MountFilesFrom(podSpec *corev1.PodSpec, container *corev1.Container, kind ObjectKind, objectName string, mountPath string, singleFileName string, data map[string]string) {

	volName := GenerateVolumeNameFromCmOrSecret(objectName)
	volSrc := corev1.VolumeSource{}
	if kind == ConfigMapObjectKind {
		volSrc.ConfigMap = &corev1.ConfigMapVolumeSource{
			LocalObjectReference: corev1.LocalObjectReference{Name: objectName},
			DefaultMode:          pointer.Int32(420),
		}
	} else if kind == SecretObjectKind {
		volSrc.Secret = &corev1.SecretVolumeSource{
			SecretName:  objectName,
			DefaultMode: pointer.Int32(420),
		}
	}

	podSpec.Volumes = append(podSpec.Volumes, corev1.Volume{Name: volName, VolumeSource: volSrc})

	if data != nil {
		for file := range data {
			if singleFileName == "" || singleFileName == file {
				vm := corev1.VolumeMount{Name: volName, MountPath: filepath.Join(mountPath, file), SubPath: file, ReadOnly: true}
				container.VolumeMounts = append(container.VolumeMounts, vm)
			}
		}
	} else {
		vm := corev1.VolumeMount{Name: volName, MountPath: filepath.Join(mountPath, objectName), ReadOnly: true}
		container.VolumeMounts = append(container.VolumeMounts, vm)
	}

}

func AddEnvVarsFrom(container *corev1.Container, kind ObjectKind, objectName string, singleVarName string) {

	if singleVarName == "" {
		envFromSrc := corev1.EnvFromSource{}
		if kind == ConfigMapObjectKind {
			envFromSrc.ConfigMapRef = &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: objectName}}
		} else if kind == SecretObjectKind {
			envFromSrc.SecretRef = &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: objectName}}
		}
		container.EnvFrom = append(container.EnvFrom, envFromSrc)
	} else {
		envVarSrc := &corev1.EnvVarSource{}
		if kind == ConfigMapObjectKind {
			envVarSrc.ConfigMapKeyRef = &corev1.ConfigMapKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: objectName,
				},
				Key: singleVarName,
			}
		} else if kind == SecretObjectKind {
			envVarSrc.SecretKeyRef = &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: objectName,
				},
				Key: singleVarName,
			}
		}
		container.Env = append(container.Env, corev1.EnvVar{
			Name:      singleVarName,
			ValueFrom: envVarSrc,
		})
	}
}
