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
)

func testAccVirtualRouterDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	virtualRouterName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appmesh_virtual_router.test"
	dataSourceName := "data.aws_appmesh_virtual_router.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVirtualRouterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualRouterDataSourceConfig(meshName, virtualRouterName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "created_date", dataSourceName, "created_date"),
					resource.TestCheckResourceAttrPair(resourceName, "last_updated_date", dataSourceName, "last_updated_date"),
					resource.TestCheckResourceAttrPair(resourceName, "mesh_name", dataSourceName, "mesh_name"),
					resource.TestCheckResourceAttrPair(resourceName, "mesh_owner", dataSourceName, "mesh_owner"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_owner", dataSourceName, "resource_owner"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.#", dataSourceName, "spec.#"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.listener.#", dataSourceName, "spec.0.listener.#"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.listener.0.port_mapping.#", dataSourceName, "spec.0.listener.0.port_mapping.#"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.listener.0.port_mapping.0.port", dataSourceName, "spec.0.listener.0.port_mapping.0.port"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.listener.0.port_mapping.0.protocol", dataSourceName, "spec.0.listener.0.port_mapping.0.protocol"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
				),
			},
		},
	})
}

func testAccVirtualRouterDataSourceConfig(meshName, vrName string) string {
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

  tags = {
    Name = %[2]q
  }
}

data "aws_appmesh_virtual_router" "test" {
  name      = aws_appmesh_virtual_router.test.name
  mesh_name = aws_appmesh_mesh.test.name
}
`, meshName, vrName)
}
