// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appmesh_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/appmesh"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccRouteDataSource_http2Route(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_appmesh_route.test"
	dataSourceName := "data.aws_appmesh_route.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vrName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vnName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRouteDataSourceConfig_http2Route(meshName, vrName, vnName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrCreatedDate, dataSourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrLastUpdatedDate, dataSourceName, names.AttrLastUpdatedDate),
					resource.TestCheckResourceAttrPair(resourceName, "mesh_name", dataSourceName, "mesh_name"),
					resource.TestCheckResourceAttrPair(resourceName, "virtual_router_name", dataSourceName, "virtual_router_name"),
					resource.TestCheckResourceAttrPair(resourceName, "mesh_owner", dataSourceName, "mesh_owner"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrName, dataSourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, acctest.CtResourceOwner, dataSourceName, acctest.CtResourceOwner),
					resource.TestCheckResourceAttrPair(resourceName, "spec.#", dataSourceName, "spec.#"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.grpc_route.#", dataSourceName, "spec.0.grpc_route.#"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.http_route.#", dataSourceName, "spec.0.http_route.#"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.http2_route.#", dataSourceName, "spec.0.http2_route.#"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.priority", dataSourceName, "spec.0.priority"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.tcp_route.#", dataSourceName, "spec.0.tcp_route.#"),
					resource.TestCheckResourceAttrPair(resourceName, acctest.CtTagsPercent, dataSourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func testAccRouteDataSource_httpRoute(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_appmesh_route.test"
	dataSourceName := "data.aws_appmesh_route.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vrName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vnName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRouteDataSourceConfig_httpRoute(meshName, vrName, vnName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrCreatedDate, dataSourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrLastUpdatedDate, dataSourceName, names.AttrLastUpdatedDate),
					resource.TestCheckResourceAttrPair(resourceName, "mesh_name", dataSourceName, "mesh_name"),
					resource.TestCheckResourceAttrPair(resourceName, "virtual_router_name", dataSourceName, "virtual_router_name"),
					resource.TestCheckResourceAttrPair(resourceName, "mesh_owner", dataSourceName, "mesh_owner"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrName, dataSourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, acctest.CtResourceOwner, dataSourceName, acctest.CtResourceOwner),
					resource.TestCheckResourceAttrPair(resourceName, "spec.#", dataSourceName, "spec.#"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.grpc_route.#", dataSourceName, "spec.0.grpc_route.#"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.http_route.#", dataSourceName, "spec.0.http_route.#"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.http2_route.#", dataSourceName, "spec.0.http2_route.#"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.priority", dataSourceName, "spec.0.priority"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.tcp_route.#", dataSourceName, "spec.0.tcp_route.#"),
					resource.TestCheckResourceAttrPair(resourceName, acctest.CtTagsPercent, dataSourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func testAccRouteDataSource_grpcRoute(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_appmesh_route.test"
	dataSourceName := "data.aws_appmesh_route.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vrName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vnName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRouteDataSourceConfig_grpcRoute(meshName, vrName, vnName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrCreatedDate, dataSourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrLastUpdatedDate, dataSourceName, names.AttrLastUpdatedDate),
					resource.TestCheckResourceAttrPair(resourceName, "mesh_name", dataSourceName, "mesh_name"),
					resource.TestCheckResourceAttrPair(resourceName, "virtual_router_name", dataSourceName, "virtual_router_name"),
					resource.TestCheckResourceAttrPair(resourceName, "mesh_owner", dataSourceName, "mesh_owner"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrName, dataSourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, acctest.CtResourceOwner, dataSourceName, acctest.CtResourceOwner),
					resource.TestCheckResourceAttrPair(resourceName, "spec.#", dataSourceName, "spec.#"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.grpc_route.#", dataSourceName, "spec.0.grpc_route.#"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.http_route.#", dataSourceName, "spec.0.http_route.#"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.http2_route.#", dataSourceName, "spec.0.http2_route.#"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.priority", dataSourceName, "spec.0.priority"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.tcp_route.#", dataSourceName, "spec.0.tcp_route.#"),
					resource.TestCheckResourceAttrPair(resourceName, acctest.CtTagsPercent, dataSourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func testAccRouteDataSource_tcpRoute(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_appmesh_route.test"
	dataSourceName := "data.aws_appmesh_route.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vrName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vnName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRouteDataSourceConfig_tcpRoute(meshName, vrName, vnName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrCreatedDate, dataSourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrLastUpdatedDate, dataSourceName, names.AttrLastUpdatedDate),
					resource.TestCheckResourceAttrPair(resourceName, "mesh_name", dataSourceName, "mesh_name"),
					resource.TestCheckResourceAttrPair(resourceName, "virtual_router_name", dataSourceName, "virtual_router_name"),
					resource.TestCheckResourceAttrPair(resourceName, "mesh_owner", dataSourceName, "mesh_owner"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrName, dataSourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, acctest.CtResourceOwner, dataSourceName, acctest.CtResourceOwner),
					resource.TestCheckResourceAttrPair(resourceName, "spec.#", dataSourceName, "spec.#"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.grpc_route.#", dataSourceName, "spec.0.grpc_route.#"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.http_route.#", dataSourceName, "spec.0.http_route.#"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.http2_route.#", dataSourceName, "spec.0.http2_route.#"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.priority", dataSourceName, "spec.0.priority"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.tcp_route.#", dataSourceName, "spec.0.tcp_route.#"),
					resource.TestCheckResourceAttrPair(resourceName, acctest.CtTagsPercent, dataSourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func testAccRouteDataSourceConfig_base(meshName, vrName, vnName string) string {
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
        protocol = "http"
      }
    }
  }
}

resource "aws_appmesh_virtual_node" "test" {
  name      = %[3]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {}
}
`, meshName, vrName, vnName)
}

func testAccRouteDataSourceConfig_httpRoute(meshName, vrName, vnName, rName string) string {
	return acctest.ConfigCompose(testAccRouteDataSourceConfig_base(meshName, vrName, vnName), fmt.Sprintf(`
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
          virtual_node = aws_appmesh_virtual_node.test.name
          weight       = 100
        }
      }
    }
  }

  tags = {
    Name = %[1]q
  }
}

data "aws_appmesh_route" "test" {
  name                = aws_appmesh_route.test.name
  mesh_name           = aws_appmesh_route.test.mesh_name
  virtual_router_name = aws_appmesh_route.test.virtual_router_name
}
`, rName))
}

func testAccRouteDataSourceConfig_http2Route(meshName, vrName, vnName, rName string) string {
	return acctest.ConfigCompose(testAccRouteDataSourceConfig_base(meshName, vrName, vnName), fmt.Sprintf(`
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
          virtual_node = aws_appmesh_virtual_node.test.name
          weight       = 100
        }
      }
    }
  }
}

data "aws_appmesh_route" "test" {
  name                = aws_appmesh_route.test.name
  mesh_name           = aws_appmesh_route.test.mesh_name
  virtual_router_name = aws_appmesh_route.test.virtual_router_name
}
`, rName))
}

func testAccRouteDataSourceConfig_grpcRoute(meshName, vrName, vnName, rName string) string {
	return acctest.ConfigCompose(testAccRouteDataSourceConfig_base(meshName, vrName, vnName), fmt.Sprintf(`
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
          virtual_node = aws_appmesh_virtual_node.test.name
          weight       = 100
        }
      }
    }
  }

  tags = {
    Name = %[1]q
  }
}

data "aws_appmesh_route" "test" {
  name                = aws_appmesh_route.test.name
  mesh_name           = aws_appmesh_route.test.mesh_name
  virtual_router_name = aws_appmesh_route.test.virtual_router_name
}
`, rName))
}

func testAccRouteDataSourceConfig_tcpRoute(meshName, vrName, vnName, rName string) string {
	return acctest.ConfigCompose(testAccRouteDataSourceConfig_base(meshName, vrName, vnName), fmt.Sprintf(`
resource "aws_appmesh_route" "test" {
  name                = %[1]q
  mesh_name           = aws_appmesh_mesh.test.id
  virtual_router_name = aws_appmesh_virtual_router.test.name

  spec {
    tcp_route {
      action {
        weighted_target {
          virtual_node = aws_appmesh_virtual_node.test.name
          weight       = 100
        }
      }
    }
  }
}

data "aws_appmesh_route" "test" {
  name                = aws_appmesh_route.test.name
  mesh_name           = aws_appmesh_route.test.mesh_name
  virtual_router_name = aws_appmesh_route.test.virtual_router_name
}
`, rName))
}
