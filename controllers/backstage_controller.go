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

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"janus-idp.io/backstage-operator/pkg/model"

	"k8s.io/apimachinery/pkg/types"

	bs "janus-idp.io/backstage-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	BackstageAppLabel = "janus-idp.io/app"
)

// BackstageReconciler reconciles a Backstage object
type BackstageReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	// If true, Backstage Controller always sync the state of runtime objects created
	// otherwise, runtime objects can be re-configured independently
	OwnsRuntime bool

	// Namespace allows to restrict the reconciliation to this particular namespace,
	// and ignore requests from other namespaces.
	// This is mostly useful for our tests, to overcome a limitation of EnvTest about namespace deletion.
	Namespace string

	IsOpenShift bool

	PsqlImage string

	BackstageImage string
}

//+kubebuilder:rbac:groups=janus-idp.io,resources=backstages,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=janus-idp.io,resources=backstages/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=janus-idp.io,resources=backstages/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=configmaps;secrets;persistentvolumes;persistentvolumeclaims;services,verbs=get;watch;create;update;list;delete
//+kubebuilder:rbac:groups="apps",resources=deployments,verbs=get;watch;create;update;list;delete
//+kubebuilder:rbac:groups="apps",resources=statefulsets,verbs=get;watch;create;update;list;delete
//+kubebuilder:rbac:groups="route.openshift.io",resources=routes;routes/custom-host,verbs=get;watch;create;update;list;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Backstage object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *BackstageReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	lg := log.FromContext(ctx)

	lg.V(1).Info(fmt.Sprintf("starting reconciliation (namespace: %q)", req.NamespacedName))

	// Ignore requests for other namespaces, if specified.
	// This is mostly useful for our tests, to overcome a limitation of EnvTest about namespace deletion.
	// More details on https://book.kubebuilder.io/reference/envtest.html#namespace-usage-limitation
	if r.Namespace != "" && req.Namespace != r.Namespace {
		return ctrl.Result{}, nil
	}

	backstage := bs.Backstage{}
	if err := r.Get(ctx, req.NamespacedName, &backstage); err != nil {
		if errors.IsNotFound(err) {
			lg.Info("backstage gone from the namespace")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("failed to load backstage deployment from the cluster: %w", err)
	}

	// This helps to:
	// 1. Preliminary read and prepare some config objects from the specs (configMaps, Secrets...)
	// 2. Make some validation to fail fast
	spec, err := r.preprocessSpec(ctx, backstage.Spec, req.Namespace)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to preprocess backstage spec: %w", err)
	}

	// This creates array of model objects to be reconsiled
	bsModel, err := model.InitObjects(ctx, backstage, spec, r.OwnsRuntime, r.IsOpenShift)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to initialize backstage model: %w", err)
	}

	//TODO, do it on model? (need to send Scheme to InitObjects just for this)
	if r.OwnsRuntime {
		for _, obj := range bsModel.Objects {
			if err = controllerutil.SetControllerReference(&backstage, obj.Object(), r.Scheme); err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to set owner reference: %s", err)
			}
		}
	}

	err = r.applyObjects(ctx, bsModel.Objects)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to apply backstage objects: %w", err)
	}

	//TODO: it is just a placeholder for the time
	r.setRunningStatus(ctx, &backstage, req.Namespace)
	r.setSyncStatus(&backstage)
	err = r.Status().Update(ctx, &backstage)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to set status: %w", err)
		//log.FromContext(ctx).Error(err, "unable to update backstage.status")
	}

	return ctrl.Result{}, nil
}

func (r *BackstageReconciler) applyObjects(ctx context.Context, objects []model.BackstageObject) error {

	lg := log.FromContext(ctx)

	for _, obj := range objects {

		if err := r.Get(ctx, types.NamespacedName{Name: obj.Object().GetName(), Namespace: r.Namespace}, obj.EmptyObject()); err != nil {
			if !errors.IsNotFound(err) {
				return fmt.Errorf("failed to get object: %w", err)
			}

			if err := r.Create(ctx, obj.Object()); err != nil {
				if errors.IsAlreadyExists(err) {
					lg.V(1).Info("Already created by other reconcilation", "", obj.Object().GetName())
					continue
				}
				return fmt.Errorf("failed to create object %s: %w", obj.Object().GetName(), err)
			}

			lg.V(1).Info("Create object ", "obj", obj.Object().GetName())
			continue
		}

		if err := r.Update(ctx, obj.Object()); err != nil {
			return fmt.Errorf("failed to update object %s: %w", obj.Object().GetName(), err)
		}
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *BackstageReconciler) SetupWithManager(mgr ctrl.Manager, log logr.Logger) error {

	builder := ctrl.NewControllerManagedBy(mgr).
		For(&bs.Backstage{})

	if r.OwnsRuntime {
		builder.Owns(&appsv1.Deployment{}).
			Owns(&corev1.Service{}).
			Owns(&appsv1.StatefulSet{})
	}

	return builder.Complete(r)
}
