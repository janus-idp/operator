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

func RouteName(backstageName string) string {
	return utils.GenerateRuntimeObjectName(backstageName, "route")
}

func (b *BackstageRoute) patchRoute(specified bsv1alpha1.Route) {

	osroute := b.route

	if len(specified.Host) > 0 {
		osroute.Spec.Host = specified.Host
	}
	if len(specified.Subdomain) > 0 {
		osroute.Spec.Subdomain = specified.Subdomain
	}
	if specified.TLS == nil {
		return
	}
	if osroute.Spec.TLS == nil {
		osroute.Spec.TLS = &openshift.TLSConfig{
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
		osroute.Spec.TLS.Certificate = specified.TLS.Certificate
	}
	if len(specified.TLS.Key) > 0 {
		osroute.Spec.TLS.Key = specified.TLS.Key
	}
	if len(specified.TLS.Certificate) > 0 {
		osroute.Spec.TLS.Certificate = specified.TLS.Certificate
	}
	if len(specified.TLS.CACertificate) > 0 {
		osroute.Spec.TLS.CACertificate = specified.TLS.CACertificate
	}
	if len(specified.TLS.ExternalCertificateSecretName) > 0 {
		osroute.Spec.TLS.ExternalCertificate = &openshift.LocalObjectReference{
			Name: specified.TLS.ExternalCertificateSecretName,
		}
	}
	return
}

func init() {
	registerConfig("route.yaml", BackstageRouteFactory{}, ForOpenshift)
}

// implementation of BackstageObject interface
func (b *BackstageRoute) Object() client.Object {
	return b.route
}

// implementation of BackstageObject interface
func (b *BackstageRoute) EmptyObject() client.Object {
	return &openshift.Route{}
}

// implementation of BackstageObject interface
func (b *BackstageRoute) addToModel(model *RuntimeModel, backstageMeta bsv1alpha1.Backstage, ownsRuntime bool) {
	model.route = b
	model.setObject(b)

	b.route.SetName(RouteName(backstageMeta.Name))
}

// implementation of BackstageObject interface
func (b *BackstageRoute) validate(model *RuntimeModel) error {
	b.route.Spec.To.Name = model.backstageService.service.Name
	return nil
}
