// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logs_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflogs "github.com/hashicorp/terraform-provider-aws/internal/service/logs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func TestAccLogsQueryDefinition_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v cloudwatchlogs.QueryDefinition
	resourceName := "aws_cloudwatch_query_definition.test"
	queryName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	expectedQueryString := `fields @timestamp, @message
| sort @timestamp desc
| limit 20
`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudwatchlogs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueryDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueryDefinitionConfig_basic(queryName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueryDefinitionExists(ctx, resourceName, &v),
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
	ctx := acctest.Context(t)
	var v cloudwatchlogs.QueryDefinition
	resourceName := "aws_cloudwatch_query_definition.test"
	queryName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudwatchlogs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueryDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueryDefinitionConfig_basic(queryName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueryDefinitionExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tflogs.ResourceQueryDefinition(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLogsQueryDefinition_rename(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 cloudwatchlogs.QueryDefinition
	resourceName := "aws_cloudwatch_query_definition.test"
	queryName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	updatedQueryName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudwatchlogs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueryDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueryDefinitionConfig_basic(queryName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueryDefinitionExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "name", queryName),
				),
			},
			{
				Config: testAccQueryDefinitionConfig_basic(updatedQueryName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueryDefinitionExists(ctx, resourceName, &v2),
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
	ctx := acctest.Context(t)
	var v1, v2 cloudwatchlogs.QueryDefinition
	resourceName := "aws_cloudwatch_query_definition.test"
	queryName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudwatchlogs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueryDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueryDefinitionConfig_logGroups(queryName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueryDefinitionExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "name", queryName),
					resource.TestCheckResourceAttr(resourceName, "log_group_names.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "log_group_names.0", "aws_cloudwatch_log_group.test.0", "name"),
				),
			},
			{
				Config: testAccQueryDefinitionConfig_logGroups(queryName, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueryDefinitionExists(ctx, resourceName, &v2),
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

func testAccCheckQueryDefinitionExists(ctx context.Context, n string, v *cloudwatchlogs.QueryDefinition) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No CloudWatch Logs Query Definition ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LogsConn(ctx)

		output, err := tflogs.FindQueryDefinitionByTwoPartKey(ctx, conn, rs.Primary.Attributes["name"], rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckQueryDefinitionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LogsConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudwatch_query_definition" {
				continue
			}

			_, err := tflogs.FindQueryDefinitionByTwoPartKey(ctx, conn, rs.Primary.Attributes["name"], rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudWatch Logs Query Definition still exists: %s", rs.Primary.ID)
		}

		return nil
	}
}

func testAccQueryDefinitionConfig_basic(rName string) string {
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

func testAccQueryDefinitionConfig_logGroups(rName string, count int) string {
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
