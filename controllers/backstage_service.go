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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	DefaultBackstageService = `
apiVersion: v1
kind: Service
metadata:
  name: backstage
spec:
  type: NodePort
  selector:
    backstage.io/app:  # placeholder for 'backstage-<cr-name>'
  ports:
    - name: http
      port: 80
      targetPort: http
`
)

// selector for deploy.spec.template.spec.meta.label
// targetPort: http for deploy.spec.template.spec.containers.ports.name=http
func (r *BackstageReconciler) applyBackstageService(ctx context.Context, backstage bs.Backstage, ns string) error {

	lg := log.FromContext(ctx)

	service := &corev1.Service{}
	err := r.readConfigMapOrDefault(ctx, backstage.Spec.RawRuntimeConfig.BackstageConfigName, "service", ns, DefaultBackstageService, service)
	if err != nil {
		return err
	}

	setBackstageAppLabel(service.Spec.Selector, backstage)

	err = r.Get(ctx, types.NamespacedName{Name: service.Name, Namespace: ns}, service)
	if err != nil {
		if errors.IsNotFound(err) {
		} else {
			return fmt.Errorf("failed to get backstage service, reason: %s", err)
		}
	} else {
		lg.Info("CR update is ignored for the time")
		return nil
	}

	r.labels(&service.ObjectMeta, backstage)
	err = r.Create(ctx, service)
	if err != nil {
		return fmt.Errorf("failed to create backstage service, reason: %s", err)
	}
	return nil
}
