// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package guardduty_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/guardduty"
	awstypes "github.com/aws/aws-sdk-go-v2/service/guardduty/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfguardduty "github.com/hashicorp/terraform-provider-aws/internal/service/guardduty"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccThreatIntelSet_basic(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := fmt.Sprintf("tf-test-%s", sdkacctest.RandString(5))
	keyName1 := fmt.Sprintf("tf-%s", sdkacctest.RandString(5))
	keyName2 := fmt.Sprintf("tf-%s", sdkacctest.RandString(5))
	threatintelsetName1 := fmt.Sprintf("tf-%s", sdkacctest.RandString(5))
	threatintelsetName2 := fmt.Sprintf("tf-%s", sdkacctest.RandString(5))
	resourceName := "aws_guardduty_threatintelset.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GuardDutyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckThreatIntelSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccThreatIntelSetConfig_basic(bucketName, keyName1, threatintelsetName1, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThreatIntelSetExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "guardduty", regexache.MustCompile("detector/.+/threatintelset/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, threatintelsetName1),
					resource.TestCheckResourceAttr(resourceName, "activate", acctest.CtTrue),
					resource.TestMatchResourceAttr(resourceName, names.AttrLocation, regexache.MustCompile(fmt.Sprintf("%s/%s$", bucketName, keyName1))),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccThreatIntelSetConfig_basic(bucketName, keyName2, threatintelsetName2, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThreatIntelSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, threatintelsetName2),
					resource.TestCheckResourceAttr(resourceName, "activate", acctest.CtFalse),
					resource.TestMatchResourceAttr(resourceName, names.AttrLocation, regexache.MustCompile(fmt.Sprintf("%s/%s$", bucketName, keyName2))),
				),
			},
		},
	})
}

func testAccThreatIntelSet_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_guardduty_threatintelset.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GuardDutyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckThreatIntelSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccThreatIntelSetConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThreatIntelSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccThreatIntelSetConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThreatIntelSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccThreatIntelSetConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThreatIntelSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckThreatIntelSetDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).GuardDutyClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_guardduty_threatintelset" {
				continue
			}

			threatIntelSetId, detectorId, err := tfguardduty.DecodeThreatIntelSetID(rs.Primary.ID)
			if err != nil {
				return err
			}
			input := &guardduty.GetThreatIntelSetInput{
				ThreatIntelSetId: aws.String(threatIntelSetId),
				DetectorId:       aws.String(detectorId),
			}

			resp, err := conn.GetThreatIntelSet(ctx, input)
			if err != nil {
				if errs.IsAErrorMessageContains[*awstypes.BadRequestException](err, "The request is rejected because the input detectorId is not owned by the current account.") {
					return nil
				}
				return err
			}

			if resp.Status == awstypes.ThreatIntelSetStatusDeletePending || resp.Status == awstypes.ThreatIntelSetStatusDeleted {
				return nil
			}

			return fmt.Errorf("Expected GuardDuty ThreatIntelSet to be destroyed, %s found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckThreatIntelSetExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		threatIntelSetId, detectorId, err := tfguardduty.DecodeThreatIntelSetID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &guardduty.GetThreatIntelSetInput{
			DetectorId:       aws.String(detectorId),
			ThreatIntelSetId: aws.String(threatIntelSetId),
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GuardDutyClient(ctx)
		_, err = conn.GetThreatIntelSet(ctx, input)
		return err
	}
}

func testAccThreatIntelSetConfig_base(rName string) string {
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

func testAccThreatIntelSetConfig_basic(bucketName, keyName, threatintelsetName string, activate bool) string {
	return acctest.ConfigCompose(testAccThreatIntelSetConfig_base(bucketName),
		fmt.Sprintf(`
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

resource "aws_guardduty_threatintelset" "test" {
  name        = %[2]q
  detector_id = aws_guardduty_detector.test.id
  format      = "TXT"
  location    = "https://s3.amazonaws.com/${aws_s3_object.test.bucket}/${aws_s3_object.test.key}"
  activate    = %[3]t
}
`, keyName, threatintelsetName, activate))
}

func testAccThreatIntelSetConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccThreatIntelSetConfig_base(rName),
		fmt.Sprintf(`
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

resource "aws_guardduty_threatintelset" "test" {
  activate    = true
  detector_id = aws_guardduty_detector.test.id
  format      = "TXT"
  location    = "https://s3.amazonaws.com/${aws_s3_object.test.bucket}/${aws_s3_object.test.key}"
  name        = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccThreatIntelSetConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccThreatIntelSetConfig_base(rName),
		fmt.Sprintf(`
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

resource "aws_guardduty_threatintelset" "test" {
  activate    = true
  detector_id = aws_guardduty_detector.test.id
  format      = "TXT"
  location    = "https://s3.amazonaws.com/${aws_s3_object.test.bucket}/${aws_s3_object.test.key}"
  name        = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
