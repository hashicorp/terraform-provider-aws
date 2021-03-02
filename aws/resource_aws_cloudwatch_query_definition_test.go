package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudwatchlogs/finder"
	"testing"
)

func TestAccAWSCloudWatchQueryDefinition_basic(t *testing.T) {
	ident := "basic"
	resourceName := fmt.Sprintf("aws_cloudwatch_query_definition.%s", ident)
	queryName := "test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSCloudWatchQueryDefinitionDestroy(resourceName, queryName),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchQueryDefinitionConfig(queryName, ident),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloudWatchQueryDefinitionExists(resourceName, queryName),
					resource.TestCheckResourceAttr(resourceName, "name", queryName),
				),
			}, {
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_modified"},
				ImportStateIdPrefix:     fmt.Sprintf("%s_", queryName),
			},
		},
	})
}

func TestAccAWSCloudWatchQueryDefinition_disappears(t *testing.T) {
	ident := "disappears"
	resourceName := fmt.Sprintf("aws_cloudwatch_query_definition.%s", ident)
	queryName := "test-disappears"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSCloudWatchQueryDefinitionDestroy(resourceName, queryName),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchQueryDefinitionConfig(queryName, ident),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloudWatchQueryDefinitionExists(resourceName, queryName),
					testAccCheckAWSCloudWatchQueryDefinitionDisappears(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSCloudWatchQueryDefinition_update(t *testing.T) {
	ident := "update"
	resourceName := fmt.Sprintf("aws_cloudwatch_query_definition.%s", ident)
	queryName := "testupdate"
	updatedQueryName := "test-update"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSCloudWatchQueryDefinitionDestroy(resourceName, queryName),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchQueryDefinitionConfig(queryName, ident),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloudWatchQueryDefinitionExists(resourceName, queryName),
					resource.TestCheckResourceAttr(resourceName, "name", queryName),
				),
			}, {
				Config: testAccAWSCloudWatchQueryDefinitionConfig(updatedQueryName, ident),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloudWatchQueryDefinitionExists(resourceName, updatedQueryName),
					resource.TestCheckResourceAttr(resourceName, "name", updatedQueryName),
				),
			},
		},
	})
}

func testAccCheckAWSCloudWatchQueryDefinitionDisappears(rName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).cloudwatchlogsconn

		rs, ok := s.RootModule().Resources[rName]
		if !ok {
			return fmt.Errorf("resource not found: %s", rName)
		}

		qID := rs.Primary.ID

		input := &cloudwatchlogs.DeleteQueryDefinitionInput{QueryDefinitionId: &qID}
		_, err := conn.DeleteQueryDefinition(input)

		return err
	}
}

func testAccCheckAWSCloudWatchQueryDefinitionExists(rName, qName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rName]
		if !ok {
			return fmt.Errorf("Not found: %s", rName)
		}
		conn := testAccProvider.Meta().(*AWSClient).cloudwatchlogsconn

		result, err := finder.QueryDefinition(conn, qName, rs.Primary.ID)

		if err != nil {
			return err
		}

		if result == nil {
			return fmt.Errorf("query with name %s, id %s not found", qName, rs.Primary.ID)
		}

		if got, want := aws.StringValue(result.QueryDefinitionId), rs.Primary.ID; got != want {
			return fmt.Errorf("want query id: %s, got %s", want, got)
		}

		return nil
	}
}

func testAccAWSCloudWatchQueryDefinitionConfig(name, ident string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_query_definition" "%s" {
	name = "%s"
    log_groups = []
	query = <<EOF
fields @timestamp, @message
| sort @timestamp desc
| limit 20
EOF
}
`, ident, name)
}

func testAccCheckAWSCloudWatchQueryDefinitionDestroy(rName, qName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).cloudwatchlogsconn

		rs, ok := s.RootModule().Resources[rName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", rName)
		}

		result, err := finder.QueryDefinition(conn, qName, rs.Primary.ID)
		if err != nil {
			return err
		}

		if result != nil {
			return fmt.Errorf("query definition %s - %s still exists", qName, rs.Primary.ID)
		}

		return nil
	}
}
