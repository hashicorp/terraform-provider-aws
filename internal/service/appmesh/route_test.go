// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appmesh_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/appmesh"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappmesh "github.com/hashicorp/terraform-provider-aws/internal/service/appmesh"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccRoute_grpcRoute(t *testing.T) {
	ctx := acctest.Context(t)
	var r appmesh.RouteData
	resourceName := "aws_appmesh_route.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vrName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn1Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn2Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRouteConfig_grpcRoute(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.0.weighted_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.metadata.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.grpc_route.0.match.0.metadata.*", map[string]string{
						"invert":       acctest.CtFalse,
						"match.#":      acctest.Ct0,
						names.AttrName: "X-Testing1",
					}),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.method_name", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.service_name", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.grpc_retry_events.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.grpc_retry_events.*", "deadline-exceeded"),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.grpc_retry_events.*", "resource-exhausted"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.http_retry_events.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.http_retry_events.*", "server-error"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.max_retries", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.per_retry_timeout.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.per_retry_timeout.0.unit", "s"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.per_retry_timeout.0.value", "15"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.tcp_retry_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.timeout.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				Config: testAccRouteConfig_grpcRouteUpdated(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.0.weighted_target.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.metadata.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.grpc_route.0.match.0.metadata.*", map[string]string{
						"invert":       acctest.CtTrue,
						"match.#":      acctest.Ct0,
						names.AttrName: "X-Testing1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.grpc_route.0.match.0.metadata.*", map[string]string{
						"invert":                acctest.CtFalse,
						"match.#":               acctest.Ct1,
						"match.0.range.#":       acctest.Ct1,
						"match.0.range.0.end":   "7",
						"match.0.range.0.start": acctest.Ct2,
						names.AttrName:          "X-Testing2",
					}),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.method_name", "test"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.service_name", "test.local"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.grpc_retry_events.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.grpc_retry_events.*", "cancelled"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.http_retry_events.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.http_retry_events.*", "gateway-error"),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.http_retry_events.*", "client-error"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.max_retries", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.per_retry_timeout.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.per_retry_timeout.0.unit", "ms"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.per_retry_timeout.0.value", "250000"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.tcp_retry_events.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.tcp_retry_events.*", "connection-error"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.timeout.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				Config: testAccRouteConfig_grpcRouteUpdatedWithZeroWeight(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.0.weighted_target.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.metadata.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.grpc_route.0.match.0.metadata.*", map[string]string{
						"invert":       acctest.CtTrue,
						"match.#":      acctest.Ct0,
						names.AttrName: "X-Testing1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.grpc_route.0.match.0.metadata.*", map[string]string{
						"invert":                acctest.CtFalse,
						"match.#":               acctest.Ct1,
						"match.0.range.#":       acctest.Ct1,
						"match.0.range.0.end":   "7",
						"match.0.range.0.start": acctest.Ct2,
						names.AttrName:          "X-Testing2",
					}),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.method_name", "test"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.service_name", "test.local"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.grpc_retry_events.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.grpc_retry_events.*", "cancelled"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.http_retry_events.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.http_retry_events.*", "gateway-error"),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.http_retry_events.*", "client-error"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.max_retries", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.per_retry_timeout.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.per_retry_timeout.0.unit", "ms"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.per_retry_timeout.0.value", "250000"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.tcp_retry_events.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.tcp_retry_events.*", "connection-error"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.timeout.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				Config: testAccRouteConfig_grpcRouteWithMaxRetriesZero(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.0.weighted_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.metadata.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.grpc_route.0.match.0.metadata.*", map[string]string{
						"invert":       acctest.CtFalse,
						"match.#":      acctest.Ct0,
						names.AttrName: "X-Testing1",
					}),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.method_name", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.service_name", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.grpc_retry_events.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.grpc_retry_events.*", "deadline-exceeded"),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.grpc_retry_events.*", "resource-exhausted"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.http_retry_events.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.http_retry_events.*", "server-error"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.max_retries", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.per_retry_timeout.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.per_retry_timeout.0.unit", "s"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.per_retry_timeout.0.value", "15"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.tcp_retry_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.timeout.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
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

func testAccRoute_grpcRouteWithPortMatch(t *testing.T) {
	ctx := acctest.Context(t)
	var r appmesh.RouteData
	resourceName := "aws_appmesh_route.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vrName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn1Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn2Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRouteConfig_grpcRouteWithPortMatch(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.0.weighted_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.0.weighted_target.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.metadata.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.grpc_route.0.match.0.metadata.*", map[string]string{
						"invert":       acctest.CtFalse,
						"match.#":      acctest.Ct0,
						names.AttrName: "X-Testing1",
					}),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.method_name", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.service_name", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.grpc_retry_events.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.grpc_retry_events.*", "deadline-exceeded"),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.grpc_retry_events.*", "resource-exhausted"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.http_retry_events.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.http_retry_events.*", "server-error"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.max_retries", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.per_retry_timeout.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.per_retry_timeout.0.unit", "s"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.per_retry_timeout.0.value", "15"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.tcp_retry_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.timeout.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				Config: testAccRouteConfig_grpcRouteWithPortMatchUpdated(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.0.weighted_target.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.0.weighted_target.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.0.weighted_target.1.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.metadata.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.grpc_route.0.match.0.metadata.*", map[string]string{
						"invert":       acctest.CtTrue,
						"match.#":      acctest.Ct0,
						names.AttrName: "X-Testing1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.grpc_route.0.match.0.metadata.*", map[string]string{
						"invert":                acctest.CtFalse,
						"match.#":               acctest.Ct1,
						"match.0.range.#":       acctest.Ct1,
						"match.0.range.0.end":   "7",
						"match.0.range.0.start": acctest.Ct2,
						names.AttrName:          "X-Testing2",
					}),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.method_name", "test"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.service_name", "test.local"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.grpc_retry_events.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.grpc_retry_events.*", "cancelled"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.http_retry_events.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.http_retry_events.*", "gateway-error"),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.http_retry_events.*", "client-error"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.max_retries", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.per_retry_timeout.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.per_retry_timeout.0.unit", "ms"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.per_retry_timeout.0.value", "250000"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.tcp_retry_events.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.tcp_retry_events.*", "connection-error"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.timeout.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				Config: testAccRouteConfig_grpcRouteUpdatedWithZeroWeight(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.0.weighted_target.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.metadata.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.grpc_route.0.match.0.metadata.*", map[string]string{
						"invert":       acctest.CtTrue,
						"match.#":      acctest.Ct0,
						names.AttrName: "X-Testing1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.grpc_route.0.match.0.metadata.*", map[string]string{
						"invert":                acctest.CtFalse,
						"match.#":               acctest.Ct1,
						"match.0.range.#":       acctest.Ct1,
						"match.0.range.0.end":   "7",
						"match.0.range.0.start": acctest.Ct2,
						names.AttrName:          "X-Testing2",
					}),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.method_name", "test"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.service_name", "test.local"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.grpc_retry_events.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.grpc_retry_events.*", "cancelled"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.http_retry_events.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.http_retry_events.*", "gateway-error"),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.http_retry_events.*", "client-error"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.max_retries", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.per_retry_timeout.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.per_retry_timeout.0.unit", "ms"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.per_retry_timeout.0.value", "250000"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.tcp_retry_events.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.tcp_retry_events.*", "connection-error"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.timeout.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				Config: testAccRouteConfig_grpcRouteWithMaxRetriesZero(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.0.weighted_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.metadata.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.grpc_route.0.match.0.metadata.*", map[string]string{
						"invert":       acctest.CtFalse,
						"match.#":      acctest.Ct0,
						names.AttrName: "X-Testing1",
					}),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.method_name", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.service_name", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.grpc_retry_events.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.grpc_retry_events.*", "deadline-exceeded"),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.grpc_retry_events.*", "resource-exhausted"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.http_retry_events.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.http_retry_events.*", "server-error"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.max_retries", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.per_retry_timeout.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.per_retry_timeout.0.unit", "s"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.per_retry_timeout.0.value", "15"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.0.tcp_retry_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.timeout.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
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
	ctx := acctest.Context(t)
	var r appmesh.RouteData
	resourceName := "aws_appmesh_route.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vrName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn1Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn2Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRouteConfig_grpcRouteWithTimeout(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.0.weighted_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.metadata.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.grpc_route.0.match.0.metadata.*", map[string]string{
						"invert":       acctest.CtFalse,
						"match.#":      acctest.Ct0,
						names.AttrName: "X-Testing1",
					}),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.method_name", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.service_name", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.timeout.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.timeout.0.idle.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.timeout.0.idle.0.unit", "ms"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.timeout.0.idle.0.value", "250000"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.timeout.0.per_request.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				Config: testAccRouteConfig_grpcRouteWithTimeoutUpdated(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.0.weighted_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.metadata.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.grpc_route.0.match.0.metadata.*", map[string]string{
						"invert":       acctest.CtFalse,
						"match.#":      acctest.Ct0,
						names.AttrName: "X-Testing1",
					}),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.method_name", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.service_name", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.timeout.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.timeout.0.idle.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.timeout.0.idle.0.unit", "s"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.timeout.0.idle.0.value", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.timeout.0.per_request.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.timeout.0.per_request.0.unit", "s"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.timeout.0.per_request.0.value", "5"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
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
	ctx := acctest.Context(t)
	var r appmesh.RouteData
	resourceName := "aws_appmesh_route.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vrName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn1Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn2Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRouteConfig_grpcRouteWithEmptyMatch(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.0.weighted_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.metadata.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.method_name", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.service_name", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.retry_policy.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.timeout.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
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
	ctx := acctest.Context(t)
	var r appmesh.RouteData
	resourceName := "aws_appmesh_route.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vrName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn1Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn2Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRouteConfig_http2Route(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.weighted_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.header.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.http2_route.0.match.0.header.*", map[string]string{
						"invert":       acctest.CtFalse,
						"match.#":      acctest.Ct0,
						names.AttrName: "X-Testing1",
					}),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.method", "POST"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.path.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.port", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.prefix", "/"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.query_parameter.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.scheme", "http"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.retry_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.http_retry_events.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.http_retry_events.*", "server-error"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.max_retries", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.per_retry_timeout.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.per_retry_timeout.0.unit", "s"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.per_retry_timeout.0.value", "15"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.tcp_retry_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.timeout.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				Config: testAccRouteConfig_http2RouteUpdated(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.weighted_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.header.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.http2_route.0.match.0.header.*", map[string]string{
						"invert":       acctest.CtTrue,
						"match.#":      acctest.Ct0,
						names.AttrName: "X-Testing1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.http2_route.0.match.0.header.*", map[string]string{
						"invert":                acctest.CtFalse,
						"match.#":               acctest.Ct1,
						"match.0.range.#":       acctest.Ct1,
						"match.0.range.0.end":   "7",
						"match.0.range.0.start": acctest.Ct2,
						names.AttrName:          "X-Testing2",
					}),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.method", "PUT"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.prefix", "/path"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.scheme", "https"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.retry_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.http_retry_events.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.http_retry_events.*", "gateway-error"),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.http_retry_events.*", "client-error"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.max_retries", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.per_retry_timeout.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.per_retry_timeout.0.unit", "ms"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.per_retry_timeout.0.value", "250000"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.tcp_retry_events.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.tcp_retry_events.*", "connection-error"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.timeout.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
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

func testAccRoute_http2RouteWithPathMatch(t *testing.T) {
	ctx := acctest.Context(t)
	var r appmesh.RouteData
	resourceName := "aws_appmesh_route.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vrName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn1Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn2Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRouteConfig_http2RouteWithPathMatch(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.weighted_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.weighted_target.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.method", "POST"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.scheme", "http"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.path.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.path.0.exact", "/test"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.path.0.regex", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.port", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.query_parameter.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.retry_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.http_retry_events.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.http_retry_events.*", "server-error"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.max_retries", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.per_retry_timeout.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.per_retry_timeout.0.unit", "s"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.per_retry_timeout.0.value", "15"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.tcp_retry_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.timeout.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
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

func testAccRoute_http2RouteWithPortMatch(t *testing.T) {
	ctx := acctest.Context(t)
	var r appmesh.RouteData
	resourceName := "aws_appmesh_route.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vrName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn1Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn2Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRouteConfig_http2RouteWithPortMatch(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.weighted_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.weighted_target.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.header.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.http2_route.0.match.0.header.*", map[string]string{
						"invert":       acctest.CtFalse,
						"match.#":      acctest.Ct0,
						names.AttrName: "X-Testing1",
					}),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.method", "POST"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.prefix", "/"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.scheme", "http"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.retry_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.http_retry_events.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.http_retry_events.*", "server-error"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.max_retries", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.per_retry_timeout.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.per_retry_timeout.0.unit", "s"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.per_retry_timeout.0.value", "15"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.tcp_retry_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.timeout.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				Config: testAccRouteConfig_http2RouteWithPortMatchUpdated(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.weighted_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.weighted_target.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.header.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.http2_route.0.match.0.header.*", map[string]string{
						"invert":       acctest.CtTrue,
						"match.#":      acctest.Ct0,
						names.AttrName: "X-Testing1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.http2_route.0.match.0.header.*", map[string]string{
						"invert":                acctest.CtFalse,
						"match.#":               acctest.Ct1,
						"match.0.range.#":       acctest.Ct1,
						"match.0.range.0.end":   "7",
						"match.0.range.0.start": acctest.Ct2,
						names.AttrName:          "X-Testing2",
					}),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.method", "PUT"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.prefix", "/path"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.scheme", "https"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.retry_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.http_retry_events.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.http_retry_events.*", "gateway-error"),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.http_retry_events.*", "client-error"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.max_retries", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.per_retry_timeout.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.per_retry_timeout.0.unit", "ms"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.per_retry_timeout.0.value", "250000"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.tcp_retry_events.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.http2_route.0.retry_policy.0.tcp_retry_events.*", "connection-error"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.timeout.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
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
	ctx := acctest.Context(t)
	var r appmesh.RouteData
	resourceName := "aws_appmesh_route.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vrName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn1Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn2Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRouteConfig_http2RouteWithTimeout(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.weighted_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.header.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.http2_route.0.match.0.header.*", map[string]string{
						"invert":       acctest.CtFalse,
						"match.#":      acctest.Ct0,
						names.AttrName: "X-Testing1",
					}),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.method", "POST"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.prefix", "/"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.scheme", "http"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.retry_policy.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.timeout.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.timeout.0.idle.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.timeout.0.idle.0.unit", "ms"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.timeout.0.idle.0.value", "250000"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.timeout.0.per_request.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				Config: testAccRouteConfig_http2RouteWithTimeoutUpdated(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.weighted_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.header.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.http2_route.0.match.0.header.*", map[string]string{
						"invert":       acctest.CtFalse,
						"match.#":      acctest.Ct0,
						names.AttrName: "X-Testing1",
					}),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.method", "POST"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.prefix", "/"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.scheme", "http"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.retry_policy.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.timeout.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.timeout.0.idle.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.timeout.0.idle.0.unit", "s"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.timeout.0.idle.0.value", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.timeout.0.per_request.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.timeout.0.per_request.0.unit", "s"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.timeout.0.per_request.0.value", "5"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
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
	ctx := acctest.Context(t)
	var r appmesh.RouteData
	resourceName := "aws_appmesh_route.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vrName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn1Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn2Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRouteConfig_http(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.weighted_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.method", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.path.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.port", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.prefix", "/"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.query_parameter.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.scheme", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.timeout.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.priority", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				Config: testAccRouteConfig_httpUpdated(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.weighted_target.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.method", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.prefix", "/path"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.query_parameter.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.scheme", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.timeout.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.priority", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				Config: testAccRouteConfig_httpUpdatedZeroWeight(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.weighted_target.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.method", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.prefix", "/path"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.query_parameter.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.scheme", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.timeout.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.priority", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
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

func testAccRoute_httpRouteWithPortMatch(t *testing.T) {
	ctx := acctest.Context(t)
	var r appmesh.RouteData
	resourceName := "aws_appmesh_route.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vrName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn1Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn2Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRouteConfig_httpWithPortMatch(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.weighted_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.weighted_target.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.method", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.prefix", "/"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.scheme", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.timeout.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.priority", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				Config: testAccRouteConfig_httpWithPortMatchUpdated(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.weighted_target.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.weighted_target.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.weighted_target.1.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.method", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.prefix", "/path"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.scheme", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.timeout.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.priority", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				Config: testAccRouteConfig_httpUpdatedZeroWeight(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.weighted_target.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.method", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.prefix", "/path"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.scheme", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.timeout.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.priority", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
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

func testAccRoute_httpRouteWithQueryParameterMatch(t *testing.T) {
	ctx := acctest.Context(t)
	var r appmesh.RouteData
	resourceName := "aws_appmesh_route.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vrName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn1Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn2Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRouteConfig_httpWithQueryParameterMatch(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.weighted_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.weighted_target.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.method", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.prefix", "/"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.scheme", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.port", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.query_parameter.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.http_route.0.match.0.query_parameter.*", map[string]string{
						"match.#":       acctest.Ct1,
						"match.0.exact": "xact",
						names.AttrName:  "param1",
					}),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.timeout.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.priority", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
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
	ctx := acctest.Context(t)
	var r appmesh.RouteData
	resourceName := "aws_appmesh_route.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vrName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn1Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn2Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRouteConfig_httpTimeout(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.weighted_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.method", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.prefix", "/"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.scheme", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.timeout.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.timeout.0.idle.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.timeout.0.idle.0.unit", "ms"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.timeout.0.idle.0.value", "250000"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.timeout.0.per_request.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.priority", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				Config: testAccRouteConfig_httpTimeoutUpdated(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.weighted_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.method", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.prefix", "/"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.scheme", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.timeout.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.timeout.0.idle.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.timeout.0.idle.0.unit", "s"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.timeout.0.idle.0.value", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.timeout.0.per_request.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.timeout.0.per_request.0.unit", "s"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.timeout.0.per_request.0.value", "5"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.priority", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
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
	ctx := acctest.Context(t)
	var r appmesh.RouteData
	resourceName := "aws_appmesh_route.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vrName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn1Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn2Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRouteConfig_tcp(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.priority", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.action.0.weighted_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.timeout.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				Config: testAccRouteConfig_tcpUpdated(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.priority", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.action.0.weighted_target.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.timeout.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				Config: testAccRouteConfig_tcpUpdatedZeroWeight(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.priority", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.action.0.weighted_target.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.timeout.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
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

func testAccRoute_tcpRouteWithPortMatch(t *testing.T) {
	ctx := acctest.Context(t)
	var r appmesh.RouteData
	resourceName := "aws_appmesh_route.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vrName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn1Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn2Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRouteConfig_tcpWithPortMatch(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.priority", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.action.0.weighted_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.action.0.weighted_target.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.match.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.timeout.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				Config: testAccRouteConfig_tcpWithPortMatchUpdated(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.priority", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.action.0.weighted_target.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.action.0.weighted_target.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.action.0.weighted_target.1.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.match.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.timeout.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				Config: testAccRouteConfig_tcpUpdatedZeroWeight(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.priority", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.action.0.weighted_target.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.timeout.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
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
	ctx := acctest.Context(t)
	var r appmesh.RouteData
	resourceName := "aws_appmesh_route.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vrName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn1Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn2Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRouteConfig_tcpTimeout(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.priority", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.action.0.weighted_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.timeout.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.timeout.0.idle.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.timeout.0.idle.0.unit", "ms"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.timeout.0.idle.0.value", "250000"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				Config: testAccRouteConfig_tcpTimeoutUpdated(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.priority", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.action.0.weighted_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.timeout.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.timeout.0.idle.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.timeout.0.idle.0.unit", "s"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.0.timeout.0.idle.0.value", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
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
	ctx := acctest.Context(t)
	var r appmesh.RouteData
	resourceName := "aws_appmesh_route.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vrName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn1Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn2Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRouteConfig_httpHeader(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.weighted_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.http_route.0.match.0.header.*", map[string]string{
						"invert":       acctest.CtFalse,
						"match.#":      acctest.Ct0,
						names.AttrName: "X-Testing1",
					}),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.method", "POST"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.prefix", "/"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.priority", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				Config: testAccRouteConfig_httpHeaderUpdated(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.weighted_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.http_route.0.match.0.header.*", map[string]string{
						"invert":       acctest.CtTrue,
						"match.#":      acctest.Ct0,
						names.AttrName: "X-Testing1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.http_route.0.match.0.header.*", map[string]string{
						"invert":                acctest.CtFalse,
						"match.#":               acctest.Ct1,
						"match.0.range.#":       acctest.Ct1,
						"match.0.range.0.end":   "7",
						"match.0.range.0.start": acctest.Ct2,
						names.AttrName:          "X-Testing2",
					}),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.method", "PUT"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.prefix", "/path"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.priority", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
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
	ctx := acctest.Context(t)
	var r appmesh.RouteData
	resourceName := "aws_appmesh_route.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vrName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn1Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn2Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRouteConfig_routePriority(meshName, vrName, vn1Name, vn2Name, rName, 42),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.weighted_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.method", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.prefix", "/"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.scheme", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.priority", "42"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				Config: testAccRouteConfig_routePriority(meshName, vrName, vn1Name, vn2Name, rName, 1000),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.weighted_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.method", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.prefix", "/"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.scheme", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.priority", "1000"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
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
	ctx := acctest.Context(t)
	var r appmesh.RouteData
	resourceName := "aws_appmesh_route.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vrName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn1Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn2Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRouteConfig_httpRetryPolicy(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.weighted_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.method", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.prefix", "/"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.scheme", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.0.http_retry_events.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.http_route.0.retry_policy.0.http_retry_events.*", "server-error"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.0.max_retries", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.0.per_retry_timeout.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.0.per_retry_timeout.0.unit", "s"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.0.per_retry_timeout.0.value", "15"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.0.tcp_retry_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				Config: testAccRouteConfig_httpMaxRetriesZero(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.weighted_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.method", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.prefix", "/"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.scheme", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.0.http_retry_events.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.http_route.0.retry_policy.0.http_retry_events.*", "server-error"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.0.max_retries", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.0.per_retry_timeout.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.0.per_retry_timeout.0.unit", "s"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.0.per_retry_timeout.0.value", "15"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.0.tcp_retry_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
				),
			},
			{
				Config: testAccRouteConfig_httpRetryPolicyUpdated(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "virtual_router_name", vrName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.weighted_target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.method", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.prefix", "/"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.scheme", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.0.http_retry_events.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.http_route.0.retry_policy.0.http_retry_events.*", "gateway-error"),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.http_route.0.retry_policy.0.http_retry_events.*", "client-error"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.0.max_retries", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.0.per_retry_timeout.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.0.per_retry_timeout.0.unit", "ms"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.0.per_retry_timeout.0.value", "250000"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.retry_policy.0.tcp_retry_events.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.http_route.0.retry_policy.0.tcp_retry_events.*", "connection-error"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.tcp_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s/route/%s", meshName, vrName, rName)),
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

func testAccRoute_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var r appmesh.RouteData
	resourceName := "aws_appmesh_route.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vrName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn1Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vn2Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRouteConfig_grpcRoute(meshName, vrName, vn1Name, vn2Name, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &r),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfappmesh.ResourceRoute(), resourceName),
				),
				ExpectNonEmptyPlan: true,
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

		return fmt.Sprintf("%s/%s/%s", rs.Primary.Attributes["mesh_name"], rs.Primary.Attributes["virtual_router_name"], rs.Primary.Attributes[names.AttrName]), nil
	}
}

func testAccCheckRouteDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppMeshConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appmesh_route" {
				continue
			}

			_, err := tfappmesh.FindRouteByFourPartKey(ctx, conn, rs.Primary.Attributes["mesh_name"], rs.Primary.Attributes["mesh_owner"], rs.Primary.Attributes["virtual_router_name"], rs.Primary.Attributes[names.AttrName])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("App Mesh Route %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckRouteExists(ctx context.Context, n string, v *appmesh.RouteData) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppMeshConn(ctx)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No App Mesh Route ID is set")
		}

		output, err := tfappmesh.FindRouteByFourPartKey(ctx, conn, rs.Primary.Attributes["mesh_name"], rs.Primary.Attributes["mesh_owner"], rs.Primary.Attributes["virtual_router_name"], rs.Primary.Attributes[names.AttrName])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccRouteConfig_base(meshName, vrName, vrProtocol, vn1Name, vn2Name string) string {
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

resource "aws_appmesh_virtual_node" "test1" {
  name      = %[4]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    listener {
      port_mapping {
        port     = 8080
        protocol = %[3]q
      }
    }

    service_discovery {
      dns {
        hostname = "test1.simpleapp.local"
      }
    }
  }
}

resource "aws_appmesh_virtual_node" "test2" {
  name      = %[5]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    listener {
      port_mapping {
        port     = 8080
        protocol = %[3]q
      }
    }

    service_discovery {
      dns {
        hostname = "test2.simpleapp.local"
      }
    }
  }
}
`, meshName, vrName, vrProtocol, vn1Name, vn2Name)
}

func testAccRouteConfig_grpcRoute(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfig_base(meshName, vrName, "grpc", vn1Name, vn2Name), fmt.Sprintf(`
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
          virtual_node = aws_appmesh_virtual_node.test1.name
          weight       = 100
        }
      }
    }
  }
}
`, rName))
}

func testAccRouteConfig_grpcRouteUpdated(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfig_base(meshName, vrName, "grpc", vn1Name, vn2Name), fmt.Sprintf(`
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
          virtual_node = aws_appmesh_virtual_node.test1.name
          weight       = 90
        }

        weighted_target {
          virtual_node = aws_appmesh_virtual_node.test2.name
          weight       = 10
        }
      }
    }
  }
}
`, rName))
}

func testAccRouteConfig_grpcRouteWithPortMatch(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfig_base(meshName, vrName, "grpc", vn1Name, vn2Name), fmt.Sprintf(`
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
        port = 8080
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
          virtual_node = aws_appmesh_virtual_node.test1.name
          weight       = 100
          port         = 8080
        }
      }
    }
  }
}
`, rName))
}

func testAccRouteConfig_grpcRouteWithPortMatchUpdated(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfig_base(meshName, vrName, "grpc", vn1Name, vn2Name), fmt.Sprintf(`
resource "aws_appmesh_route" "test" {
  name                = %[1]q
  mesh_name           = aws_appmesh_mesh.test.id
  virtual_router_name = aws_appmesh_virtual_router.test.name

  spec {
    grpc_route {
      match {
        method_name  = "test"
        service_name = "test.local"
        port         = 8080
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
          virtual_node = aws_appmesh_virtual_node.test1.name
          weight       = 90
          port         = 8080
        }

        weighted_target {
          virtual_node = aws_appmesh_virtual_node.test2.name
          weight       = 10
          port         = 8080
        }
      }
    }
  }
}
`, rName))
}

func testAccRouteConfig_grpcRouteUpdatedWithZeroWeight(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfig_base(meshName, vrName, "grpc", vn1Name, vn2Name), fmt.Sprintf(`
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
          virtual_node = aws_appmesh_virtual_node.test1.name
          weight       = 99
        }

        weighted_target {
          virtual_node = aws_appmesh_virtual_node.test2.name
          weight       = 0
        }
      }
    }
  }
}
`, rName))
}

func testAccRouteConfig_grpcRouteWithTimeout(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfig_base(meshName, vrName, "grpc", vn1Name, vn2Name), fmt.Sprintf(`
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
          virtual_node = aws_appmesh_virtual_node.test1.name
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
	return acctest.ConfigCompose(testAccRouteConfig_base(meshName, vrName, "grpc", vn1Name, vn2Name), fmt.Sprintf(`
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
          virtual_node = aws_appmesh_virtual_node.test1.name
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
	return acctest.ConfigCompose(testAccRouteConfig_base(meshName, vrName, "grpc", vn1Name, vn2Name), fmt.Sprintf(`
resource "aws_appmesh_route" "test" {
  name                = %[1]q
  mesh_name           = aws_appmesh_mesh.test.id
  virtual_router_name = aws_appmesh_virtual_router.test.name

  spec {
    grpc_route {
      match {}

      action {
        weighted_target {
          virtual_node = aws_appmesh_virtual_node.test1.name
          weight       = 100
        }
      }
    }
  }
}
`, rName))
}

func testAccRouteConfig_grpcRouteWithMaxRetriesZero(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfig_base(meshName, vrName, "grpc", vn1Name, vn2Name), fmt.Sprintf(`
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
          virtual_node = aws_appmesh_virtual_node.test1.name
          weight       = 100
        }
      }
    }
  }
}
`, rName))
}

func testAccRouteConfig_http2Route(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfig_base(meshName, vrName, "http2", vn1Name, vn2Name), fmt.Sprintf(`
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
          virtual_node = aws_appmesh_virtual_node.test1.name
          weight       = 100
        }
      }
    }
  }
}
`, rName))
}

func testAccRouteConfig_http2RouteUpdated(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfig_base(meshName, vrName, "http2", vn1Name, vn2Name), fmt.Sprintf(`
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
          virtual_node = aws_appmesh_virtual_node.test1.name
          weight       = 100
        }
      }
    }
  }
}
`, rName))
}

func testAccRouteConfig_http2RouteWithPathMatch(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfig_base(meshName, vrName, "http2", vn1Name, vn2Name), fmt.Sprintf(`
resource "aws_appmesh_route" "test" {
  name                = %[1]q
  mesh_name           = aws_appmesh_mesh.test.id
  virtual_router_name = aws_appmesh_virtual_router.test.name

  spec {
    http2_route {
      match {
        method = "POST"
        scheme = "http"

        path {
          exact = "/test"
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
          virtual_node = aws_appmesh_virtual_node.test1.name
          weight       = 100
          port         = 8080
        }
      }
    }
  }
}
`, rName))
}

func testAccRouteConfig_http2RouteWithPortMatch(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfig_base(meshName, vrName, "http2", vn1Name, vn2Name), fmt.Sprintf(`
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
        port   = 8080
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
          virtual_node = aws_appmesh_virtual_node.test1.name
          weight       = 100
          port         = 8080
        }
      }
    }
  }
}
`, rName))
}

func testAccRouteConfig_http2RouteWithPortMatchUpdated(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfig_base(meshName, vrName, "http2", vn1Name, vn2Name), fmt.Sprintf(`
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
        port   = 8080
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
          virtual_node = aws_appmesh_virtual_node.test1.name
          weight       = 100
          port         = 8080
        }
      }
    }
  }
}
`, rName))
}

func testAccRouteConfig_http2RouteWithTimeout(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfig_base(meshName, vrName, "http2", vn1Name, vn2Name), fmt.Sprintf(`
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
          virtual_node = aws_appmesh_virtual_node.test1.name
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
	return acctest.ConfigCompose(testAccRouteConfig_base(meshName, vrName, "http2", vn1Name, vn2Name), fmt.Sprintf(`
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
          virtual_node = aws_appmesh_virtual_node.test1.name
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
	return acctest.ConfigCompose(testAccRouteConfig_base(meshName, vrName, "http", vn1Name, vn2Name), fmt.Sprintf(`
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
          virtual_node = aws_appmesh_virtual_node.test1.name
          weight       = 100
        }
      }
    }
  }
}
`, rName))
}

func testAccRouteConfig_httpUpdated(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfig_base(meshName, vrName, "http", vn1Name, vn2Name), fmt.Sprintf(`
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
          virtual_node = aws_appmesh_virtual_node.test1.name
          weight       = 90
        }

        weighted_target {
          virtual_node = aws_appmesh_virtual_node.test2.name
          weight       = 10
        }
      }
    }
  }
}
`, rName))
}

func testAccRouteConfig_httpWithPortMatch(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfig_base(meshName, vrName, "http", vn1Name, vn2Name), fmt.Sprintf(`
resource "aws_appmesh_route" "test" {
  name                = %[1]q
  mesh_name           = aws_appmesh_mesh.test.id
  virtual_router_name = aws_appmesh_virtual_router.test.name

  spec {
    http_route {
      match {
        prefix = "/"
        port   = 8080
      }

      action {
        weighted_target {
          virtual_node = aws_appmesh_virtual_node.test1.name
          weight       = 100
          port         = 8080
        }
      }
    }
  }
}
`, rName))
}

func testAccRouteConfig_httpWithPortMatchUpdated(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfig_base(meshName, vrName, "http", vn1Name, vn2Name), fmt.Sprintf(`
resource "aws_appmesh_route" "test" {
  name                = %[1]q
  mesh_name           = aws_appmesh_mesh.test.id
  virtual_router_name = aws_appmesh_virtual_router.test.name

  spec {
    http_route {
      match {
        prefix = "/path"
        port   = 8080
      }

      action {
        weighted_target {
          virtual_node = aws_appmesh_virtual_node.test1.name
          weight       = 90
          port         = 8080
        }

        weighted_target {
          virtual_node = aws_appmesh_virtual_node.test2.name
          weight       = 10
          port         = 8080
        }
      }
    }
  }
}
`, rName))
}

func testAccRouteConfig_httpWithQueryParameterMatch(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfig_base(meshName, vrName, "http", vn1Name, vn2Name), fmt.Sprintf(`
resource "aws_appmesh_route" "test" {
  name                = %[1]q
  mesh_name           = aws_appmesh_mesh.test.id
  virtual_router_name = aws_appmesh_virtual_router.test.name

  spec {
    http_route {
      match {
        prefix = "/"

        query_parameter {
          name = "param1"

          match {
            exact = "xact"
          }
        }
      }

      action {
        weighted_target {
          virtual_node = aws_appmesh_virtual_node.test1.name
          weight       = 100
          port         = 8080
        }
      }
    }
  }
}
`, rName))
}

func testAccRouteConfig_httpUpdatedZeroWeight(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfig_base(meshName, vrName, "http", vn1Name, vn2Name), fmt.Sprintf(`
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
          virtual_node = aws_appmesh_virtual_node.test1.name
          weight       = 99
        }

        weighted_target {
          virtual_node = aws_appmesh_virtual_node.test2.name
          weight       = 0
        }
      }
    }
  }
}
`, rName))
}

func testAccRouteConfig_httpTimeout(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfig_base(meshName, vrName, "http", vn1Name, vn2Name), fmt.Sprintf(`
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
          virtual_node = aws_appmesh_virtual_node.test1.name
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
	return acctest.ConfigCompose(testAccRouteConfig_base(meshName, vrName, "http", vn1Name, vn2Name), fmt.Sprintf(`
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
          virtual_node = aws_appmesh_virtual_node.test1.name
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
	return acctest.ConfigCompose(testAccRouteConfig_base(meshName, vrName, "tcp", vn1Name, vn2Name), fmt.Sprintf(`
resource "aws_appmesh_route" "test" {
  name                = %[1]q
  mesh_name           = aws_appmesh_mesh.test.id
  virtual_router_name = aws_appmesh_virtual_router.test.name

  spec {
    tcp_route {
      action {
        weighted_target {
          virtual_node = aws_appmesh_virtual_node.test1.name
          weight       = 100
        }
      }
    }
  }
}
`, rName))
}

func testAccRouteConfig_tcpUpdated(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfig_base(meshName, vrName, "tcp", vn1Name, vn2Name), fmt.Sprintf(`
resource "aws_appmesh_route" "test" {
  name                = %[1]q
  mesh_name           = aws_appmesh_mesh.test.id
  virtual_router_name = aws_appmesh_virtual_router.test.name

  spec {
    tcp_route {
      action {
        weighted_target {
          virtual_node = aws_appmesh_virtual_node.test1.name
          weight       = 90
        }

        weighted_target {
          virtual_node = aws_appmesh_virtual_node.test2.name
          weight       = 10
        }
      }
    }
  }
}
`, rName))
}

func testAccRouteConfig_tcpWithPortMatch(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfig_base(meshName, vrName, "tcp", vn1Name, vn2Name), fmt.Sprintf(`
resource "aws_appmesh_route" "test" {
  name                = %[1]q
  mesh_name           = aws_appmesh_mesh.test.id
  virtual_router_name = aws_appmesh_virtual_router.test.name

  spec {
    tcp_route {
      action {
        weighted_target {
          virtual_node = aws_appmesh_virtual_node.test1.name
          weight       = 100
          port         = 8080
        }
      }
      match {
        port = 8080
      }
    }
  }
}
`, rName))
}

func testAccRouteConfig_tcpWithPortMatchUpdated(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfig_base(meshName, vrName, "tcp", vn1Name, vn2Name), fmt.Sprintf(`
resource "aws_appmesh_route" "test" {
  name                = %[1]q
  mesh_name           = aws_appmesh_mesh.test.id
  virtual_router_name = aws_appmesh_virtual_router.test.name

  spec {
    tcp_route {
      action {
        weighted_target {
          virtual_node = aws_appmesh_virtual_node.test1.name
          weight       = 90
          port         = 8080
        }

        weighted_target {
          virtual_node = aws_appmesh_virtual_node.test2.name
          weight       = 10
          port         = 8080
        }
      }
      match {
        port = 8080
      }
    }
  }
}
`, rName))
}

func testAccRouteConfig_tcpUpdatedZeroWeight(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfig_base(meshName, vrName, "tcp", vn1Name, vn2Name), fmt.Sprintf(`
resource "aws_appmesh_route" "test" {
  name                = %[1]q
  mesh_name           = aws_appmesh_mesh.test.id
  virtual_router_name = aws_appmesh_virtual_router.test.name

  spec {
    tcp_route {
      action {
        weighted_target {
          virtual_node = aws_appmesh_virtual_node.test1.name
          weight       = 99
        }

        weighted_target {
          virtual_node = aws_appmesh_virtual_node.test2.name
          weight       = 0
        }
      }
    }
  }
}
`, rName))
}

func testAccRouteConfig_tcpTimeout(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfig_base(meshName, vrName, "tcp", vn1Name, vn2Name), fmt.Sprintf(`
resource "aws_appmesh_route" "test" {
  name                = %[1]q
  mesh_name           = aws_appmesh_mesh.test.id
  virtual_router_name = aws_appmesh_virtual_router.test.name

  spec {
    tcp_route {
      action {
        weighted_target {
          virtual_node = aws_appmesh_virtual_node.test1.name
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
	return acctest.ConfigCompose(testAccRouteConfig_base(meshName, vrName, "tcp", vn1Name, vn2Name), fmt.Sprintf(`
resource "aws_appmesh_route" "test" {
  name                = %[1]q
  mesh_name           = aws_appmesh_mesh.test.id
  virtual_router_name = aws_appmesh_virtual_router.test.name

  spec {
    tcp_route {
      action {
        weighted_target {
          virtual_node = aws_appmesh_virtual_node.test1.name
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

func testAccRouteConfig_httpHeader(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfig_base(meshName, vrName, "http", vn1Name, vn2Name), fmt.Sprintf(`
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
          virtual_node = aws_appmesh_virtual_node.test1.name
          weight       = 100
        }
      }
    }
  }
}
`, rName))
}

func testAccRouteConfig_httpHeaderUpdated(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfig_base(meshName, vrName, "http", vn1Name, vn2Name), fmt.Sprintf(`
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
          virtual_node = aws_appmesh_virtual_node.test1.name
          weight       = 100
        }
      }
    }
  }
}
`, rName))
}

func testAccRouteConfig_routePriority(meshName, vrName, vn1Name, vn2Name, rName string, priority int) string {
	return acctest.ConfigCompose(testAccRouteConfig_base(meshName, vrName, "http", vn1Name, vn2Name), fmt.Sprintf(`
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
          virtual_node = aws_appmesh_virtual_node.test1.name
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
	return acctest.ConfigCompose(testAccRouteConfig_base(meshName, vrName, "http", vn1Name, vn2Name), fmt.Sprintf(`
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
          virtual_node = aws_appmesh_virtual_node.test1.name
          weight       = 100
        }
      }
    }
  }
}
`, rName))
}

func testAccRouteConfig_httpMaxRetriesZero(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfig_base(meshName, vrName, "http", vn1Name, vn2Name), fmt.Sprintf(`
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
          virtual_node = aws_appmesh_virtual_node.test1.name
          weight       = 100
        }
      }
    }
  }
}
`, rName))
}

func testAccRouteConfig_httpRetryPolicyUpdated(meshName, vrName, vn1Name, vn2Name, rName string) string {
	return acctest.ConfigCompose(testAccRouteConfig_base(meshName, vrName, "http", vn1Name, vn2Name), fmt.Sprintf(`
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
          virtual_node = aws_appmesh_virtual_node.test1.name
          weight       = 100
        }
      }
    }
  }
}
`, rName))
}
