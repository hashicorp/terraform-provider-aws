// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3files_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3files"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfs3files "github.com/hashicorp/terraform-provider-aws/internal/service/s3files"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3FilesFileSystem_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3files.GetFileSystemOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3files_file_system.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3FilesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileSystemDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFileSystemConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFileSystemExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "file_system_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "file_system_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"accept_bucket_warning"},
			},
		},
	})
}

func TestAccS3FilesFileSystem_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3files.GetFileSystemOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3files_file_system.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3FilesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileSystemDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFileSystemConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFileSystemExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfs3files.ResourceFileSystem, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckFileSystemDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).S3FilesClient(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3files_file_system" {
				continue
			}
			_, err := tfs3files.FindFileSystemByID(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}
			return fmt.Errorf("S3 Files File System %s still exists", rs.Primary.ID)
		}
		return nil
	}
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

func testAccFileSystemConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "s3files.amazonaws.com"
      }
    }]
  })
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = ["s3:*"]
      Resource = [aws_s3_bucket.test.arn, "${aws_s3_bucket.test.arn}/*"]
    }]
  })
}
`, rName)
}

func testAccFileSystemConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccFileSystemConfig_base(rName), `
resource "aws_s3files_file_system" "test" {
  bucket   = aws_s3_bucket.test.arn
  role_arn = aws_iam_role.test.arn

  depends_on = [aws_iam_role_policy.test]
}
`)
}

func TestAccS3FilesFileSystem_prefix(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3files.GetFileSystemOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3files_file_system.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3FilesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileSystemDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFileSystemConfig_prefix(rName, "data/"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFileSystemExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "prefix", "data/"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"accept_bucket_warning"},
			},
		},
	})
}

func TestAccS3FilesFileSystem_kmsKeyId(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3files.GetFileSystemOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3files_file_system.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3FilesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileSystemDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFileSystemConfig_kmsKeyId(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFileSystemExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "kms_key_id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"accept_bucket_warning"},
			},
		},
	})
}

func testAccFileSystemConfig_prefix(rName, prefix string) string {
	return acctest.ConfigCompose(testAccFileSystemConfig_base(rName), fmt.Sprintf(`
resource "aws_s3files_file_system" "test" {
  bucket   = aws_s3_bucket.test.arn
  role_arn = aws_iam_role.test.arn
  prefix   = %[1]q

  depends_on = [aws_iam_role_policy.test]
}
`, prefix))
}

func testAccFileSystemConfig_kmsKeyId(rName string) string {
	return acctest.ConfigCompose(testAccFileSystemConfig_base(rName), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description = %[1]q
}

resource "aws_s3files_file_system" "test" {
  bucket     = aws_s3_bucket.test.arn
  role_arn   = aws_iam_role.test.arn
  kms_key_id = aws_kms_key.test.arn

  depends_on = [aws_iam_role_policy.test]
}
`, rName))
}
