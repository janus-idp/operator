package model

import (
	"path/filepath"

	bsv1alpha1 "janus-idp.io/backstage-operator/api/v1alpha1"
	"janus-idp.io/backstage-operator/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const defaultDir = "/test/dir"

type AppConfig struct {
	//path      string
	configMap *corev1.ConfigMap
}

func newAppConfig() *AppConfig {
	return &AppConfig{configMap: &corev1.ConfigMap{}}
}

func (b *AppConfig) Object() client.Object {
	return b.configMap
}

func (b *AppConfig) initMetainfo(backstageMeta bsv1alpha1.Backstage, ownsRuntime bool) {
	initMetainfo(b, backstageMeta, ownsRuntime)
	b.configMap.SetName(utils.GenerateRuntimeObjectName(backstageMeta.Name, "default-appconfig"))
}

func (b *AppConfig) updateBackstagePod(pod *backstagePod) {
	path := defaultDir
	for k := range b.configMap.Data {
		path = filepath.Join(path, k)
	}
	pod.addAppConfig(b.configMap.Name, path)
}

func (b *AppConfig) addToModel(model *runtimeModel) {
	// nothing to add
}
