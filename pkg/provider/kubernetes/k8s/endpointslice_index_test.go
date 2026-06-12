package k8s

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestEndpointSliceServiceNameIndexFunc(t *testing.T) {
	testCases := []struct {
		desc string
		obj  any
		want []string
	}{
		{
			desc: "EndpointSlice with service name",
			obj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Labels: map[string]string{
						discoveryv1.LabelServiceName: "whoami",
					},
				},
			},
			want: []string{"default/whoami"},
		},
		{
			desc: "EndpointSlice without service name",
			obj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
				},
			},
		},
		{
			desc: "Wrong type",
			obj:  &corev1.Service{},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			got, err := EndpointSliceServiceNameIndexFunc(test.obj)
			assert.NoError(t, err)
			assert.Equal(t, test.want, got)
		})
	}
}
