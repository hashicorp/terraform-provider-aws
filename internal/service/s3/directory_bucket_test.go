// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3_test

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
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3DirectoryBucket_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_directory_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectoryBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryBucketConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDirectoryBucketExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "s3express", regexache.MustCompile(fmt.Sprintf(`bucket/%s--.*-x-s3`, rName))),
					resource.TestCheckResourceAttr(resourceName, "data_redundancy", "SingleAvailabilityZone"),
					resource.TestCheckResourceAttr(resourceName, "location.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "location.0.name"),
					resource.TestCheckResourceAttr(resourceName, "location.0.type", "AvailabilityZone"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "Directory"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
		},
	})
}

func TestAccS3DirectoryBucket_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_directory_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectoryBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryBucketConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDirectoryBucketExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfs3.ResourceDirectoryBucket, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3DirectoryBucket_forceDestroy(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_directory_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectoryBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryBucketConfig_forceDestroy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDirectoryBucketExists(ctx, resourceName),
					testAccCheckBucketAddObjects(ctx, resourceName, "data.txt", "prefix/more_data.txt"),
				),
			},
		},
	})
}

func testAccCheckDirectoryBucketDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3ExpressClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3_directory_bucket" {
				continue
			}

			err := tfs3.FindBucket(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Bucket %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDirectoryBucketExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3ExpressClient(ctx)

		return tfs3.FindBucket(ctx, conn, rs.Primary.ID)
	}
}

func testAccConfigAvailableAZsDirectoryBucket() string {
	// https://docs.aws.amazon.com/AmazonS3/latest/userguide/s3-express-networking.html#s3-express-endpoints.
	return acctest.ConfigAvailableAZsNoOptInExclude("use1-az1", "use1-az2", "use1-az3", "usw2-az2", "apne1-az2")
}

func testAccDirectoryBucketConfig_base(rName string) string {
	return acctest.ConfigCompose(testAccConfigAvailableAZsDirectoryBucket(), fmt.Sprintf(`
locals {
  location_name = data.aws_availability_zones.available.zone_ids[0]
  bucket        = "%[1]s--${local.location_name}--x-s3"
}
`, rName))
}

func testAccDirectoryBucketConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccDirectoryBucketConfig_base(rName), `
resource "aws_s3_directory_bucket" "test" {
  bucket = local.bucket

  location {
    name = local.location_name
  }
}
`)
}

func testAccDirectoryBucketConfig_forceDestroy(rName string) string {
	return acctest.ConfigCompose(testAccDirectoryBucketConfig_base(rName), `
resource "aws_s3_directory_bucket" "test" {
  bucket = local.bucket

  location {
    name = local.location_name
  }

  force_destroy = true
}
`)
}
