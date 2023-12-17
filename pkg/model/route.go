//
// Copyright (c) 2023 Red Hat, Inc.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package model

import (
	openshift "github.com/openshift/api/route/v1"
	bsv1alpha1 "janus-idp.io/backstage-operator/api/v1alpha1"
	"janus-idp.io/backstage-operator/pkg/utils"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type BackstageRouteFactory struct{}

func (f BackstageRouteFactory) newBackstageObject() BackstageObject {
	return &BackstageRoute{route: &openshift.Route{}}
}

type BackstageRoute struct {
	route *openshift.Route
}

//func newRoute() *BackstageRoute {
//	return &BackstageRoute{route: &openshift.Route{}}
//}

func (b *BackstageRoute) Object() client.Object {
	return b.route
}

func (b *BackstageRoute) EmptyObject() client.Object {
	return &openshift.Route{}
}

func (b *BackstageRoute) initMetainfo(backstageMeta bsv1alpha1.Backstage, ownsRuntime bool) {
	initMetainfo(b, backstageMeta, ownsRuntime)
	b.route.SetName(utils.GenerateRuntimeObjectName(backstageMeta.Name, "route"))
	b.route.Spec.To.Name = b.route.Name
}

func (b *BackstageRoute) addToModel(model *runtimeModel) {
	// nothing to add
}
