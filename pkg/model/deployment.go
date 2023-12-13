package model

import (
	"fmt"

	bsv1alpha1 "janus-idp.io/backstage-operator/api/v1alpha1"

	"janus-idp.io/backstage-operator/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type BackstageDeployment struct {
	deployment *appsv1.Deployment
	pod        *backstagePod
}

func newBackstageDeployment() *BackstageDeployment {
	return &BackstageDeployment{deployment: &appsv1.Deployment{}}
}

func (b *BackstageDeployment) Object() client.Object {
	return b.deployment
}

func (b *BackstageDeployment) initMetainfo(backstageMeta bsv1alpha1.Backstage, ownsRuntime bool) {
	initMetainfo(b, backstageMeta, ownsRuntime)
	b.deployment.SetName(utils.GenerateRuntimeObjectName(backstageMeta.Name, "deployment"))
	utils.GenerateLabel(&b.deployment.Spec.Template.ObjectMeta.Labels, backstageAppLabel, fmt.Sprintf("backstage-%s", backstageMeta.Name))
	utils.GenerateLabel(&b.deployment.Spec.Selector.MatchLabels, backstageAppLabel, fmt.Sprintf("backstage-%s", backstageMeta.Name))
}

func (b *BackstageDeployment) setReplicas(replicas *int32) {
	if replicas != nil {
		b.deployment.Spec.Replicas = replicas
	}
}
