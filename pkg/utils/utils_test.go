package utils

import (
	"regexp"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/util/validation"
)

func TestToKubernetesResourceName(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "should return lowercase",
			in:   "Lorem-Ipsum",
			want: "lorem-ipsum",
		},
		{
			name: "should replace invalid characters with a dash",
			in:   "Kube-root CA.crt",
			want: "kube-root-ca-crt",
		},
		{
			name: "should remove trailing and leading invalid characters",
			in:   "_ some/file/ path ?",
			want: "some-file-path",
		},
		{
			name: "all-numeric string should remain unchanged",
			in:   "123456789",
			want: "123456789",
		},
		{
			name: "should return an empty string for input with only invalid characters",
			in:   "#?@  *&^------ %% --",
			want: "",
		},
		{
			name: "should truncate up to the maximum length",
			in:   strings.Repeat("my_file.sh" /* 10 characters*/, 10),
			want: strings.Repeat("my-file-sh", 6) + "my",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToKubernetesResourceName(tt.in); got != tt.want {
				t.Errorf("ToKubernetesResourceName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func FuzzToKubernetesResourceName(f *testing.F) {
	// seed corpus
	testcases := []string{"Hello, world", "!12345", ">> .Lorem Ipsum !#"}
	for _, tc := range testcases {
		f.Add(tc)
	}
	re := regexp.MustCompile(`[a-zA-Z0-9]`)
	f.Fuzz(func(t *testing.T, in string) {
		k8sName := ToKubernetesResourceName(in)
		if len(k8sName) > maxK8sResourceNameLength {
			t.Errorf("ToKubernetesResourceName produced a string longer than %d characters for %q as input", maxK8sResourceNameLength, in)
			return
		}
		if k8sName == "" {
			// we should get an empty output if input only consists of invalid characters
			validChars := re.FindString(in)
			if validChars != "" {
				t.Errorf("ToKubernetesResourceName produced an empty string for %q, which contains some valid characters", in)
			}
			return
		}
		errs := validation.IsDNS1123Label(k8sName)
		if len(errs) > 0 {
			t.Errorf("ToKubernetesResourceName produced non-compliant resource name for %q, errors: %s",
				in, strings.Join(errs, ", "))
		}
	})
}
