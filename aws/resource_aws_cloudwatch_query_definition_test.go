package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"testing"
)

func TestAccAWSCloudWatchQueryDefinition_basic(t *testing.T) {
	resourceName := "aws_cloudwatch_query_definition.query"
	queryName := "test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSCloudWatchQueryDefinitionDestroy(resourceName, queryName),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchQueryDefinitionConfig(queryName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloudWatchQueryDefinitionExists(resourceName, queryName),
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
	resourceName := "aws_cloudwatch_query_definition.query"
	queryName := "test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSCloudWatchQueryDefinitionDestroy(resourceName, queryName),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchQueryDefinitionConfig(queryName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloudWatchQueryDefinitionExists(resourceName, queryName),
					testAccCheckAWSCloudWatchQueryDefinitionDisappears(resourceName),
				),
				ExpectNonEmptyPlan: true,
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
		params := &cloudwatchlogs.DescribeQueryDefinitionsInput{QueryDefinitionNamePrefix: aws.String(qName)}
		resp, err := conn.DescribeQueryDefinitions(params)

		if err != nil {
			return err
		}

		if len(resp.QueryDefinitions) != 1 {
			return fmt.Errorf("expected 1 query result, got %d", len(resp.QueryDefinitions))
		}

		if got, want := *resp.QueryDefinitions[0].QueryDefinitionId, rs.Primary.ID; got != want {
			return fmt.Errorf("want query id: %s, got %s", want, got)
		}

		return nil
	}
}

func testAccAWSCloudWatchQueryDefinitionConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_query_definition" "query" {
	name = "%s"
    log_groups = []
	query = <<EOF
fields @timestamp, @message
| sort @timestamp desc
| limit 20
EOF
}
`, name)
}

func testAccCheckAWSCloudWatchQueryDefinitionDestroy(rName, qName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).cloudwatchlogsconn

		rs, ok := s.RootModule().Resources[rName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", rName)
		}

		params := &cloudwatchlogs.DescribeQueryDefinitionsInput{QueryDefinitionNamePrefix: aws.String(qName)}
		resp, err := conn.DescribeQueryDefinitions(params)

		if err != nil {
			return err
		}

		if len(resp.QueryDefinitions) != 0 {
			return fmt.Errorf("query definition still exists: %s", rs.Primary.ID)
		}

		return nil
	}
}
