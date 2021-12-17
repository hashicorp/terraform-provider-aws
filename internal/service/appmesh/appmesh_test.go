package appmesh_test

import (
	"testing"
)

func TestAccAppMesh_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"GatewayRoute": {
			"basic":      testAccGatewayRoute_basic,
			"disappears": testAccGatewayRoute_disappears,
			"grpcRoute":  testAccGatewayRoute_GRPCRoute,
			"httpRoute":  testAccGatewayRoute_HTTPRoute,
			"http2Route": testAccGatewayRoute_HTTP2Route,
			"tags":       testAccGatewayRoute_Tags,
		},
		"Mesh": {
			"basic":        testAccMesh_basic,
			"egressFilter": testAccMesh_egressFilter,
			"tags":         testAccMesh_tags,
		},
		"Route": {
			"grpcRoute":           testAccRoute_grpcRoute,
			"grpcRouteEmptyMatch": testAccRoute_grpcRouteEmptyMatch,
			"grpcRouteTimeout":    testAccRoute_grpcRouteTimeout,
			"http2Route":          testAccRoute_http2Route,
			"http2RouteTimeout":   testAccRoute_http2RouteTimeout,
			"httpHeader":          testAccRoute_httpHeader,
			"httpRetryPolicy":     testAccRoute_httpRetryPolicy,
			"httpRoute":           testAccRoute_httpRoute,
			"httpRouteTimeout":    testAccRoute_httpRouteTimeout,
			"routePriority":       testAccRoute_routePriority,
			"tcpRoute":            testAccRoute_tcpRoute,
			"tcpRouteTimeout":     testAccRoute_tcpRouteTimeout,
			"tags":                testAccRoute_tags,
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
			"logging":                    testAccVirtualNode_logging,
			"tags":                       testAccVirtualNode_tags,
		},
		"VirtualRouter": {
			"basic": testAccVirtualRouter_basic,
			"tags":  testAccVirtualRouter_tags,
		},
		"VirtualService": {
			"virtualNode":   testAccVirtualService_virtualNode,
			"virtualRouter": testAccVirtualService_virtualRouter,
			"tags":          testAccVirtualService_tags,
		},
	}

	for group, m := range testCases {
		m := m
		t.Run(group, func(t *testing.T) {
			for name, tc := range m {
				tc := tc
				t.Run(name, func(t *testing.T) {
					tc(t)
				})
			}
		})
	}
}
