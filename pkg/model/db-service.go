package model

import (
	"fmt"

	bsv1alpha1 "janus-idp.io/backstage-operator/api/v1alpha1"
	"janus-idp.io/backstage-operator/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DbService struct {
	service *corev1.Service
}

func newDbService() *DbService {
	return &DbService{service: &corev1.Service{}}
}

func (s *DbService) Object() client.Object {
	return s.service
}

func (s *DbService) initMetainfo(backstageMeta bsv1alpha1.Backstage, ownsRuntime bool) {
	initMetainfo(s, backstageMeta, ownsRuntime)
	s.service.SetName(fmt.Sprintf("%s-db-service", backstageMeta.Name))
	utils.GenerateLabel(&s.service.Spec.Selector, backstageAppLabel, fmt.Sprintf("backstage-db-%s", backstageMeta.Name))
}
