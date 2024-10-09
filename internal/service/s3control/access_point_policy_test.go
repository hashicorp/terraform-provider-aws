// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3control_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3control "github.com/hashicorp/terraform-provider-aws/internal/service/s3control"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3ControlAccessPointPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_s3control_access_point_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPointPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPointPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "has_public_access_policy", acctest.CtFalse),
					resource.TestMatchResourceAttr(resourceName, names.AttrPolicy, regexache.MustCompile(`s3:GetObjectTagging`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAccessPointPolicyImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccS3ControlAccessPointPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_s3control_access_point_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPointPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPointPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointPolicyExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfs3control.ResourceAccessPointPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3ControlAccessPointPolicy_disappears_AccessPoint(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_s3control_access_point_policy.test"
	accessPointResourceName := "aws_s3_access_point.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPointPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPointPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointPolicyExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfs3control.ResourceAccessPoint(), accessPointResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3ControlAccessPointPolicy_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_s3control_access_point_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPointPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPointPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "has_public_access_policy", acctest.CtFalse),
					resource.TestMatchResourceAttr(resourceName, names.AttrPolicy, regexache.MustCompile(`s3:GetObjectTagging`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAccessPointPolicyImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccAccessPointPolicyConfig_updated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "has_public_access_policy", acctest.CtFalse),
					resource.TestMatchResourceAttr(resourceName, names.AttrPolicy, regexache.MustCompile(`s3:GetObjectLegalHold`)),
				),
			},
		},
	})
}

func testAccAccessPointPolicyImportStateIdFunc(n string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return "", fmt.Errorf("Not found: %s", n)
		}

		return rs.Primary.Attributes["access_point_arn"], nil
	}
}

func testAccCheckAccessPointPolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3ControlClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3control_access_point_policy" {
				continue
			}

			accountID, name, err := tfs3control.AccessPointParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, _, err = tfs3control.FindAccessPointPolicyAndStatusByTwoPartKey(ctx, conn, accountID, name)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Access Point Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAccessPointPolicyExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		accountID, name, err := tfs3control.AccessPointParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3ControlClient(ctx)

		_, _, err = tfs3control.FindAccessPointPolicyAndStatusByTwoPartKey(ctx, conn, accountID, name)

		return err
	}
}

func testAccAccessPointPolicyConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_access_point" "test" {
  bucket = aws_s3_bucket.test.id
  name   = %[1]q

  public_access_block_configuration {
    block_public_acls       = true
    block_public_policy     = false
    ignore_public_acls      = true
    restrict_public_buckets = false
  }

  lifecycle {
    ignore_changes = [policy]
  }
}
`, rName)
}

func testAccAccessPointPolicyConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccAccessPointPolicyConfig_base(rName), `
resource "aws_s3control_access_point_policy" "test" {
  access_point_arn = aws_s3_access_point.test.arn

  policy = jsonencode({
    Version = "2008-10-17"
    Statement = [{
      Effect = "Allow"
      Action = "s3:GetObjectTagging"
      Principal = {
        AWS = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
      }
      Resource = "${aws_s3_access_point.test.arn}/object/*"
    }]
  })
}
`)
}

func testAccAccessPointPolicyConfig_updated(rName string) string {
	return acctest.ConfigCompose(testAccAccessPointPolicyConfig_base(rName), `
resource "aws_s3control_access_point_policy" "test" {
  access_point_arn = aws_s3_access_point.test.arn

  policy = jsonencode({
    Version = "2008-10-17"
    Statement = [{
      Effect = "Allow"
      Action = [
        "s3:GetObjectLegalHold",
        "s3:GetObjectRetention",
      ]
      Principal = {
        AWS = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
      }
      Resource = "${aws_s3_access_point.test.arn}/object/prefix/*"
    }]
  })
}
`)
}
