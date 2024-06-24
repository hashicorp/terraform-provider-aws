// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3control_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3control/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3control "github.com/hashicorp/terraform-provider-aws/internal/service/s3control"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3ControlMultiRegionAccessPoint_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.MultiRegionAccessPointReport
	resourceName := "aws_s3control_multi_region_access_point.test"
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionNot(t, names.USGovCloudPartitionID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMultiRegionAccessPointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMultiRegionAccessPointConfig_basic(bucketName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMultiRegionAccessPointExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrAccountID),
					resource.TestMatchResourceAttr(resourceName, names.AttrAlias, regexache.MustCompile(`^[a-z][0-9a-z]*[.]mrap$`)),
					acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrARN, "s3", regexache.MustCompile(`accesspoint\/[a-z][0-9a-z]*[.]mrap$`)),
					acctest.MatchResourceAttrGlobalHostname(resourceName, names.AttrDomainName, "accesspoint.s3-global", regexache.MustCompile(`^[a-z][0-9a-z]*[.]mrap`)),
					resource.TestCheckResourceAttr(resourceName, "details.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "details.0.name", rName),
					resource.TestCheckResourceAttr(resourceName, "details.0.public_access_block.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "details.0.public_access_block.0.block_public_acls", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "details.0.public_access_block.0.block_public_policy", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "details.0.public_access_block.0.ignore_public_acls", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "details.0.public_access_block.0.restrict_public_buckets", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "details.0.region.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "details.0.region.*", map[string]string{
						names.AttrBucket:    bucketName,
						"bucket_account_id": acctest.AccountID(),
						names.AttrRegion:    acctest.Region(),
					}),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.MultiRegionAccessPointStatusReady)),
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

func TestAccS3ControlMultiRegionAccessPoint_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.MultiRegionAccessPointReport
	resourceName := "aws_s3control_multi_region_access_point.test"
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionNot(t, names.USGovCloudPartitionID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMultiRegionAccessPointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMultiRegionAccessPointConfig_basic(bucketName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiRegionAccessPointExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfs3control.ResourceMultiRegionAccessPoint(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3ControlMultiRegionAccessPoint_PublicAccessBlock(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.MultiRegionAccessPointReport
	resourceName := "aws_s3control_multi_region_access_point.test"
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionNot(t, names.USGovCloudPartitionID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMultiRegionAccessPointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMultiRegionAccessPointConfig_publicBlock(bucketName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiRegionAccessPointExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "details.0.public_access_block.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "details.0.public_access_block.0.block_public_acls", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "details.0.public_access_block.0.block_public_policy", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "details.0.public_access_block.0.ignore_public_acls", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "details.0.public_access_block.0.restrict_public_buckets", acctest.CtFalse),
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

func TestAccS3ControlMultiRegionAccessPoint_name(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 types.MultiRegionAccessPointReport
	resourceName := "aws_s3control_multi_region_access_point.test"
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionNot(t, names.USGovCloudPartitionID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMultiRegionAccessPointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMultiRegionAccessPointConfig_basic(bucketName, rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiRegionAccessPointExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "details.0.name", rName1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMultiRegionAccessPointConfig_basic(bucketName, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiRegionAccessPointExists(ctx, resourceName, &v2),
					testAccCheckMultiRegionAccessPointRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "details.0.name", rName2),
				),
			},
		},
	})
}

func TestAccS3ControlMultiRegionAccessPoint_threeRegions(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.MultiRegionAccessPointReport
	resourceName := "aws_s3control_multi_region_access_point.test"
	bucket1Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucket2Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucket3Name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 3)
			acctest.PreCheckPartitionNot(t, names.USGovCloudPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 3),
		CheckDestroy:             testAccCheckMultiRegionAccessPointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMultiRegionAccessPointConfig_three(bucket1Name, bucket2Name, bucket3Name, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiRegionAccessPointExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "details.0.region.#", acctest.Ct3),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "details.0.region.*", map[string]string{
						names.AttrBucket: bucket1Name,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "details.0.region.*", map[string]string{
						names.AttrBucket: bucket2Name,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "details.0.region.*", map[string]string{
						names.AttrBucket: bucket3Name,
					}),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.MultiRegionAccessPointStatusReady)),
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

func TestAccS3ControlMultiRegionAccessPoint_putAndGetObject(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.MultiRegionAccessPointReport
	resourceName := "aws_s3control_multi_region_access_point.test"
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionNot(t, names.USGovCloudPartitionID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMultiRegionAccessPointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMultiRegionAccessPointConfig_putAndGetObject(bucketName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiRegionAccessPointExists(ctx, resourceName, &v),
				),
			},
		},
	})
}

func testAccCheckMultiRegionAccessPointDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3ControlClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3control_multi_region_access_point" {
				continue
			}

			accountID, name, err := tfs3control.MultiRegionAccessPointParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = tfs3control.FindMultiRegionAccessPointByTwoPartKey(ctx, conn, accountID, name)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Multi-Region Access Point %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckMultiRegionAccessPointExists(ctx context.Context, n string, v *types.MultiRegionAccessPointReport) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		accountID, name, err := tfs3control.MultiRegionAccessPointParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3ControlClient(ctx)

		output, err := tfs3control.FindMultiRegionAccessPointByTwoPartKey(ctx, conn, accountID, name)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

// Multi-Region Access Point aliases are unique throughout time and aren’t based on the name or configuration of a Multi-Region Access Point.
// If you create a Multi-Region Access Point, and then delete it and create another one with the same name and configuration, the
// second Multi-Region Access Point will have a different alias than the first. (https://docs.aws.amazon.com/AmazonS3/latest/userguide/CreatingMultiRegionAccessPoints.html#multi-region-access-point-naming)
func testAccCheckMultiRegionAccessPointRecreated(before, after *types.MultiRegionAccessPointReport) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.Alias), aws.ToString(after.Alias); before == after {
			return fmt.Errorf("S3 Multi-Region Access Point (%s) not recreated", before)
		}

		return nil
	}
}

func testAccMultiRegionAccessPointConfig_basic(bucketName, multiRegionAccessPointName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3control_multi_region_access_point" "test" {
  details {
    name = %[2]q

    region {
      bucket = aws_s3_bucket.test.id
    }
  }
}
`, bucketName, multiRegionAccessPointName)
}

func testAccMultiRegionAccessPointConfig_publicBlock(bucketName, multiRegionAccessPointName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3control_multi_region_access_point" "test" {
  details {
    name = %[2]q

    public_access_block {
      block_public_acls       = false
      block_public_policy     = false
      ignore_public_acls      = false
      restrict_public_buckets = false
    }

    region {
      bucket = aws_s3_bucket.test.id
    }
  }
}
`, bucketName, multiRegionAccessPointName)
}

func testAccMultiRegionAccessPointConfig_three(bucketName1, bucketName2, bucketName3, multiRegionAccessPointName string) string {
	return acctest.ConfigCompose(acctest.ConfigMultipleRegionProvider(3), fmt.Sprintf(`
resource "aws_s3_bucket" "test1" {
  provider = aws

  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_bucket" "test2" {
  provider = awsalternate

  bucket        = %[2]q
  force_destroy = true
}

resource "aws_s3_bucket" "test3" {
  provider = awsthird

  bucket        = %[3]q
  force_destroy = true
}

resource "aws_s3control_multi_region_access_point" "test" {
  provider = aws

  details {
    name = %[4]q

    region {
      bucket = aws_s3_bucket.test1.id
    }

    region {
      bucket = aws_s3_bucket.test2.id
    }

    region {
      bucket = aws_s3_bucket.test3.id
    }
  }
}
`, bucketName1, bucketName2, bucketName3, multiRegionAccessPointName))
}

func testAccMultiRegionAccessPointConfig_putAndGetObject(bucketName, multiRegionAccessPointName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3control_multi_region_access_point" "test" {
  details {
    name = %[2]q

    region {
      bucket = aws_s3_bucket.test.id
    }
  }
}

resource "aws_s3_object" "test" {
  bucket  = aws_s3control_multi_region_access_point.test.arn
  key     = "%[1]s-key"
  content = "Hello World"

  tags = {
    Name = %[2]q
  }
}

# Ensure that we can GET through the bucket.
data "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.bucket
  key    = aws_s3_object.test.key
}
`, bucketName, multiRegionAccessPointName)
}
