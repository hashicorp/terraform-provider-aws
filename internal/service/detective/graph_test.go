package detective_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/detective"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdetective "github.com/hashicorp/terraform-provider-aws/internal/service/detective"
)

func testAccDetectiveGraph_basic(t *testing.T) {
	var graphOutput detective.Graph
	resourceName := "aws_detective_graph.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDetectiveGraphDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, detective.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccDetectiveGraphConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDetectiveGraphExists(resourceName, &graphOutput),
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

func testAccDetectiveGraph_tags(t *testing.T) {
	var graph1, graph2 detective.Graph
	resourceName := "aws_detective_graph.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDetectiveGraphDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, detective.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccDetectiveGraphConfigTags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDetectiveGraphExists(resourceName, &graph1),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_time"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", "value1"),
				),
			},
			{
				Config: testAccDetectiveGraphConfigTags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDetectiveGraphExists(resourceName, &graph2),
					testAccCheckDetectiveGraphNotRecreated(&graph1, &graph2),
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
				Config: testAccDetectiveGraphConfigTags1("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDetectiveGraphExists(resourceName, &graph2),
					testAccCheckDetectiveGraphNotRecreated(&graph1, &graph2),
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

func testAccDetectiveGraph_disappears(t *testing.T) {
	var graphOutput detective.Graph
	resourceName := "aws_detective_graph.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDetectiveGraphDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, detective.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccDetectiveGraphConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDetectiveGraphExists(resourceName, &graphOutput),
					acctest.CheckResourceDisappears(acctest.Provider, tfdetective.ResourceGraph(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDetectiveGraphDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DetectiveConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_detective_graph" {
			continue
		}

		resp, err := tfdetective.FindDetectiveGraphByArn(conn, context.Background(), rs.Primary.ID)

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

func testAccCheckDetectiveGraphExists(resourceName string, graph *detective.Graph) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DetectiveConn
		resp, err := tfdetective.FindDetectiveGraphByArn(conn, context.Background(), rs.Primary.ID)

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

func testAccCheckDetectiveGraphNotRecreated(before, after *detective.Graph) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.StringValue(before.Arn), aws.StringValue(after.Arn); before != after {
			return fmt.Errorf("detective graph (%s/%s) recreated", before, after)
		}

		return nil
	}
}

func testAccDetectiveGraphConfigBasic() string {
	return `
resource "aws_detective_graph" "test" {}
`
}

func testAccDetectiveGraphConfigTags1(tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_detective_graph" "test" {
  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1)
}

func testAccDetectiveGraphConfigTags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_detective_graph" "test" {
  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}
