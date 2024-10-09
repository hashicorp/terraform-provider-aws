// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2SpotDatafeedSubscription_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:      testAccSpotDatafeedSubscription_basic,
		acctest.CtDisappears: testAccSpotDatafeedSubscription_disappears,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccSpotDatafeedSubscription_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var subscription awstypes.SpotDatafeedSubscription
	resourceName := "aws_spot_datafeed_subscription.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSpotDatafeedSubscription(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotDatafeedSubscriptionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotDatafeedSubscriptionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSpotDatafeedSubscriptionExists(ctx, resourceName, &subscription),
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

func testAccSpotDatafeedSubscription_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var subscription awstypes.SpotDatafeedSubscription
	resourceName := "aws_spot_datafeed_subscription.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSpotDatafeedSubscription(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotDatafeedSubscriptionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotDatafeedSubscriptionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSpotDatafeedSubscriptionExists(ctx, resourceName, &subscription),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceSpotDataFeedSubscription(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSpotDatafeedSubscriptionExists(ctx context.Context, n string, v *awstypes.SpotDatafeedSubscription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Spot Datafeed Subscription ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindSpotDatafeedSubscription(ctx, conn)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckSpotDatafeedSubscriptionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_spot_datafeed_subscription" {
				continue
			}

			_, err := tfec2.FindSpotDatafeedSubscription(ctx, conn)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 Spot Datafeed Subscription %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccPreCheckSpotDatafeedSubscription(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

	_, err := conn.DescribeSpotDatafeedSubscription(ctx, &ec2.DescribeSpotDatafeedSubscriptionInput{})

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInvalidSpotDatafeedNotFound) {
		return
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccSpotDatafeedSubscriptionConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_canonical_user_id" "current" {}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_public_access_block" "test" {
  bucket = aws_s3_bucket.test.bucket

  block_public_acls       = false
  block_public_policy     = false
  ignore_public_acls      = false
  restrict_public_buckets = false
}

resource "aws_s3_bucket_ownership_controls" "test" {
  bucket = aws_s3_bucket.test.bucket
  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  access_control_policy {
    grant {
      grantee {
        id   = data.aws_canonical_user_id.current.id
        type = "CanonicalUser"
      }
      permission = "FULL_CONTROL"
    }

    grant {
      grantee {
        id   = "c4c1ede66af53448b93c283ce9448c4ba468c9432aa01d700d3878632f77d2d0" # EC2 Account
        type = "CanonicalUser"
      }
      permission = "FULL_CONTROL"
    }

    owner {
      id = data.aws_canonical_user_id.current.id
    }
  }
  depends_on = [
    aws_s3_bucket_public_access_block.test,
    aws_s3_bucket_ownership_controls.test
  ]
}

resource "aws_spot_datafeed_subscription" "test" {
  # Must have bucket grants configured
  depends_on = [aws_s3_bucket_acl.test]

  bucket = aws_s3_bucket.test.bucket
}
`, rName)
}
