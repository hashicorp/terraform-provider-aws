// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"context"
	"fmt"
	"slices"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRDSExportTask_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var exportTask types.ExportTask
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_rds_export_task.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
		CheckDestroy: testAccCheckExportTaskDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExportTaskConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExportTaskExists(ctx, t, resourceName, &exportTask),
					resource.TestCheckResourceAttr(resourceName, "export_task_identifier", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrID, rName),
					resource.TestCheckResourceAttrPair(resourceName, "source_arn", "aws_db_snapshot.test", "db_snapshot_arn"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrS3BucketName, "aws_s3_bucket.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrIAMRoleARN, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyID, "aws_kms_key.test", names.AttrARN),
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

func TestAccRDSExportTask_optional(t *testing.T) {
	ctx := acctest.Context(t)
	var exportTask types.ExportTask
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_rds_export_task.test"
	s3Prefix := "test_prefix/test-export"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
		CheckDestroy: testAccCheckExportTaskDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExportTaskConfig_optional(rName, s3Prefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExportTaskExists(ctx, t, resourceName, &exportTask),
					resource.TestCheckResourceAttr(resourceName, "export_task_identifier", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrID, rName),
					resource.TestCheckResourceAttrPair(resourceName, "source_arn", "aws_db_snapshot.test", "db_snapshot_arn"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrS3BucketName, "aws_s3_bucket.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrIAMRoleARN, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyID, "aws_kms_key.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "export_only.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "export_only.0", names.AttrDatabase),
					resource.TestCheckResourceAttr(resourceName, "s3_prefix", s3Prefix),
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

func testAccCheckExportTaskDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).RDSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_rds_export_task" {
				continue
			}

			output, err := tfrds.FindExportTaskByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			// COMPLETE and FAILED statuses are valid because the resource is simply removed from
			// state in these scenarios. In-progress tasks should be cancelled upon destroy, so CANCELED
			// is also valid.
			if status := aws.ToString(output.Status); slices.Contains([]string{"COMPLETE", "FAILED", "CANCELED"}, status) {
				continue
			}

			return fmt.Errorf("RDS ExportTask %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckExportTaskExists(ctx context.Context, t *testing.T, n string, v *types.ExportTask) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).RDSClient(ctx)

		output, err := tfrds.FindExportTaskByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccExportTaskConfig_base(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigRandomPassword(),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "export.rds.amazonaws.com"
        }
      },
    ]
  })
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = [
      "s3:ListAllMyBuckets",
    ]
    resources = [
      "*"
    ]
  }
  statement {
    actions = [
      "s3:GetBucketLocation",
      "s3:ListBucket",
    ]
    resources = [
      aws_s3_bucket.test.arn,
    ]
  }
  statement {
    actions = [
      "s3:GetObject",
      "s3:PutObject",
      "s3:DeleteObject",
    ]
    resources = [
      "${aws_s3_bucket.test.arn}/*"
    ]
  }
}

resource "aws_iam_policy" "test" {
  name   = %[1]q
  policy = data.aws_iam_policy_document.test.json
}

resource "aws_iam_role_policy_attachment" "test-attach" {
  role       = aws_iam_role.test.name
  policy_arn = aws_iam_policy.test.arn
}

resource "aws_kms_key" "test" {
  deletion_window_in_days = 10
}

resource "aws_db_instance" "test" {
  identifier          = %[1]q
  allocated_storage   = 10
  db_name             = "test"
  engine              = "mysql"
  instance_class      = "db.t3.micro"
  username            = "foo"
  password_wo         = ephemeral.aws_secretsmanager_random_password.test.random_password
  password_wo_version = 1
  skip_final_snapshot = true
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.test.identifier
  db_snapshot_identifier = %[1]q
}
`, rName))
}

func testAccExportTaskConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccExportTaskConfig_base(rName),
		fmt.Sprintf(`
resource "aws_rds_export_task" "test" {
  export_task_identifier = %[1]q
  source_arn             = aws_db_snapshot.test.db_snapshot_arn
  s3_bucket_name         = aws_s3_bucket.test.id
  iam_role_arn           = aws_iam_role.test.arn
  kms_key_id             = aws_kms_key.test.arn
}
`, rName))
}

func testAccExportTaskConfig_optional(rName, s3Prefix string) string {
	return acctest.ConfigCompose(
		testAccExportTaskConfig_base(rName),
		fmt.Sprintf(`
resource "aws_rds_export_task" "test" {
  export_task_identifier = %[1]q
  source_arn             = aws_db_snapshot.test.db_snapshot_arn
  s3_bucket_name         = aws_s3_bucket.test.id
  iam_role_arn           = aws_iam_role.test.arn
  kms_key_id             = aws_kms_key.test.arn

  export_only = ["database"]
  s3_prefix   = %[2]q
}
`, rName, s3Prefix))
}
