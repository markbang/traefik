package k8s

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestStripManagedFields(t *testing.T) {
	obj := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			ManagedFields: []metav1.ManagedFieldsEntry{{
				Manager: "kubectl",
			}},
		},
	}

	got, err := StripManagedFields(obj)
	assert.NoError(t, err)
	assert.Same(t, obj, got)
	assert.Empty(t, obj.ManagedFields)
}

func TestStripManagedFieldsIgnoresNonObjects(t *testing.T) {
	obj := "not-a-kubernetes-object"

	got, err := StripManagedFields(obj)
	assert.NoError(t, err)
	assert.Equal(t, obj, got)
}
