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

package controller

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	bs "backstage.io/backstage-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func (r *BackstageReconciler) applyLocalDbDeployment(ctx context.Context, backstage bs.Backstage, ns string) error {

	//lg := log.FromContext(ctx)

	deployment := &appsv1.Deployment{}
	err := r.readConfigMapOrDefault(ctx, backstage.Spec.RawRuntimeConfig.LocalDbConfigName, "db-deployment.yaml", ns, deployment)
	if err != nil {
		return err
	}

	setBackstageLocalDbLabel(deployment.Spec.Template.ObjectMeta.Labels, backstage)
	setBackstageLocalDbLabel(deployment.Spec.Selector.MatchLabels, backstage)

	//deployment.Namespace = ns
	err = r.Get(ctx, types.NamespacedName{Name: deployment.Name, Namespace: ns}, deployment)
	if err != nil {
		if errors.IsNotFound(err) {

		} else {
			return fmt.Errorf("failed to get deployment, reason: %s", err)
		}
	} else {
		//lg.Info("CR update is ignored for the time")
		return nil
	}

	r.labels(&deployment.ObjectMeta, backstage)

	if r.OwnsRuntime {
		if err := controllerutil.SetControllerReference(&backstage, deployment, r.Scheme); err != nil {
			return fmt.Errorf("failed to set owner reference: %s", err)
		}
	}

	err = r.Create(ctx, deployment)
	if err != nil {
		return fmt.Errorf("failed to create deplyment, reason: %s", err)
	}

	return nil
}

func (r *BackstageReconciler) applyLocalDbService(ctx context.Context, backstage bs.Backstage, ns string) error {

	//lg := log.FromContext(ctx)

	service := &corev1.Service{}
	err := r.readConfigMapOrDefault(ctx, backstage.Spec.RawRuntimeConfig.LocalDbConfigName, "db-service.yaml", ns, service)
	if err != nil {
		return err
	}

	setBackstageLocalDbLabel(service.Spec.Selector, backstage)

	err = r.Get(ctx, types.NamespacedName{Name: service.Name, Namespace: ns}, service)
	if err != nil {
		if errors.IsNotFound(err) {
		} else {
			return fmt.Errorf("failed to get service, reason: %s", err)
		}
	} else {
		//lg.Info("CR update is ignored for the time")
		return nil
	}

	r.labels(&service.ObjectMeta, backstage)
	if r.OwnsRuntime {
		if err := controllerutil.SetControllerReference(&backstage, service, r.Scheme); err != nil {
			return fmt.Errorf("failed to set owner reference: %s", err)
		}
	}

	err = r.Create(ctx, service)
	if err != nil {
		return fmt.Errorf("failed to create service, reason: %s", err)
	}

	return nil
}
