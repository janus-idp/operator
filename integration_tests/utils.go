package integration_tests

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	//. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func generateConfigMap(ctx context.Context, k8sClient client.Client, name, namespace string, data map[string]string) {
	//data := map[string]string{}
	//for k, v := range data {
	//	data[k] = fmt.Sprintf("value-%s", v)
	//}
	Expect(k8sClient.Create(ctx, &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: data,
	})).To(Not(HaveOccurred()))
}

func generateSecret(ctx context.Context, k8sClient client.Client, name, namespace string, keys []string) {
	data := map[string]string{}
	for _, v := range keys {
		data[v] = fmt.Sprintf("value-%s", v)
	}
	Expect(k8sClient.Create(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		StringData: data,
	})).To(Not(HaveOccurred()))
}

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
