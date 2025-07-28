// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccSpotDataFeedSubscriptionDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_spot_datafeed_subscription.test"
	dataSourceName := "data.aws_spot_datafeed_subscription.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotDatafeedSubscriptionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotDataFeedSubscriptionDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrBucket, resourceName, names.AttrBucket),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrPrefix, resourceName, names.AttrPrefix),
				),
			},
		},
	})
}

func testAccSpotDataFeedSubscriptionDataSourceConfig_basic(rName string) string {
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

data "aws_spot_datafeed_subscription" "test" {
  depends_on = [aws_spot_datafeed_subscription.test]
}`, rName)
}
