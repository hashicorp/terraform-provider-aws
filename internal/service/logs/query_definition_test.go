// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package logs_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tflogs "github.com/hashicorp/terraform-provider-aws/internal/service/logs"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLogsQueryDefinition_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.QueryDefinition
	resourceName := "aws_cloudwatch_query_definition.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	expectedQueryString := `fields @timestamp, @message
| sort @timestamp desc
| limit 20
`

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueryDefinitionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/QueryDefinition/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueryDefinitionExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "query_string", expectedQueryString),
					resource.TestCheckResourceAttr(resourceName, "log_group_names.#", "0"),
					resource.TestMatchResourceAttr(resourceName, "query_definition_id", regexache.MustCompile(verify.UUIDRegexPattern)),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/QueryDefinition/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccLogsQueryDefinition_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.QueryDefinition
	resourceName := "aws_cloudwatch_query_definition.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueryDefinitionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/QueryDefinition/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueryDefinitionExists(ctx, t, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tflogs.ResourceQueryDefinition(), resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccLogsQueryDefinition_rename(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.QueryDefinition
	resourceName := "aws_cloudwatch_query_definition.test"
	queryName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	updatedQueryName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueryDefinitionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccQueryDefinitionConfig_basic(queryName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueryDefinitionExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, queryName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: testAccQueryDefinitionConfig_basic(updatedQueryName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueryDefinitionExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, updatedQueryName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccLogsQueryDefinition_logGroups(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.QueryDefinition
	resourceName := "aws_cloudwatch_query_definition.test"
	queryName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueryDefinitionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccQueryDefinitionConfig_logGroups(queryName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueryDefinitionExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, queryName),
					resource.TestCheckResourceAttr(resourceName, "log_group_names.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "log_group_names.0", "aws_cloudwatch_log_group.test.0", names.AttrName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: testAccQueryDefinitionConfig_logGroups(queryName, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueryDefinitionExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, queryName),
					resource.TestCheckResourceAttr(resourceName, "log_group_names.#", "5"),
					resource.TestCheckResourceAttrPair(resourceName, "log_group_names.0", "aws_cloudwatch_log_group.test.0", names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, "log_group_names.1", "aws_cloudwatch_log_group.test.1", names.AttrName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccLogsQueryDefinition_logGroupARNs(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.QueryDefinition
	resourceName := "aws_cloudwatch_query_definition.test"
	queryName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueryDefinitionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccQueryDefinitionConfig_logGroupARNs(queryName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueryDefinitionExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, queryName),
					resource.TestCheckResourceAttr(resourceName, "log_group_names.#", "2"),
					resource.TestCheckResourceAttrPair(resourceName, "log_group_names.0", "aws_cloudwatch_log_group.test.0", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "log_group_names.1", "aws_cloudwatch_log_group.test.1", names.AttrARN),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func testAccCheckQueryDefinitionExists(ctx context.Context, t *testing.T, n string, v *types.QueryDefinition) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).LogsClient(ctx)

		output, err := tflogs.FindQueryDefinitionByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrName], rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckQueryDefinitionDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).LogsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudwatch_query_definition" {
				continue
			}

			_, err := tflogs.FindQueryDefinitionByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrName], rs.Primary.ID)

			if retry.NotFound(err) {
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

func testAccQueryDefinitionConfig_logGroupARNs(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_query_definition" "test" {
  name = %[1]q

  log_group_names = [
    aws_cloudwatch_log_group.test[0].arn,
    aws_cloudwatch_log_group.test[1].arn,
  ]

  query_string = <<EOF
fields @timestamp, @message
| sort @timestamp desc
| limit 20
EOF
}

resource "aws_cloudwatch_log_group" "test" {
  count = 2

  name = "%[1]s-${count.index}"
}
`, rName)
}
