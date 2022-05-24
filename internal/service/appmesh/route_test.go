package appmesh_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func testAccRoute_grpcRoute(t *testing.T) {
	var r appmesh.RouteData
	resourceName := "aws_appmesh_route.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vrName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn1Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn2Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appmesh.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouteConfig_grpcRoute(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.0.weighted_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.metadata.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.grpc_route.0.match.0.metadata.*", map[string]string{
						"invert":  "false",
						"match.#": "0",
						"name":    "X-Testing1",
					}),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.method_name", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.service_name", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.grpc_retry_events.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.grpc_retry_events.*", "deadline-exceeded"),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.grpc_retry_events.*", "resource-exhausted"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.http_retry_events.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.http_retry_events.*", "server-error"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.max_retries", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.per_retry_timeout.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.per_retry_timeout.0.unit", "s"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.per_retry_timeout.0.value", "15"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.tcp_retry_events.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.timeout.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				Config: testAccRouteConfig_grpcRouteUpdated(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.0.weighted_target.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.metadata.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.grpc_route.0.match.0.metadata.*", map[string]string{
						"invert":  "true",
						"match.#": "0",
						"name":    "X-Testing1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.grpc_route.0.match.0.metadata.*", map[string]string{
						"invert":                "false",
						"match.#":               "1",
						"match.0.range.#":       "1",
						"match.0.range.0.end":   "7",
						"match.0.range.0.start": "2",
						"name":                  "X-Testing2",
					}),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.method_name", "test"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.service_name", "test.local"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.grpc_retry_events.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.grpc_retry_events.*", "cancelled"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.http_retry_events.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.http_retry_events.*", "gateway-error"),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.http_retry_events.*", "client-error"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.max_retries", "3"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.per_retry_timeout.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.per_retry_timeout.0.unit", "ms"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.per_retry_timeout.0.value", "250000"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.tcp_retry_events.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.tcp_retry_events.*", "connection-error"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.timeout.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				Config: testAccRouteConfig_grpcRouteUpdatedWithZeroWeight(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.0.weighted_target.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.metadata.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.grpc_route.0.match.0.metadata.*", map[string]string{
						"invert":  "true",
						"match.#": "0",
						"name":    "X-Testing1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.grpc_route.0.match.0.metadata.*", map[string]string{
						"invert":                "false",
						"match.#":               "1",
						"match.0.range.#":       "1",
						"match.0.range.0.end":   "7",
						"match.0.range.0.start": "2",
						"name":                  "X-Testing2",
					}),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.method_name", "test"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.service_name", "test.local"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.grpc_retry_events.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.grpc_retry_events.*", "cancelled"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.http_retry_events.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.http_retry_events.*", "gateway-error"),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.http_retry_events.*", "client-error"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.max_retries", "3"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.per_retry_timeout.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.per_retry_timeout.0.unit", "ms"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.per_retry_timeout.0.value", "250000"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.tcp_retry_events.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.tcp_retry_events.*", "connection-error"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.timeout.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				Config: testAccRouteConfig_grpcRouteWithMaxRetriesZero(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.0.weighted_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.metadata.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.grpc_route.0.match.0.metadata.*", map[string]string{
						"invert":  "false",
						"match.#": "0",
						"name":    "X-Testing1",
					}),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.method_name", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.service_name", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.grpc_retry_events.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.grpc_retry_events.*", "deadline-exceeded"),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.grpc_retry_events.*", "resource-exhausted"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.http_retry_events.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.http_retry_events.*", "server-error"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.max_retries", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.per_retry_timeout.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.per_retry_timeout.0.unit", "s"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.per_retry_timeout.0.value", "15"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.tcp_retry_events.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.timeout.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccRoute_grpcRouteTimeout(t *testing.T) {
	var r appmesh.RouteData
	resourceName := "aws_appmesh_route.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vrName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn1Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn2Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appmesh.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouteConfig_grpcRouteWithTimeout(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.0.weighted_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.metadata.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.grpc_route.0.match.0.metadata.*", map[string]string{
						"invert":  "false",
						"match.#": "0",
						"name":    "X-Testing1",
					}),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.method_name", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.service_name", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.timeout.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.timeout.0.idle.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.timeout.0.idle.0.unit", "ms"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.timeout.0.idle.0.value", "250000"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.timeout.0.per_request.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				Config: testAccRouteConfig_grpcRouteWithTimeoutUpdated(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.0.weighted_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.metadata.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.grpc_route.0.match.0.metadata.*", map[string]string{
						"invert":  "false",
						"match.#": "0",
						"name":    "X-Testing1",
					}),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.method_name", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.service_name", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.timeout.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.timeout.0.idle.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.timeout.0.idle.0.unit", "s"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.timeout.0.idle.0.value", "10"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.timeout.0.per_request.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.timeout.0.per_request.0.unit", "s"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.timeout.0.per_request.0.value", "5"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccRoute_grpcRouteEmptyMatch(t *testing.T) {
	var r appmesh.RouteData
	resourceName := "aws_appmesh_route.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vrName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn1Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn2Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appmesh.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouteConfig_grpcRouteWithEmptyMatch(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.0.weighted_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.metadata.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.method_name", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.service_name", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.timeout.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccRoute_http2Route(t *testing.T) {
	var r appmesh.RouteData
	resourceName := "aws_appmesh_route.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vrName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn1Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn2Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appmesh.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouteConfig_http2Route(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.weighted_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.header.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.http2_route.0.match.0.header.*", map[string]string{
						"invert":  "false",
						"match.#": "0",
						"name":    "X-Testing1",
					}),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.method", "POST"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.prefix", "/"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.scheme", "http"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.retry_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.http_retry_events.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.http_retry_events.*", "server-error"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.max_retries", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.per_retry_timeout.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.per_retry_timeout.0.unit", "s"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.per_retry_timeout.0.value", "15"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.tcp_retry_events.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.timeout.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				Config: testAccRouteConfig_http2RouteUpdated(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.weighted_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.header.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.http2_route.0.match.0.header.*", map[string]string{
						"invert":  "true",
						"match.#": "0",
						"name":    "X-Testing1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.http2_route.0.match.0.header.*", map[string]string{
						"invert":                "false",
						"match.#":               "1",
						"match.0.range.#":       "1",
						"match.0.range.0.end":   "7",
						"match.0.range.0.start": "2",
						"name":                  "X-Testing2",
					}),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.method", "PUT"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.prefix", "/path"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.scheme", "https"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.retry_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.http_retry_events.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.http_retry_events.*", "gateway-error"),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.http_retry_events.*", "client-error"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.max_retries", "3"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.per_retry_timeout.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.per_retry_timeout.0.unit", "ms"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.per_retry_timeout.0.value", "250000"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.tcp_retry_events.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.tcp_retry_events.*", "connection-error"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.timeout.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccRoute_http2RouteTimeout(t *testing.T) {
	var r appmesh.RouteData
	resourceName := "aws_appmesh_route.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vrName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn1Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn2Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appmesh.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouteConfig_http2RouteWithTimeout(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.weighted_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.header.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.http2_route.0.match.0.header.*", map[string]string{
						"invert":  "false",
						"match.#": "0",
						"name":    "X-Testing1",
					}),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.method", "POST"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.prefix", "/"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.scheme", "http"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.retry_policy.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.timeout.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.timeout.0.idle.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.timeout.0.idle.0.unit", "ms"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.timeout.0.idle.0.value", "250000"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.timeout.0.per_request.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				Config: testAccRouteConfig_http2RouteWithTimeoutUpdated(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.weighted_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.header.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.http2_route.0.match.0.header.*", map[string]string{
						"invert":  "false",
						"match.#": "0",
						"name":    "X-Testing1",
					}),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.method", "POST"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.prefix", "/"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.scheme", "http"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.retry_policy.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.timeout.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.timeout.0.idle.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.timeout.0.idle.0.unit", "s"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.timeout.0.idle.0.value", "10"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.timeout.0.per_request.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.timeout.0.per_request.0.unit", "s"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.timeout.0.per_request.0.value", "5"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccRoute_httpRoute(t *testing.T) {
	var r appmesh.RouteData
	resourceName := "aws_appmesh_route.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vrName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn1Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn2Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appmesh.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouteConfig_http(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.weighted_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.method", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.prefix", "/"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.scheme", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.timeout.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.priority", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				Config: testAccRouteConfig_httpUpdated(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.weighted_target.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.method", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.prefix", "/path"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.scheme", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.timeout.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.priority", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				Config: testAccRouteConfig_httpUpdatedZeroWeight(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.weighted_target.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.method", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.prefix", "/path"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.scheme", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.timeout.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.priority", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccRoute_httpRouteTimeout(t *testing.T) {
	var r appmesh.RouteData
	resourceName := "aws_appmesh_route.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vrName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn1Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn2Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appmesh.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouteConfig_httpTimeout(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.weighted_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.method", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.prefix", "/"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.scheme", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.timeout.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.timeout.0.idle.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.timeout.0.idle.0.unit", "ms"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.timeout.0.idle.0.value", "250000"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.timeout.0.per_request.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.priority", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				Config: testAccRouteConfig_httpTimeoutUpdated(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.weighted_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.method", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.prefix", "/"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.scheme", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.timeout.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.timeout.0.idle.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.timeout.0.idle.0.unit", "s"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.timeout.0.idle.0.value", "10"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.timeout.0.per_request.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.timeout.0.per_request.0.unit", "s"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.timeout.0.per_request.0.value", "5"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.priority", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccRoute_tcpRoute(t *testing.T) {
	var r appmesh.RouteData
	resourceName := "aws_appmesh_route.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vrName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn1Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn2Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appmesh.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouteConfig_tcp(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.priority", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.action.0.weighted_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.timeout.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				Config: testAccRouteConfig_tcpUpdated(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.priority", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.action.0.weighted_target.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.timeout.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				Config: testAccRouteConfig_tcpUpdatedZeroWeight(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.priority", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.action.0.weighted_target.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.timeout.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccRoute_tcpRouteTimeout(t *testing.T) {
	var r appmesh.RouteData
	resourceName := "aws_appmesh_route.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vrName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn1Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn2Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appmesh.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouteConfig_tcpTimeout(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.priority", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.action.0.weighted_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.timeout.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.timeout.0.idle.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.timeout.0.idle.0.unit", "ms"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.timeout.0.idle.0.value", "250000"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				Config: testAccRouteConfig_tcpTimeoutUpdated(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.priority", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.action.0.weighted_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.timeout.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.timeout.0.idle.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.timeout.0.idle.0.unit", "s"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.timeout.0.idle.0.value", "10"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccRoute_tags(t *testing.T) {
	var r appmesh.RouteData
	resourceName := "aws_appmesh_route.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vrName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn1Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn2Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appmesh.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouteConfig_tags(meshName, vrName, vn1Name, vn2Name, rName, "foo", "bar", "good", "bad"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "bar"),
					resource.TestCheckResourceAttr(resourceName, "tags.good", "bad"),
				),
			},
			{
				Config: testAccRouteConfig_tags(meshName, vrName, vn1Name, vn2Name, rName, "foo2", "bar", "good", "bad2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.foo2", "bar"),
					resource.TestCheckResourceAttr(resourceName, "tags.good", "bad2"),
				),
			},
			{
				Config: testAccRouteConfig_http(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccRoute_httpHeader(t *testing.T) {
	var r appmesh.RouteData
	resourceName := "aws_appmesh_route.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vrName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn1Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn2Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appmesh.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouteConfig_httpHeader(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.weighted_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.http_route.0.match.0.header.*", map[string]string{
						"invert":  "false",
						"match.#": "0",
						"name":    "X-Testing1",
					}),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.method", "POST"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.prefix", "/"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.priority", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				Config: testAccRouteConfig_httpHeaderUpdated(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.weighted_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.http_route.0.match.0.header.*", map[string]string{
						"invert":  "true",
						"match.#": "0",
						"name":    "X-Testing1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.http_route.0.match.0.header.*", map[string]string{
						"invert":                "false",
						"match.#":               "1",
						"match.0.range.#":       "1",
						"match.0.range.0.end":   "7",
						"match.0.range.0.start": "2",
						"name":                  "X-Testing2",
					}),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.method", "PUT"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.prefix", "/path"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.priority", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccRoute_routePriority(t *testing.T) {
	var r appmesh.RouteData
	resourceName := "aws_appmesh_route.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vrName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn1Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn2Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appmesh.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouteConfig_routePriority(meshName, vrName, vn1Name, vn2Name, rName, 42),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.weighted_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.method", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.prefix", "/"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.scheme", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.priority", "42"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				Config: testAccRouteConfig_routePriority(meshName, vrName, vn1Name, vn2Name, rName, 1000),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.weighted_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.method", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.prefix", "/"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.scheme", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.priority", "1000"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccRoute_httpRetryPolicy(t *testing.T) {
	var r appmesh.RouteData
	resourceName := "aws_appmesh_route.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vrName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn1Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn2Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appmesh.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouteConfig_httpRetryPolicy(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.weighted_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.method", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.prefix", "/"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.scheme", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.0.http_retry_events.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.http_route.0.retry_policy.0.http_retry_events.*", "server-error"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.0.max_retries", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.0.per_retry_timeout.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.0.per_retry_timeout.0.unit", "s"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.0.per_retry_timeout.0.value", "15"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.0.tcp_retry_events.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				Config: testAccRouteConfig_httpMaxRetriesZero(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.weighted_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.method", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.prefix", "/"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.scheme", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.0.http_retry_events.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.http_route.0.retry_policy.0.http_retry_events.*", "server-error"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.0.max_retries", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.0.per_retry_timeout.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.0.per_retry_timeout.0.unit", "s"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.0.per_retry_timeout.0.value", "15"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.0.tcp_retry_events.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				Config: testAccRouteConfig_httpRetryPolicyUpdated(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.weighted_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.method", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.prefix", "/"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.scheme", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.0.http_retry_events.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.http_route.0.retry_policy.0.http_retry_events.*", "gateway-error"),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.http_route.0.retry_policy.0.http_retry_events.*", "client-error"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.0.max_retries", "3"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.0.per_retry_timeout.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.0.per_retry_timeout.0.unit", "ms"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.0.per_retry_timeout.0.value", "250000"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.0.tcp_retry_events.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.http_route.0.retry_policy.0.tcp_retry_events.*", "connection-error"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccRouteImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s/%s", rs.Primary.Attributes["mesh_name"], rs.Primary.Attributes["virtual_router_name"], rs.Primary.Attributes["name"]), nil
	}
}

func testAccCheckRouteDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AppMeshConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appmesh_route" {
			continue
		}

		_, err := conn.DescribeRoute(&appmesh.DescribeRouteInput{
			MeshName:          aws.String(rs.Primary.Attributes["mesh_name"]),
			RouteName:         aws.String(rs.Primary.Attributes["name"]),
			VirtualRouterName: aws.String(rs.Primary.Attributes["virtual_router_name"]),
		})
		if tfawserr.ErrCodeEquals(err, appmesh.ErrCodeNotFoundException) {
			continue
		}
		if err != nil {
			return err
		}
		return fmt.Errorf("still exist.")
	}

	return nil
}

func testAccCheckRouteExists(name string, v *appmesh.RouteData) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppMeshConn

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		resp, err := conn.DescribeRoute(&appmesh.DescribeRouteInput{
			MeshName:          aws.String(rs.Primary.Attributes["mesh_name"]),
			RouteName:         aws.String(rs.Primary.Attributes["name"]),
			VirtualRouterName: aws.String(rs.Primary.Attributes["virtual_router_name"]),
		})
		if err != nil {
			return err
		}

		*v = *resp.Route

		return nil
	}
}

func testAccRouteConfigBase(meshName, vrName, vrProtocol, vn1Name, vn2Name string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}

resource "aws_appmesh_virtual_router" "test" {
  name      = %[2]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    listener {
      port_mapping {
        port     = 8080
        protocol = %[3]q
      }
    }
  }
}

resource "aws_appmesh_virtual_node" "foo" {
  name      = %[4]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {}
}

resource "aws_appmesh_virtual_node" "bar" {
  name      = %[5]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {}
}
`, meshName, vrName, vrProtocol, vn1Name, vn2Name)
}

func testAccRouteConfig_grpcRoute(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfigBase(meshName, vrName, "grpc", vn1Name, vn2Name), fmt.Sprintf(`
resource "aws_appmesh_route" "test" {
  name                = %[1]q
  mesh_name           = aws_appmesh_mesh.test.id
  virtual_router_name = aws_appmesh_virtual_router.test.name

  spec {
    grpc_route {
      match {
        metadata {
          name = "X-Testing1"
        }
      }

      retry_policy {
        grpc_retry_events = [
          "deadline-exceeded",
          "resource-exhausted",
        ]

        http_retry_events = [
          "server-error",
        ]

        max_retries = 1

        per_retry_timeout {
          unit  = "s"
          value = 15
        }
      }

      action {
        weighted_target {
          virtual_node = aws_appmesh_virtual_node.foo.name
          weight       = 100
        }
      }
    }
  }
}
`, rName))
}

func testAccRouteConfig_grpcRouteUpdated(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfigBase(meshName, vrName, "grpc", vn1Name, vn2Name), fmt.Sprintf(`
resource "aws_appmesh_route" "test" {
  name                = %[1]q
  mesh_name           = aws_appmesh_mesh.test.id
  virtual_router_name = aws_appmesh_virtual_router.test.name

  spec {
    grpc_route {
      match {
        method_name  = "test"
        service_name = "test.local"

        metadata {
          name   = "X-Testing1"
          invert = true
        }

        metadata {
          name   = "X-Testing2"
          invert = false

          match {
            range {
              start = 2
              end   = 7
            }
          }
        }
      }

      retry_policy {
        grpc_retry_events = [
          "cancelled",
        ]

        http_retry_events = [
          "client-error",
          "gateway-error",
        ]

        max_retries = 3

        per_retry_timeout {
          unit  = "ms"
          value = 250000
        }

        tcp_retry_events = [
          "connection-error",
        ]
      }

      action {
        weighted_target {
          virtual_node = aws_appmesh_virtual_node.foo.name
          weight       = 90
        }

        weighted_target {
          virtual_node = aws_appmesh_virtual_node.bar.name
          weight       = 10
        }
      }
    }
  }
}
`, rName))
}

func testAccRouteConfig_grpcRouteUpdatedWithZeroWeight(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfigBase(meshName, vrName, "grpc", vn1Name, vn2Name), fmt.Sprintf(`
resource "aws_appmesh_route" "test" {
  name                = %[1]q
  mesh_name           = aws_appmesh_mesh.test.id
  virtual_router_name = aws_appmesh_virtual_router.test.name

  spec {
    grpc_route {
      match {
        method_name  = "test"
        service_name = "test.local"

        metadata {
          name   = "X-Testing1"
          invert = true
        }

        metadata {
          name   = "X-Testing2"
          invert = false

          match {
            range {
              start = 2
              end   = 7
            }
          }
        }
      }

      retry_policy {
        grpc_retry_events = [
          "cancelled",
        ]

        http_retry_events = [
          "client-error",
          "gateway-error",
        ]

        max_retries = 3

        per_retry_timeout {
          unit  = "ms"
          value = 250000
        }

        tcp_retry_events = [
          "connection-error",
        ]
      }

      action {
        weighted_target {
          virtual_node = aws_appmesh_virtual_node.foo.name
          weight       = 99
        }

        weighted_target {
          virtual_node = aws_appmesh_virtual_node.bar.name
          weight       = 0
        }
      }
    }
  }
}
`, rName))
}

func testAccRouteConfig_grpcRouteWithTimeout(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfigBase(meshName, vrName, "grpc", vn1Name, vn2Name), fmt.Sprintf(`
resource "aws_appmesh_route" "test" {
  name                = %[1]q
  mesh_name           = aws_appmesh_mesh.test.id
  virtual_router_name = aws_appmesh_virtual_router.test.name

  spec {
    grpc_route {
      match {
        metadata {
          name = "X-Testing1"
        }
      }

      action {
        weighted_target {
          virtual_node = aws_appmesh_virtual_node.foo.name
          weight       = 100
        }
      }

      timeout {
        idle {
          unit  = "ms"
          value = 250000
        }
      }
    }
  }
}
`, rName))
}

func testAccRouteConfig_grpcRouteWithTimeoutUpdated(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfigBase(meshName, vrName, "grpc", vn1Name, vn2Name), fmt.Sprintf(`
resource "aws_appmesh_route" "test" {
  name                = %[1]q
  mesh_name           = aws_appmesh_mesh.test.id
  virtual_router_name = aws_appmesh_virtual_router.test.name

  spec {
    grpc_route {
      match {
        metadata {
          name = "X-Testing1"
        }
      }

      action {
        weighted_target {
          virtual_node = aws_appmesh_virtual_node.foo.name
          weight       = 100
        }
      }

      timeout {
        idle {
          unit  = "s"
          value = 10
        }

        per_request {
          unit  = "s"
          value = 5
        }
      }
    }
  }
}
`, rName))
}

func testAccRouteConfig_grpcRouteWithEmptyMatch(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfigBase(meshName, vrName, "grpc", vn1Name, vn2Name), fmt.Sprintf(`
resource "aws_appmesh_route" "test" {
  name                = %[1]q
  mesh_name           = aws_appmesh_mesh.test.id
  virtual_router_name = aws_appmesh_virtual_router.test.name

  spec {
    grpc_route {
      match {}

      action {
        weighted_target {
          virtual_node = aws_appmesh_virtual_node.foo.name
          weight       = 100
        }
      }
    }
  }
}
`, rName))
}

func testAccRouteConfig_grpcRouteWithMaxRetriesZero(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfigBase(meshName, vrName, "grpc", vn1Name, vn2Name), fmt.Sprintf(`
resource "aws_appmesh_route" "test" {
  name                = %[1]q
  mesh_name           = aws_appmesh_mesh.test.id
  virtual_router_name = aws_appmesh_virtual_router.test.name

  spec {
    grpc_route {
      match {
        metadata {
          name = "X-Testing1"
        }
      }

      retry_policy {
        grpc_retry_events = [
          "deadline-exceeded",
          "resource-exhausted",
        ]

        http_retry_events = [
          "server-error",
        ]

        max_retries = 0

        per_retry_timeout {
          unit  = "s"
          value = 15
        }
      }

      action {
        weighted_target {
          virtual_node = aws_appmesh_virtual_node.foo.name
          weight       = 100
        }
      }
    }
  }
}
`, rName))
}

func testAccRouteConfig_http2Route(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfigBase(meshName, vrName, "http2", vn1Name, vn2Name), fmt.Sprintf(`
resource "aws_appmesh_route" "test" {
  name                = %[1]q
  mesh_name           = aws_appmesh_mesh.test.id
  virtual_router_name = aws_appmesh_virtual_router.test.name

  spec {
    http2_route {
      match {
        prefix = "/"
        method = "POST"
        scheme = "http"

        header {
          name = "X-Testing1"
        }
      }

      retry_policy {
        http_retry_events = [
          "server-error",
        ]

        max_retries = 1

        per_retry_timeout {
          unit  = "s"
          value = 15
        }
      }

      action {
        weighted_target {
          virtual_node = aws_appmesh_virtual_node.foo.name
          weight       = 100
        }
      }
    }
  }
}
`, rName))
}

func testAccRouteConfig_http2RouteUpdated(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfigBase(meshName, vrName, "http2", vn1Name, vn2Name), fmt.Sprintf(`
resource "aws_appmesh_route" "test" {
  name                = %[1]q
  mesh_name           = aws_appmesh_mesh.test.id
  virtual_router_name = aws_appmesh_virtual_router.test.name

  spec {
    http2_route {
      match {
        prefix = "/path"
        method = "PUT"
        scheme = "https"

        header {
          name   = "X-Testing1"
          invert = true
        }

        header {
          name   = "X-Testing2"
          invert = false

          match {
            range {
              start = 2
              end   = 7
            }
          }
        }
      }

      retry_policy {
        http_retry_events = [
          "client-error",
          "gateway-error",
        ]

        max_retries = 3

        per_retry_timeout {
          unit  = "ms"
          value = 250000
        }

        tcp_retry_events = [
          "connection-error",
        ]
      }

      action {
        weighted_target {
          virtual_node = aws_appmesh_virtual_node.foo.name
          weight       = 100
        }
      }
    }
  }
}
`, rName))
}

func testAccRouteConfig_http2RouteWithTimeout(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfigBase(meshName, vrName, "http2", vn1Name, vn2Name), fmt.Sprintf(`
resource "aws_appmesh_route" "test" {
  name                = %[1]q
  mesh_name           = aws_appmesh_mesh.test.id
  virtual_router_name = aws_appmesh_virtual_router.test.name

  spec {
    http2_route {
      match {
        prefix = "/"
        method = "POST"
        scheme = "http"

        header {
          name = "X-Testing1"
        }
      }

      action {
        weighted_target {
          virtual_node = aws_appmesh_virtual_node.foo.name
          weight       = 100
        }
      }

      timeout {
        idle {
          unit  = "ms"
          value = 250000
        }
      }
    }
  }
}
`, rName))
}

func testAccRouteConfig_http2RouteWithTimeoutUpdated(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfigBase(meshName, vrName, "http2", vn1Name, vn2Name), fmt.Sprintf(`
resource "aws_appmesh_route" "test" {
  name                = %[1]q
  mesh_name           = aws_appmesh_mesh.test.id
  virtual_router_name = aws_appmesh_virtual_router.test.name

  spec {
    http2_route {
      match {
        prefix = "/"
        method = "POST"
        scheme = "http"

        header {
          name = "X-Testing1"
        }
      }

      action {
        weighted_target {
          virtual_node = aws_appmesh_virtual_node.foo.name
          weight       = 100
        }
      }

      timeout {
        idle {
          unit  = "s"
          value = 10
        }

        per_request {
          unit  = "s"
          value = 5
        }
      }
    }
  }
}
`, rName))
}

func testAccRouteConfig_http(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfigBase(meshName, vrName, "http", vn1Name, vn2Name), fmt.Sprintf(`
resource "aws_appmesh_route" "test" {
  name                = %[1]q
  mesh_name           = aws_appmesh_mesh.test.id
  virtual_router_name = aws_appmesh_virtual_router.test.name

  spec {
    http_route {
      match {
        prefix = "/"
      }

      action {
        weighted_target {
          virtual_node = aws_appmesh_virtual_node.foo.name
          weight       = 100
        }
      }
    }
  }
}
`, rName))
}

func testAccRouteConfig_httpUpdated(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfigBase(meshName, vrName, "http", vn1Name, vn2Name), fmt.Sprintf(`
resource "aws_appmesh_route" "test" {
  name                = %[1]q
  mesh_name           = aws_appmesh_mesh.test.id
  virtual_router_name = aws_appmesh_virtual_router.test.name

  spec {
    http_route {
      match {
        prefix = "/path"
      }

      action {
        weighted_target {
          virtual_node = aws_appmesh_virtual_node.foo.name
          weight       = 90
        }

        weighted_target {
          virtual_node = aws_appmesh_virtual_node.bar.name
          weight       = 10
        }
      }
    }
  }
}
`, rName))
}

func testAccRouteConfig_httpUpdatedZeroWeight(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfigBase(meshName, vrName, "http", vn1Name, vn2Name), fmt.Sprintf(`
resource "aws_appmesh_route" "test" {
  name                = %[1]q
  mesh_name           = aws_appmesh_mesh.test.id
  virtual_router_name = aws_appmesh_virtual_router.test.name

  spec {
    http_route {
      match {
        prefix = "/path"
      }

      action {
        weighted_target {
          virtual_node = aws_appmesh_virtual_node.foo.name
          weight       = 99
        }

        weighted_target {
          virtual_node = aws_appmesh_virtual_node.bar.name
          weight       = 0
        }
      }
    }
  }
}
`, rName))
}

func testAccRouteConfig_httpTimeout(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfigBase(meshName, vrName, "http", vn1Name, vn2Name), fmt.Sprintf(`
resource "aws_appmesh_route" "test" {
  name                = %[1]q
  mesh_name           = aws_appmesh_mesh.test.id
  virtual_router_name = aws_appmesh_virtual_router.test.name

  spec {
    http_route {
      match {
        prefix = "/"
      }

      action {
        weighted_target {
          virtual_node = aws_appmesh_virtual_node.foo.name
          weight       = 100
        }
      }

      timeout {
        idle {
          unit  = "ms"
          value = 250000
        }
      }
    }
  }
}
`, rName))
}

func testAccRouteConfig_httpTimeoutUpdated(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfigBase(meshName, vrName, "http", vn1Name, vn2Name), fmt.Sprintf(`
resource "aws_appmesh_route" "test" {
  name                = %[1]q
  mesh_name           = aws_appmesh_mesh.test.id
  virtual_router_name = aws_appmesh_virtual_router.test.name

  spec {
    http_route {
      match {
        prefix = "/"
      }

      action {
        weighted_target {
          virtual_node = aws_appmesh_virtual_node.foo.name
          weight       = 100
        }
      }

      timeout {
        idle {
          unit  = "s"
          value = 10
        }

        per_request {
          unit  = "s"
          value = 5
        }
      }
    }
  }
}
`, rName))
}

func testAccRouteConfig_tcp(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfigBase(meshName, vrName, "tcp", vn1Name, vn2Name), fmt.Sprintf(`
resource "aws_appmesh_route" "test" {
  name                = %[1]q
  mesh_name           = aws_appmesh_mesh.test.id
  virtual_router_name = aws_appmesh_virtual_router.test.name

  spec {
    tcp_route {
      action {
        weighted_target {
          virtual_node = aws_appmesh_virtual_node.foo.name
          weight       = 100
        }
      }
    }
  }
}
`, rName))
}

func testAccRouteConfig_tcpUpdated(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfigBase(meshName, vrName, "tcp", vn1Name, vn2Name), fmt.Sprintf(`
resource "aws_appmesh_route" "test" {
  name                = %[1]q
  mesh_name           = aws_appmesh_mesh.test.id
  virtual_router_name = aws_appmesh_virtual_router.test.name

  spec {
    tcp_route {
      action {
        weighted_target {
          virtual_node = aws_appmesh_virtual_node.foo.name
          weight       = 90
        }

        weighted_target {
          virtual_node = aws_appmesh_virtual_node.bar.name
          weight       = 10
        }
      }
    }
  }
}
`, rName))
}

func testAccRouteConfig_tcpUpdatedZeroWeight(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfigBase(meshName, vrName, "tcp", vn1Name, vn2Name), fmt.Sprintf(`
resource "aws_appmesh_route" "test" {
  name                = %[1]q
  mesh_name           = aws_appmesh_mesh.test.id
  virtual_router_name = aws_appmesh_virtual_router.test.name

  spec {
    tcp_route {
      action {
        weighted_target {
          virtual_node = aws_appmesh_virtual_node.foo.name
          weight       = 99
        }

        weighted_target {
          virtual_node = aws_appmesh_virtual_node.bar.name
          weight       = 0
        }
      }
    }
  }
}
`, rName))
}

func testAccRouteConfig_tcpTimeout(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfigBase(meshName, vrName, "tcp", vn1Name, vn2Name), fmt.Sprintf(`
resource "aws_appmesh_route" "test" {
  name                = %[1]q
  mesh_name           = aws_appmesh_mesh.test.id
  virtual_router_name = aws_appmesh_virtual_router.test.name

  spec {
    tcp_route {
      action {
        weighted_target {
          virtual_node = aws_appmesh_virtual_node.foo.name
          weight       = 100
        }
      }

      timeout {
        idle {
          unit  = "ms"
          value = 250000
        }
      }
    }
  }
}
`, rName))
}

func testAccRouteConfig_tcpTimeoutUpdated(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfigBase(meshName, vrName, "tcp", vn1Name, vn2Name), fmt.Sprintf(`
resource "aws_appmesh_route" "test" {
  name                = %[1]q
  mesh_name           = aws_appmesh_mesh.test.id
  virtual_router_name = aws_appmesh_virtual_router.test.name

  spec {
    tcp_route {
      action {
        weighted_target {
          virtual_node = aws_appmesh_virtual_node.foo.name
          weight       = 100
        }
      }

      timeout {
        idle {
          unit  = "s"
          value = 10
        }
      }
    }
  }
}
`, rName))
}

func testAccRouteConfig_tags(meshName, vrName, vn1Name, vn2Name, rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccRouteConfigBase(meshName, vrName, "http", vn1Name, vn2Name), fmt.Sprintf(`
resource "aws_appmesh_route" "test" {
  name                = %[1]q
  mesh_name           = aws_appmesh_mesh.test.id
  virtual_router_name = aws_appmesh_virtual_router.test.name

  spec {
    http_route {
      match {
        prefix = "/"
      }

      action {
        weighted_target {
          virtual_node = aws_appmesh_virtual_node.foo.name
          weight       = 100
        }
      }
    }
  }

  tags = {
    %[2]s = %[3]q
    %[4]s = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccRouteConfig_httpHeader(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfigBase(meshName, vrName, "http", vn1Name, vn2Name), fmt.Sprintf(`
resource "aws_appmesh_route" "test" {
  name                = %[1]q
  mesh_name           = aws_appmesh_mesh.test.id
  virtual_router_name = aws_appmesh_virtual_router.test.name

  spec {
    http_route {
      match {
        prefix = "/"
        method = "POST"

        header {
          name = "X-Testing1"
        }
      }

      action {
        weighted_target {
          virtual_node = aws_appmesh_virtual_node.foo.name
          weight       = 100
        }
      }
    }
  }
}
`, rName))
}

func testAccRouteConfig_httpHeaderUpdated(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfigBase(meshName, vrName, "http", vn1Name, vn2Name), fmt.Sprintf(`
resource "aws_appmesh_route" "test" {
  name                = %[1]q
  mesh_name           = aws_appmesh_mesh.test.id
  virtual_router_name = aws_appmesh_virtual_router.test.name

  spec {
    http_route {
      match {
        prefix = "/path"
        method = "PUT"

        header {
          name   = "X-Testing1"
          invert = true
        }

        header {
          name   = "X-Testing2"
          invert = false

          match {
            range {
              start = 2
              end   = 7
            }
          }
        }
      }

      action {
        weighted_target {
          virtual_node = aws_appmesh_virtual_node.foo.name
          weight       = 100
        }
      }
    }
  }
}
`, rName))
}

func testAccRouteConfig_routePriority(meshName, vrName, vn1Name, vn2Name, rName string, priority int) string {
	return acctest.ConfigCompose(testAccRouteConfigBase(meshName, vrName, "http", vn1Name, vn2Name), fmt.Sprintf(`
resource "aws_appmesh_route" "test" {
  name                = %[1]q
  mesh_name           = aws_appmesh_mesh.test.id
  virtual_router_name = aws_appmesh_virtual_router.test.name

  spec {
    http_route {
      match {
        prefix = "/"
      }

      action {
        weighted_target {
          virtual_node = aws_appmesh_virtual_node.foo.name
          weight       = 100
        }
      }
    }

    priority = %[2]d
  }
}
`, rName, priority))
}

func testAccRouteConfig_httpRetryPolicy(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfigBase(meshName, vrName, "http", vn1Name, vn2Name), fmt.Sprintf(`
resource "aws_appmesh_route" "test" {
  name                = %[1]q
  mesh_name           = aws_appmesh_mesh.test.id
  virtual_router_name = aws_appmesh_virtual_router.test.name

  spec {
    http_route {
      match {
        prefix = "/"
      }

      retry_policy {
        http_retry_events = [
          "server-error",
        ]

        max_retries = 1

        per_retry_timeout {
          unit  = "s"
          value = 15
        }
      }

      action {
        weighted_target {
          virtual_node = aws_appmesh_virtual_node.foo.name
          weight       = 100
        }
      }
    }
  }
}
`, rName))
}

func testAccRouteConfig_httpMaxRetriesZero(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfigBase(meshName, vrName, "http", vn1Name, vn2Name), fmt.Sprintf(`
resource "aws_appmesh_route" "test" {
  name                = %[1]q
  mesh_name           = aws_appmesh_mesh.test.id
  virtual_router_name = aws_appmesh_virtual_router.test.name

  spec {
    http_route {
      match {
        prefix = "/"
      }

      retry_policy {
        http_retry_events = [
          "server-error",
        ]

        max_retries = 0

        per_retry_timeout {
          unit  = "s"
          value = 15
        }
      }

      action {
        weighted_target {
          virtual_node = aws_appmesh_virtual_node.foo.name
          weight       = 100
        }
      }
    }
  }
}
`, rName))
}

func testAccRouteConfig_httpRetryPolicyUpdated(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfigBase(meshName, vrName, "http", vn1Name, vn2Name), fmt.Sprintf(`
resource "aws_appmesh_route" "test" {
  name                = %[1]q
  mesh_name           = aws_appmesh_mesh.test.id
  virtual_router_name = aws_appmesh_virtual_router.test.name

  spec {
    http_route {
      match {
        prefix = "/"
      }

      retry_policy {
        http_retry_events = [
          "client-error",
          "gateway-error",
        ]

        max_retries = 3

        per_retry_timeout {
          unit  = "ms"
          value = 250000
        }

        tcp_retry_events = [
          "connection-error",
        ]
      }

      action {
        weighted_target {
          virtual_node = aws_appmesh_virtual_node.foo.name
          weight       = 100
        }
      }
    }
  }
}
`, rName))
}
