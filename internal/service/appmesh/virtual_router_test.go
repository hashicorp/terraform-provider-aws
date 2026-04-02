// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package appmesh_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/appmesh/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfappmesh "github.com/hashicorp/terraform-provider-aws/internal/service/appmesh"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccVirtualRouter_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var vr awstypes.VirtualRouterData
	resourceName := "aws_appmesh_virtual_router.test"
	meshName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	vrName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppMeshEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVirtualRouterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualRouterConfig_basic(meshName, vrName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVirtualRouterExists(ctx, t, resourceName, &vr),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, vrName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "http"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s", meshName, vrName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				Config: testAccVirtualRouterConfig_updated(meshName, vrName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVirtualRouterExists(ctx, t, resourceName, &vr),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, vrName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8081"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
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
	var vr awstypes.VirtualRouterData
	resourceName := "aws_appmesh_virtual_router.test"
	meshName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	vrName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppMeshEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVirtualRouterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualRouterConfig_multiListener(meshName, vrName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVirtualRouterExists(ctx, t, resourceName, &vr),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, vrName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.port_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.port_mapping.0.port", "8081"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.port_mapping.0.protocol", "http2"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s", meshName, vrName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				Config: testAccVirtualRouterConfig_multiListenerUpdated(meshName, vrName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVirtualRouterExists(ctx, t, resourceName, &vr),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, vrName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8081"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.port_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.port_mapping.0.port", "8082"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.port_mapping.0.protocol", "http2"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.2.port_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.2.port_mapping.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.2.port_mapping.0.protocol", "grpc"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
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
	var vr awstypes.VirtualRouterData
	resourceName := "aws_appmesh_virtual_router.test"
	meshName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	vrName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppMeshEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVirtualRouterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualRouterConfig_basic(meshName, vrName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVirtualRouterExists(ctx, t, resourceName, &vr),
					acctest.CheckSDKResourceDisappears(ctx, t, tfappmesh.ResourceVirtualRouter(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckVirtualRouterDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).AppMeshClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appmesh_virtual_router" {
				continue
			}

			_, err := tfappmesh.FindVirtualRouterByThreePartKey(ctx, conn, rs.Primary.Attributes["mesh_name"], rs.Primary.Attributes["mesh_owner"], rs.Primary.Attributes[names.AttrName])

			if retry.NotFound(err) {
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

func testAccCheckVirtualRouterExists(ctx context.Context, t *testing.T, n string, v *awstypes.VirtualRouterData) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).AppMeshClient(ctx)

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
