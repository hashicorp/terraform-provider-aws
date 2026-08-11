// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3files_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/service/s3files"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3FilesFileSystemPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_s3files_file_system_policy.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3FilesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileSystemPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFileSystemPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystemPolicyExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrFileSystemID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccFileSystemPolicyImportStateIdFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: names.AttrFileSystemID,
			},
		},
	})
}

func TestAccS3FilesFileSystemPolicy_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_s3files_file_system_policy.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3FilesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileSystemPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFileSystemPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystemPolicyExists(ctx, t, resourceName),
				),
			},
			{
				Config: testAccFileSystemPolicyConfig_updated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystemPolicyExists(ctx, t, resourceName),
				),
			},
		},
	})
}

func testAccCheckFileSystemPolicyExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).S3FilesClient(ctx)

		_, err := s3files.FindFileSystemPolicyByID(ctx, conn, rs.Primary.Attributes[names.AttrFileSystemID])

		return err
	}
}

func testAccFileSystemPolicyImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.Attributes[names.AttrFileSystemID], nil
	}
}

func testAccCheckFileSystemPolicyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).S3FilesClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3files_file_system_policy" {
				continue
			}

			_, err := s3files.FindFileSystemPolicyByID(ctx, conn, rs.Primary.Attributes[names.AttrFileSystemID])

			if err == nil {
				return fmt.Errorf("S3 Files File System Policy %s still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccFileSystemPolicyConfig_base(rName string) string {
	return acctest.ConfigCompose(
		testAccFileSystemConfig_base(rName),
		`
resource "aws_s3files_file_system" "test" {
  bucket   = aws_s3_bucket.test.arn
  role_arn = aws_iam_role.test.arn

  depends_on = [aws_s3_bucket_versioning.test]
}
`)
}

func testAccFileSystemPolicyConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccFileSystemPolicyConfig_base(rName),
		`
resource "aws_s3files_file_system_policy" "test" {
  file_system_id = aws_s3files_file_system.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          AWS = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
        }
        Action   = "s3files:ClientMount"
        Resource = "*"
      }
    ]
  })
}
`)
}

func testAccFileSystemPolicyConfig_updated(rName string) string {
	return acctest.ConfigCompose(
		testAccFileSystemPolicyConfig_base(rName),
		`
resource "aws_s3files_file_system_policy" "test" {
  file_system_id = aws_s3files_file_system.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          AWS = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
        }
        Action   = ["s3files:ClientMount", "s3files:ClientWrite"]
        Resource = "*"
      }
    ]
  })
}
`)
}
