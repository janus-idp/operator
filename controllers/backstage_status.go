package controller

import (
	"context"

	bs "janus-idp.io/backstage-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// sets the RuntimeRunning condition
func (r *BackstageReconciler) setRunningStatus(ctx context.Context, backstage *bs.Backstage, ns string) {

	meta.SetStatusCondition(&backstage.Status.Conditions, v1.Condition{
		Type:               bs.RuntimeConditionRunning,
		Status:             "Unknown",
		LastTransitionTime: v1.Time{},
		Reason:             "Unknown",
		Message:            "Runtime in unknown status",
	})
}

// sets the RuntimeSyncedWithConfig condition
func (r *BackstageReconciler) setSyncStatus(backstage *bs.Backstage) {

	status := v1.ConditionUnknown
	reason := "Unknown"
	message := "Sync in unknown status"
	if r.OwnsRuntime {
		status = v1.ConditionTrue
		reason = "Synced"
		message = "Backstage syncs runtime"
	}

	meta.SetStatusCondition(&backstage.Status.Conditions, v1.Condition{
		Type:               bs.RuntimeConditionSynced,
		Status:             status,
		LastTransitionTime: v1.Time{},
		Reason:             reason,
		Message:            message,
	})
}
