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

func testAccVirtualRouter_basic(t *testing.T) {
	var vr appmesh.VirtualRouterData
	resourceName := "aws_appmesh_virtual_router.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vrName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appmesh.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVirtualRouterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualRouterConfig_basic(meshName, vrName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVirtualRouterExists(resourceName, &vr),
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
				),
			},
			{
				Config: testAccVirtualRouterConfig_updated(meshName, vrName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVirtualRouterExists(resourceName, &vr),
					resource.TestCheckResourceAttr(resourceName, "name", vrName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8081"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "http"),
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
	var vr appmesh.VirtualRouterData
	resourceName := "aws_appmesh_virtual_router.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vrName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appmesh.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVirtualRouterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualRouterConfig_tags(meshName, vrName, "foo", "bar", "good", "bad"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVirtualRouterExists(
						resourceName, &vr),
					resource.TestCheckResourceAttr(
						resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(
						resourceName, "tags.foo", "bar"),
					resource.TestCheckResourceAttr(
						resourceName, "tags.good", "bad"),
				),
			},
			{
				Config: testAccVirtualRouterConfig_tags(meshName, vrName, "foo2", "bar", "good", "bad2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVirtualRouterExists(
						resourceName, &vr),
					resource.TestCheckResourceAttr(
						resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(
						resourceName, "tags.foo2", "bar"),
					resource.TestCheckResourceAttr(
						resourceName, "tags.good", "bad2"),
				),
			},
			{
				Config: testAccVirtualRouterConfig_basic(meshName, vrName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVirtualRouterExists(
						resourceName, &vr),
					resource.TestCheckResourceAttr(
						resourceName, "tags.%", "0"),
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

func testAccCheckVirtualRouterDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AppMeshConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appmesh_virtual_router" {
			continue
		}

		_, err := conn.DescribeVirtualRouter(&appmesh.DescribeVirtualRouterInput{
			MeshName:          aws.String(rs.Primary.Attributes["mesh_name"]),
			VirtualRouterName: aws.String(rs.Primary.Attributes["name"]),
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

func testAccCheckVirtualRouterExists(name string, v *appmesh.VirtualRouterData) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppMeshConn

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		resp, err := conn.DescribeVirtualRouter(&appmesh.DescribeVirtualRouterInput{
			MeshName:          aws.String(rs.Primary.Attributes["mesh_name"]),
			VirtualRouterName: aws.String(rs.Primary.Attributes["name"]),
		})
		if err != nil {
			return err
		}

		*v = *resp.VirtualRouter

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

func testAccVirtualRouterConfig_tags(meshName, vrName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
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
