// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3files_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/s3files"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3files/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfs3files "github.com/hashicorp/terraform-provider-aws/internal/service/s3files"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3FilesFileSystem_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var fileSystem s3files.GetFileSystemOutput
	resourceName := "aws_s3files_file_system.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3FilesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileSystemDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFileSystemConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystemExists(ctx, t, resourceName, &fileSystem),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "s3files", regexache.MustCompile(`file-system/.+`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.LifeCycleStateAvailable)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationTime),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrOwnerID),
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

func TestAccS3FilesFileSystem_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var fileSystem s3files.GetFileSystemOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3files_file_system.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3FilesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileSystemDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFileSystemConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystemExists(ctx, t, resourceName, &fileSystem),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfs3files.ResourceFileSystem, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3FilesFileSystem_kmsKey(t *testing.T) {
	ctx := acctest.Context(t)
	var fileSystem s3files.GetFileSystemOutput
	resourceName := "aws_s3files_file_system.test"
	kmsKeyResourceName := "aws_kms_key.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3FilesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileSystemDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFileSystemConfig_kmsKey(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystemExists(ctx, t, resourceName, &fileSystem),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyID, kmsKeyResourceName, names.AttrARN),
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

func TestAccS3FilesFileSystem_acceptBucketWarning(t *testing.T) {
	ctx := acctest.Context(t)
	var fileSystem s3files.GetFileSystemOutput
	resourceName := "aws_s3files_file_system.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3FilesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileSystemDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFileSystemConfig_acceptBucketWarning(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystemExists(ctx, t, resourceName, &fileSystem),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "s3files", regexache.MustCompile(`file-system/.+`)),
					resource.TestCheckResourceAttr(resourceName, "accept_bucket_warning", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.LifeCycleStateAvailable)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationTime),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrOwnerID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"accept_bucket_warning",
				},
			},
		},
	})
}

func testAccCheckFileSystemExists(ctx context.Context, t *testing.T, n string, v *s3files.GetFileSystemOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).S3FilesClient(ctx)

		output, err := tfs3files.FindFileSystemByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckFileSystemDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).S3FilesClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3files_file_system" {
				continue
			}

			_, err := tfs3files.FindFileSystemByID(ctx, conn, rs.Primary.ID)

			if err == nil {
				return fmt.Errorf("S3 Files File System %s still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccFileSystemConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}
data "aws_region" "current" {}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_versioning" "test" {
  bucket = aws_s3_bucket.test.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "AllowS3FilesAssumeRole"
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "elasticfilesystem.amazonaws.com"
        }
        Condition = {
          StringEquals = {
            "aws:SourceAccount" = data.aws_caller_identity.current.account_id
          }
          ArnLike = {
            "aws:SourceArn" = "arn:${data.aws_partition.current.partition}:s3files:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:file-system/*"
          }
        }
      }
    ]
  })
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "S3BucketPermissions"
        Effect = "Allow"
        Action = [
          "s3:ListBucket",
          "s3:ListBucketVersions"
        ]
        Resource = aws_s3_bucket.test.arn
        Condition = {
          StringEquals = {
            "aws:ResourceAccount" = data.aws_caller_identity.current.account_id
          }
        }
      },
      {
        Sid    = "S3ObjectPermissions"
        Effect = "Allow"
        Action = [
          "s3:AbortMultipartUpload",
          "s3:DeleteObject*",
          "s3:GetObject*",
          "s3:List*",
          "s3:PutObject*"
        ]
        Resource = "${aws_s3_bucket.test.arn}/*"
        Condition = {
          StringEquals = {
            "aws:ResourceAccount" = data.aws_caller_identity.current.account_id
          }
        }
      },
      {
        Sid    = "UseKmsKeyWithS3Files"
        Effect = "Allow"
        Action = [
          "kms:GenerateDataKey",
          "kms:Encrypt",
          "kms:Decrypt",
          "kms:ReEncryptFrom",
          "kms:ReEncryptTo"
        ]
        Resource = "arn:${data.aws_partition.current.partition}:kms:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:*"
        Condition = {
          StringLike = {
            "kms:ViaService" = "s3.${data.aws_region.current.name}.amazonaws.com"
            "kms:EncryptionContext:aws:s3:arn" = [
              aws_s3_bucket.test.arn,
              "${aws_s3_bucket.test.arn}/*"
            ]
          }
        }
      },
      {
        Sid    = "EventBridgeManage"
        Effect = "Allow"
        Action = [
          "events:DeleteRule",
          "events:DisableRule",
          "events:EnableRule",
          "events:PutRule",
          "events:PutTargets",
          "events:RemoveTargets"
        ]
        Resource = "arn:${data.aws_partition.current.partition}:events:*:*:rule/DO-NOT-DELETE-S3-Files*"
        Condition = {
          StringEquals = {
            "events:ManagedBy" = "elasticfilesystem.amazonaws.com"
          }
        }
      },
      {
        Sid    = "EventBridgeRead"
        Effect = "Allow"
        Action = [
          "events:DescribeRule",
          "events:ListRuleNamesByTarget",
          "events:ListRules",
          "events:ListTargetsByRule"
        ]
        Resource = "arn:${data.aws_partition.current.partition}:events:*:*:rule/*"
      }
    ]
  })
}
`, rName)
}

func testAccFileSystemConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccFileSystemConfig_base(rName),
		fmt.Sprintf(`
resource "aws_s3files_file_system" "test" {
  bucket   = aws_s3_bucket.test.arn
  role_arn = aws_iam_role.test.arn

  depends_on = [aws_s3_bucket_versioning.test]

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccFileSystemConfig_kmsKey(rName string) string {
	return acctest.ConfigCompose(
		testAccFileSystemConfig_base(rName),
		fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_s3files_file_system" "test" {
  bucket     = aws_s3_bucket.test.arn
  role_arn   = aws_iam_role.test.arn
  kms_key_id = aws_kms_key.test.arn

  depends_on = [aws_s3_bucket_versioning.test]

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccFileSystemConfig_acceptBucketWarning(rName string) string {
	return acctest.ConfigCompose(
		testAccFileSystemConfig_base(rName),
		fmt.Sprintf(`
resource "aws_s3files_file_system" "test" {
  bucket   = aws_s3_bucket.test.arn
  role_arn = aws_iam_role.test.arn

  accept_bucket_warning = true

  depends_on = [aws_s3_bucket_versioning.test]

  tags = {
    Name = %[1]q
  }
}
`, rName))
}
