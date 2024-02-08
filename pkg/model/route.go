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

func (f BackstageRouteFactory) newBackstageObject() RuntimeObject {
	return &BackstageRoute{ /*route: &openshift.Route{}*/ }
}

type BackstageRoute struct {
	route *openshift.Route
}

func RouteName(backstageName string) string {
	return utils.GenerateRuntimeObjectName(backstageName, "route")
}

func (b *BackstageRoute) setRoute(specified bsv1alpha1.Route) {

	b.route = &openshift.Route{}

	if len(specified.Host) > 0 {
		b.route.Spec.Host = specified.Host
	}
	if len(specified.Subdomain) > 0 {
		b.route.Spec.Subdomain = specified.Subdomain
	}
	if specified.TLS == nil {
		return
	}
	if b.route.Spec.TLS == nil {
		b.route.Spec.TLS = &openshift.TLSConfig{
			Termination:                   openshift.TLSTerminationEdge,
			InsecureEdgeTerminationPolicy: openshift.InsecureEdgeTerminationPolicyRedirect,
			Certificate:                   specified.TLS.Certificate,
			Key:                           specified.TLS.Key,
			CACertificate:                 specified.TLS.CACertificate,
			ExternalCertificate: &openshift.LocalObjectReference{
				Name: specified.TLS.ExternalCertificateSecretName,
			},
		}
		return
	}
	if len(specified.TLS.Certificate) > 0 {
		b.route.Spec.TLS.Certificate = specified.TLS.Certificate
	}
	if len(specified.TLS.Key) > 0 {
		b.route.Spec.TLS.Key = specified.TLS.Key
	}
	if len(specified.TLS.Certificate) > 0 {
		b.route.Spec.TLS.Certificate = specified.TLS.Certificate
	}
	if len(specified.TLS.CACertificate) > 0 {
		b.route.Spec.TLS.CACertificate = specified.TLS.CACertificate
	}
	if len(specified.TLS.ExternalCertificateSecretName) > 0 {
		b.route.Spec.TLS.ExternalCertificate = &openshift.LocalObjectReference{
			Name: specified.TLS.ExternalCertificateSecretName,
		}
	}
}

func init() {
	registerConfig("route.yaml", BackstageRouteFactory{})
}

// implementation of RuntimeObject interface
func (b *BackstageRoute) Object() client.Object {
	return b.route
}

func (b *BackstageRoute) setObject(object client.Object) {
	b.route = nil
	if object != nil {
		b.route = object.(*openshift.Route)
	}
}

// implementation of RuntimeObject interface
func (b *BackstageRoute) EmptyObject() client.Object {
	return &openshift.Route{}
}

// implementation of RuntimeObject interface
func (b *BackstageRoute) addToModel(model *BackstageModel, backstage bsv1alpha1.Backstage, ownsRuntime bool) error {
	if (b.route == nil && !backstage.Spec.IsRouteEnabled()) || !model.isOpenshift {
		// no route
		return nil
	}

	// load from spec
	if backstage.Spec.IsRouteEnabled() && !backstage.Spec.IsRouteEmpty() {
		//if model.route == nil {
		//	br := BackstageRoute{route: &openshift.Route{}}
		//	br.addToModel(model, backstageMeta, ownsRuntime)
		//}
		b.setRoute(*backstage.Spec.Application.Route)
	}

	b.route.SetName(RouteName(backstage.Name))
	model.route = b
	model.setRuntimeObject(b)

	return nil
}

// implementation of RuntimeObject interface
func (b *BackstageRoute) validate(model *BackstageModel) error {
	b.route.Spec.To.Name = model.backstageService.service.Name
	return nil
}
