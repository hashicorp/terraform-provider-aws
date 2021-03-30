package aws

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudwatchlogs/finder"
)

func TestAccAWSCloudWatchQueryDefinition_basic(t *testing.T) {
	var v cloudwatchlogs.QueryDefinition
	resourceName := "aws_cloudwatch_query_definition.test"
	queryName := acctest.RandomWithPrefix("tf-acc-test")

	expectedQueryString := `fields @timestamp, @message
| sort @timestamp desc
| limit 20
`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudwatchlogs.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSCloudWatchQueryDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchQueryDefinitionConfig_Basic(queryName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloudWatchQueryDefinitionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", queryName),
					resource.TestCheckResourceAttr(resourceName, "query_string", expectedQueryString),
					resource.TestCheckResourceAttr(resourceName, "log_group_names.#", "0"),
					resource.TestMatchResourceAttr(resourceName, "query_definition_id", regexp.MustCompile(uuidRegexPattern)),
				),
			}, {
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSCloudWatchQueryDefinition_disappears(t *testing.T) {
	var v cloudwatchlogs.QueryDefinition
	resourceName := "aws_cloudwatch_query_definition.test"
	queryName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, cloudwatchlogs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchQueryDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchQueryDefinitionConfig_Basic(queryName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloudWatchQueryDefinitionExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCloudWatchQueryDefinition(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSCloudWatchQueryDefinition_Rename(t *testing.T) {
	var v cloudwatchlogs.QueryDefinition
	resourceName := "aws_cloudwatch_query_definition.test"
	queryName := acctest.RandomWithPrefix("tf-acc-test")
	updatedQueryName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudwatchlogs.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSCloudWatchQueryDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchQueryDefinitionConfig_Basic(queryName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloudWatchQueryDefinitionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", queryName),
				),
			}, {
				Config: testAccAWSCloudWatchQueryDefinitionConfig_Basic(updatedQueryName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloudWatchQueryDefinitionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", updatedQueryName),
				),
			},
		},
	})
}

func testAccCheckAWSCloudWatchQueryDefinitionExists(rName string, v *cloudwatchlogs.QueryDefinition) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rName]
		if !ok {
			return fmt.Errorf("Not found: %s", rName)
		}
		conn := testAccProvider.Meta().(*AWSClient).cloudwatchlogsconn

		result, err := finder.QueryDefinitionByResourceID(context.Background(), conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if result == nil {
			return fmt.Errorf("CloudWatch query definition (%s) not found", rs.Primary.ID)
		}

		*v = *result

		return nil
	}
}

func testAccCheckAWSCloudWatchQueryDefinitionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cloudwatchlogsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudwatch_query_definition" {
			continue
		}

		output, err := finder.QueryDefinitionByResourceID(context.Background(), conn, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error reading CloudWatch query definition (%s): %w", rs.Primary.ID, err)
		}

		if output != nil {
			return fmt.Errorf("CloudWatch query definition (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccAWSCloudWatchQueryDefinitionConfig_Basic(name string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_query_definition" "test" {
  name = %[1]q

  query_string = <<EOF
fields @timestamp, @message
| sort @timestamp desc
| limit 20
EOF
}
`, name)
}
