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

func testAccGatewayRoute_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v appmesh.GatewayRouteData
	resourceName := "aws_appmesh_gateway_route.test"
	vsResourceName := "aws_appmesh_virtual_service.test.0"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vgName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	grName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayRouteConfig_httpRoute(meshName, vgName, grName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayRouteExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, grName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.target.0.virtual_service.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.http_route.0.action.0.target.0.virtual_service.0.virtual_service_name", vsResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.hostname.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.path.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.port", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.prefix", "/"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.query_parameter.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.priority", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "virtual_gateway_name", vgName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s/gatewayRoute/%s", meshName, vgName, grName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccGatewayRouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccGatewayRoute_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v appmesh.GatewayRouteData
	resourceName := "aws_appmesh_gateway_route.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vgName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	grName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayRouteConfig_httpRoute(meshName, vgName, grName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayRouteExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfappmesh.ResourceGatewayRoute(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccGatewayRoute_grpcRoute(t *testing.T) {
	ctx := acctest.Context(t)
	var v appmesh.GatewayRouteData
	resourceName := "aws_appmesh_gateway_route.test"
	vs1ResourceName := "aws_appmesh_virtual_service.test.0"
	vs2ResourceName := "aws_appmesh_virtual_service.test.1"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vgName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	grName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayRouteConfig_grpcRoute(meshName, vgName, grName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayRouteExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, grName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.0.target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.0.target.0.virtual_service.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.grpc_route.0.action.0.target.0.virtual_service.0.virtual_service_name", vs1ResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.service_name", "test1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.priority", "7"),
					resource.TestCheckResourceAttr(resourceName, "virtual_gateway_name", vgName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s/gatewayRoute/%s", meshName, vgName, grName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				Config: testAccGatewayRouteConfig_grpcRouteUpdated(meshName, vgName, grName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayRouteExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, grName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.0.target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.0.target.0.virtual_service.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.grpc_route.0.action.0.target.0.virtual_service.0.virtual_service_name", vs2ResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.service_name", "test2"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.priority", "77"),
					resource.TestCheckResourceAttr(resourceName, "virtual_gateway_name", vgName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s/gatewayRoute/%s", meshName, vgName, grName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccGatewayRouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccGatewayRoute_grpcRouteWithPort(t *testing.T) {
	ctx := acctest.Context(t)
	var v appmesh.GatewayRouteData
	resourceName := "aws_appmesh_gateway_route.test"
	vs1ResourceName := "aws_appmesh_virtual_service.test.0"
	vs2ResourceName := "aws_appmesh_virtual_service.test.1"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vgName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	grName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayRouteConfig_grpcRouteWithPort(meshName, vgName, grName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayRouteExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, grName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.0.target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.0.target.0.virtual_service.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.grpc_route.0.action.0.target.0.virtual_service.0.virtual_service_name", vs1ResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.service_name", "test1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "virtual_gateway_name", vgName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s/gatewayRoute/%s", meshName, vgName, grName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				Config: testAccGatewayRouteConfig_grpcRouteWithPortUpdated(meshName, vgName, grName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayRouteExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, grName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.0.target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.0.target.0.virtual_service.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.grpc_route.0.action.0.target.0.virtual_service.0.virtual_service_name", vs2ResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.service_name", "test2"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "virtual_gateway_name", vgName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s/gatewayRoute/%s", meshName, vgName, grName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccGatewayRouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccGatewayRoute_grpcRouteTargetPort(t *testing.T) {
	ctx := acctest.Context(t)
	var v appmesh.GatewayRouteData
	resourceName := "aws_appmesh_gateway_route.test"
	vsMultiResourceName := "aws_appmesh_virtual_service.multi_test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vgName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	grName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayRouteConfig_grpcRouteTargetPort(meshName, vgName, grName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayRouteExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, grName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.0.target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.0.target.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.0.target.0.virtual_service.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.grpc_route.0.action.0.target.0.virtual_service.0.virtual_service_name", vsMultiResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.service_name", "multi-test"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "virtual_gateway_name", vgName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s/gatewayRoute/%s", meshName, vgName, grName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				Config: testAccGatewayRouteConfig_grpcRouteTargetPortUpdated(meshName, vgName, grName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayRouteExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, grName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.0.target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.0.target.0.port", "8081"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.0.target.0.virtual_service.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.grpc_route.0.action.0.target.0.virtual_service.0.virtual_service_name", vsMultiResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.service_name", "multi-test"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "virtual_gateway_name", vgName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s/gatewayRoute/%s", meshName, vgName, grName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccGatewayRouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccGatewayRoute_httpRoute(t *testing.T) {
	ctx := acctest.Context(t)
	var v appmesh.GatewayRouteData
	resourceName := "aws_appmesh_gateway_route.test"
	vs1ResourceName := "aws_appmesh_virtual_service.test.0"
	vs2ResourceName := "aws_appmesh_virtual_service.test.1"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vgName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	grName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayRouteConfig_httpRoute(meshName, vgName, grName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayRouteExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, grName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.target.0.virtual_service.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.http_route.0.action.0.target.0.virtual_service.0.virtual_service_name", vs1ResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.hostname.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.path.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.port", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.prefix", "/"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.query_parameter.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "virtual_gateway_name", vgName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s/gatewayRoute/%s", meshName, vgName, grName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				Config: testAccGatewayRouteConfig_httpRouteUpdated(meshName, vgName, grName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayRouteExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, grName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.target.0.virtual_service.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.http_route.0.action.0.target.0.virtual_service.0.virtual_service_name", vs2ResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.hostname.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.path.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.port", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.prefix", "/users"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.query_parameter.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "virtual_gateway_name", vgName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s/gatewayRoute/%s", meshName, vgName, grName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				Config: testAccGatewayRouteConfig_httpRouteMatchHostname(meshName, vgName, grName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayRouteExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, grName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.target.0.virtual_service.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.http_route.0.action.0.target.0.virtual_service.0.virtual_service_name", vs1ResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.hostname.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.hostname.0.exact", "test.example.com"),
					resource.TestCheckResourceAttr(resourceName, "virtual_gateway_name", vgName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s/gatewayRoute/%s", meshName, vgName, grName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				Config: testAccGatewayRouteConfig_httpRouteRewrite(meshName, vgName, grName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayRouteExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, grName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.target.0.virtual_service.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.http_route.0.action.0.target.0.virtual_service.0.virtual_service_name", vs1ResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.rewrite.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.rewrite.0.hostname.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.rewrite.0.hostname.0.default_target_hostname", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.rewrite.0.prefix.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.rewrite.0.prefix.0.default_prefix", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.prefix", "/"),
					resource.TestCheckResourceAttr(resourceName, "virtual_gateway_name", vgName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s/gatewayRoute/%s", meshName, vgName, grName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				Config: testAccGatewayRouteConfig_httpRouteRewriteWithPath(meshName, vgName, grName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayRouteExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, grName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.target.0.virtual_service.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.http_route.0.action.0.target.0.virtual_service.0.virtual_service_name", vs1ResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.rewrite.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.rewrite.0.hostname.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.rewrite.0.prefix.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.rewrite.0.path.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.rewrite.0.path.0.exact", "/rewrite_path"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.path.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.path.0.exact", "/match_path"),
					resource.TestCheckResourceAttr(resourceName, "virtual_gateway_name", vgName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s/gatewayRoute/%s", meshName, vgName, grName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccGatewayRouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccGatewayRoute_httpRouteTargetPort(t *testing.T) {
	ctx := acctest.Context(t)
	var v appmesh.GatewayRouteData
	resourceName := "aws_appmesh_gateway_route.test"
	vsMultiResourceName := "aws_appmesh_virtual_service.multi_test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vgName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	grName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayRouteConfig_httpRouteTargetPort(meshName, vgName, grName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayRouteExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, grName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.target.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.target.0.virtual_service.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.http_route.0.action.0.target.0.virtual_service.0.virtual_service_name", vsMultiResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.prefix", "/"),
					resource.TestCheckResourceAttr(resourceName, "virtual_gateway_name", vgName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s/gatewayRoute/%s", meshName, vgName, grName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				Config: testAccGatewayRouteConfig_httpRouteTargetPortUpdated(meshName, vgName, grName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayRouteExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, grName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.target.0.port", "8081"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.target.0.virtual_service.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.http_route.0.action.0.target.0.virtual_service.0.virtual_service_name", vsMultiResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.prefix", "/users"),
					resource.TestCheckResourceAttr(resourceName, "virtual_gateway_name", vgName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s/gatewayRoute/%s", meshName, vgName, grName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccGatewayRouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccGatewayRoute_httpRouteWithPath(t *testing.T) {
	ctx := acctest.Context(t)
	var v appmesh.GatewayRouteData
	resourceName := "aws_appmesh_gateway_route.test"
	vsResourceName := "aws_appmesh_virtual_service.test.0"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vgName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	grName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayRouteConfig_httpRouteWithPath(meshName, vgName, grName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayRouteExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, grName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.target.0.virtual_service.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.http_route.0.action.0.target.0.virtual_service.0.virtual_service_name", vsResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.header.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.hostname.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.path.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.path.0.exact", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.path.0.regex", "/.*"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.port", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.query_parameter.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "virtual_gateway_name", vgName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s/gatewayRoute/%s", meshName, vgName, grName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccGatewayRouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccGatewayRoute_httpRouteWithPort(t *testing.T) {
	ctx := acctest.Context(t)
	var v appmesh.GatewayRouteData
	resourceName := "aws_appmesh_gateway_route.test"
	vs1ResourceName := "aws_appmesh_virtual_service.test.0"
	vs2ResourceName := "aws_appmesh_virtual_service.test.1"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vgName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	grName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayRouteConfig_httpRouteWithPort(meshName, vgName, grName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayRouteExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, grName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.target.0.virtual_service.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.http_route.0.action.0.target.0.virtual_service.0.virtual_service_name", vs1ResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.prefix", "/"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "virtual_gateway_name", vgName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s/gatewayRoute/%s", meshName, vgName, grName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				Config: testAccGatewayRouteConfig_httpRouteWithPortUpdated(meshName, vgName, grName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayRouteExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, grName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.target.0.virtual_service.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.http_route.0.action.0.target.0.virtual_service.0.virtual_service_name", vs2ResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.prefix", "/users"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "virtual_gateway_name", vgName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s/gatewayRoute/%s", meshName, vgName, grName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				Config: testAccGatewayRouteConfig_httpRouteMatchHostname(meshName, vgName, grName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayRouteExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, grName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.target.0.virtual_service.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.http_route.0.action.0.target.0.virtual_service.0.virtual_service_name", vs1ResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.hostname.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.hostname.0.exact", "test.example.com"),
					resource.TestCheckResourceAttr(resourceName, "virtual_gateway_name", vgName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s/gatewayRoute/%s", meshName, vgName, grName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				Config: testAccGatewayRouteConfig_httpRouteRewrite(meshName, vgName, grName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayRouteExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, grName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.target.0.virtual_service.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.http_route.0.action.0.target.0.virtual_service.0.virtual_service_name", vs1ResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.rewrite.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.rewrite.0.hostname.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.rewrite.0.hostname.0.default_target_hostname", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.rewrite.0.prefix.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.rewrite.0.prefix.0.default_prefix", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.prefix", "/"),
					resource.TestCheckResourceAttr(resourceName, "virtual_gateway_name", vgName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s/gatewayRoute/%s", meshName, vgName, grName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				Config: testAccGatewayRouteConfig_httpRouteRewriteWithPath(meshName, vgName, grName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayRouteExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, grName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.target.0.virtual_service.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.http_route.0.action.0.target.0.virtual_service.0.virtual_service_name", vs1ResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.rewrite.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.rewrite.0.hostname.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.rewrite.0.prefix.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.rewrite.0.path.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.rewrite.0.path.0.exact", "/rewrite_path"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.path.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.path.0.exact", "/match_path"),
					resource.TestCheckResourceAttr(resourceName, "virtual_gateway_name", vgName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s/gatewayRoute/%s", meshName, vgName, grName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccGatewayRouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccGatewayRoute_http2Route(t *testing.T) {
	ctx := acctest.Context(t)
	var v appmesh.GatewayRouteData
	resourceName := "aws_appmesh_gateway_route.test"
	vs1ResourceName := "aws_appmesh_virtual_service.test.0"
	vs2ResourceName := "aws_appmesh_virtual_service.test.1"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vgName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	grName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayRouteConfig_http2Route(meshName, vgName, grName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayRouteExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, grName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.target.0.virtual_service.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.http2_route.0.action.0.target.0.virtual_service.0.virtual_service_name", vs1ResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.header.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.http2_route.0.match.0.header.*", map[string]string{
						"invert":       acctest.CtFalse,
						"match.#":      acctest.Ct0,
						names.AttrName: "X-Testing1",
					}),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.hostname.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.path.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.port", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.prefix", "/"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.query_parameter.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "virtual_gateway_name", vgName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s/gatewayRoute/%s", meshName, vgName, grName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				Config: testAccGatewayRouteConfig_http2RouteUpdated(meshName, vgName, grName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayRouteExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, grName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.target.0.virtual_service.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.http2_route.0.action.0.target.0.virtual_service.0.virtual_service_name", vs2ResourceName, names.AttrName),
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
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.hostname.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.path.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.port", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.prefix", "/users"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.query_parameter.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "virtual_gateway_name", vgName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s/gatewayRoute/%s", meshName, vgName, grName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				Config: testAccGatewayRouteConfig_http2RouteMatchHostname(meshName, vgName, grName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayRouteExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, grName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.target.0.virtual_service.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.http2_route.0.action.0.target.0.virtual_service.0.virtual_service_name", vs1ResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.hostname.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.hostname.0.exact", "test.example.com"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "virtual_gateway_name", vgName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s/gatewayRoute/%s", meshName, vgName, grName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				Config: testAccGatewayRouteConfig_http2RouteRewrite(meshName, vgName, grName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayRouteExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, grName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.target.0.virtual_service.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.http2_route.0.action.0.target.0.virtual_service.0.virtual_service_name", vs1ResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.rewrite.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.rewrite.0.hostname.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.rewrite.0.hostname.0.default_target_hostname", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.rewrite.0.prefix.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.rewrite.0.prefix.0.default_prefix", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.prefix", "/"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "virtual_gateway_name", vgName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s/gatewayRoute/%s", meshName, vgName, grName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				Config: testAccGatewayRouteConfig_http2RouteRewriteWithPath(meshName, vgName, grName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayRouteExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, grName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.target.0.virtual_service.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.http2_route.0.action.0.target.0.virtual_service.0.virtual_service_name", vs1ResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.rewrite.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.rewrite.0.hostname.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.rewrite.0.prefix.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.rewrite.0.path.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.rewrite.0.path.0.exact", "/rewrite_path"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.path.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.path.0.exact", "/match_path"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "virtual_gateway_name", vgName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s/gatewayRoute/%s", meshName, vgName, grName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccGatewayRouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccGatewayRoute_http2RouteTargetPort(t *testing.T) {
	ctx := acctest.Context(t)
	var v appmesh.GatewayRouteData
	resourceName := "aws_appmesh_gateway_route.test"
	vsMultiResourceName := "aws_appmesh_virtual_service.multi_test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vgName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	grName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayRouteConfig_http2RouteTargetPort(meshName, vgName, grName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayRouteExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, grName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.target.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.target.0.virtual_service.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.http2_route.0.action.0.target.0.virtual_service.0.virtual_service_name", vsMultiResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.prefix", "/"),
					resource.TestCheckResourceAttr(resourceName, "virtual_gateway_name", vgName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s/gatewayRoute/%s", meshName, vgName, grName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				Config: testAccGatewayRouteConfig_http2RouteTargetPortUpdated(meshName, vgName, grName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayRouteExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, grName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.target.0.port", "8081"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.target.0.virtual_service.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.http2_route.0.action.0.target.0.virtual_service.0.virtual_service_name", vsMultiResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.prefix", "/users"),
					resource.TestCheckResourceAttr(resourceName, "virtual_gateway_name", vgName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s/gatewayRoute/%s", meshName, vgName, grName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccGatewayRouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccGatewayRoute_http2RouteWithPort(t *testing.T) {
	ctx := acctest.Context(t)
	var v appmesh.GatewayRouteData
	resourceName := "aws_appmesh_gateway_route.test"
	vs1ResourceName := "aws_appmesh_virtual_service.test.0"
	vs2ResourceName := "aws_appmesh_virtual_service.test.1"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vgName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	grName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayRouteConfig_http2RouteWithPort(meshName, vgName, grName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayRouteExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, grName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.target.0.virtual_service.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.http2_route.0.action.0.target.0.virtual_service.0.virtual_service_name", vs1ResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.prefix", "/"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "virtual_gateway_name", vgName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s/gatewayRoute/%s", meshName, vgName, grName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				Config: testAccGatewayRouteConfig_http2RouteWithPortUpdated(meshName, vgName, grName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayRouteExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, grName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.target.0.virtual_service.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.http2_route.0.action.0.target.0.virtual_service.0.virtual_service_name", vs2ResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.prefix", "/users"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "virtual_gateway_name", vgName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s/gatewayRoute/%s", meshName, vgName, grName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				Config: testAccGatewayRouteConfig_http2RouteMatchHostname(meshName, vgName, grName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayRouteExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, grName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.target.0.virtual_service.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.http2_route.0.action.0.target.0.virtual_service.0.virtual_service_name", vs1ResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.hostname.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.hostname.0.exact", "test.example.com"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "virtual_gateway_name", vgName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s/gatewayRoute/%s", meshName, vgName, grName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				Config: testAccGatewayRouteConfig_http2RouteRewrite(meshName, vgName, grName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayRouteExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, grName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.target.0.virtual_service.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.http2_route.0.action.0.target.0.virtual_service.0.virtual_service_name", vs1ResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.rewrite.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.rewrite.0.hostname.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.rewrite.0.hostname.0.default_target_hostname", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.rewrite.0.prefix.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.rewrite.0.prefix.0.default_prefix", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.prefix", "/"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "virtual_gateway_name", vgName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s/gatewayRoute/%s", meshName, vgName, grName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				Config: testAccGatewayRouteConfig_http2RouteRewriteWithPath(meshName, vgName, grName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayRouteExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, grName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.target.0.virtual_service.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.http2_route.0.action.0.target.0.virtual_service.0.virtual_service_name", vs1ResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.rewrite.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.rewrite.0.hostname.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.rewrite.0.prefix.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.rewrite.0.path.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.rewrite.0.path.0.exact", "/rewrite_path"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.path.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.path.0.exact", "/match_path"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "virtual_gateway_name", vgName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s/gatewayRoute/%s", meshName, vgName, grName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccGatewayRouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccGatewayRoute_http2RouteWithQueryParameter(t *testing.T) {
	ctx := acctest.Context(t)
	var v appmesh.GatewayRouteData
	resourceName := "aws_appmesh_gateway_route.test"
	vsResourceName := "aws_appmesh_virtual_service.test.0"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vgName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	grName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayRouteConfig_http2RouteWithQueryParameter(meshName, vgName, grName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayRouteExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, grName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.target.0.virtual_service.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.http2_route.0.action.0.target.0.virtual_service.0.virtual_service_name", vsResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.header.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.hostname.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.path.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.port", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.prefix", "/"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.query_parameter.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.http2_route.0.match.0.query_parameter.*", map[string]string{
						"match.#":       acctest.Ct1,
						"match.0.exact": "xact",
						names.AttrName:  "param1",
					}),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "virtual_gateway_name", vgName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s/gatewayRoute/%s", meshName, vgName, grName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccGatewayRouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccGatewayRouteImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s/%s", rs.Primary.Attributes["mesh_name"], rs.Primary.Attributes["virtual_gateway_name"], rs.Primary.Attributes[names.AttrName]), nil
	}
}

func testAccCheckGatewayRouteDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppMeshConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appmesh_gateway_route" {
				continue
			}

			_, err := tfappmesh.FindGatewayRouteByFourPartKey(ctx, conn, rs.Primary.Attributes["mesh_name"], rs.Primary.Attributes["mesh_owner"], rs.Primary.Attributes["virtual_gateway_name"], rs.Primary.Attributes[names.AttrName])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("App Mesh Gateway Route %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckGatewayRouteExists(ctx context.Context, n string, v *appmesh.GatewayRouteData) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppMeshConn(ctx)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No App Mesh Gateway Route ID is set")
		}

		output, err := tfappmesh.FindGatewayRouteByFourPartKey(ctx, conn, rs.Primary.Attributes["mesh_name"], rs.Primary.Attributes["mesh_owner"], rs.Primary.Attributes["virtual_gateway_name"], rs.Primary.Attributes[names.AttrName])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccGatewayRouteConfig_base(meshName, vgName, protocol string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}

resource "aws_appmesh_virtual_gateway" "test" {
  name      = %[2]q
  mesh_name = aws_appmesh_mesh.test.name

  spec {
    listener {
      port_mapping {
        port     = 8080
        protocol = "%[3]s"
      }
    }
  }
}

resource "aws_appmesh_virtual_service" "test" {
  count = 2

  name      = "%[2]s-${count.index}"
  mesh_name = aws_appmesh_mesh.test.name

  spec {}
}
`, meshName, vgName, protocol)
}

func testAccGatewayRouteConfig_ServiceNodeMultipleListeners(vnName string, protocol string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_virtual_node" "test" {
  name      = "%[1]s-multi-listener-vn"
  mesh_name = aws_appmesh_mesh.test.name

  spec {
    listener {
      port_mapping {
        port     = 8080
        protocol = "%[2]s"
      }
    }

    listener {
      port_mapping {
        port     = 8081
        protocol = "%[2]s"
      }
    }

    service_discovery {
      dns {
        hostname = "serviceb.simpleapp.local"
      }
    }
  }
}

resource "aws_appmesh_virtual_service" "multi_test" {
  name      = "%[1]s-multi-listener"
  mesh_name = aws_appmesh_mesh.test.name

  spec {
    provider {
      virtual_node {
        virtual_node_name = aws_appmesh_virtual_node.test.name
      }
    }
  }
}
`, vnName, protocol)
}

func testAccGatewayRouteConfig_grpcRoute(meshName, vgName, grName string) string {
	return acctest.ConfigCompose(testAccGatewayRouteConfig_base(meshName, vgName, "grpc"), fmt.Sprintf(`
resource "aws_appmesh_gateway_route" "test" {
  name                 = %[1]q
  mesh_name            = aws_appmesh_mesh.test.name
  virtual_gateway_name = aws_appmesh_virtual_gateway.test.name

  spec {
    grpc_route {
      action {
        target {
          virtual_service {
            virtual_service_name = aws_appmesh_virtual_service.test[0].name
          }
        }
      }

      match {
        service_name = "test1"
      }
    }

    priority = 7
  }
}
`, grName))
}

func testAccGatewayRouteConfig_grpcRouteUpdated(meshName, vgName, grName string) string {
	return acctest.ConfigCompose(testAccGatewayRouteConfig_base(meshName, vgName, "grpc"), fmt.Sprintf(`
resource "aws_appmesh_gateway_route" "test" {
  name                 = %[1]q
  mesh_name            = aws_appmesh_mesh.test.name
  virtual_gateway_name = aws_appmesh_virtual_gateway.test.name

  spec {
    grpc_route {
      action {
        target {
          virtual_service {
            virtual_service_name = aws_appmesh_virtual_service.test[1].name
          }
        }
      }

      match {
        service_name = "test2"
      }
    }

    priority = 77
  }
}
`, grName))
}

func testAccGatewayRouteConfig_grpcRouteTargetPort(meshName, vgName, grName string) string {
	return acctest.ConfigCompose(
		testAccGatewayRouteConfig_base(meshName, vgName, "grpc"),
		testAccGatewayRouteConfig_ServiceNodeMultipleListeners(vgName, "grpc"),
		fmt.Sprintf(`
resource "aws_appmesh_gateway_route" "test" {
  name                 = %[1]q
  mesh_name            = aws_appmesh_mesh.test.name
  virtual_gateway_name = aws_appmesh_virtual_gateway.test.name

  spec {
    grpc_route {
      action {
        target {
          virtual_service {
            virtual_service_name = aws_appmesh_virtual_service.multi_test.name
          }
          port = 8080
        }
      }

      match {
        service_name = "multi-test"
      }
    }
  }
}
`, grName))
}

func testAccGatewayRouteConfig_grpcRouteTargetPortUpdated(meshName, vgName, grName string) string {
	return acctest.ConfigCompose(
		testAccGatewayRouteConfig_base(meshName, vgName, "grpc"),
		testAccGatewayRouteConfig_ServiceNodeMultipleListeners(vgName, "grpc"),
		fmt.Sprintf(`
resource "aws_appmesh_gateway_route" "test" {
  name                 = %[1]q
  mesh_name            = aws_appmesh_mesh.test.name
  virtual_gateway_name = aws_appmesh_virtual_gateway.test.name

  spec {
    grpc_route {
      action {
        target {
          virtual_service {
            virtual_service_name = aws_appmesh_virtual_service.multi_test.name
          }
          port = 8081
        }
      }

      match {
        service_name = "multi-test"
      }
    }
  }
}
`, grName))
}

func testAccGatewayRouteConfig_grpcRouteWithPort(meshName, vgName, grName string) string {
	return acctest.ConfigCompose(testAccGatewayRouteConfig_base(meshName, vgName, "grpc"), fmt.Sprintf(`
resource "aws_appmesh_gateway_route" "test" {
  name                 = %[1]q
  mesh_name            = aws_appmesh_mesh.test.name
  virtual_gateway_name = aws_appmesh_virtual_gateway.test.name

  spec {
    grpc_route {
      action {
        target {
          virtual_service {
            virtual_service_name = aws_appmesh_virtual_service.test[0].name
          }
        }
      }

      match {
        service_name = "test1"
        port         = 8080
      }
    }
  }
}
`, grName))
}

func testAccGatewayRouteConfig_grpcRouteWithPortUpdated(meshName, vgName, grName string) string {
	return acctest.ConfigCompose(testAccGatewayRouteConfig_base(meshName, vgName, "grpc"), fmt.Sprintf(`
resource "aws_appmesh_gateway_route" "test" {
  name                 = %[1]q
  mesh_name            = aws_appmesh_mesh.test.name
  virtual_gateway_name = aws_appmesh_virtual_gateway.test.name

  spec {
    grpc_route {
      action {
        target {
          virtual_service {
            virtual_service_name = aws_appmesh_virtual_service.test[1].name
          }
        }
      }

      match {
        service_name = "test2"
        port         = 8080
      }
    }
  }
}
`, grName))
}

func testAccGatewayRouteConfig_httpRoute(meshName, vgName, grName string) string {
	return acctest.ConfigCompose(testAccGatewayRouteConfig_base(meshName, vgName, "http"), fmt.Sprintf(`
resource "aws_appmesh_gateway_route" "test" {
  name                 = %[1]q
  mesh_name            = aws_appmesh_mesh.test.name
  virtual_gateway_name = aws_appmesh_virtual_gateway.test.name

  spec {
    http_route {
      action {
        target {
          virtual_service {
            virtual_service_name = aws_appmesh_virtual_service.test[0].name
          }
        }
      }

      match {
        prefix = "/"
      }
    }
  }
}
`, grName))
}

func testAccGatewayRouteConfig_httpRouteUpdated(meshName, vgName, grName string) string {
	return acctest.ConfigCompose(testAccGatewayRouteConfig_base(meshName, vgName, "http"), fmt.Sprintf(`
resource "aws_appmesh_gateway_route" "test" {
  name                 = %[1]q
  mesh_name            = aws_appmesh_mesh.test.name
  virtual_gateway_name = aws_appmesh_virtual_gateway.test.name

  spec {
    http_route {
      action {
        target {
          virtual_service {
            virtual_service_name = aws_appmesh_virtual_service.test[1].name
          }
        }
      }

      match {
        prefix = "/users"
      }
    }
  }
}
`, grName))
}

func testAccGatewayRouteConfig_httpRouteTargetPort(meshName, vgName, grName string) string {
	return acctest.ConfigCompose(
		testAccGatewayRouteConfig_base(meshName, vgName, "http"),
		testAccGatewayRouteConfig_ServiceNodeMultipleListeners(vgName, "http"),
		fmt.Sprintf(`
resource "aws_appmesh_gateway_route" "test" {
  name                 = %[1]q
  mesh_name            = aws_appmesh_mesh.test.name
  virtual_gateway_name = aws_appmesh_virtual_gateway.test.name

  spec {
    http_route {
      action {
        target {
          virtual_service {
            virtual_service_name = aws_appmesh_virtual_service.multi_test.name
          }
          port = 8080
        }
      }

      match {
        prefix = "/"
      }
    }
  }
}
`, grName))
}

func testAccGatewayRouteConfig_httpRouteTargetPortUpdated(meshName, vgName, grName string) string {
	return acctest.ConfigCompose(
		testAccGatewayRouteConfig_base(meshName, vgName, "http"),
		testAccGatewayRouteConfig_ServiceNodeMultipleListeners(vgName, "http"),
		fmt.Sprintf(`
resource "aws_appmesh_gateway_route" "test" {
  name                 = %[1]q
  mesh_name            = aws_appmesh_mesh.test.name
  virtual_gateway_name = aws_appmesh_virtual_gateway.test.name

  spec {
    http_route {
      action {
        target {
          virtual_service {
            virtual_service_name = aws_appmesh_virtual_service.multi_test.name
          }
          port = 8081
        }
      }

      match {
        prefix = "/users"
      }
    }
  }
}
`, grName))
}

func testAccGatewayRouteConfig_httpRouteWithPath(meshName, vgName, grName string) string {
	return acctest.ConfigCompose(testAccGatewayRouteConfig_base(meshName, vgName, "http"), fmt.Sprintf(`
resource "aws_appmesh_gateway_route" "test" {
  name                 = %[1]q
  mesh_name            = aws_appmesh_mesh.test.name
  virtual_gateway_name = aws_appmesh_virtual_gateway.test.name

  spec {
    http_route {
      action {
        target {
          virtual_service {
            virtual_service_name = aws_appmesh_virtual_service.test[0].name
          }
        }
      }

      match {
        path {
          regex = "/.*"
        }
      }
    }
  }
}
`, grName))
}

func testAccGatewayRouteConfig_httpRouteWithPort(meshName, vgName, grName string) string {
	return acctest.ConfigCompose(testAccGatewayRouteConfig_base(meshName, vgName, "http"), fmt.Sprintf(`
resource "aws_appmesh_gateway_route" "test" {
  name                 = %[1]q
  mesh_name            = aws_appmesh_mesh.test.name
  virtual_gateway_name = aws_appmesh_virtual_gateway.test.name

  spec {
    http_route {
      action {
        target {
          virtual_service {
            virtual_service_name = aws_appmesh_virtual_service.test[0].name
          }
        }
      }

      match {
        prefix = "/"
        port   = 8080
      }
    }
  }
}
`, grName))
}

func testAccGatewayRouteConfig_httpRouteWithPortUpdated(meshName, vgName, grName string) string {
	return acctest.ConfigCompose(testAccGatewayRouteConfig_base(meshName, vgName, "http"), fmt.Sprintf(`
resource "aws_appmesh_gateway_route" "test" {
  name                 = %[1]q
  mesh_name            = aws_appmesh_mesh.test.name
  virtual_gateway_name = aws_appmesh_virtual_gateway.test.name

  spec {
    http_route {
      action {
        target {
          virtual_service {
            virtual_service_name = aws_appmesh_virtual_service.test[1].name
          }
        }
      }

      match {
        prefix = "/users"
        port   = 8080
      }
    }
  }
}
`, grName))
}

func testAccGatewayRouteConfig_httpRouteMatchHostname(meshName, vgName, grName string) string {
	return acctest.ConfigCompose(testAccGatewayRouteConfig_base(meshName, vgName, "http"), fmt.Sprintf(`
resource "aws_appmesh_gateway_route" "test" {
  name                 = %[1]q
  mesh_name            = aws_appmesh_mesh.test.name
  virtual_gateway_name = aws_appmesh_virtual_gateway.test.name

  spec {
    http_route {
      action {
        target {
          virtual_service {
            virtual_service_name = aws_appmesh_virtual_service.test[0].name
          }
        }
      }

      match {
        hostname {
          exact = "test.example.com"
        }
      }
    }
  }
}
`, grName))
}

func testAccGatewayRouteConfig_httpRouteRewrite(meshName, vgName, grName string) string {
	return acctest.ConfigCompose(testAccGatewayRouteConfig_base(meshName, vgName, "http"), fmt.Sprintf(`
resource "aws_appmesh_gateway_route" "test" {
  name                 = %[1]q
  mesh_name            = aws_appmesh_mesh.test.name
  virtual_gateway_name = aws_appmesh_virtual_gateway.test.name

  spec {
    http_route {
      action {
        target {
          virtual_service {
            virtual_service_name = aws_appmesh_virtual_service.test[0].name
          }
        }
        rewrite {
          hostname {
            default_target_hostname = "DISABLED"
          }
          prefix {
            default_prefix = "DISABLED"
          }
        }
      }

      match {
        prefix = "/"
      }
    }
  }
}
`, grName))
}

func testAccGatewayRouteConfig_httpRouteRewriteWithPath(meshName, vgName, grName string) string {
	return acctest.ConfigCompose(testAccGatewayRouteConfig_base(meshName, vgName, "http"), fmt.Sprintf(`
resource "aws_appmesh_gateway_route" "test" {
  name                 = %[1]q
  mesh_name            = aws_appmesh_mesh.test.name
  virtual_gateway_name = aws_appmesh_virtual_gateway.test.name

  spec {
    http_route {
      action {
        target {
          virtual_service {
            virtual_service_name = aws_appmesh_virtual_service.test[0].name
          }
        }
        rewrite {
          path {
            exact = "/rewrite_path"
          }
        }
      }

      match {
        path {
          exact = "/match_path"
        }
      }
    }
  }
}
`, grName))
}

func testAccGatewayRouteConfig_http2Route(meshName, vgName, grName string) string {
	return acctest.ConfigCompose(testAccGatewayRouteConfig_base(meshName, vgName, "http2"), fmt.Sprintf(`
resource "aws_appmesh_gateway_route" "test" {
  name                 = %[1]q
  mesh_name            = aws_appmesh_mesh.test.name
  virtual_gateway_name = aws_appmesh_virtual_gateway.test.name

  spec {
    http2_route {
      action {
        target {
          virtual_service {
            virtual_service_name = aws_appmesh_virtual_service.test[0].name
          }
        }
      }

      match {
        prefix = "/"

        header {
          name = "X-Testing1"
        }
      }
    }
  }
}
`, grName))
}

func testAccGatewayRouteConfig_http2RouteUpdated(meshName, vgName, grName string) string {
	return acctest.ConfigCompose(testAccGatewayRouteConfig_base(meshName, vgName, "http2"), fmt.Sprintf(`
resource "aws_appmesh_gateway_route" "test" {
  name                 = %[1]q
  mesh_name            = aws_appmesh_mesh.test.name
  virtual_gateway_name = aws_appmesh_virtual_gateway.test.name

  spec {
    http2_route {
      action {
        target {
          virtual_service {
            virtual_service_name = aws_appmesh_virtual_service.test[1].name
          }
        }
      }

      match {
        prefix = "/users"

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
    }
  }
}
`, grName))
}

func testAccGatewayRouteConfig_http2RouteTargetPort(meshName, vgName, grName string) string {
	return acctest.ConfigCompose(
		testAccGatewayRouteConfig_base(meshName, vgName, "http2"),
		testAccGatewayRouteConfig_ServiceNodeMultipleListeners(vgName, "http2"),
		fmt.Sprintf(`
resource "aws_appmesh_gateway_route" "test" {
  name                 = %[1]q
  mesh_name            = aws_appmesh_mesh.test.name
  virtual_gateway_name = aws_appmesh_virtual_gateway.test.name

  spec {
    http2_route {
      action {
        target {
          virtual_service {
            virtual_service_name = aws_appmesh_virtual_service.multi_test.name
          }
          port = 8080
        }
      }

      match {
        prefix = "/"
      }
    }
  }
}
`, grName))
}

func testAccGatewayRouteConfig_http2RouteTargetPortUpdated(meshName, vgName, grName string) string {
	return acctest.ConfigCompose(
		testAccGatewayRouteConfig_base(meshName, vgName, "http2"),
		testAccGatewayRouteConfig_ServiceNodeMultipleListeners(vgName, "http2"),
		fmt.Sprintf(`
resource "aws_appmesh_gateway_route" "test" {
  name                 = %[1]q
  mesh_name            = aws_appmesh_mesh.test.name
  virtual_gateway_name = aws_appmesh_virtual_gateway.test.name

  spec {
    http2_route {
      action {
        target {
          virtual_service {
            virtual_service_name = aws_appmesh_virtual_service.multi_test.name
          }
          port = 8081
        }
      }

      match {
        prefix = "/users"
      }
    }
  }
}
`, grName))
}

func testAccGatewayRouteConfig_http2RouteWithPort(meshName, vgName, grName string) string {
	return acctest.ConfigCompose(testAccGatewayRouteConfig_base(meshName, vgName, "http2"), fmt.Sprintf(`
resource "aws_appmesh_gateway_route" "test" {
  name                 = %[1]q
  mesh_name            = aws_appmesh_mesh.test.name
  virtual_gateway_name = aws_appmesh_virtual_gateway.test.name

  spec {
    http2_route {
      action {
        target {
          virtual_service {
            virtual_service_name = aws_appmesh_virtual_service.test[0].name
          }
        }
      }

      match {
        prefix = "/"
        port   = 8080
      }
    }
  }
}
`, grName))
}

func testAccGatewayRouteConfig_http2RouteWithPortUpdated(meshName, vgName, grName string) string {
	return acctest.ConfigCompose(testAccGatewayRouteConfig_base(meshName, vgName, "http2"), fmt.Sprintf(`
resource "aws_appmesh_gateway_route" "test" {
  name                 = %[1]q
  mesh_name            = aws_appmesh_mesh.test.name
  virtual_gateway_name = aws_appmesh_virtual_gateway.test.name

  spec {
    http2_route {
      action {
        target {
          virtual_service {
            virtual_service_name = aws_appmesh_virtual_service.test[1].name
          }
        }
      }

      match {
        prefix = "/users"
        port   = 8080
      }
    }
  }
}
`, grName))
}

func testAccGatewayRouteConfig_http2RouteWithQueryParameter(meshName, vgName, grName string) string {
	return acctest.ConfigCompose(testAccGatewayRouteConfig_base(meshName, vgName, "http2"), fmt.Sprintf(`
resource "aws_appmesh_gateway_route" "test" {
  name                 = %[1]q
  mesh_name            = aws_appmesh_mesh.test.name
  virtual_gateway_name = aws_appmesh_virtual_gateway.test.name

  spec {
    http2_route {
      action {
        target {
          virtual_service {
            virtual_service_name = aws_appmesh_virtual_service.test[0].name
          }
        }
      }

      match {
        prefix = "/"

        query_parameter {
          name = "param1"

          match {
            exact = "xact"
          }
        }
      }
    }
  }
}
`, grName))
}

func testAccGatewayRouteConfig_http2RouteMatchHostname(meshName, vgName, grName string) string {
	return acctest.ConfigCompose(testAccGatewayRouteConfig_base(meshName, vgName, "http2"), fmt.Sprintf(`
resource "aws_appmesh_gateway_route" "test" {
  name                 = %[1]q
  mesh_name            = aws_appmesh_mesh.test.name
  virtual_gateway_name = aws_appmesh_virtual_gateway.test.name

  spec {
    http2_route {
      action {
        target {
          virtual_service {
            virtual_service_name = aws_appmesh_virtual_service.test[0].name
          }
        }
      }

      match {
        hostname {
          exact = "test.example.com"
        }
      }
    }
  }
}
`, grName))
}

func testAccGatewayRouteConfig_http2RouteRewrite(meshName, vgName, grName string) string {
	return acctest.ConfigCompose(testAccGatewayRouteConfig_base(meshName, vgName, "http2"), fmt.Sprintf(`
resource "aws_appmesh_gateway_route" "test" {
  name                 = %[1]q
  mesh_name            = aws_appmesh_mesh.test.name
  virtual_gateway_name = aws_appmesh_virtual_gateway.test.name

  spec {
    http2_route {
      action {
        target {
          virtual_service {
            virtual_service_name = aws_appmesh_virtual_service.test[0].name
          }
        }
        rewrite {
          hostname {
            default_target_hostname = "DISABLED"
          }
          prefix {
            default_prefix = "DISABLED"
          }
        }
      }

      match {
        prefix = "/"
      }
    }
  }
}
`, grName))
}

func testAccGatewayRouteConfig_http2RouteRewriteWithPath(meshName, vgName, grName string) string {
	return acctest.ConfigCompose(testAccGatewayRouteConfig_base(meshName, vgName, "http2"), fmt.Sprintf(`
resource "aws_appmesh_gateway_route" "test" {
  name                 = %[1]q
  mesh_name            = aws_appmesh_mesh.test.name
  virtual_gateway_name = aws_appmesh_virtual_gateway.test.name

  spec {
    http2_route {
      action {
        target {
          virtual_service {
            virtual_service_name = aws_appmesh_virtual_service.test[0].name
          }
        }
        rewrite {
          path {
            exact = "/rewrite_path"
          }
        }
      }

      match {
        path {
          exact = "/match_path"
        }
      }
    }
  }
}
`, grName))
}
