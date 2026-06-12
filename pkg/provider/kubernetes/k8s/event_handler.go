package k8s

import (
	"reflect"

	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatev1 "sigs.k8s.io/gateway-api/apis/v1"
	gatev1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gatev1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

// ResourceEventHandler handles Add, Update or Delete Events for resources.
type ResourceEventHandler struct {
	Ev chan<- any
}

// OnAdd is called on Add Events.
func (reh *ResourceEventHandler) OnAdd(obj any, _ bool) {
	eventHandlerFunc(reh.Ev, obj)
}

// OnUpdate is called on Update Events.
// Ignores useless changes.
func (reh *ResourceEventHandler) OnUpdate(oldObj, newObj any) {
	if objChanged(oldObj, newObj) {
		eventHandlerFunc(reh.Ev, newObj)
	}
}

// OnDelete is called on Delete Events.
func (reh *ResourceEventHandler) OnDelete(obj any) {
	eventHandlerFunc(reh.Ev, obj)
}

// eventHandlerFunc will pass the obj on to the events channel or drop it.
// This is so passing the events along won't block in the case of high volume.
// The events are only used for signaling anyway so dropping a few is ok.
func eventHandlerFunc(events chan<- any, obj any) {
	select {
	case events <- obj:
	default:
	}
}

func objChanged(oldObj, newObj any) bool {
	if oldObj == nil || newObj == nil {
		return true
	}

	if oldObj.(metav1.Object).GetResourceVersion() == newObj.(metav1.Object).GetResourceVersion() {
		return false
	}

	if _, ok := oldObj.(*discoveryv1.EndpointSlice); ok {
		return endpointSliceChanged(oldObj.(*discoveryv1.EndpointSlice), newObj.(*discoveryv1.EndpointSlice))
	}

	if _, ok := oldObj.(*corev1.Node); ok {
		return nodeChanged(oldObj.(*corev1.Node), newObj.(*corev1.Node))
	}

	switch oldObj := oldObj.(type) {
	case *corev1.Namespace:
		return namespaceChanged(oldObj, newObj.(*corev1.Namespace))
	case *corev1.Service:
		return serviceChanged(oldObj, newObj.(*corev1.Service))
	case *corev1.Secret:
		return secretChanged(oldObj, newObj.(*corev1.Secret))
	case *corev1.ConfigMap:
		return configMapChanged(oldObj, newObj.(*corev1.ConfigMap))
	}

	switch oldObj := oldObj.(type) {
	case *gatev1.GatewayClass:
		return gatewayClassChanged(oldObj, newObj.(*gatev1.GatewayClass))
	case *gatev1.Gateway:
		return gatewayChanged(oldObj, newObj.(*gatev1.Gateway))
	case *gatev1.HTTPRoute:
		return httpRouteChanged(oldObj, newObj.(*gatev1.HTTPRoute))
	case *gatev1.GRPCRoute:
		return grpcRouteChanged(oldObj, newObj.(*gatev1.GRPCRoute))
	case *gatev1.TLSRoute:
		return tlsRouteChanged(oldObj, newObj.(*gatev1.TLSRoute))
	case *gatev1.BackendTLSPolicy:
		return backendTLSPolicyChanged(oldObj, newObj.(*gatev1.BackendTLSPolicy))
	case *gatev1alpha2.TCPRoute:
		return tcpRouteChanged(oldObj, newObj.(*gatev1alpha2.TCPRoute))
	case *gatev1beta1.ReferenceGrant:
		return referenceGrantChanged(oldObj, newObj.(*gatev1beta1.ReferenceGrant))
	}

	return true
}

// In some Kubernetes versions leader election is done by updating an endpoint annotation every second,
// if there are no changes to the endpoints addresses, ports, and there are no addresses defined for an endpoint
// the event can safely be ignored and won't cause unnecessary config reloads.
// TODO: check if Kubernetes is still using EndpointSlice for leader election, which seems to not be the case anymore.
func endpointSliceChanged(a, b *discoveryv1.EndpointSlice) bool {
	if a.AddressType != b.AddressType {
		return true
	}

	if a.Labels[discoveryv1.LabelServiceName] != b.Labels[discoveryv1.LabelServiceName] {
		return true
	}

	if len(a.Ports) != len(b.Ports) {
		return true
	}

	for i, aport := range a.Ports {
		bport := b.Ports[i]
		if !ptrEqual(aport.Name, bport.Name) {
			return true
		}
		if !ptrEqual(aport.Port, bport.Port) {
			return true
		}
		if !ptrEqual(aport.Protocol, bport.Protocol) {
			return true
		}
		if !ptrEqual(aport.AppProtocol, bport.AppProtocol) {
			return true
		}
	}

	if len(a.Endpoints) != len(b.Endpoints) {
		return true
	}

	for i, ea := range a.Endpoints {
		eb := b.Endpoints[i]
		if endpointChanged(ea, eb) {
			return true
		}
	}

	return false
}

func endpointChanged(a, b discoveryv1.Endpoint) bool {
	if !ptrEqual(a.Conditions.Ready, b.Conditions.Ready) {
		return true
	}
	if !ptrEqual(a.Conditions.Serving, b.Conditions.Serving) {
		return true
	}
	if !ptrEqual(a.Conditions.Terminating, b.Conditions.Terminating) {
		return true
	}

	if len(a.Addresses) != len(b.Addresses) {
		return true
	}

	for i, aaddr := range a.Addresses {
		baddr := b.Addresses[i]
		if aaddr != baddr {
			return true
		}
	}

	return false
}

func ptrEqual[T comparable](a, b *T) bool {
	if a == nil || b == nil {
		return a == b
	}

	return *a == *b
}

func nodeChanged(a, b *corev1.Node) bool {
	return !sameNodeAddresses(a.Status.Addresses, b.Status.Addresses)
}

func sameNodeAddresses(a, b []corev1.NodeAddress) bool {
	aAddresses := nodeAddressSet(a)
	bAddresses := nodeAddressSet(b)
	if len(aAddresses) != len(bAddresses) {
		return false
	}

	for address := range aAddresses {
		if _, ok := bAddresses[address]; !ok {
			return false
		}
	}

	return true
}

func nodeAddressSet(addresses []corev1.NodeAddress) map[corev1.NodeAddress]struct{} {
	result := map[corev1.NodeAddress]struct{}{}
	for _, address := range addresses {
		if address.Type != corev1.NodeInternalIP && address.Type != corev1.NodeExternalIP {
			continue
		}
		result[address] = struct{}{}
	}

	return result
}

func namespaceChanged(a, b *corev1.Namespace) bool {
	return !reflect.DeepEqual(a.Labels, b.Labels)
}

func serviceChanged(a, b *corev1.Service) bool {
	return a.Generation != b.Generation ||
		!reflect.DeepEqual(a.Spec, b.Spec) ||
		!reflect.DeepEqual(a.Annotations, b.Annotations) ||
		!reflect.DeepEqual(a.Status.LoadBalancer, b.Status.LoadBalancer)
}

func secretChanged(a, b *corev1.Secret) bool {
	return a.Type != b.Type ||
		!ptrEqual(a.Immutable, b.Immutable) ||
		!reflect.DeepEqual(a.Data, b.Data)
}

func configMapChanged(a, b *corev1.ConfigMap) bool {
	return !ptrEqual(a.Immutable, b.Immutable) ||
		!reflect.DeepEqual(a.Data, b.Data) ||
		!reflect.DeepEqual(a.BinaryData, b.BinaryData)
}

func gatewayClassChanged(a, b *gatev1.GatewayClass) bool {
	return a.Generation != b.Generation || !reflect.DeepEqual(a.Spec, b.Spec)
}

func gatewayChanged(a, b *gatev1.Gateway) bool {
	return a.Generation != b.Generation || !reflect.DeepEqual(a.Spec, b.Spec)
}

func httpRouteChanged(a, b *gatev1.HTTPRoute) bool {
	return a.Generation != b.Generation || !reflect.DeepEqual(a.Spec, b.Spec)
}

func grpcRouteChanged(a, b *gatev1.GRPCRoute) bool {
	return a.Generation != b.Generation || !reflect.DeepEqual(a.Spec, b.Spec)
}

func tlsRouteChanged(a, b *gatev1.TLSRoute) bool {
	return a.Generation != b.Generation || !reflect.DeepEqual(a.Spec, b.Spec)
}

func backendTLSPolicyChanged(a, b *gatev1.BackendTLSPolicy) bool {
	return a.Generation != b.Generation || !reflect.DeepEqual(a.Spec, b.Spec)
}

func tcpRouteChanged(a, b *gatev1alpha2.TCPRoute) bool {
	return a.Generation != b.Generation || !reflect.DeepEqual(a.Spec, b.Spec)
}

func referenceGrantChanged(a, b *gatev1beta1.ReferenceGrant) bool {
	return a.Generation != b.Generation || !reflect.DeepEqual(a.Spec, b.Spec)
}
