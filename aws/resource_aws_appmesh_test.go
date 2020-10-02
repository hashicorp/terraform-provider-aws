package aws

import (
	"testing"
)

func TestAccAWSAppmesh_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Mesh": {
			"basic":        testAccAwsAppmeshMesh_basic,
			"egressFilter": testAccAwsAppmeshMesh_egressFilter,
			"tags":         testAccAwsAppmeshMesh_tags,
		},
		"Route": {
			"grpcRoute":       testAccAwsAppmeshRoute_grpcRoute,
			"http2Route":      testAccAwsAppmeshRoute_http2Route,
			"httpHeader":      testAccAwsAppmeshRoute_httpHeader,
			"httpRetryPolicy": testAccAwsAppmeshRoute_httpRetryPolicy,
			"httpRoute":       testAccAwsAppmeshRoute_httpRoute,
			"routePriority":   testAccAwsAppmeshRoute_routePriority,
			"tcpRoute":        testAccAwsAppmeshRoute_tcpRoute,
			"tags":            testAccAwsAppmeshRoute_tags,
		},
		"VirtualNode": {
			"basic":                    testAccAwsAppmeshVirtualNode_basic,
			"backendDefaults":          testAccAwsAppmeshVirtualNode_backendDefaults,
			"clientPolicyAcm":          testAccAwsAppmeshVirtualNode_clientPolicyAcm,
			"clientPolicyFile":         testAccAwsAppmeshVirtualNode_clientPolicyFile,
			"cloudMapServiceDiscovery": testAccAwsAppmeshVirtualNode_cloudMapServiceDiscovery,
			"listenerHealthChecks":     testAccAwsAppmeshVirtualNode_listenerHealthChecks,
			"logging":                  testAccAwsAppmeshVirtualNode_logging,
			"tls":                      testAccAwsAppmeshVirtualNode_tls,
			"tags":                     testAccAwsAppmeshVirtualNode_tags,
		},
		"VirtualRouter": {
			"basic": testAccAwsAppmeshVirtualRouter_basic,
			"tags":  testAccAwsAppmeshVirtualRouter_tags,
		},
		"VirtualService": {
			"virtualNode":   testAccAwsAppmeshVirtualService_virtualNode,
			"virtualRouter": testAccAwsAppmeshVirtualService_virtualRouter,
			"tags":          testAccAwsAppmeshVirtualService_tags,
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
