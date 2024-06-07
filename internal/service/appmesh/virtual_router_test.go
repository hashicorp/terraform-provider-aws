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

func testAccVirtualRouter_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var vr appmesh.VirtualRouterData
	resourceName := "aws_appmesh_virtual_router.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vrName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVirtualRouterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualRouterConfig_basic(meshName, vrName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVirtualRouterExists(ctx, resourceName, &vr),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, vrName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "http"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s", meshName, vrName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				Config: testAccVirtualRouterConfig_updated(meshName, vrName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVirtualRouterExists(ctx, resourceName, &vr),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, vrName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8081"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccVirtualRouterImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccVirtualRouter_multiListener(t *testing.T) {
	ctx := acctest.Context(t)
	var vr appmesh.VirtualRouterData
	resourceName := "aws_appmesh_virtual_router.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vrName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVirtualRouterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualRouterConfig_multiListener(meshName, vrName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVirtualRouterExists(ctx, resourceName, &vr),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, vrName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.port_mapping.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.port_mapping.0.port", "8081"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.port_mapping.0.protocol", "http2"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s", meshName, vrName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				Config: testAccVirtualRouterConfig_multiListenerUpdated(meshName, vrName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVirtualRouterExists(ctx, resourceName, &vr),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, vrName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8081"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.port_mapping.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.port_mapping.0.port", "8082"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.port_mapping.0.protocol", "http2"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.2.port_mapping.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.2.port_mapping.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.2.port_mapping.0.protocol", "grpc"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccVirtualRouterImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccVirtualRouter_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var vr appmesh.VirtualRouterData
	resourceName := "aws_appmesh_virtual_router.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vrName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVirtualRouterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualRouterConfig_basic(meshName, vrName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVirtualRouterExists(ctx, resourceName, &vr),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfappmesh.ResourceVirtualRouter(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckVirtualRouterDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppMeshConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appmesh_virtual_router" {
				continue
			}

			_, err := tfappmesh.FindVirtualRouterByThreePartKey(ctx, conn, rs.Primary.Attributes["mesh_name"], rs.Primary.Attributes["mesh_owner"], rs.Primary.Attributes[names.AttrName])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("App Mesh Virtual Router %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckVirtualRouterExists(ctx context.Context, n string, v *appmesh.VirtualRouterData) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppMeshConn(ctx)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No App Mesh Virtual Router ID is set")
		}

		output, err := tfappmesh.FindVirtualRouterByThreePartKey(ctx, conn, rs.Primary.Attributes["mesh_name"], rs.Primary.Attributes["mesh_owner"], rs.Primary.Attributes[names.AttrName])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccVirtualRouterImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["mesh_name"], rs.Primary.Attributes[names.AttrName]), nil
	}
}

func testAccVirtualRouterConfig_basic(meshName, vrName string) string {
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
`, meshName, vrName)
}

func testAccVirtualRouterConfig_updated(meshName, vrName string) string {
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
        port     = 8081
        protocol = "http"
      }
    }
  }
}
`, meshName, vrName)
}

func testAccVirtualRouterConfig_multiListener(meshName, vrName string) string {
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
    listener {
      port_mapping {
        port     = 8081
        protocol = "http2"
      }
    }
  }
}
`, meshName, vrName)
}

func testAccVirtualRouterConfig_multiListenerUpdated(meshName, vrName string) string {
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
        port     = 8081
        protocol = "http"
      }
    }
    listener {
      port_mapping {
        port     = 8082
        protocol = "http2"
      }
    }
    listener {
      port_mapping {
        port     = 8080
        protocol = "grpc"
      }
    }
  }
}
`, meshName, vrName)
}
