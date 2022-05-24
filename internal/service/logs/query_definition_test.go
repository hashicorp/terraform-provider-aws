package logs_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflogs "github.com/hashicorp/terraform-provider-aws/internal/service/logs"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func TestAccLogsQueryDefinition_basic(t *testing.T) {
	var v cloudwatchlogs.QueryDefinition
	resourceName := "aws_cloudwatch_query_definition.test"
	queryName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	expectedQueryString := `fields @timestamp, @message
| sort @timestamp desc
| limit 20
`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudwatchlogs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckQueryDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccQueryDefinitionConfig_Basic(queryName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueryDefinitionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", queryName),
					resource.TestCheckResourceAttr(resourceName, "query_string", expectedQueryString),
					resource.TestCheckResourceAttr(resourceName, "log_group_names.#", "0"),
					resource.TestMatchResourceAttr(resourceName, "query_definition_id", regexp.MustCompile(verify.UUIDRegexPattern)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccQueryDefinitionImportStateID(&v),
			},
		},
	})
}

func testAccQueryDefinitionImportStateID(v *cloudwatchlogs.QueryDefinition) resource.ImportStateIdFunc {
	return func(*terraform.State) (string, error) {
		id := arn.ARN{
			AccountID: acctest.AccountID(),
			Partition: acctest.Partition(),
			Region:    acctest.Region(),
			Service:   cloudwatchlogs.ServiceName,
			Resource:  fmt.Sprintf("query-definition:%s", aws.StringValue(v.QueryDefinitionId)),
		}

		return id.String(), nil
	}
}

func TestAccLogsQueryDefinition_disappears(t *testing.T) {
	var v cloudwatchlogs.QueryDefinition
	resourceName := "aws_cloudwatch_query_definition.test"
	queryName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudwatchlogs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckQueryDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccQueryDefinitionConfig_Basic(queryName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueryDefinitionExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tflogs.ResourceQueryDefinition(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLogsQueryDefinition_rename(t *testing.T) {
	var v1, v2 cloudwatchlogs.QueryDefinition
	resourceName := "aws_cloudwatch_query_definition.test"
	queryName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	updatedQueryName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudwatchlogs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckQueryDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccQueryDefinitionConfig_Basic(queryName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueryDefinitionExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "name", queryName),
				),
			},
			{
				Config: testAccQueryDefinitionConfig_Basic(updatedQueryName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueryDefinitionExists(resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "name", updatedQueryName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccQueryDefinitionImportStateID(&v2),
			},
		},
	})
}

func TestAccLogsQueryDefinition_logGroups(t *testing.T) {
	var v1, v2 cloudwatchlogs.QueryDefinition
	resourceName := "aws_cloudwatch_query_definition.test"
	queryName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudwatchlogs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckQueryDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccQueryDefinitionConfig_LogGroups(queryName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueryDefinitionExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "name", queryName),
					resource.TestCheckResourceAttr(resourceName, "log_group_names.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "log_group_names.0", "aws_cloudwatch_log_group.test.0", "name"),
				),
			},
			{
				Config: testAccQueryDefinitionConfig_LogGroups(queryName, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueryDefinitionExists(resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "name", queryName),
					resource.TestCheckResourceAttr(resourceName, "log_group_names.#", "5"),
					resource.TestCheckResourceAttrPair(resourceName, "log_group_names.0", "aws_cloudwatch_log_group.test.0", "name"),
					resource.TestCheckResourceAttrPair(resourceName, "log_group_names.1", "aws_cloudwatch_log_group.test.1", "name"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccQueryDefinitionImportStateID(&v2),
			},
		},
	})
}

func testAccCheckQueryDefinitionExists(rName string, v *cloudwatchlogs.QueryDefinition) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rName]
		if !ok {
			return fmt.Errorf("Not found: %s", rName)
		}
		conn := acctest.Provider.Meta().(*conns.AWSClient).LogsConn

		result, err := tflogs.FindQueryDefinition(context.Background(), conn, "", rs.Primary.ID)

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

func testAccCheckQueryDefinitionDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).LogsConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudwatch_query_definition" {
			continue
		}

		result, err := tflogs.FindQueryDefinition(context.Background(), conn, "", rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error reading CloudWatch query definition (%s): %w", rs.Primary.ID, err)
		}

		if result != nil {
			return fmt.Errorf("CloudWatch query definition (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccQueryDefinitionConfig_Basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_query_definition" "test" {
  name = %[1]q

  query_string = <<EOF
fields @timestamp, @message
| sort @timestamp desc
| limit 20
EOF
}
`, rName)
}

func testAccQueryDefinitionConfig_LogGroups(rName string, count int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_query_definition" "test" {
  name = %[1]q

  log_group_names = aws_cloudwatch_log_group.test[*].name

  query_string = <<EOF
fields @timestamp, @message
| sort @timestamp desc
| limit 20
EOF
}

resource "aws_cloudwatch_log_group" "test" {
  count = %[2]d

  name = "%[1]s-${count.index}"
}
`, rName, count)
}
