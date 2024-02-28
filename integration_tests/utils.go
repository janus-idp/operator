package integration_tests

import (
	corev1 "k8s.io/api/core/v1"
)

func getEnvFromSecret(container corev1.Container, name string) *corev1.EnvFromSource {
	for _, from := range container.EnvFrom {
		if from.SecretRef.Name == name {
			return &from
		}
	}
	return nil
}

func findVolume(vols []corev1.Volume, name string) (corev1.Volume, bool) {
	list := findElementsByPredicate(vols, func(vol corev1.Volume) bool {
		return vol.Name == name
	})
	if len(list) == 0 {
		return corev1.Volume{}, false
	}
	return list[0], true
}

func findElementsByPredicate[T any](l []T, predicate func(t T) bool) (result []T) {
	for _, v := range l {
		if predicate(v) {
			result = append(result, v)
		}
	}
	return result
}
