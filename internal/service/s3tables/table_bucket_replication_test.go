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

func TestAccS3TablesTableBucketReplication_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3tables.GetTableBucketReplicationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3tables_table_bucket_replication.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3TablesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableBucketReplicationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTableBucketReplicationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableBucketReplicationExists(ctx, t, resourceName, &v),
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
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "table_bucket_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "table_bucket_arn",
				ImportStateVerifyIgnore:              []string{"version_token"},
			},
		},
	})
}

func TestAccS3TablesTableBucketReplication_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3tables.GetTableBucketReplicationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3tables_table_bucket_replication.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3TablesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableBucketReplicationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTableBucketReplicationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableBucketReplicationExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfs3tables.ResourceTableBucketReplication, resourceName),
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

func TestAccS3TablesTableBucketReplication_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3tables.GetTableBucketReplicationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3tables_table_bucket_replication.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3TablesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableBucketReplicationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTableBucketReplicationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableBucketReplicationExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: testAccTableBucketReplicationConfig_updated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableBucketReplicationExists(ctx, t, resourceName, &v),
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

func testAccCheckTableBucketReplicationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).S3TablesClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3tables_table_bucket_replication" {
				continue
			}

			_, err := tfs3tables.FindTableBucketReplicationByARN(ctx, conn, rs.Primary.Attributes["table_bucket_arn"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Tables Table Bucket Replication %s still exists", rs.Primary.Attributes["table_bucket_arn"])
		}

		return nil
	}
}

func testAccCheckTableBucketReplicationExists(ctx context.Context, t *testing.T, n string, v *s3tables.GetTableBucketReplicationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).S3TablesClient(ctx)

		output, err := tfs3tables.FindTableBucketReplicationByARN(ctx, conn, rs.Primary.Attributes["table_bucket_arn"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccTableBucketReplicationConfig_base(rName string) string {
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
`, rName)
}

func testAccTableBucketReplicationConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccTableBucketReplicationConfig_base(rName), `
resource "aws_s3tables_table_bucket_replication" "test" {
  table_bucket_arn = aws_s3tables_table_bucket.source.arn
  role             = aws_iam_role.test.arn

  rule {
    destination {
      destination_table_bucket_arn = aws_s3tables_table_bucket.target.arn
    }
  }
}
`)
}

func testAccTableBucketReplicationConfig_updated(rName string) string {
	return acctest.ConfigCompose(testAccTableBucketReplicationConfig_base(rName), fmt.Sprintf(`
resource "aws_s3tables_table_bucket" "target2" {
  name = "%[1]s-target2"
}

resource "aws_s3tables_table_bucket_replication" "test" {
  table_bucket_arn = aws_s3tables_table_bucket.source.arn
  role             = aws_iam_role.test.arn

  rule {
    destination {
      destination_table_bucket_arn = aws_s3tables_table_bucket.target2.arn
    }
  }
}
`, rName))
}
