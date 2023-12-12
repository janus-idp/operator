package model

import (
	bsv1alpha1 "janus-idp.io/backstage-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type AppConfig struct {
	path      string
	configMap corev1.ConfigMap
}

func newAppConfig() *AppConfig {
	return &AppConfig{configMap: corev1.ConfigMap{}}
}

func (b *AppConfig) Object() client.Object {
	return &b.configMap
}

func (b *AppConfig) initMetainfo(backstageMeta bsv1alpha1.Backstage, ownsRuntime bool) {
	initMetainfo(b, backstageMeta, ownsRuntime)
}

func (b *AppConfig) updateBackstagePod(pod *backstagePod) {
	pod.addAppConfig(b.configMap.Name, b.path)
}
