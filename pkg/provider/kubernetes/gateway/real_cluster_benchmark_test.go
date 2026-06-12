package gateway

import (
	"bufio"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/provider/kubernetes/k8s"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubefake "k8s.io/client-go/kubernetes/fake"
	kscheme "k8s.io/client-go/kubernetes/scheme"
	gatev1 "sigs.k8s.io/gateway-api/apis/v1"
)

const realClusterSnapshotEnv = "TRAEFIK_GATEWAY_CLUSTER_SNAPSHOT"

func BenchmarkRealClusterSnapshotLoadConfiguration(b *testing.B) {
	p := newRealClusterSnapshotProvider(b)

	var httpRouters, httpServices, httpRouteStatuses int
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		conf, statusReport, err := p.loadConfigurationFromGateways(b.Context())
		require.NoError(b, err)

		httpRouters = len(conf.HTTP.Routers)
		httpServices = len(conf.HTTP.Services)
		httpRouteStatuses = len(statusReport.httpRoutes)
	}

	b.ReportMetric(float64(httpRouters), "http_routers")
	b.ReportMetric(float64(httpServices), "http_services")
	b.ReportMetric(float64(httpRouteStatuses), "httproute_statuses")
}

func BenchmarkRealClusterSnapshotNoisyUpdates(b *testing.B) {
	k8sObjects, gwObjects := readRealClusterSnapshotObjects(b)

	httpRoute := firstObject[*gatev1.HTTPRoute](b, gwObjects)
	oldHTTPRoute := httpRoute.DeepCopy()
	oldHTTPRoute.ResourceVersion = "1"
	newHTTPRoute := httpRoute.DeepCopy()
	newHTTPRoute.ResourceVersion = "2"
	newHTTPRoute.Status.Parents = append(newHTTPRoute.Status.Parents, gatev1.RouteParentStatus{
		ControllerName: controllerName,
		ParentRef: gatev1.ParentReference{
			Name: "traefik-gateway",
		},
	})

	service := firstObject[*corev1.Service](b, k8sObjects)
	oldService := service.DeepCopy()
	oldService.ResourceVersion = "1"
	newService := service.DeepCopy()
	newService.ResourceVersion = "2"
	newService.Labels = map[string]string{"snapshot-benchmark": "metadata-only"}

	b.Run("HTTPRouteStatusOnly", func(b *testing.B) {
		events := make(chan any, 1)
		handler := &k8s.ResourceEventHandler{Ev: events}
		handler.OnUpdate(oldHTTPRoute, newHTTPRoute)
		require.Empty(b, events)

		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			handler.OnUpdate(oldHTTPRoute, newHTTPRoute)
		}

		require.Empty(b, events)
	})

	b.Run("ServiceMetadataOnly", func(b *testing.B) {
		events := make(chan any, 1)
		handler := &k8s.ResourceEventHandler{Ev: events}
		handler.OnUpdate(oldService, newService)
		require.Empty(b, events)

		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			handler.OnUpdate(oldService, newService)
		}

		require.Empty(b, events)
	})
}

func TestRealClusterSnapshotLoadConfiguration(t *testing.T) {
	p := newRealClusterSnapshotProvider(t)

	conf, statusReport, err := p.loadConfigurationFromGateways(t.Context())
	require.NoError(t, err)

	t.Logf("httpRouters=%d httpServices=%d httpRouteStatuses=%d", len(conf.HTTP.Routers), len(conf.HTTP.Services), len(statusReport.httpRoutes))
}

func newRealClusterSnapshotProvider(t testing.TB) *Provider {
	t.Helper()

	k8sObjects, gwObjects := readRealClusterSnapshotObjects(t)
	kubeClient := kubefake.NewClientset(k8sObjects...)
	gwClient := newGatewaySimpleClientSet(t, gwObjects...)

	client := newClientImpl(kubeClient, gwClient)
	stopCh := make(chan struct{})
	t.Cleanup(func() {
		close(stopCh)
	})

	_, err := client.WatchAll(nil, stopCh)
	require.NoError(t, err)

	return &Provider{
		EntryPoints: map[string]Entrypoint{
			"web":       {Address: ":8000"},
			"websecure": {Address: ":8443", HasHTTPTLSConf: true},
		},
		client: client,
	}
}

func readRealClusterSnapshotObjects(t testing.TB) ([]runtime.Object, []runtime.Object) {
	t.Helper()

	path := os.Getenv(realClusterSnapshotEnv)
	if path == "" {
		t.Skipf("%s is not set", realClusterSnapshotEnv)
	}

	file, err := os.Open(path)
	require.NoError(t, err)
	defer file.Close()

	var k8sObjects []runtime.Object
	var gwObjects []runtime.Object

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 1024*1024), 64*1024*1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		obj, gvk, err := kscheme.Codecs.UniversalDeserializer().Decode(line, nil, nil)
		require.NoError(t, err)

		if gvk.Group == groupGateway {
			gwObjects = append(gwObjects, obj)
			continue
		}

		k8sObjects = append(k8sObjects, obj)
	}
	require.NoError(t, scanner.Err())
	require.NotEmpty(t, gwObjects)

	return k8sObjects, gwObjects
}

func firstObject[T any](t testing.TB, objects []runtime.Object) T {
	t.Helper()

	for _, obj := range objects {
		if typed, ok := obj.(T); ok {
			return typed
		}
	}

	var zero T
	require.Failf(t, "object not found", "object type %T", zero)
	return zero
}

var _ = metav1.NamespaceAll
