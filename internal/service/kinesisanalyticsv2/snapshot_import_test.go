// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package kinesisanalyticsv2_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kinesisanalyticsv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kinesisanalyticsv2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfkinesisanalyticsv2 "github.com/hashicorp/terraform-provider-aws/internal/service/kinesisanalyticsv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// TestAccKinesisAnalyticsV2SnapshotImport_basic exports an application's snapshot to S3,
// imports it as a new snapshot, and starts a second application from the import.
func TestAccKinesisAnalyticsV2SnapshotImport_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.SnapshotInfoDetails
	resourceName := "aws_kinesisanalyticsv2_snapshot_import.test"
	restoredResourceName := "aws_kinesisanalyticsv2_application.restored"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSnapshotDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotImportConfig_source(rName),
				Check:  testAccCheckSnapshotExported(ctx, t, "aws_kinesisanalyticsv2_snapshot.test", "aws_iam_role.transfer", "aws_s3_bucket.transfer"),
			},
			{
				Config: testAccSnapshotImportConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSnapshotExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "kinesisanalytics", "snapshot/"+rName+"-imported"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationTime),
					resource.TestCheckResourceAttr(resourceName, names.AttrFormat, "CANONICAL"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName+"-imported"),
					resource.TestCheckResourceAttrSet(resourceName, "size_in_bytes"),
					resource.TestCheckResourceAttr(restoredResourceName, names.AttrStatus, "RUNNING"),
					resource.TestCheckResourceAttrPair(restoredResourceName, "application_configuration.0.run_configuration.0.application_restore_configuration.0.snapshot_arn", resourceName, names.AttrARN),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportStateVerifyIgnore:              []string{"destination_encryption_configuration", "import_source_configuration"},
			},
			// After a real import the unreadable source arguments are null in state.
			// Forget the snapshot, import it with an import block, and check the apply adopts the configuration rather than replacing the snapshot.
			{
				Config: testAccSnapshotImportConfig_forgotten(rName),
			},
			{
				Config: testAccSnapshotImportConfig_imported(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSnapshotExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "import_source_configuration.0.source_type", "S3"),
				),
			},
		},
	})
}

func TestAccKinesisAnalyticsV2SnapshotImport_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.SnapshotInfoDetails
	resourceName := "aws_kinesisanalyticsv2_snapshot_import.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSnapshotDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotImportConfig_source(rName),
				Check:  testAccCheckSnapshotExported(ctx, t, "aws_kinesisanalyticsv2_snapshot.test", "aws_iam_role.transfer", "aws_s3_bucket.transfer"),
			},
			{
				Config: testAccSnapshotImportConfig_importOnly(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSnapshotExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfkinesisanalyticsv2.ResourceSnapshotImport, resourceName),
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

// testAccCheckSnapshotExported exports a snapshot to the root of a bucket and waits for the export to succeed.
// Terraform does not manage exports, so the test drives the API directly.
func testAccCheckSnapshotExported(ctx context.Context, t *testing.T, snapshotResourceName, roleResourceName, bucketResourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		attr := func(n, key string) (string, error) {
			rs, ok := s.RootModule().Resources[n]
			if !ok {
				return "", fmt.Errorf("Not found: %s", n)
			}
			return rs.Primary.Attributes[key], nil
		}

		snapshotName, err := attr(snapshotResourceName, names.AttrName)
		if err != nil {
			return err
		}
		roleARN, err := attr(roleResourceName, names.AttrARN)
		if err != nil {
			return err
		}
		bucketARN, err := attr(bucketResourceName, names.AttrARN)
		if err != nil {
			return err
		}

		conn := acctest.ProviderMeta(ctx, t).KinesisAnalyticsV2Client(ctx)

		input := kinesisanalyticsv2.ExportSnapshotInput{
			SnapshotName: aws.String(snapshotName),
			ExportDestinationConfiguration: &awstypes.ExportDestinationConfiguration{
				DestinationType: awstypes.StorageLocationTypeS3,
				ExportRoleArn:   aws.String(roleARN),
				DestinationEncryptionConfiguration: &awstypes.SnapshotEncryptionConfiguration{
					KeyType: awstypes.KeyTypeAwsOwnedKey,
				},
				DestinationS3PrefixLocation: &awstypes.S3PrefixLocation{
					BucketARN: aws.String(bucketARN),
					KeyPrefix: aws.String("exports/"),
				},
			},
		}

		output, err := conn.ExportSnapshot(ctx, &input)
		if err != nil {
			return fmt.Errorf("exporting Kinesis Analytics v2 Snapshot (%s): %w", snapshotName, err)
		}

		operationID := aws.ToString(output.SnapshotOperationId)
		stateConf := &retry.StateChangeConf{
			Pending: enum.Slice(awstypes.SnapshotOperationStatusInProgress),
			Target:  enum.Slice(awstypes.SnapshotOperationStatusSuccessful),
			Refresh: func(ctx context.Context) (any, string, error) {
				input := kinesisanalyticsv2.DescribeSnapshotOperationInput{
					SnapshotName:        aws.String(snapshotName),
					SnapshotOperationId: aws.String(operationID),
				}
				output, err := conn.DescribeSnapshotOperation(ctx, &input)
				if err != nil {
					return nil, "", err
				}
				return output, string(output.SnapshotOperationInfoDetails.SnapshotOperationStatus), nil
			},
			Timeout: 30 * time.Minute,
			Delay:   10 * time.Second,
		}

		if _, err := stateConf.WaitForStateContext(ctx); err != nil {
			return fmt.Errorf("waiting for Kinesis Analytics v2 Snapshot (%s) export (%s): %w", snapshotName, operationID, err)
		}

		return nil
	}
}

// testAccSnapshotImportConfig_source is a running application with a snapshot to export,
// and a bucket and role for the export and the import.
func testAccSnapshotImportConfig_source(rName string) string {
	return acctest.ConfigCompose(testAccSnapshotConfig_basic(rName), fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_s3_bucket" "transfer" {
  bucket        = "%[1]s-transfer"
  force_destroy = true
}

resource "aws_iam_role" "transfer" {
  name = "%[1]s-transfer"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Service = "kinesisanalytics.amazonaws.com" }
      Action    = "sts:AssumeRole"
      Condition = {
        StringEquals = { "aws:SourceAccount" = data.aws_caller_identity.current.account_id }
      }
    }]
  })
}

resource "aws_iam_role_policy" "transfer" {
  role = aws_iam_role.transfer.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = ["s3:GetObject", "s3:GetObjectAttributes", "s3:PutObject", "s3:AbortMultipartUpload", "s3:ListMultipartUploadParts"]
        Resource = "${aws_s3_bucket.transfer.arn}/*"
      },
      {
        Effect   = "Allow"
        Action   = ["s3:ListBucket", "s3:ListBucketMultipartUploads"]
        Resource = aws_s3_bucket.transfer.arn
      },
    ]
  })
}
`, rName))
}

// testAccSnapshotImportConfig_exportedPrefix finds the directory an export wrote, which is named for the export operation.
func testAccSnapshotImportConfig_exportedPrefix() string {
	return `
data "aws_s3_objects" "exported" {
  bucket = aws_s3_bucket.transfer.bucket
  prefix = "exports/"
}

locals {
  exported_prefix = one([for k in data.aws_s3_objects.exported.keys : trimsuffix(k, "_metadata") if endswith(k, "/_metadata")])
}
`
}

func testAccSnapshotImportConfig_importOnly(rName string) string {
	return acctest.ConfigCompose(
		testAccSnapshotImportConfig_source(rName),
		testAccSnapshotImportConfig_exportedPrefix(),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_snapshot_import" "test" {
  name = "%[1]s-imported"

  import_source_configuration {
    source_type     = "S3"
    import_role_arn = aws_iam_role.transfer.arn

    source_s3_prefix_location {
      bucket_arn = aws_s3_bucket.transfer.arn
      key_prefix = local.exported_prefix
    }
  }

  depends_on = [aws_iam_role_policy.transfer]
}
`, rName))
}

// testAccSnapshotImportConfig_forgotten removes the imported snapshot from state without deleting it.
// The application restored from it is destroyed, since it references the snapshot.
func testAccSnapshotImportConfig_forgotten(rName string) string {
	return acctest.ConfigCompose(
		testAccSnapshotImportConfig_source(rName),
		testAccSnapshotImportConfig_exportedPrefix(),
		`
removed {
  from = aws_kinesisanalyticsv2_snapshot_import.test

  lifecycle {
    destroy = false
  }
}
`)
}

func testAccSnapshotImportConfig_imported(rName string) string {
	return acctest.ConfigCompose(testAccSnapshotImportConfig_importOnly(rName), fmt.Sprintf(`
import {
  to = aws_kinesisanalyticsv2_snapshot_import.test
  id = "%[1]s-imported"
}
`, rName))
}

func testAccSnapshotImportConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccSnapshotImportConfig_importOnly(rName),
		testAccSnapshotConfig_enhancedFlinkApplication("restored", rName+"-restored", `
        application_restore_type = "RESTORE_FROM_CUSTOM_SNAPSHOT"
        snapshot_arn             = aws_kinesisanalyticsv2_snapshot_import.test.arn
`),
	)
}
