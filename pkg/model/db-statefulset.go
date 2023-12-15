package model

import (
	"fmt"

	bsv1alpha1 "janus-idp.io/backstage-operator/api/v1alpha1"
	"janus-idp.io/backstage-operator/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DbStatefulSet struct {
	statefulSet *appsv1.StatefulSet
}

func newDbStatefulSet() *DbStatefulSet {
	return &DbStatefulSet{statefulSet: &appsv1.StatefulSet{}}
}

func (b *DbStatefulSet) Object() client.Object {
	return b.statefulSet
}

func (b *DbStatefulSet) initMetainfo(backstageMeta bsv1alpha1.Backstage, ownsRuntime bool) {
	initMetainfo(b, backstageMeta, ownsRuntime)
	b.statefulSet.SetName(utils.GenerateRuntimeObjectName(backstageMeta.Name, "db-statefulset"))
	utils.GenerateLabel(&b.statefulSet.Spec.Template.ObjectMeta.Labels, backstageAppLabel, fmt.Sprintf("backstage-db-%s", backstageMeta.Name))
	utils.GenerateLabel(&b.statefulSet.Spec.Selector.MatchLabels, backstageAppLabel, fmt.Sprintf("backstage-db-%s", backstageMeta.Name))
}

func (b *DbStatefulSet) addToModel(model *runtimeModel) {
	model.localDbStatefulSet = b
}
