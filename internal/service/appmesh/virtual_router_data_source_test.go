// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package appmesh_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccVirtualRouterDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	virtualRouterName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	meshName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_appmesh_virtual_router.test"
	dataSourceName := "data.aws_appmesh_virtual_router.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppMeshEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVirtualRouterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualRouterDataSourceConfig(meshName, virtualRouterName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrCreatedDate, dataSourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrLastUpdatedDate, dataSourceName, names.AttrLastUpdatedDate),
					resource.TestCheckResourceAttrPair(resourceName, "mesh_name", dataSourceName, "mesh_name"),
					resource.TestCheckResourceAttrPair(resourceName, "mesh_owner", dataSourceName, "mesh_owner"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrName, dataSourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, acctest.CtResourceOwner, dataSourceName, acctest.CtResourceOwner),
					resource.TestCheckResourceAttrPair(resourceName, "spec.#", dataSourceName, "spec.#"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.listener.#", dataSourceName, "spec.0.listener.#"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.listener.0.port_mapping.#", dataSourceName, "spec.0.listener.0.port_mapping.#"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.listener.0.port_mapping.0.port", dataSourceName, "spec.0.listener.0.port_mapping.0.port"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.listener.0.port_mapping.0.protocol", dataSourceName, "spec.0.listener.0.port_mapping.0.protocol"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
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
}

data "aws_appmesh_virtual_router" "test" {
  name      = aws_appmesh_virtual_router.test.name
  mesh_name = aws_appmesh_mesh.test.name
}
`, meshName, vrName)
}
