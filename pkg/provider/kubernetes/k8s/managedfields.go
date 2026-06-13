package k8s

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// StripManagedFields removes metadata.managedFields before objects enter informer caches.
func StripManagedFields(obj any) (any, error) {
	object, ok := obj.(metav1.Object)
	if ok {
		object.SetManagedFields(nil)
	}

	return obj, nil
}
