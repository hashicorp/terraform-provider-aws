// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3control_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/s3control"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3control "github.com/hashicorp/terraform-provider-aws/internal/service/s3control"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccS3ControlBucket_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3control_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, s3control.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "s3-outposts", regexp.MustCompile(fmt.Sprintf("outpost/[^/]+/bucket/%s", rName))),
					resource.TestCheckResourceAttr(resourceName, "bucket", rName),
					resource.TestCheckResourceAttrSet(resourceName, "creation_date"),
					resource.TestCheckResourceAttrPair(resourceName, "outpost_id", "data.aws_outposts_outpost.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccS3ControlBucket_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3control_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, s3control.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfs3control.ResourceBucket(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3ControlBucket_tags(t *testing.T) {
	acctest.Skip(t, "S3 Control Bucket resource tagging requires additional eventual consistency handling, see also: https://github.com/hashicorp/terraform-provider-aws/issues/15572")
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3control_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, s3control.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccBucketConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckBucketDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3ControlConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3control_bucket" {
				continue
			}

			parsedArn, err := arn.Parse(rs.Primary.ID)

			if err != nil {
				return err
			}

			_, err = tfs3control.FindBucketByTwoPartKey(ctx, conn, parsedArn.AccountID, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Control Bucket %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckBucketExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No S3 Control Bucket ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3ControlConn(ctx)

		parsedArn, err := arn.Parse(rs.Primary.ID)

		if err != nil {
			return err
		}

		_, err = tfs3control.FindBucketByTwoPartKey(ctx, conn, parsedArn.AccountID, rs.Primary.ID)

		return err
	}
}

func testAccBucketConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

resource "aws_s3control_bucket" "test" {
  bucket     = %[1]q
  outpost_id = data.aws_outposts_outpost.test.id
}
`, rName)
}

func testAccBucketConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

resource "aws_s3control_bucket" "test" {
  bucket     = %[1]q
  outpost_id = data.aws_outposts_outpost.test.id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccBucketConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

resource "aws_s3control_bucket" "test" {
  bucket     = %[1]q
  outpost_id = data.aws_outposts_outpost.test.id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
