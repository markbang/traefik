package k8s

import (
	discoveryv1 "k8s.io/api/discovery/v1"
	"k8s.io/client-go/tools/cache"
)

// EndpointSliceServiceNameIndex is the informer index name for EndpointSlices by Service.
const EndpointSliceServiceNameIndex = "serviceName"

// EndpointSliceServiceNameIndexers indexes EndpointSlices by namespace and Service name.
var EndpointSliceServiceNameIndexers = cache.Indexers{
	EndpointSliceServiceNameIndex: EndpointSliceServiceNameIndexFunc,
}

// EndpointSliceServiceNameIndexKey returns the index key for a Service's EndpointSlices.
func EndpointSliceServiceNameIndexKey(namespace, serviceName string) string {
	return namespace + "/" + serviceName
}

// EndpointSliceServiceNameIndexFunc indexes EndpointSlices by namespace and Service name.
func EndpointSliceServiceNameIndexFunc(obj any) ([]string, error) {
	endpointSlice, ok := obj.(*discoveryv1.EndpointSlice)
	if !ok {
		return nil, nil
	}

	serviceName := endpointSlice.Labels[discoveryv1.LabelServiceName]
	if serviceName == "" {
		return nil, nil
	}

	return []string{EndpointSliceServiceNameIndexKey(endpointSlice.Namespace, serviceName)}, nil
}
