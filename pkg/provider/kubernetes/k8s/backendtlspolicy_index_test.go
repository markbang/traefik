package k8s

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatev1 "sigs.k8s.io/gateway-api/apis/v1"
)

const (
	testBackendTLSPolicyGroupCore   = "core"
	testBackendTLSPolicyKindService = "Service"
)

func TestBackendTLSPolicyServiceNameIndexFunc(t *testing.T) {
	testCases := []struct {
		desc string
		obj  any
		want []string
	}{
		{
			desc: "BackendTLSPolicy with service targetRefs",
			obj: &gatev1.BackendTLSPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
				},
				Spec: gatev1.BackendTLSPolicySpec{
					TargetRefs: []gatev1.LocalPolicyTargetReferenceWithSectionName{
						{
							LocalPolicyTargetReference: gatev1.LocalPolicyTargetReference{
								Group: gatev1.Group(testBackendTLSPolicyGroupCore),
								Kind:  gatev1.Kind(testBackendTLSPolicyKindService),
								Name:  "svc-a",
							},
						},
						{
							LocalPolicyTargetReference: gatev1.LocalPolicyTargetReference{
								Group: gatev1.Group(testBackendTLSPolicyGroupCore),
								Kind:  gatev1.Kind(testBackendTLSPolicyKindService),
								Name:  "svc-a",
							},
						},
						{
							LocalPolicyTargetReference: gatev1.LocalPolicyTargetReference{
								Group: gatev1.Group(testBackendTLSPolicyGroupCore),
								Kind:  gatev1.Kind(testBackendTLSPolicyKindService),
								Name:  "svc-b",
							},
						},
					},
				},
			},
			want: []string{"default/svc-a", "default/svc-b"},
		},
		{
			desc: "BackendTLSPolicy without service targetRefs",
			obj: &gatev1.BackendTLSPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
				},
				Spec: gatev1.BackendTLSPolicySpec{
					TargetRefs: []gatev1.LocalPolicyTargetReferenceWithSectionName{{
						LocalPolicyTargetReference: gatev1.LocalPolicyTargetReference{
							Group: gatev1.Group("gateway.networking.k8s.io"),
							Kind:  gatev1.Kind("Gateway"),
							Name:  "gw",
						},
					}},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			got, err := BackendTLSPolicyServiceNameIndexFunc(test.obj)
			assert.NoError(t, err)
			assert.Equal(t, test.want, got)
		})
	}
}
