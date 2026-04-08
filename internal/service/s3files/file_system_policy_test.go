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

func TestAccS3FilesFileSystemPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3files.GetFileSystemPolicyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3files_file_system_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3FilesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileSystemPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFileSystemPolicyConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFileSystemPolicyExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
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

func TestAccS3FilesFileSystemPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3files.GetFileSystemPolicyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3files_file_system_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3FilesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileSystemPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFileSystemPolicyConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFileSystemPolicyExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfs3files.ResourceFileSystemPolicy, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckFileSystemPolicyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).S3FilesClient(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3files_file_system_policy" {
				continue
			}
			_, err := tfs3files.FindFileSystemPolicyByID(ctx, conn, rs.Primary.Attributes["file_system_id"])
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}
			return fmt.Errorf("S3 Files File System Policy %s still exists", rs.Primary.Attributes["file_system_id"])
		}
		return nil
	}
}

func testAccCheckFileSystemPolicyExists(ctx context.Context, t *testing.T, n string, v *s3files.GetFileSystemPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		conn := acctest.ProviderMeta(ctx, t).S3FilesClient(ctx)
		output, err := tfs3files.FindFileSystemPolicyByID(ctx, conn, rs.Primary.Attributes["file_system_id"])
		if err != nil {
			return err
		}
		*v = *output
		return nil
	}
}

func testAccFileSystemPolicyConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccFileSystemConfig_basic(rName), `
resource "aws_s3files_file_system_policy" "test" {
  file_system_id = aws_s3files_file_system.test.file_system_id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { AWS = "*" }
      Action    = ["s3files:ClientMount"]
      Resource  = aws_s3files_file_system.test.file_system_arn
    }]
  })
}
`)
}

func TestAccS3FilesFileSystemPolicy_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3files.GetFileSystemPolicyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3files_file_system_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3FilesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileSystemPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFileSystemPolicyConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFileSystemPolicyExists(ctx, t, resourceName, &v),
				),
			},
			{
				Config: testAccFileSystemPolicyConfig_updated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFileSystemPolicyExists(ctx, t, resourceName, &v),
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

func testAccFileSystemPolicyConfig_updated(rName string) string {
	return acctest.ConfigCompose(testAccFileSystemConfig_basic(rName), `
resource "aws_s3files_file_system_policy" "test" {
  file_system_id = aws_s3files_file_system.test.file_system_id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { AWS = "*" }
      Action    = ["s3files:ClientMount", "s3files:ClientWrite"]
      Resource  = aws_s3files_file_system.test.file_system_arn
    }]
  })
}
`)
}
