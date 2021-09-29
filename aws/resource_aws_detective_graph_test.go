package aws

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/detective"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAwsDetectiveGraph_basic(t *testing.T) {
	var graphOutput detective.Graph
	resourceName := "aws_detective_graph.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsDetectiveGraphDestroy,
		ErrorCheck:        testAccErrorCheck(t, detective.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsDetectiveGraphConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDetectiveGraphExists(resourceName, &graphOutput),
					testAccCheckResourceAttrRfc3339(resourceName, "created_time"),
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

func TestAccAwsDetectiveGraph_WithTags(t *testing.T) {
	var graphOutput detective.Graph
	resourceName := "aws_detective_graph.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsDetectiveGraphDestroy,
		ErrorCheck:        testAccErrorCheck(t, detective.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsDetectiveGraphConfigWithTags(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDetectiveGraphExists(resourceName, &graphOutput),
					testAccCheckResourceAttrRfc3339(resourceName, "created_time"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key", "value"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Key", "value"),
				),
			},
			{
				Config: testAccAwsDetectiveGraphConfigWithTagsUpdate(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDetectiveGraphExists(resourceName, &graphOutput),
					testAccCheckResourceAttrRfc3339(resourceName, "created_time"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key", "value"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Key", "value"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Key2", "value2"),
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

func TestAccAwsDetectiveGraph_disappears(t *testing.T) {
	var graphOutput detective.Graph
	resourceName := "aws_detective_graph.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsDetectiveGraphDestroy,
		ErrorCheck:        testAccErrorCheck(t, detective.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsDetectiveGraphConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDetectiveGraphExists(resourceName, &graphOutput),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsDetectiveGraph(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAwsDetectiveGraphDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).detectiveconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_detective_graph" {
			continue
		}

		resp, err := getDetectiveGraphArn(conn, context.Background(), rs.Primary.ID)

		if tfawserr.ErrCodeEquals(err, detective.ErrCodeResourceNotFoundException) || resp == nil {
			continue
		}

		if err != nil {
			return err
		}

		if resp != nil {
			return fmt.Errorf("detective graph %q still exists", rs.Primary.ID)
		}
	}

	return nil

}

func testAccCheckAwsDetectiveGraphExists(resourceName string, graph *detective.Graph) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).detectiveconn
		resp, err := getDetectiveGraphArn(conn, context.Background(), rs.Primary.ID)

		if err != nil {
			return err
		}

		if resp == nil {
			return fmt.Errorf("detective graph %q does not exist", rs.Primary.ID)
		}

		*graph = *resp

		return nil
	}
}

func testAccAwsDetectiveGraphConfigBasic() string {
	return `
resource "aws_detective_graph" "test" {}
`
}

func testAccAwsDetectiveGraphConfigWithTags() string {
	return `
resource "aws_detective_graph" "test" {
  tags = {
    Key = "value"
  }
}
`
}

func testAccAwsDetectiveGraphConfigWithTagsUpdate() string {
	return `
resource "aws_detective_graph" "test" {
  tags = {
    Key  = "value"
    Key2 = "value2"
  }
}
`
}
