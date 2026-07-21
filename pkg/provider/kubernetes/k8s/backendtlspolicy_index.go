package k8s

import (
	"k8s.io/client-go/tools/cache"
	gatev1 "sigs.k8s.io/gateway-api/apis/v1"
)

const (
	backendTLSPolicyGroupCore   = "core"
	backendTLSPolicyKindService = "Service"
)

// BackendTLSPolicyServiceNameIndex is the informer index name for BackendTLSPolicies by Service.
const BackendTLSPolicyServiceNameIndex = "serviceName"

// BackendTLSPolicyServiceNameIndexers indexes BackendTLSPolicies by namespace and Service name.
var BackendTLSPolicyServiceNameIndexers = cache.Indexers{
	BackendTLSPolicyServiceNameIndex: BackendTLSPolicyServiceNameIndexFunc,
}

// BackendTLSPolicyServiceNameIndexKey returns the index key for a Service's BackendTLSPolicies.
func BackendTLSPolicyServiceNameIndexKey(namespace, serviceName string) string {
	return namespace + "/" + serviceName
}

// BackendTLSPolicyServiceNameIndexFunc indexes BackendTLSPolicies by namespace and Service name.
func BackendTLSPolicyServiceNameIndexFunc(obj any) ([]string, error) {
	policy, ok := obj.(*gatev1.BackendTLSPolicy)
	if !ok {
		return nil, nil
	}

	keys := make([]string, 0, len(policy.Spec.TargetRefs))
	seen := make(map[string]struct{}, len(policy.Spec.TargetRefs))
	for _, ref := range policy.Spec.TargetRefs {
		if (ref.Group != "" && string(ref.Group) != backendTLSPolicyGroupCore) || string(ref.Kind) != backendTLSPolicyKindService || ref.Name == "" {
			continue
		}

		key := BackendTLSPolicyServiceNameIndexKey(policy.Namespace, string(ref.Name))
		if _, ok := seen[key]; ok {
			continue
		}

		seen[key] = struct{}{}
		keys = append(keys, key)
	}

	if len(keys) == 0 {
		return nil, nil
	}

	return keys, nil
}
