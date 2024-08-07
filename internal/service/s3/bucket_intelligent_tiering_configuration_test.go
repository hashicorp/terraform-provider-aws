// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3BucketIntelligentTieringConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var itc types.IntelligentTieringConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_intelligent_tiering_configuration.test"
	bucketResourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketIntelligentTieringConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketIntelligentTieringConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketIntelligentTieringConfigurationExists(ctx, resourceName, &itc),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, bucketResourceName, names.AttrBucket),
					resource.TestCheckResourceAttr(resourceName, "filter.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "Enabled"),
					resource.TestCheckResourceAttr(resourceName, "tiering.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tiering.0.access_tier", "DEEP_ARCHIVE_ACCESS"),
					resource.TestCheckResourceAttr(resourceName, "tiering.0.days", "180"),
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

func TestAccS3BucketIntelligentTieringConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var itc types.IntelligentTieringConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_intelligent_tiering_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketIntelligentTieringConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketIntelligentTieringConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketIntelligentTieringConfigurationExists(ctx, resourceName, &itc),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfs3.ResourceBucketIntelligentTieringConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3BucketIntelligentTieringConfiguration_Filter(t *testing.T) {
	ctx := acctest.Context(t)
	var itc types.IntelligentTieringConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_intelligent_tiering_configuration.test"
	bucketResourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketIntelligentTieringConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketIntelligentTieringConfigurationConfig_filterPrefix(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketIntelligentTieringConfigurationExists(ctx, resourceName, &itc),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, bucketResourceName, names.AttrBucket),
					resource.TestCheckResourceAttr(resourceName, "filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", "p1/"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "Disabled"),
					resource.TestCheckResourceAttr(resourceName, "tiering.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "tiering.*", map[string]string{
						"access_tier": "DEEP_ARCHIVE_ACCESS",
						"days":        "180",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketIntelligentTieringConfigurationConfig_filterPrefixAndTag(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketIntelligentTieringConfigurationExists(ctx, resourceName, &itc),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, bucketResourceName, names.AttrBucket),
					resource.TestCheckResourceAttr(resourceName, "filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", "p2/"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.Environment", "test"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "Enabled"),
					resource.TestCheckResourceAttr(resourceName, "tiering.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "tiering.*", map[string]string{
						"access_tier": "ARCHIVE_ACCESS",
						"days":        "90",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "tiering.*", map[string]string{
						"access_tier": "DEEP_ARCHIVE_ACCESS",
						"days":        "360",
					}),
				),
			},
			{
				Config: testAccBucketIntelligentTieringConfigurationConfig_filterTag(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketIntelligentTieringConfigurationExists(ctx, resourceName, &itc),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, bucketResourceName, names.AttrBucket),
					resource.TestCheckResourceAttr(resourceName, "filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.Environment", "acctest"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "Disabled"),
					resource.TestCheckResourceAttr(resourceName, "tiering.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "tiering.*", map[string]string{
						"access_tier": "DEEP_ARCHIVE_ACCESS",
						"days":        "270",
					}),
				),
			},
			{
				Config: testAccBucketIntelligentTieringConfigurationConfig_filterPrefixAndTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketIntelligentTieringConfigurationExists(ctx, resourceName, &itc),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, bucketResourceName, names.AttrBucket),
					resource.TestCheckResourceAttr(resourceName, "filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", "p3/"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.Environment1", "test"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.Environment2", "acctest"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "Enabled"),
					resource.TestCheckResourceAttr(resourceName, "tiering.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "tiering.*", map[string]string{
						"access_tier": "ARCHIVE_ACCESS",
						"days":        "365",
					}),
				),
			},
			{
				Config: testAccBucketIntelligentTieringConfigurationConfig_filterTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketIntelligentTieringConfigurationExists(ctx, resourceName, &itc),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, bucketResourceName, names.AttrBucket),
					resource.TestCheckResourceAttr(resourceName, "filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.Environment1", "acctest"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.Environment2", "test"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "Enabled"),
					resource.TestCheckResourceAttr(resourceName, "tiering.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "tiering.*", map[string]string{
						"access_tier": "DEEP_ARCHIVE_ACCESS",
						"days":        "365",
					}),
				),
			},
		},
	})
}

func TestAccS3BucketIntelligentTieringConfiguration_directoryBucket(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketIntelligentTieringConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccBucketIntelligentTieringConfigurationConfig_directoryBucket(rName),
				ExpectError: regexache.MustCompile(`directory buckets are not supported`),
			},
		},
	})
}

func testAccCheckBucketIntelligentTieringConfigurationExists(ctx context.Context, n string, v *types.IntelligentTieringConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		bucket, name, err := tfs3.BucketIntelligentTieringConfigurationParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)

		output, err := tfs3.FindIntelligentTieringConfiguration(ctx, conn, bucket, name)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckBucketIntelligentTieringConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3_bucket_intelligent_tiering_configuration" {
				continue
			}

			bucket, name, err := tfs3.BucketIntelligentTieringConfigurationParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = tfs3.FindIntelligentTieringConfiguration(ctx, conn, bucket, name)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Intelligent-Tiering Configuration %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccBucketIntelligentTieringConfigurationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_intelligent_tiering_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %[1]q

  tiering {
    access_tier = "DEEP_ARCHIVE_ACCESS"
    days        = 180
  }
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}
`, rName)
}

func testAccBucketIntelligentTieringConfigurationConfig_filterPrefix(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_intelligent_tiering_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %[1]q

  status = "Disabled"

  filter {
    prefix = "p1/"
  }

  tiering {
    access_tier = "DEEP_ARCHIVE_ACCESS"
    days        = 180
  }
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}
`, rName)
}

func testAccBucketIntelligentTieringConfigurationConfig_filterPrefixAndTag(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_intelligent_tiering_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %[1]q

  status = "Enabled"

  filter {
    prefix = "p2/"

    tags = {
      Environment = "test"
    }
  }

  tiering {
    access_tier = "ARCHIVE_ACCESS"
    days        = 90
  }

  tiering {
    access_tier = "DEEP_ARCHIVE_ACCESS"
    days        = 360
  }
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}
`, rName)
}

func testAccBucketIntelligentTieringConfigurationConfig_filterTag(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_intelligent_tiering_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %[1]q

  status = "Disabled"

  filter {
    tags = {
      Environment = "acctest"
    }
  }

  tiering {
    access_tier = "DEEP_ARCHIVE_ACCESS"
    days        = 270
  }
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}
`, rName)
}

func testAccBucketIntelligentTieringConfigurationConfig_filterPrefixAndTags(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_intelligent_tiering_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %[1]q

  filter {
    prefix = "p3/"

    tags = {
      Environment1 = "test"
      Environment2 = "acctest"
    }
  }

  tiering {
    access_tier = "ARCHIVE_ACCESS"
    days        = 365
  }
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}
`, rName)
}

func testAccBucketIntelligentTieringConfigurationConfig_filterTags(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_intelligent_tiering_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %[1]q

  filter {
    tags = {
      Environment1 = "acctest"
      Environment2 = "test"
    }
  }

  tiering {
    access_tier = "DEEP_ARCHIVE_ACCESS"
    days        = 365
  }
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}
`, rName)
}

func testAccBucketIntelligentTieringConfigurationConfig_directoryBucket(rName string) string {
	return acctest.ConfigCompose(testAccDirectoryBucketConfig_base(rName), fmt.Sprintf(`
resource "aws_s3_directory_bucket" "test" {
  bucket = local.bucket

  location {
    name = local.location_name
  }
}

resource "aws_s3_bucket_intelligent_tiering_configuration" "test" {
  bucket = aws_s3_directory_bucket.test.bucket
  name   = %[1]q

  tiering {
    access_tier = "DEEP_ARCHIVE_ACCESS"
    days        = 180
  }
}
`, rName))
}
