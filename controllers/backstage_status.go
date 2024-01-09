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
	bs "janus-idp.io/backstage-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// sets the RuntimeRunning condition
func (r *BackstageReconciler) setRunningStatus(backstage *bs.Backstage) {

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
