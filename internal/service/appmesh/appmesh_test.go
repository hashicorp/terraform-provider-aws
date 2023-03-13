package appmesh_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccAppMesh_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"GatewayRoute": {
			"basic":              testAccGatewayRoute_basic,
			"disappears":         testAccGatewayRoute_disappears,
			"grpcRoute":          testAccGatewayRoute_GRPCRoute,
			"grpcRouteWithPort":  testAccGatewayRoute_GRPCRouteWithPort,
			"httpRoute":          testAccGatewayRoute_HTTPRoute,
			"httpRouteWithPort":  testAccGatewayRoute_HTTPRouteWithPort,
			"http2Route":         testAccGatewayRoute_HTTP2Route,
			"http2RouteWithPort": testAccGatewayRoute_HTTP2RouteWithPort,
			"tags":               testAccGatewayRoute_Tags,
		},
		"Mesh": {
			"basic":        testAccMesh_basic,
			"egressFilter": testAccMesh_egressFilter,
			"tags":         testAccMesh_tags,
		},
		"Route": {
			"disappears":              testAccRoute_disappears,
			"grpcRoute":               testAccRoute_grpcRoute,
			"grpcRouteWithPortMatch":  testAccRoute_grpcRouteWithPortMatch,
			"grpcRouteEmptyMatch":     testAccRoute_grpcRouteEmptyMatch,
			"grpcRouteTimeout":        testAccRoute_grpcRouteTimeout,
			"http2Route":              testAccRoute_http2Route,
			"http2RouteWithPortMatch": testAccRoute_http2RouteWithPortMatch,
			"http2RouteTimeout":       testAccRoute_http2RouteTimeout,
			"httpHeader":              testAccRoute_httpHeader,
			"httpRetryPolicy":         testAccRoute_httpRetryPolicy,
			"httpRoute":               testAccRoute_httpRoute,
			"httpRouteWithPortMatch":  testAccRoute_httpRouteWithPortMatch,
			"httpRouteTimeout":        testAccRoute_httpRouteTimeout,
			"routePriority":           testAccRoute_routePriority,
			"tcpRoute":                testAccRoute_tcpRoute,
			"tcpRouteWithPortMatch":   testAccRoute_tcpRouteWithPortMatch,
			"tcpRouteTimeout":         testAccRoute_tcpRouteTimeout,
			"tags":                    testAccRoute_tags,
		},
		"VirtualGateway": {
			"basic":                      testAccVirtualGateway_basic,
			"disappears":                 testAccVirtualGateway_disappears,
			"backendDefaults":            testAccVirtualGateway_BackendDefaults,
			"backendDefaultsCertificate": testAccVirtualGateway_BackendDefaultsCertificate,
			"listenerConnectionPool":     testAccVirtualGateway_ListenerConnectionPool,
			"listenerHealthChecks":       testAccVirtualGateway_ListenerHealthChecks,
			"listenerTls":                testAccVirtualGateway_ListenerTLS,
			"listenerValidation":         testAccVirtualGateway_ListenerValidation,
			"multiListenerValidation":    testAccVirtualGateway_MultiListenerValidation,
			"logging":                    testAccVirtualGateway_Logging,
			"tags":                       testAccVirtualGateway_Tags,
		},
		"VirtualNode": {
			"basic":                      testAccVirtualNode_basic,
			"disappears":                 testAccVirtualNode_disappears,
			"backendClientPolicyAcm":     testAccVirtualNode_backendClientPolicyACM,
			"backendClientPolicyFile":    testAccVirtualNode_backendClientPolicyFile,
			"backendDefaults":            testAccVirtualNode_backendDefaults,
			"backendDefaultsCertificate": testAccVirtualNode_backendDefaultsCertificate,
			"cloudMapServiceDiscovery":   testAccVirtualNode_cloudMapServiceDiscovery,
			"listenerConnectionPool":     testAccVirtualNode_listenerConnectionPool,
			"listenerOutlierDetection":   testAccVirtualNode_listenerOutlierDetection,
			"listenerHealthChecks":       testAccVirtualNode_listenerHealthChecks,
			"listenerTimeout":            testAccVirtualNode_listenerTimeout,
			"listenerTls":                testAccVirtualNode_listenerTLS,
			"listenerValidation":         testAccVirtualNode_listenerValidation,
			"multiListenerValidation":    testAccVirtualNode_multiListenerValidation,
			"logging":                    testAccVirtualNode_logging,
			"tags":                       testAccVirtualNode_tags,
		},
		"VirtualRouter": {
			"basic":         testAccVirtualRouter_basic,
			"disappears":    testAccVirtualRouter_disappears,
			"multiListener": testAccVirtualRouter_multiListener,
			"tags":          testAccVirtualRouter_tags,
		},
		"VirtualService": {
			"disappears":    testAccVirtualService_disappears,
			"virtualNode":   testAccVirtualService_virtualNode,
			"virtualRouter": testAccVirtualService_virtualRouter,
			"tags":          testAccVirtualService_tags,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
