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
)

func testAccVirtualRouter_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var vr appmesh.VirtualRouterData
	resourceName := "aws_appmesh_virtual_router.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vrName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVirtualRouterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualRouterConfig_basic(meshName, vrName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVirtualRouterExists(ctx, resourceName, &vr),
					resource.TestCheckResourceAttr(resourceName, "name", vrName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "http"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s", meshName, vrName)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccVirtualRouterConfig_updated(meshName, vrName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVirtualRouterExists(ctx, resourceName, &vr),
					resource.TestCheckResourceAttr(resourceName, "name", vrName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8081"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", meshName, vrName),
				ImportState:       true,
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
		ErrorCheck:               acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVirtualRouterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualRouterConfig_multiListener(meshName, vrName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVirtualRouterExists(ctx, resourceName, &vr),
					resource.TestCheckResourceAttr(resourceName, "name", vrName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.port_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.port_mapping.0.port", "8081"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.port_mapping.0.protocol", "http2"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualRouter/%s", meshName, vrName)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccVirtualRouterConfig_multiListenerUpdated(meshName, vrName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVirtualRouterExists(ctx, resourceName, &vr),
					resource.TestCheckResourceAttr(resourceName, "name", vrName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
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
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", meshName, vrName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccVirtualRouter_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var vr appmesh.VirtualRouterData
	resourceName := "aws_appmesh_virtual_router.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vrName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVirtualRouterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualRouterConfig_tags1(meshName, vrName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVirtualRouterExists(ctx, resourceName, &vr),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", meshName, vrName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVirtualRouterConfig_tags2(meshName, vrName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVirtualRouterExists(ctx, resourceName, &vr),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccVirtualRouterConfig_tags1(meshName, vrName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVirtualRouterExists(ctx, resourceName, &vr),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
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
		ErrorCheck:               acctest.ErrorCheck(t, appmesh.EndpointsID),
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

			_, err := tfappmesh.FindVirtualRouterByThreePartKey(ctx, conn, rs.Primary.Attributes["mesh_name"], rs.Primary.Attributes["mesh_owner"], rs.Primary.Attributes["name"])

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

		output, err := tfappmesh.FindVirtualRouterByThreePartKey(ctx, conn, rs.Primary.Attributes["mesh_name"], rs.Primary.Attributes["mesh_owner"], rs.Primary.Attributes["name"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
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

func testAccVirtualRouterConfig_tags1(meshName, vrName, tagKey1, tagValue1 string) string {
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
    %[3]s = %[4]q
  }
}
`, meshName, vrName, tagKey1, tagValue1)
}

func testAccVirtualRouterConfig_tags2(meshName, vrName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
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
    %[3]s = %[4]q
    %[5]s = %[6]q
  }
}
`, meshName, vrName, tagKey1, tagValue1, tagKey2, tagValue2)
}
