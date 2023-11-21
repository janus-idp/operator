/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package controller

import (
	"context"
	"fmt"

	bs "backstage.io/backstage-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	DefaultBackstageDeployment = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: backstage
spec:
  replicas: 1
  selector:
    matchLabels:
      backstage.io/app:  # placeholder for 'backstage-<cr-name>'
  template:
    metadata:
      labels:
        backstage.io/app:  # placeholder for 'backstage-<cr-name>'
    spec:
      containers:
        - name: backstage
          image: ghcr.io/backstage/backstage
          imagePullPolicy: IfNotPresent
          ports:
            - name: http
              containerPort: 7007
          envFrom:
            - secretRef:
                name: postgres-secrets
#            - secretRef:
#                name: backstage-secrets


`
)

func (r *BackstageReconciler) applyBackstageDeployment(ctx context.Context, backstage bs.Backstage, ns string) error {

	lg := log.FromContext(ctx)

	deployment := &appsv1.Deployment{}
	err := r.readConfigMapOrDefault(ctx, backstage.Spec.RawRuntimeConfig.BackstageConfigName, "deploy", ns, DefaultBackstageDeployment, deployment)
	if err != nil {
		return err
	}
	setBackstageAppLabel(deployment.Spec.Template.ObjectMeta.Labels, backstage)
	setBackstageAppLabel(deployment.Spec.Selector.MatchLabels, backstage)
	//if deployment.Spec.Template.ObjectMeta.Labels == nil {
	//	deployment.Spec.Template.ObjectMeta.Labels = map[string]string{}
	//}
	//deployment.Spec.Template.ObjectMeta.Labels[BackstageAppLabel] = backstageAppId(backstage)
	//if deployment.Spec.Selector.MatchLabels == nil {
	//	deployment.Spec.Selector.MatchLabels = map[string]string{}
	//}
	//deployment.Spec.Selector.MatchLabels[BackstageAppLabel] = backstageAppId(backstage)

	err = r.Get(ctx, types.NamespacedName{Name: deployment.Name, Namespace: ns}, deployment)
	if err != nil {
		if errors.IsNotFound(err) {

		} else {
			return fmt.Errorf("failed to get backstage deployment, reason: %s", err)
		}
	} else {
		lg.Info("CR update is ignored for the time")
		return nil
	}

	r.labels(&deployment.ObjectMeta, backstage)
	err = r.Create(ctx, deployment)
	if err != nil {
		return fmt.Errorf("failed to create backstage deplyment, reason: %s", err)
	}
	return nil
}
