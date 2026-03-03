// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package guardduty_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfguardduty "github.com/hashicorp/terraform-provider-aws/internal/service/guardduty"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccIPSet_basic(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := fmt.Sprintf("tf-test-%s", sdkacctest.RandString(5))
	keyName1 := fmt.Sprintf("tf-%s", sdkacctest.RandString(5))
	keyName2 := fmt.Sprintf("tf-%s", sdkacctest.RandString(5))
	ipsetName1 := fmt.Sprintf("tf-%s", sdkacctest.RandString(5))
	ipsetName2 := fmt.Sprintf("tf-%s", sdkacctest.RandString(5))
	resourceName := "aws_guardduty_ipset.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GuardDutyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIPSetConfig_basic(bucketName, keyName1, ipsetName1, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPSetExists(ctx, t, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "guardduty", regexache.MustCompile("detector/.+/ipset/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ipsetName1),
					resource.TestCheckResourceAttr(resourceName, "activate", acctest.CtTrue),
					resource.TestMatchResourceAttr(resourceName, names.AttrLocation, regexache.MustCompile(fmt.Sprintf("%s/%s$", bucketName, keyName1))),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccIPSetConfig_basic(bucketName, keyName2, ipsetName2, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPSetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ipsetName2),
					resource.TestCheckResourceAttr(resourceName, "activate", acctest.CtFalse),
					resource.TestMatchResourceAttr(resourceName, names.AttrLocation, regexache.MustCompile(fmt.Sprintf("%s/%s$", bucketName, keyName2))),
				),
			},
		},
	})
}

func testAccCheckIPSetDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).GuardDutyClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_guardduty_ipset" {
				continue
			}

			_, err := tfguardduty.FindIPSetByTwoPartKey(ctx, conn, rs.Primary.Attributes["detector_id"], rs.Primary.Attributes["ip_set_id"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("GuardDuty IPSet %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckIPSetExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).GuardDutyClient(ctx)

		_, err := tfguardduty.FindIPSetByTwoPartKey(ctx, conn, rs.Primary.Attributes["detector_id"], rs.Primary.Attributes["ip_set_id"])

		return err
	}
}

func testAccIPSetConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_bucket_ownership_controls" "test" {
  bucket = aws_s3_bucket.test.id
  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}

resource "aws_s3_bucket_public_access_block" "test" {
  bucket = aws_s3_bucket.test.id

  block_public_acls       = false
  block_public_policy     = false
  ignore_public_acls      = false
  restrict_public_buckets = false
}

resource "aws_s3_bucket_acl" "test" {
  depends_on = [
    aws_s3_bucket_ownership_controls.test,
    aws_s3_bucket_public_access_block.test,
  ]

  bucket = aws_s3_bucket.test.id
  acl    = "public-read"
}
`, rName)
}

func testAccIPSetConfig_basic(bucketName, keyName, ipsetName string, activate bool) string {
	return acctest.ConfigCompose(testAccIPSetConfig_base(bucketName), fmt.Sprintf(`
resource "aws_guardduty_detector" "test" {}

resource "aws_s3_object" "test" {
  acl     = "public-read"
  content = "10.0.0.0/8\n"
  bucket  = aws_s3_bucket.test.id
  key     = %[1]q

  depends_on = [
    aws_s3_bucket_acl.test,
  ]
}

resource "aws_guardduty_ipset" "test" {
  name        = %[2]q
  detector_id = aws_guardduty_detector.test.id
  format      = "TXT"
  location    = "https://s3.amazonaws.com/${aws_s3_object.test.bucket}/${aws_s3_object.test.key}"
  activate    = %[3]t
}
`, keyName, ipsetName, activate))
}
