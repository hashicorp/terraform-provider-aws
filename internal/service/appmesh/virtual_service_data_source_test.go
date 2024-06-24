// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appmesh_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/appmesh"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccVirtualServiceDataSource_virtualNode(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appmesh_virtual_service.test"
	dataSourceName := "data.aws_appmesh_virtual_service.test"
	vsName := fmt.Sprintf("tf-acc-test-%d.mesh.local", sdkacctest.RandInt())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualServiceDataSourceConfig_virtualNode(rName, vsName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrCreatedDate, dataSourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrLastUpdatedDate, dataSourceName, names.AttrLastUpdatedDate),
					resource.TestCheckResourceAttrPair(resourceName, "mesh_name", dataSourceName, "mesh_name"),
					resource.TestCheckResourceAttrPair(resourceName, "mesh_owner", dataSourceName, "mesh_owner"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrName, dataSourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, acctest.CtResourceOwner, dataSourceName, acctest.CtResourceOwner),
					resource.TestCheckResourceAttrPair(resourceName, "spec.#", dataSourceName, "spec.#"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.provider.#", dataSourceName, "spec.0.provider.#"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.provider.0.virtual_node.#", dataSourceName, "spec.0.provider.0.virtual_node.#"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.provider.0.virtual_node.0.virtual_node_name", dataSourceName, "spec.0.provider.0.virtual_node.0.virtual_node_name"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.provider.0.virtual_router.#", dataSourceName, "spec.0.provider.0.virtual_router.#"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
			},
		},
	})
}

func testAccVirtualServiceDataSource_virtualRouter(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appmesh_virtual_service.test"
	dataSourceName := "data.aws_appmesh_virtual_service.test"
	vsName := fmt.Sprintf("tf-acc-test-%d.mesh.local", sdkacctest.RandInt())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualServiceDataSourceConfig_virtualRouter(rName, vsName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrCreatedDate, dataSourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrLastUpdatedDate, dataSourceName, names.AttrLastUpdatedDate),
					resource.TestCheckResourceAttrPair(resourceName, "mesh_name", dataSourceName, "mesh_name"),
					resource.TestCheckResourceAttrPair(resourceName, "mesh_owner", dataSourceName, "mesh_owner"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrName, dataSourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, acctest.CtResourceOwner, dataSourceName, acctest.CtResourceOwner),
					resource.TestCheckResourceAttrPair(resourceName, "spec.#", dataSourceName, "spec.#"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.provider.#", dataSourceName, "spec.0.provider.#"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.provider.0.virtual_node.#", dataSourceName, "spec.0.provider.0.virtual_node.#"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.provider.0.virtual_router.#", dataSourceName, "spec.0.provider.0.virtual_router.#"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.provider.0.virtual_router.0.virtual_router_name", dataSourceName, "spec.0.provider.0.virtual_router.0.virtual_router_name"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
			},
		},
	})
}

func testAccVirtualServiceDataSourceConfig_virtualNode(rName, vsName string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}

resource "aws_appmesh_virtual_node" "test" {
  name      = %[1]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {}
}

resource "aws_appmesh_virtual_service" "test" {
  name      = %[2]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    provider {
      virtual_node {
        virtual_node_name = aws_appmesh_virtual_node.test.name
      }
    }
  }
}

data "aws_appmesh_virtual_service" "test" {
  name      = aws_appmesh_virtual_service.test.name
  mesh_name = aws_appmesh_mesh.test.name
}
`, rName, vsName)
}

func testAccVirtualServiceDataSourceConfig_virtualRouter(rName, vsName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}

resource "aws_appmesh_virtual_router" "test" {
  name      = %[1]q
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

resource "aws_appmesh_virtual_service" "test" {
  name      = %[2]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    provider {
      virtual_router {
        virtual_router_name = aws_appmesh_virtual_router.test.name
      }
    }
  }
}

data "aws_appmesh_virtual_service" "test" {
  name       = aws_appmesh_virtual_service.test.name
  mesh_name  = aws_appmesh_mesh.test.name
  mesh_owner = data.aws_caller_identity.current.account_id
}
`, rName, vsName)
}
