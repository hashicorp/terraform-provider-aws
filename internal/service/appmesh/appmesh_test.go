// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appmesh_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccAppMesh_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"GatewayRoute": {
			"basic":                        testAccGatewayRoute_basic,
			"disappears":                   testAccGatewayRoute_disappears,
			"grpcRoute":                    testAccGatewayRoute_grpcRoute,
			"grpcRouteTargetPort":          testAccGatewayRoute_grpcRouteTargetPort,
			"grpcRouteWithPort":            testAccGatewayRoute_grpcRouteWithPort,
			"httpRoute":                    testAccGatewayRoute_httpRoute,
			"httpRouteTargetPort":          testAccGatewayRoute_httpRouteTargetPort,
			"httpRouteWithPath":            testAccGatewayRoute_httpRouteWithPath,
			"httpRouteWithPort":            testAccGatewayRoute_httpRouteWithPort,
			"http2Route":                   testAccGatewayRoute_http2Route,
			"http2RouteTargetPort":         testAccGatewayRoute_http2RouteTargetPort,
			"http2RouteWithPort":           testAccGatewayRoute_http2RouteWithPort,
			"http2RouteWithQueryParameter": testAccGatewayRoute_http2RouteWithQueryParameter,
			"tags":                         testAccGatewayRoute_tags,
			"dataSourceBasic":              testAccGatewayRouteDataSource_basic,
		},
		"Mesh": {
			"basic":                    testAccMesh_basic,
			"disappears":               testAccMesh_disappears,
			"egressFilter":             testAccMesh_egressFilter,
			"tags":                     testAccMesh_tags,
			"dataSourceBasic":          testAccMeshDataSource_basic,
			"dataSourceMeshOwner":      testAccMeshDataSource_meshOwner,
			"dataSourceSpecAndTagsSet": testAccMeshDataSource_specAndTagsSet,
			"dataSourceShared":         testAccMeshDataSource_shared,
		},
		"Route": {
			"disappears":                       testAccRoute_disappears,
			"grpcRoute":                        testAccRoute_grpcRoute,
			"grpcRouteWithPortMatch":           testAccRoute_grpcRouteWithPortMatch,
			"grpcRouteEmptyMatch":              testAccRoute_grpcRouteEmptyMatch,
			"grpcRouteTimeout":                 testAccRoute_grpcRouteTimeout,
			"http2Route":                       testAccRoute_http2Route,
			"http2RouteWithPathMatch":          testAccRoute_http2RouteWithPathMatch,
			"http2RouteWithPortMatch":          testAccRoute_http2RouteWithPortMatch,
			"http2RouteTimeout":                testAccRoute_http2RouteTimeout,
			"httpHeader":                       testAccRoute_httpHeader,
			"httpRetryPolicy":                  testAccRoute_httpRetryPolicy,
			"httpRoute":                        testAccRoute_httpRoute,
			"httpRouteWithPortMatch":           testAccRoute_httpRouteWithPortMatch,
			"httpRouteWithQueryParameterMatch": testAccRoute_httpRouteWithQueryParameterMatch,
			"httpRouteTimeout":                 testAccRoute_httpRouteTimeout,
			"routePriority":                    testAccRoute_routePriority,
			"tcpRoute":                         testAccRoute_tcpRoute,
			"tcpRouteWithPortMatch":            testAccRoute_tcpRouteWithPortMatch,
			"tcpRouteTimeout":                  testAccRoute_tcpRouteTimeout,
			"tags":                             testAccRoute_tags,
			"dataSourceHTTP2Route":             testAccRouteDataSource_http2Route,
			"dataSourceHTTPRoute":              testAccRouteDataSource_httpRoute,
			"dataSourceGRPCRoute":              testAccRouteDataSource_grpcRoute,
			"dataSourceTCPRoute":               testAccRouteDataSource_tcpRoute,
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
			"dataSourceBasic":            testAccVirtualGatewayDataSource_basic,
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
			"dataSourceBasic":            testAccVirtualNodeDataSource_basic,
		},
		"VirtualRouter": {
			"basic":           testAccVirtualRouter_basic,
			"disappears":      testAccVirtualRouter_disappears,
			"multiListener":   testAccVirtualRouter_multiListener,
			"tags":            testAccVirtualRouter_tags,
			"dataSourceBasic": testAccVirtualRouterDataSource_basic,
		},
		"VirtualService": {
			"disappears":              testAccVirtualService_disappears,
			"virtualNode":             testAccVirtualService_virtualNode,
			"virtualRouter":           testAccVirtualService_virtualRouter,
			"tags":                    testAccVirtualService_tags,
			"dataSourceVirtualNode":   testAccVirtualServiceDataSource_virtualNode,
			"dataSourceVirtualRouter": testAccVirtualServiceDataSource_virtualRouter,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
