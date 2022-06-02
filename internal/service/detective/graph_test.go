package detective_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/detective"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdetective "github.com/hashicorp/terraform-provider-aws/internal/service/detective"
)

func testAccGraph_basic(t *testing.T) {
	var graphOutput detective.Graph
	resourceName := "aws_detective_graph.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGraphDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, detective.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphExists(resourceName, &graphOutput),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_time"),
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

func testAccGraph_tags(t *testing.T) {
	var graph1, graph2 detective.Graph
	resourceName := "aws_detective_graph.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGraphDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, detective.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphConfig_tags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphExists(resourceName, &graph1),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_time"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", "value1"),
				),
			},
			{
				Config: testAccGraphConfig_tags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphExists(resourceName, &graph2),
					testAccCheckGraphNotRecreated(&graph1, &graph2),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_time"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key2", "value2"),
				),
			},
			{
				Config: testAccGraphConfig_tags1("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphExists(resourceName, &graph2),
					testAccCheckGraphNotRecreated(&graph1, &graph2),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_time"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key2", "value2"),
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

func testAccGraph_disappears(t *testing.T) {
	var graphOutput detective.Graph
	resourceName := "aws_detective_graph.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGraphDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, detective.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGraphExists(resourceName, &graphOutput),
					acctest.CheckResourceDisappears(acctest.Provider, tfdetective.ResourceGraph(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckGraphDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DetectiveConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_detective_graph" {
			continue
		}

		resp, err := tfdetective.FindGraphByARN(conn, context.Background(), rs.Primary.ID)

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

func testAccCheckGraphExists(resourceName string, graph *detective.Graph) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DetectiveConn
		resp, err := tfdetective.FindGraphByARN(conn, context.Background(), rs.Primary.ID)

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

func testAccCheckGraphNotRecreated(before, after *detective.Graph) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.StringValue(before.Arn), aws.StringValue(after.Arn); before != after {
			return fmt.Errorf("detective graph (%s/%s) recreated", before, after)
		}

		return nil
	}
}

func testAccGraphConfig_basic() string {
	return `
resource "aws_detective_graph" "test" {}
`
}

func testAccGraphConfig_tags1(tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_detective_graph" "test" {
  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1)
}

func testAccGraphConfig_tags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_detective_graph" "test" {
  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}
