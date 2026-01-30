// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3tables_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3tables"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfs3tables "github.com/hashicorp/terraform-provider-aws/internal/service/s3tables"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3TablesTableReplication_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3tables.GetTableReplicationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3tables_table_replication.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3TablesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableReplicationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTableReplicationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableReplicationExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "table_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "table_arn",
				ImportStateVerifyIgnore:              []string{"version_token"},
			},
		},
	})
}

func TestAccS3TablesTableReplication_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3tables.GetTableReplicationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3tables_table_replication.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3TablesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableReplicationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTableReplicationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableReplicationExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfs3tables.ResourceTableReplication, resourceName),
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

func TestAccS3TablesTableReplication_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3tables.GetTableReplicationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3tables_table_replication.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3TablesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableReplicationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTableReplicationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableReplicationExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: testAccTableReplicationConfig_updated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableReplicationExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func testAccCheckTableReplicationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).S3TablesClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3tables_table_replication" {
				continue
			}

			_, err := tfs3tables.FindTableReplicationByARN(ctx, conn, rs.Primary.Attributes["table_arn"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Tables Table Bucket Replication %s still exists", rs.Primary.Attributes["table_arn"])
		}

		return nil
	}
}

func testAccCheckTableReplicationExists(ctx context.Context, t *testing.T, n string, v *s3tables.GetTableReplicationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).S3TablesClient(ctx)

		output, err := tfs3tables.FindTableReplicationByARN(ctx, conn, rs.Primary.Attributes["table_arn"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccTableReplicationConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}
data "aws_service_principal" "current" {
  service_name = "s3"
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "${data.aws_service_principal.current.name}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_s3tables_table_bucket" "source" {
  name = "%[1]s-source"
}

resource "aws_s3tables_table_bucket" "target" {
  name = "%[1]s-target"
}

resource "aws_s3tables_table" "test" {
  name             = replace(%[1]q, "-", "_")
  namespace        = aws_s3tables_namespace.test.namespace
  table_bucket_arn = aws_s3tables_namespace.test.table_bucket_arn
  format           = "ICEBERG"
}

resource "aws_s3tables_namespace" "test" {
  namespace        = replace(%[1]q, "-", "_")
  table_bucket_arn = aws_s3tables_table_bucket.source.arn

  lifecycle {
    create_before_destroy = true
  }
}
`, rName)
}

func testAccTableReplicationConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccTableReplicationConfig_base(rName), `
resource "aws_s3tables_table_replication" "test" {
  table_arn = aws_s3tables_table.test.arn
  role      = aws_iam_role.test.arn

  rule {
    destination {
      destination_table_bucket_arn = aws_s3tables_table_bucket.target.arn
    }
  }
}
`)
}

func testAccTableReplicationConfig_updated(rName string) string {
	return acctest.ConfigCompose(testAccTableReplicationConfig_base(rName), fmt.Sprintf(`
resource "aws_s3tables_table_bucket" "target2" {
  name = "%[1]s-target2"
}

resource "aws_s3tables_table_replication" "test" {
  table_arn = aws_s3tables_table.test.arn
  role      = aws_iam_role.test.arn

  rule {
    destination {
      destination_table_bucket_arn = aws_s3tables_table_bucket.target2.arn
    }
  }
}
`, rName))
}
