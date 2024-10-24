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

func TestAccS3BucketAnalyticsConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var ac types.AnalyticsConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_analytics_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketAnalyticsConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketAnalyticsConfigurationConfig_basic(rName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketAnalyticsConfigurationExists(ctx, resourceName, &ac),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, "aws_s3_bucket.test", names.AttrBucket),
					resource.TestCheckResourceAttr(resourceName, "filter.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.#", acctest.Ct0),
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

func TestAccS3BucketAnalyticsConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var ac types.AnalyticsConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_analytics_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketAnalyticsConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketAnalyticsConfigurationConfig_basic(rName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketAnalyticsConfigurationExists(ctx, resourceName, &ac),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfs3.ResourceBucketAnalyticsConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3BucketAnalyticsConfiguration_updateBasic(t *testing.T) {
	ctx := acctest.Context(t)
	var ac types.AnalyticsConfiguration
	originalACName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	originalBucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	updatedACName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	updatedBucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_analytics_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketAnalyticsConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketAnalyticsConfigurationConfig_basic(originalACName, originalBucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketAnalyticsConfigurationExists(ctx, resourceName, &ac),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, originalACName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, "aws_s3_bucket.test", names.AttrBucket),
					resource.TestCheckResourceAttr(resourceName, "filter.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.#", acctest.Ct0),
				),
			},
			{
				Config: testAccBucketAnalyticsConfigurationConfig_basic(updatedACName, originalBucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketAnalyticsConfigurationExists(ctx, resourceName, &ac),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, updatedACName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, "aws_s3_bucket.test", names.AttrBucket),
					resource.TestCheckResourceAttr(resourceName, "filter.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.#", acctest.Ct0),
				),
			},
			{
				Config: testAccBucketAnalyticsConfigurationConfig_update(updatedACName, originalBucketName, updatedBucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketAnalyticsConfigurationExists(ctx, resourceName, &ac),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, updatedACName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, "aws_s3_bucket.test_2", names.AttrBucket),
					resource.TestCheckResourceAttr(resourceName, "filter.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.#", acctest.Ct0),
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

func TestAccS3BucketAnalyticsConfiguration_WithFilter_empty(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketAnalyticsConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccBucketAnalyticsConfigurationConfig_emptyFilter(rName, rName),
				ExpectError: regexache.MustCompile(`one of .* must be specified`),
			},
		},
	})
}

func TestAccS3BucketAnalyticsConfiguration_WithFilter_prefix(t *testing.T) {
	ctx := acctest.Context(t)
	var ac types.AnalyticsConfiguration
	rInt := sdkacctest.RandInt()
	resourceName := "aws_s3_bucket_analytics_configuration.test"

	rName := fmt.Sprintf("tf-acc-test-%d", rInt)
	prefix := fmt.Sprintf("prefix-%d/", rInt)
	prefixUpdate := fmt.Sprintf("prefix-update-%d/", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketAnalyticsConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketAnalyticsConfigurationConfig_filterPrefix(rName, rName, prefix),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketAnalyticsConfigurationExists(ctx, resourceName, &ac),
					resource.TestCheckResourceAttr(resourceName, "filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", prefix),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", acctest.Ct0),
				),
			},
			{
				Config: testAccBucketAnalyticsConfigurationConfig_filterPrefix(rName, rName, prefixUpdate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketAnalyticsConfigurationExists(ctx, resourceName, &ac),
					resource.TestCheckResourceAttr(resourceName, "filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", prefixUpdate),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", acctest.Ct0),
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

func TestAccS3BucketAnalyticsConfiguration_WithFilter_singleTag(t *testing.T) {
	ctx := acctest.Context(t)
	var ac types.AnalyticsConfiguration
	rInt := sdkacctest.RandInt()
	resourceName := "aws_s3_bucket_analytics_configuration.test"

	rName := fmt.Sprintf("tf-acc-test-%d", rInt)
	tag1 := fmt.Sprintf("tag-%d", rInt)
	tag1Update := fmt.Sprintf("tag-update-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketAnalyticsConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketAnalyticsConfigurationConfig_filterSingleTag(rName, rName, tag1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketAnalyticsConfigurationExists(ctx, resourceName, &ac),
					resource.TestCheckResourceAttr(resourceName, "filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.tag1", tag1),
				),
			},
			{
				Config: testAccBucketAnalyticsConfigurationConfig_filterSingleTag(rName, rName, tag1Update),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketAnalyticsConfigurationExists(ctx, resourceName, &ac),
					resource.TestCheckResourceAttr(resourceName, "filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.tag1", tag1Update),
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

func TestAccS3BucketAnalyticsConfiguration_WithFilter_multipleTags(t *testing.T) {
	ctx := acctest.Context(t)
	var ac types.AnalyticsConfiguration
	rInt := sdkacctest.RandInt()
	resourceName := "aws_s3_bucket_analytics_configuration.test"

	rName := fmt.Sprintf("tf-acc-test-%d", rInt)
	tag1 := fmt.Sprintf("tag1-%d", rInt)
	tag1Update := fmt.Sprintf("tag1-update-%d", rInt)
	tag2 := fmt.Sprintf("tag2-%d", rInt)
	tag2Update := fmt.Sprintf("tag2-update-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketAnalyticsConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketAnalyticsConfigurationConfig_filterMultipleTags(rName, rName, tag1, tag2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketAnalyticsConfigurationExists(ctx, resourceName, &ac),
					resource.TestCheckResourceAttr(resourceName, "filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.tag1", tag1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.tag2", tag2),
				),
			},
			{
				Config: testAccBucketAnalyticsConfigurationConfig_filterMultipleTags(rName, rName, tag1Update, tag2Update),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketAnalyticsConfigurationExists(ctx, resourceName, &ac),
					resource.TestCheckResourceAttr(resourceName, "filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.tag1", tag1Update),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.tag2", tag2Update),
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

func TestAccS3BucketAnalyticsConfiguration_WithFilter_prefixAndTags(t *testing.T) {
	ctx := acctest.Context(t)
	var ac types.AnalyticsConfiguration
	rInt := sdkacctest.RandInt()
	resourceName := "aws_s3_bucket_analytics_configuration.test"

	rName := fmt.Sprintf("tf-acc-test-%d", rInt)
	prefix := fmt.Sprintf("prefix-%d/", rInt)
	prefixUpdate := fmt.Sprintf("prefix-update-%d/", rInt)
	tag1 := fmt.Sprintf("tag1-%d", rInt)
	tag1Update := fmt.Sprintf("tag1-update-%d", rInt)
	tag2 := fmt.Sprintf("tag2-%d", rInt)
	tag2Update := fmt.Sprintf("tag2-update-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketAnalyticsConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketAnalyticsConfigurationConfig_filterPrefixAndTags(rName, rName, prefix, tag1, tag2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketAnalyticsConfigurationExists(ctx, resourceName, &ac),
					resource.TestCheckResourceAttr(resourceName, "filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", prefix),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.tag1", tag1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.tag2", tag2),
				),
			},
			{
				Config: testAccBucketAnalyticsConfigurationConfig_filterPrefixAndTags(rName, rName, prefixUpdate, tag1Update, tag2Update),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketAnalyticsConfigurationExists(ctx, resourceName, &ac),
					resource.TestCheckResourceAttr(resourceName, "filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", prefixUpdate),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.tag1", tag1Update),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.tag2", tag2Update),
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

func TestAccS3BucketAnalyticsConfiguration_WithFilter_remove(t *testing.T) {
	ctx := acctest.Context(t)
	var ac types.AnalyticsConfiguration
	rInt := sdkacctest.RandInt()
	resourceName := "aws_s3_bucket_analytics_configuration.test"

	rName := fmt.Sprintf("tf-acc-test-%d", rInt)
	prefix := fmt.Sprintf("prefix-%d/", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketAnalyticsConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketAnalyticsConfigurationConfig_filterPrefix(rName, rName, prefix),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketAnalyticsConfigurationExists(ctx, resourceName, &ac),
				),
			},
			{
				Config: testAccBucketAnalyticsConfigurationConfig_basic(rName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketAnalyticsConfigurationExists(ctx, resourceName, &ac),
					resource.TestCheckResourceAttr(resourceName, "filter.#", acctest.Ct0),
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

func TestAccS3BucketAnalyticsConfiguration_WithStorageClassAnalysis_empty(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketAnalyticsConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccBucketAnalyticsConfigurationConfig_emptyStorageClassAnalysis(rName, rName),
				ExpectError: regexache.MustCompile(`Insufficient data_export blocks`),
			},
		},
	})
}

func TestAccS3BucketAnalyticsConfiguration_WithStorageClassAnalysis_default(t *testing.T) {
	ctx := acctest.Context(t)
	var ac types.AnalyticsConfiguration
	resourceName := "aws_s3_bucket_analytics_configuration.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketAnalyticsConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketAnalyticsConfigurationConfig_defaultStorageClassAnalysis(rName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketAnalyticsConfigurationExists(ctx, resourceName, &ac),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.0.data_export.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.0.data_export.0.output_schema_version", "V_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.0.data_export.0.destination.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.0.data_export.0.destination.0.s3_bucket_destination.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.0.data_export.0.destination.0.s3_bucket_destination.0.format", "CSV"),
					resource.TestCheckResourceAttrPair(resourceName, "storage_class_analysis.0.data_export.0.destination.0.s3_bucket_destination.0.bucket_arn", "aws_s3_bucket.destination", names.AttrARN),
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

func TestAccS3BucketAnalyticsConfiguration_WithStorageClassAnalysis_full(t *testing.T) {
	ctx := acctest.Context(t)
	var ac types.AnalyticsConfiguration
	resourceName := "aws_s3_bucket_analytics_configuration.test"

	rInt := sdkacctest.RandInt()
	rName := fmt.Sprintf("tf-acc-test-%d", rInt)
	prefix := fmt.Sprintf("prefix-%d/", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketAnalyticsConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketAnalyticsConfigurationConfig_fullStorageClassAnalysis(rName, rName, prefix),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketAnalyticsConfigurationExists(ctx, resourceName, &ac),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.0.data_export.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.0.data_export.0.output_schema_version", "V_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.0.data_export.0.destination.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.0.data_export.0.destination.0.s3_bucket_destination.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.0.data_export.0.destination.0.s3_bucket_destination.0.format", "CSV"),
					resource.TestCheckResourceAttrPair(resourceName, "storage_class_analysis.0.data_export.0.destination.0.s3_bucket_destination.0.bucket_arn", "aws_s3_bucket.destination", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.0.data_export.0.destination.0.s3_bucket_destination.0.prefix", prefix),
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

func TestAccS3BucketAnalyticsConfiguration_directoryBucket(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketAnalyticsConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccBucketAnalyticsConfigurationConfig_directoryBucket(rName, rName),
				ExpectError: regexache.MustCompile(`directory buckets are not supported`),
			},
		},
	})
}

func testAccCheckBucketAnalyticsConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3_bucket_analytics_configuration" {
				continue
			}

			bucket, name, err := tfs3.BucketAnalyticsConfigurationParseID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = tfs3.FindAnalyticsConfiguration(ctx, conn, bucket, name)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Bucket Analytics Configuration %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckBucketAnalyticsConfigurationExists(ctx context.Context, n string, v *types.AnalyticsConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		bucket, name, err := tfs3.BucketAnalyticsConfigurationParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)

		output, err := tfs3.FindAnalyticsConfiguration(ctx, conn, bucket, name)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccBucketAnalyticsConfigurationConfig_basic(name, bucket string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_analytics_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %[1]q
}

resource "aws_s3_bucket" "test" {
  bucket = %[2]q
}
`, name, bucket)
}

func testAccBucketAnalyticsConfigurationConfig_update(name, originalBucket, updatedBucket string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_analytics_configuration" "test" {
  bucket = aws_s3_bucket.test_2.bucket
  name   = %[1]q
}

resource "aws_s3_bucket" "test" {
  bucket = %[2]q
}

resource "aws_s3_bucket" "test_2" {
  bucket = %[3]q
}
`, name, originalBucket, updatedBucket)
}

func testAccBucketAnalyticsConfigurationConfig_emptyFilter(name, bucket string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_analytics_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %[1]q

  filter {
  }
}

resource "aws_s3_bucket" "test" {
  bucket = %[2]q
}
`, name, bucket)
}

func testAccBucketAnalyticsConfigurationConfig_filterPrefix(name, bucket, prefix string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_analytics_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %[1]q

  filter {
    prefix = %[2]q
  }
}

resource "aws_s3_bucket" "test" {
  bucket = %[3]q
}
`, name, prefix, bucket)
}

func testAccBucketAnalyticsConfigurationConfig_filterSingleTag(name, bucket, tag string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_analytics_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %[1]q

  filter {
    tags = {
      "tag1" = %[2]q
    }
  }
}

resource "aws_s3_bucket" "test" {
  bucket = %[3]q
}
`, name, tag, bucket)
}

func testAccBucketAnalyticsConfigurationConfig_filterMultipleTags(name, bucket, tag1, tag2 string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_analytics_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %[1]q

  filter {
    tags = {
      "tag1" = %[2]q
      "tag2" = %[3]q
    }
  }
}

resource "aws_s3_bucket" "test" {
  bucket = %[4]q
}
`, name, tag1, tag2, bucket)
}

func testAccBucketAnalyticsConfigurationConfig_filterPrefixAndTags(name, bucket, prefix, tag1, tag2 string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_analytics_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %[1]q

  filter {
    prefix = %[2]q

    tags = {
      "tag1" = %[3]q
      "tag2" = %[4]q
    }
  }
}

resource "aws_s3_bucket" "test" {
  bucket = %[5]q
}
`, name, prefix, tag1, tag2, bucket)
}

func testAccBucketAnalyticsConfigurationConfig_emptyStorageClassAnalysis(name, bucket string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_analytics_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %[1]q

  storage_class_analysis {
  }
}

resource "aws_s3_bucket" "test" {
  bucket = %[2]q
}
`, name, bucket)
}

func testAccBucketAnalyticsConfigurationConfig_defaultStorageClassAnalysis(name, bucket string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_analytics_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %[1]q

  storage_class_analysis {
    data_export {
      destination {
        s3_bucket_destination {
          bucket_arn = aws_s3_bucket.destination.arn
        }
      }
    }
  }
}

resource "aws_s3_bucket" "test" {
  bucket = %[2]q
}

resource "aws_s3_bucket" "destination" {
  bucket = "%[2]s-destination"
}
`, name, bucket)
}

func testAccBucketAnalyticsConfigurationConfig_fullStorageClassAnalysis(name, bucket, prefix string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_analytics_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %[1]q

  storage_class_analysis {
    data_export {
      output_schema_version = "V_1"

      destination {
        s3_bucket_destination {
          format     = "CSV"
          bucket_arn = aws_s3_bucket.destination.arn
          prefix     = %[2]q
        }
      }
    }
  }
}

resource "aws_s3_bucket" "test" {
  bucket = %[3]q
}

resource "aws_s3_bucket" "destination" {
  bucket = "%[3]s-destination"
}
`, name, prefix, bucket)
}

func testAccBucketAnalyticsConfigurationConfig_directoryBucket(bucket, name string) string {
	return acctest.ConfigCompose(testAccDirectoryBucketConfig_base(bucket), fmt.Sprintf(`
resource "aws_s3_directory_bucket" "test" {
  bucket = local.bucket

  location {
    name = local.location_name
  }
}

resource "aws_s3_bucket_analytics_configuration" "test" {
  bucket = aws_s3_directory_bucket.test.bucket
  name   = %[1]q
}
`, name))
}
