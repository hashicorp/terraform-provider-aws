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
			acctest.CtBasic:                testAccGatewayRoute_basic,
			acctest.CtDisappears:           testAccGatewayRoute_disappears,
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
			"tags":                         testAccAppMeshGatewayRoute_tagsSerial,
			"dataSourceBasic":              testAccGatewayRouteDataSource_basic,
			"dataSource_tags":              testAccAppMeshGatewayRouteDataSource_tagsSerial,
		},
		"Mesh": {
			acctest.CtBasic:       testAccMesh_basic,
			acctest.CtDisappears:  testAccMesh_disappears,
			"egressFilter":        testAccMesh_egressFilter,
			"serviceDiscovery":    testAccMesh_serviceDiscovery,
			"tags":                testAccAppMeshServiceMesh_tagsSerial,
			"dataSourceBasic":     testAccMeshDataSource_basic,
			"dataSourceMeshOwner": testAccMeshDataSource_meshOwner,
			"dataSourceSpecSet":   testAccMeshDataSource_specSet,
			"dataSourceShared":    testAccMeshDataSource_shared,
			"dataSource_tags":     testAccAppMeshServiceMeshDataSource_tagsSerial,
		},
		"Route": {
			acctest.CtDisappears:               testAccRoute_disappears,
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
			"tags":                             testAccAppMeshRoute_tagsSerial,
			"dataSourceHTTP2Route":             testAccRouteDataSource_http2Route,
			"dataSourceHTTPRoute":              testAccRouteDataSource_httpRoute,
			"dataSourceGRPCRoute":              testAccRouteDataSource_grpcRoute,
			"dataSourceTCPRoute":               testAccRouteDataSource_tcpRoute,
			"dataSource_tags":                  testAccAppMeshRouteDataSource_tagsSerial,
		},
		"VirtualGateway": {
			acctest.CtBasic:              testAccVirtualGateway_basic,
			acctest.CtDisappears:         testAccVirtualGateway_disappears,
			"backendDefaults":            testAccVirtualGateway_BackendDefaults,
			"backendDefaultsCertificate": testAccVirtualGateway_BackendDefaultsCertificate,
			"listenerConnectionPool":     testAccVirtualGateway_ListenerConnectionPool,
			"listenerHealthChecks":       testAccVirtualGateway_ListenerHealthChecks,
			"listenerTls":                testAccVirtualGateway_ListenerTLS,
			"listenerValidation":         testAccVirtualGateway_ListenerValidation,
			"multiListenerValidation":    testAccVirtualGateway_MultiListenerValidation,
			"logging":                    testAccVirtualGateway_Logging,
			"tags":                       testAccAppMeshVirtualGateway_tagsSerial,
			"dataSourceBasic":            testAccVirtualGatewayDataSource_basic,
			"dataSource_tags":            testAccAppMeshVirtualGatewayDataSource_tagsSerial,
		},
		"VirtualNode": {
			acctest.CtBasic:              testAccVirtualNode_basic,
			acctest.CtDisappears:         testAccVirtualNode_disappears,
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
			"tags":                       testAccAppMeshVirtualNode_tagsSerial,
			"dataSourceBasic":            testAccVirtualNodeDataSource_basic,
			"dataSource_tags":            testAccAppMeshVirtualNodeDataSource_tagsSerial,
		},
		"VirtualRouter": {
			acctest.CtBasic:      testAccVirtualRouter_basic,
			acctest.CtDisappears: testAccVirtualRouter_disappears,
			"multiListener":      testAccVirtualRouter_multiListener,
			"tags":               testAccAppMeshVirtualRouter_tagsSerial,
			"dataSourceBasic":    testAccVirtualRouterDataSource_basic,
			"dataSource_tags":    testAccAppMeshVirtualRouterDataSource_tagsSerial,
		},
		"VirtualService": {
			acctest.CtDisappears:      testAccVirtualService_disappears,
			"virtualNode":             testAccVirtualService_virtualNode,
			"virtualRouter":           testAccVirtualService_virtualRouter,
			"tags":                    testAccAppMeshVirtualService_tagsSerial,
			"dataSourceVirtualNode":   testAccVirtualServiceDataSource_virtualNode,
			"dataSourceVirtualRouter": testAccVirtualServiceDataSource_virtualRouter,
			"dataSource_tags":         testAccAppMeshVirtualServiceDataSource_tagsSerial,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
