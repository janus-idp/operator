package model

import (
	"fmt"

	bsv1alpha1 "janus-idp.io/backstage-operator/api/v1alpha1"
	"janus-idp.io/backstage-operator/pkg/utils"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type BackstageService struct {
	service *corev1.Service
}

func newBackstageService() *BackstageService {
	return &BackstageService{service: &corev1.Service{}}
}

func (s *BackstageService) Object() client.Object {
	return s.service
}

func (s *BackstageService) initMetainfo(backstageMeta bsv1alpha1.Backstage, ownsRuntime bool) {
	initMetainfo(s, backstageMeta, ownsRuntime)
	s.service.SetName(utils.GenerateRuntimeObjectName(backstageMeta.Name, "service"))
	utils.GenerateLabel(&s.service.Spec.Selector, backstageAppLabel, fmt.Sprintf("backstage-%s", backstageMeta.Name))
}

func (b *BackstageService) addToModel(model *runtimeModel) {
	model.backstageService = b
}
