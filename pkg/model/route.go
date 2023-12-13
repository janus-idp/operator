package model

import (
	openshift "github.com/openshift/api/route/v1"
	bsv1alpha1 "janus-idp.io/backstage-operator/api/v1alpha1"
	"janus-idp.io/backstage-operator/pkg/utils"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Route struct {
	route *openshift.Route
}

func newRoute() *Route {
	return &Route{route: &openshift.Route{}}
}

func (b *Route) Object() client.Object {
	return b.route
}

func (b *Route) initMetainfo(backstageMeta bsv1alpha1.Backstage, ownsRuntime bool) {
	initMetainfo(b, backstageMeta, ownsRuntime)
	b.route.SetName(utils.GenerateRuntimeObjectName(backstageMeta.Name, "route"))
	b.route.Spec.To.Name = b.route.Name
}
