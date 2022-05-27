package appmesh_test

import (
	"fmt"
	"regexp"
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

func testAccMesh_basic(t *testing.T) {
	var mesh appmesh.MeshData
	resourceName := "aws_appmesh_mesh.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appmesh.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMeshDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMeshConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMeshExists(resourceName, &mesh),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "appmesh", regexp.MustCompile(`mesh/.+`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccMesh_egressFilter(t *testing.T) {
	var mesh appmesh.MeshData
	resourceName := "aws_appmesh_mesh.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appmesh.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMeshDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMeshConfig_egressFilter(rName, "ALLOW_ALL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMeshExists(resourceName, &mesh),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "spec.0.egress_filter.0.type", "ALLOW_ALL"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMeshConfig_egressFilter(rName, "DROP_ALL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMeshExists(resourceName, &mesh),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "spec.0.egress_filter.0.type", "DROP_ALL"),
				),
			},
			{
				PlanOnly: true,
				Config:   testAccMeshConfig_basic(rName),
			},
		},
	})
}

func testAccMesh_tags(t *testing.T) {
	var mesh appmesh.MeshData
	resourceName := "aws_appmesh_mesh.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appmesh.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMeshDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMeshConfig_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMeshExists(resourceName, &mesh),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "bar"),
					resource.TestCheckResourceAttr(resourceName, "tags.good", "bad"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMeshConfig_updateTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMeshExists(resourceName, &mesh),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.good", "bad2"),
					resource.TestCheckResourceAttr(resourceName, "tags.fizz", "buzz"),
				),
			},
			{
				Config: testAccMeshConfig_removeTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMeshExists(resourceName, &mesh),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
				),
			},
		},
	})
}

func testAccCheckMeshDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AppMeshConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appmesh_mesh" {
			continue
		}

		_, err := conn.DescribeMesh(&appmesh.DescribeMeshInput{
			MeshName: aws.String(rs.Primary.Attributes["name"]),
		})
		if tfawserr.ErrCodeEquals(err, "NotFoundException") {
			continue
		}
		if err != nil {
			return err
		}
		return fmt.Errorf("still exist.")
	}

	return nil
}

func testAccCheckMeshExists(name string, v *appmesh.MeshData) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppMeshConn

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		resp, err := conn.DescribeMesh(&appmesh.DescribeMeshInput{
			MeshName: aws.String(rs.Primary.Attributes["name"]),
		})
		if err != nil {
			return err
		}

		*v = *resp.Mesh

		return nil
	}
}

func testAccMeshConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}
`, rName)
}

func testAccMeshConfig_egressFilter(rName, egressFilterType string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q

  spec {
    egress_filter {
      type = %[2]q
    }
  }
}
`, rName, egressFilterType)
}

func testAccMeshConfig_tags(rName string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q

  tags = {
    foo  = "bar"
    good = "bad"
  }
}
`, rName)
}

func testAccMeshConfig_updateTags(rName string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q

  tags = {
    foo  = "bar"
    good = "bad2"
    fizz = "buzz"
  }
}
`, rName)
}

func testAccMeshConfig_removeTags(rName string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q

  tags = {
    foo = "bar"
  }
}
`, rName)
}
