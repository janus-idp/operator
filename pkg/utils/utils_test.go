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

package utils

import (
	"testing"
)

func TestToRFC1123Label(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		// The inputs below are all valid names for K8s ConfigMaps or Secrets.

		{
			name: "should replace invalid characters with a dash",
			in:   "kube-root-ca.crt",
			want: "kube-root-ca-crt",
		},
		{
			name: "all-numeric string should remain unchanged",
			in:   "123456789",
			want: "123456789",
		},
		{
			name: "should truncate up to the maximum length and remove leading and trailing dashes",
			in:   "ppxkgq.df-yyatvyrgjtwivunibicne-bvyyotvonbrtfv-awylmrez.ksvqjw-z.xpgdi", /* 70 characters */
			want: "ppxkgq-df-yyatvyrgjtwivunibicne-bvyyotvonbrtfv-awylmrez-ksvqjw",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToRFC1123Label(tt.in); got != tt.want {
				t.Errorf("ToRFC1123Label() = %v, want %v", got, tt.want)
			}
		})
	}
}
