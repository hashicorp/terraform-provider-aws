// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package logs_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/observabilityadmin"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tflogs "github.com/hashicorp/terraform-provider-aws/internal/service/logs"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccS3TableSourceAssociationPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).ObservabilityAdminClient(ctx)
	_, err := conn.ListS3TableIntegrations(ctx, &observabilityadmin.ListS3TableIntegrationsInput{})
	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func TestAccLogsS3TableSourceAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_s3_table_source_association.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccS3TableSourceAssociationPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckS3TableSourceAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccS3TableSourceAssociationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckS3TableSourceAssociationExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("datasource_name"), knownvalue.StringExact("*")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("datasource_type"), knownvalue.StringExact("*")),
				},
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("Not found: %s", resourceName)
					}
					return fmt.Sprintf("%s,%s", rs.Primary.Attributes["integration_arn"], rs.Primary.Attributes[names.AttrID]), nil
				},
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLogsS3TableSourceAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_s3_table_source_association.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccS3TableSourceAssociationPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckS3TableSourceAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccS3TableSourceAssociationConfig_specific(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckS3TableSourceAssociationExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tflogs.ResourceS3TableSourceAssociation, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccCheckS3TableSourceAssociationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).LogsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudwatch_log_s3_table_source_association" {
				continue
			}

			integrationARN := rs.Primary.Attributes["integration_arn"]
			identifier := rs.Primary.Attributes[names.AttrID]

			_, err := tflogs.FindS3TableSourceAssociationByTwoPartKey(ctx, conn, integrationARN, identifier)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudWatch Logs S3 Table Source Association %s still exists", identifier)
		}

		return nil
	}
}

func testAccCheckS3TableSourceAssociationExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).LogsClient(ctx)

		integrationARN := rs.Primary.Attributes["integration_arn"]
		identifier := rs.Primary.Attributes[names.AttrID]

		_, err := tflogs.FindS3TableSourceAssociationByTwoPartKey(ctx, conn, integrationARN, identifier)

		return err
	}
}

func testAccS3TableSourceAssociationConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "logs.amazonaws.com"
      }
    }]
  })
}

resource "aws_iam_role_policy" "test" {
  role = aws_iam_role.test.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3tables:CreateTableBucket",
          "s3tables:ListTableBuckets",
          "s3tables:GetTableBucket",
          "s3tables:CreateNamespace",
          "s3tables:GetNamespace",
          "s3tables:ListNamespaces",
          "s3tables:CreateTable",
          "s3tables:GetTable",
          "s3tables:ListTables",
          "s3tables:PutTableData",
          "s3tables:GetTableData",
        ]
        Resource = "*"
      },
    ]
  })
}

resource "aws_observabilityadmin_s3_table_integration" "test" {
  role_arn = aws_iam_role.test.arn

  encryption {
    sse_algorithm = "AES256"
  }
}
`, rName)
}

func testAccS3TableSourceAssociationConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccS3TableSourceAssociationConfig_base(rName),
		`
resource "aws_cloudwatch_log_s3_table_source_association" "test" {
  integration_arn = aws_observabilityadmin_s3_table_integration.test.arn
}
`,
	)
}

func testAccS3TableSourceAssociationConfig_specific(rName string) string {
	// datasource_name and datasource_type only allow letters, numbers, and underscores.
	safeName := strings.ReplaceAll(rName, "-", "_")
	return acctest.ConfigCompose(
		testAccS3TableSourceAssociationConfig_base(rName),
		fmt.Sprintf(`
resource "aws_cloudwatch_log_s3_table_source_association" "test" {
  integration_arn = aws_observabilityadmin_s3_table_integration.test.arn
  datasource_name = %[1]q
  datasource_type = %[1]q
}
`, safeName),
	)
}
