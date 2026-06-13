package k8s

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatev1 "sigs.k8s.io/gateway-api/apis/v1"
	gatev1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

func Test_detectChanges(t *testing.T) {
	portA := int32(80)
	portB := int32(8080)
	portNameA := "http"
	portNameB := "admin"
	appProtocolA := "http"
	appProtocolB := "h2c"
	ready := true
	notReady := false
	serving := true
	terminating := true
	protocolTCP := corev1.ProtocolTCP
	protocolUDP := corev1.ProtocolUDP
	tests := []struct {
		name   string
		oldObj any
		newObj any
		want   bool
	}{
		{
			name: "With nil values",
			want: true,
		},
		{
			name:   "With empty endpointslice",
			oldObj: &discoveryv1.EndpointSlice{},
			newObj: &discoveryv1.EndpointSlice{},
		},
		{
			name:   "With old nil",
			newObj: &discoveryv1.EndpointSlice{},
			want:   true,
		},
		{
			name:   "With new nil",
			oldObj: &discoveryv1.EndpointSlice{},
			want:   true,
		},
		{
			name: "With same version",
			oldObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1",
				},
			},
			newObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1",
				},
			},
		},
		{
			name: "With different version",
			oldObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1",
				},
			},
			newObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "2",
				},
			},
		},
		{
			name: "Ingress With same version",
			oldObj: &netv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1",
				},
			},
			newObj: &netv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1",
				},
			},
		},
		{
			name: "Ingress With different version",
			oldObj: &netv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1",
				},
			},
			newObj: &netv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "2",
				},
			},
			want: true,
		},
		{
			name: "With same annotations",
			oldObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1",
					Annotations: map[string]string{
						"test-annotation": "_",
					},
				},
			},
			newObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "2",
					Annotations: map[string]string{
						"test-annotation": "_",
					},
				},
			},
		},
		{
			name: "With different annotations",
			oldObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1",
					Annotations: map[string]string{
						"test-annotation": "V",
					},
				},
			},
			newObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "2",
					Annotations: map[string]string{
						"test-annotation": "X",
					},
				},
			},
		},
		{
			name: "With same endpoints and ports",
			oldObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1",
				},
				Endpoints: []discoveryv1.Endpoint{},
				Ports:     []discoveryv1.EndpointPort{},
			},
			newObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1",
				},
				Endpoints: []discoveryv1.Endpoint{},
				Ports:     []discoveryv1.EndpointPort{},
			},
		},
		{
			name: "With same port values",
			oldObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1",
				},
				Ports: []discoveryv1.EndpointPort{{
					Name:        &portNameA,
					Port:        &portA,
					Protocol:    &protocolTCP,
					AppProtocol: &appProtocolA,
				}},
			},
			newObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "2",
				},
				Ports: []discoveryv1.EndpointPort{{
					Name:        &portNameA,
					Port:        &portA,
					Protocol:    &protocolTCP,
					AppProtocol: &appProtocolA,
				}},
			},
		},
		{
			name: "With different service name labels",
			oldObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1",
					Labels: map[string]string{
						discoveryv1.LabelServiceName: "svc-a",
					},
				},
			},
			newObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "2",
					Labels: map[string]string{
						discoveryv1.LabelServiceName: "svc-b",
					},
				},
			},
			want: true,
		},
		{
			name: "With different address types",
			oldObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1",
				},
				AddressType: discoveryv1.AddressTypeIPv4,
			},
			newObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "2",
				},
				AddressType: discoveryv1.AddressTypeIPv6,
			},
			want: true,
		},
		{
			name: "With different len of endpoints",
			oldObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1",
				},
				Endpoints: []discoveryv1.Endpoint{},
			},
			newObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "2",
				},
				Endpoints: []discoveryv1.Endpoint{{}},
			},
			want: true,
		},
		{
			name: "With different endpoints",
			oldObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1",
				},
				Endpoints: []discoveryv1.Endpoint{{
					Addresses: []string{"10.10.10.10"},
				}},
			},
			newObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "2",
				},
				Endpoints: []discoveryv1.Endpoint{{
					Addresses: []string{"10.10.10.11"},
				}},
			},
			want: true,
		},
		{
			name: "With different endpoint ready condition",
			oldObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1",
				},
				Endpoints: []discoveryv1.Endpoint{{
					Addresses: []string{"10.10.10.10"},
					Conditions: discoveryv1.EndpointConditions{
						Ready: &ready,
					},
				}},
			},
			newObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "2",
				},
				Endpoints: []discoveryv1.Endpoint{{
					Addresses: []string{"10.10.10.10"},
					Conditions: discoveryv1.EndpointConditions{
						Ready: &notReady,
					},
				}},
			},
			want: true,
		},
		{
			name: "With different endpoint serving condition",
			oldObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1",
				},
				Endpoints: []discoveryv1.Endpoint{{
					Addresses: []string{"10.10.10.10"},
				}},
			},
			newObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "2",
				},
				Endpoints: []discoveryv1.Endpoint{{
					Addresses: []string{"10.10.10.10"},
					Conditions: discoveryv1.EndpointConditions{
						Serving: &serving,
					},
				}},
			},
			want: true,
		},
		{
			name: "With different endpoint terminating condition",
			oldObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1",
				},
				Endpoints: []discoveryv1.Endpoint{{
					Addresses: []string{"10.10.10.10"},
				}},
			},
			newObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "2",
				},
				Endpoints: []discoveryv1.Endpoint{{
					Addresses: []string{"10.10.10.10"},
					Conditions: discoveryv1.EndpointConditions{
						Terminating: &terminating,
					},
				}},
			},
			want: true,
		},
		{
			name: "Node with same internal and external addresses",
			oldObj: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1",
				},
				Status: corev1.NodeStatus{
					Addresses: []corev1.NodeAddress{
						{Type: corev1.NodeInternalIP, Address: "10.0.0.1"},
						{Type: corev1.NodeExternalIP, Address: "192.0.2.1"},
						{Type: corev1.NodeHostName, Address: "node-a"},
					},
				},
			},
			newObj: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "2",
				},
				Status: corev1.NodeStatus{
					Addresses: []corev1.NodeAddress{
						{Type: corev1.NodeHostName, Address: "node-b"},
						{Type: corev1.NodeExternalIP, Address: "192.0.2.1"},
						{Type: corev1.NodeInternalIP, Address: "10.0.0.1"},
					},
				},
			},
		},
		{
			name: "Node with different internal address",
			oldObj: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1",
				},
				Status: corev1.NodeStatus{
					Addresses: []corev1.NodeAddress{
						{Type: corev1.NodeInternalIP, Address: "10.0.0.1"},
					},
				},
			},
			newObj: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "2",
				},
				Status: corev1.NodeStatus{
					Addresses: []corev1.NodeAddress{
						{Type: corev1.NodeInternalIP, Address: "10.0.0.2"},
					},
				},
			},
			want: true,
		},
		{
			name: "Service with metadata-only update",
			oldObj: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Generation:      1,
					ResourceVersion: "1",
				},
				Spec: corev1.ServiceSpec{
					ClusterIP: "10.0.0.10",
					Ports:     []corev1.ServicePort{{Port: 80}},
				},
			},
			newObj: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Generation:      1,
					ResourceVersion: "2",
					Labels:          map[string]string{"changed": "true"},
				},
				Spec: corev1.ServiceSpec{
					ClusterIP: "10.0.0.10",
					Ports:     []corev1.ServicePort{{Port: 80}},
				},
			},
		},
		{
			name: "Service with spec update",
			oldObj: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Generation:      1,
					ResourceVersion: "1",
				},
				Spec: corev1.ServiceSpec{
					ClusterIP: "10.0.0.10",
					Ports:     []corev1.ServicePort{{Port: 80}},
				},
			},
			newObj: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Generation:      2,
					ResourceVersion: "2",
				},
				Spec: corev1.ServiceSpec{
					ClusterIP: "10.0.0.10",
					Ports:     []corev1.ServicePort{{Port: 8080}},
				},
			},
			want: true,
		},
		{
			name: "Service with annotation update",
			oldObj: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Generation:      1,
					ResourceVersion: "1",
				},
				Spec: corev1.ServiceSpec{ClusterIP: "10.0.0.10"},
			},
			newObj: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Generation:      1,
					ResourceVersion: "2",
					Annotations:     map[string]string{"traefik.io/service.nativelb": "true"},
				},
				Spec: corev1.ServiceSpec{ClusterIP: "10.0.0.10"},
			},
			want: true,
		},
		{
			name: "Secret with metadata-only update",
			oldObj: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1",
				},
				Data: map[string][]byte{"tls.crt": []byte("cert")},
			},
			newObj: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "2",
					Labels:          map[string]string{"changed": "true"},
				},
				Data: map[string][]byte{"tls.crt": []byte("cert")},
			},
		},
		{
			name: "Secret with data update",
			oldObj: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1",
				},
				Data: map[string][]byte{"tls.crt": []byte("cert")},
			},
			newObj: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "2",
				},
				Data: map[string][]byte{"tls.crt": []byte("new-cert")},
			},
			want: true,
		},
		{
			name: "ConfigMap with metadata-only update",
			oldObj: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1",
				},
				Data: map[string]string{"ca.crt": "cert"},
			},
			newObj: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "2",
					Labels:          map[string]string{"changed": "true"},
				},
				Data: map[string]string{"ca.crt": "cert"},
			},
		},
		{
			name: "ConfigMap with data update",
			oldObj: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1",
				},
				Data: map[string]string{"ca.crt": "cert"},
			},
			newObj: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "2",
				},
				Data: map[string]string{"ca.crt": "new-cert"},
			},
			want: true,
		},
		{
			name: "Namespace with metadata-only update",
			oldObj: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1",
					Labels:          map[string]string{"team": "edge"},
				},
			},
			newObj: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "2",
					Labels:          map[string]string{"team": "edge"},
					Annotations:     map[string]string{"changed": "true"},
				},
			},
		},
		{
			name: "Namespace with label update",
			oldObj: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1",
					Labels:          map[string]string{"team": "edge"},
				},
			},
			newObj: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "2",
					Labels:          map[string]string{"team": "platform"},
				},
			},
			want: true,
		},
		{
			name: "HTTPRoute with status-only update",
			oldObj: &gatev1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Generation:      1,
					ResourceVersion: "1",
				},
				Spec: gatev1.HTTPRouteSpec{
					CommonRouteSpec: gatev1.CommonRouteSpec{
						ParentRefs: []gatev1.ParentReference{{Name: "gateway"}},
					},
				},
			},
			newObj: &gatev1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Generation:      1,
					ResourceVersion: "2",
				},
				Spec: gatev1.HTTPRouteSpec{
					CommonRouteSpec: gatev1.CommonRouteSpec{
						ParentRefs: []gatev1.ParentReference{{Name: "gateway"}},
					},
				},
				Status: gatev1.HTTPRouteStatus{
					RouteStatus: gatev1.RouteStatus{
						Parents: []gatev1.RouteParentStatus{{ControllerName: "traefik.io/gateway-controller"}},
					},
				},
			},
		},
		{
			name: "HTTPRoute with spec update",
			oldObj: &gatev1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Generation:      1,
					ResourceVersion: "1",
				},
				Spec: gatev1.HTTPRouteSpec{
					CommonRouteSpec: gatev1.CommonRouteSpec{
						ParentRefs: []gatev1.ParentReference{{Name: "gateway-a"}},
					},
				},
			},
			newObj: &gatev1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Generation:      2,
					ResourceVersion: "2",
				},
				Spec: gatev1.HTTPRouteSpec{
					CommonRouteSpec: gatev1.CommonRouteSpec{
						ParentRefs: []gatev1.ParentReference{{Name: "gateway-b"}},
					},
				},
			},
			want: true,
		},
		{
			name: "Gateway with status-only update",
			oldObj: &gatev1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Generation:      1,
					ResourceVersion: "1",
				},
				Spec: gatev1.GatewaySpec{
					GatewayClassName: "traefik",
					Listeners: []gatev1.Listener{{
						Name:     "web",
						Port:     80,
						Protocol: gatev1.HTTPProtocolType,
					}},
				},
			},
			newObj: &gatev1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Generation:      1,
					ResourceVersion: "2",
				},
				Spec: gatev1.GatewaySpec{
					GatewayClassName: "traefik",
					Listeners: []gatev1.Listener{{
						Name:     "web",
						Port:     80,
						Protocol: gatev1.HTTPProtocolType,
					}},
				},
				Status: gatev1.GatewayStatus{
					Listeners: []gatev1.ListenerStatus{{Name: "web", AttachedRoutes: 10}},
				},
			},
		},
		{
			name: "GatewayClass with status-only update",
			oldObj: &gatev1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Generation:      1,
					ResourceVersion: "1",
				},
				Spec: gatev1.GatewayClassSpec{
					ControllerName: "traefik.io/gateway-controller",
				},
			},
			newObj: &gatev1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Generation:      1,
					ResourceVersion: "2",
				},
				Spec: gatev1.GatewayClassSpec{
					ControllerName: "traefik.io/gateway-controller",
				},
				Status: gatev1.GatewayClassStatus{
					Conditions: []metav1.Condition{{Type: string(gatev1.GatewayClassConditionStatusAccepted)}},
				},
			},
		},
		{
			name: "ReferenceGrant with metadata-only update",
			oldObj: &gatev1beta1.ReferenceGrant{
				ObjectMeta: metav1.ObjectMeta{
					Generation:      1,
					ResourceVersion: "1",
				},
				Spec: gatev1beta1.ReferenceGrantSpec{
					From: []gatev1beta1.ReferenceGrantFrom{{Kind: "HTTPRoute", Namespace: "default"}},
					To:   []gatev1beta1.ReferenceGrantTo{{Kind: "Service"}},
				},
			},
			newObj: &gatev1beta1.ReferenceGrant{
				ObjectMeta: metav1.ObjectMeta{
					Generation:      1,
					ResourceVersion: "2",
					Labels:          map[string]string{"changed": "true"},
				},
				Spec: gatev1beta1.ReferenceGrantSpec{
					From: []gatev1beta1.ReferenceGrantFrom{{Kind: "HTTPRoute", Namespace: "default"}},
					To:   []gatev1beta1.ReferenceGrantTo{{Kind: "Service"}},
				},
			},
		},
		{
			name: "ReferenceGrant with spec update",
			oldObj: &gatev1beta1.ReferenceGrant{
				ObjectMeta: metav1.ObjectMeta{
					Generation:      1,
					ResourceVersion: "1",
				},
				Spec: gatev1beta1.ReferenceGrantSpec{
					From: []gatev1beta1.ReferenceGrantFrom{{Kind: "HTTPRoute", Namespace: "default"}},
					To:   []gatev1beta1.ReferenceGrantTo{{Kind: "Service"}},
				},
			},
			newObj: &gatev1beta1.ReferenceGrant{
				ObjectMeta: metav1.ObjectMeta{
					Generation:      2,
					ResourceVersion: "2",
				},
				Spec: gatev1beta1.ReferenceGrantSpec{
					From: []gatev1beta1.ReferenceGrantFrom{{Kind: "GRPCRoute", Namespace: "default"}},
					To:   []gatev1beta1.ReferenceGrantTo{{Kind: "Service"}},
				},
			},
			want: true,
		},
		{
			name: "With different len of ports",
			oldObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1",
				},
				Ports: []discoveryv1.EndpointPort{},
			},
			newObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "2",
				},
				Ports: []discoveryv1.EndpointPort{{}},
			},
			want: true,
		},
		{
			name: "With different port names",
			oldObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1",
				},
				Ports: []discoveryv1.EndpointPort{{
					Name: &portNameA,
				}},
			},
			newObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "2",
				},
				Ports: []discoveryv1.EndpointPort{{
					Name: &portNameB,
				}},
			},
			want: true,
		},
		{
			name: "With different ports",
			oldObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1",
				},
				Ports: []discoveryv1.EndpointPort{{
					Port: &portA,
				}},
			},
			newObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "2",
				},
				Ports: []discoveryv1.EndpointPort{{
					Port: &portB,
				}},
			},
			want: true,
		},
		{
			name: "With different port protocols",
			oldObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1",
				},
				Ports: []discoveryv1.EndpointPort{{
					Protocol: &protocolTCP,
				}},
			},
			newObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "2",
				},
				Ports: []discoveryv1.EndpointPort{{
					Protocol: &protocolUDP,
				}},
			},
			want: true,
		},
		{
			name: "With different port app protocols",
			oldObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "1",
				},
				Ports: []discoveryv1.EndpointPort{{
					AppProtocol: &appProtocolA,
				}},
			},
			newObj: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "2",
				},
				Ports: []discoveryv1.EndpointPort{{
					AppProtocol: &appProtocolB,
				}},
			},
			want: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.want, objChanged(test.oldObj, test.newObj))
		})
	}
}

func TestResourceEventHandlerIgnoresNoopUpdates(t *testing.T) {
	testCases := []struct {
		name   string
		oldObj any
		newObj any
	}{
		{
			name: "HTTPRoute status-only update",
			oldObj: &gatev1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Generation:      1,
					ResourceVersion: "1",
				},
				Spec: gatev1.HTTPRouteSpec{
					CommonRouteSpec: gatev1.CommonRouteSpec{
						ParentRefs: []gatev1.ParentReference{{Name: "gateway"}},
					},
				},
			},
			newObj: &gatev1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Generation:      1,
					ResourceVersion: "2",
				},
				Spec: gatev1.HTTPRouteSpec{
					CommonRouteSpec: gatev1.CommonRouteSpec{
						ParentRefs: []gatev1.ParentReference{{Name: "gateway"}},
					},
				},
				Status: gatev1.HTTPRouteStatus{
					RouteStatus: gatev1.RouteStatus{
						Parents: []gatev1.RouteParentStatus{{ControllerName: "traefik.io/gateway-controller"}},
					},
				},
			},
		},
		{
			name: "Service metadata-only update",
			oldObj: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Generation:      1,
					ResourceVersion: "1",
				},
				Spec: corev1.ServiceSpec{
					ClusterIP: "10.0.0.10",
					Ports:     []corev1.ServicePort{{Port: 80}},
				},
			},
			newObj: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Generation:      1,
					ResourceVersion: "2",
					Labels:          map[string]string{"changed": "true"},
				},
				Spec: corev1.ServiceSpec{
					ClusterIP: "10.0.0.10",
					Ports:     []corev1.ServicePort{{Port: 80}},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			events := make(chan any, 1)
			handler := &ResourceEventHandler{Ev: events}

			handler.OnUpdate(test.oldObj, test.newObj)

			assert.Empty(t, events)
		})
	}
}

func TestResourceEventHandlerEmitsMaterialUpdates(t *testing.T) {
	events := make(chan any, 1)
	handler := &ResourceEventHandler{Ev: events}

	oldObj := &gatev1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Generation:      1,
			ResourceVersion: "1",
		},
		Spec: gatev1.HTTPRouteSpec{
			CommonRouteSpec: gatev1.CommonRouteSpec{
				ParentRefs: []gatev1.ParentReference{{Name: "gateway-a"}},
			},
		},
	}
	newObj := &gatev1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Generation:      2,
			ResourceVersion: "2",
		},
		Spec: gatev1.HTTPRouteSpec{
			CommonRouteSpec: gatev1.CommonRouteSpec{
				ParentRefs: []gatev1.ParentReference{{Name: "gateway-b"}},
			},
		},
	}

	handler.OnUpdate(oldObj, newObj)

	assert.Same(t, newObj, <-events)
}
