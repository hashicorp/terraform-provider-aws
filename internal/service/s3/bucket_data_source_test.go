// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3BucketDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	region := acctest.Region()
	hostedZoneID, _ := tfs3.HostedZoneIDForRegion(region)
	resourceName := "aws_s3_bucket.test"
	dataSourceName := "data.aws_s3_bucket.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					testAccCheckBucketDomainName(ctx, dataSourceName, "bucket_domain_name", rName),
					resource.TestCheckResourceAttr(dataSourceName, "bucket_region", region),
					resource.TestCheckResourceAttr(dataSourceName, "bucket_regional_domain_name", testAccBucketRegionalDomainName(rName, region)),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrHostedZoneID, hostedZoneID),
					resource.TestCheckNoResourceAttr(dataSourceName, "website_endpoint"),
				),
			},
		},
	})
}

func TestAccS3BucketDataSource_website(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket.test"
	websiteConfigurationResourceName := "aws_s3_bucket_website_configuration.test"
	dataSourceName := "data.aws_s3_bucket.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketDataSourceConfig_website(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "website_domain", websiteConfigurationResourceName, "website_domain"),
					resource.TestCheckResourceAttrPair(dataSourceName, "website_endpoint", websiteConfigurationResourceName, "website_endpoint"),
				),
			},
		},
	})
}

func TestAccS3BucketDataSource_accessPointARN(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	accessPointResourceName := "aws_s3_access_point.test"
	dataSourceName := "data.aws_s3_bucket.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketDataSourceConfig_accessPointARN(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, accessPointResourceName, names.AttrARN),
				),
			},
		},
	})
}

func TestAccS3BucketDataSource_accessPointAlias(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_s3_bucket.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketDataSourceConfig_accessPointAlias(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckResourceAttrGlobalARNNoAccountFormat(dataSourceName, names.AttrARN, "s3", "{bucket}"),
				),
			},
		},
	})
}

func TestAccS3BucketDataSource_crossRegion(t *testing.T) {
	ctx := acctest.Context(t)
	var providers []*schema.Provider
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		Steps: []resource.TestStep{
			{
				// Attempt to read a bucket created in a different Region.
				Config:      testAccBucketDataSourceConfig_crossRegion(rName),
				ExpectError: regexache.MustCompile(`empty result`),
			},
		},
	})
}

func testAccBucketDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

data "aws_s3_bucket" "test" {
  bucket = aws_s3_bucket.test.id
}
`, rName)
}

func testAccBucketDataSourceConfig_website(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_website_configuration" "test" {
  bucket = aws_s3_bucket.test.id
  index_document {
    suffix = "index.html"
  }
  error_document {
    key = "error.html"
  }
}

data "aws_s3_bucket" "test" {
  # Must have bucket website configured first
  bucket = aws_s3_bucket_website_configuration.test.id
}
`, rName)
}

func testAccBucketDataSourceConfig_accessPointARN(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_versioning" "test" {
  bucket = aws_s3_bucket.test.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_access_point" "test" {
  # Must have bucket versioning enabled first
  bucket = aws_s3_bucket_versioning.test.bucket
  name   = %[1]q
}

data "aws_s3_bucket" "test" {
  bucket = aws_s3_access_point.test.arn
}
`, rName)
}

func testAccBucketDataSourceConfig_accessPointAlias(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_versioning" "test" {
  bucket = aws_s3_bucket.test.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_access_point" "test" {
  # Must have bucket versioning enabled first
  bucket = aws_s3_bucket_versioning.test.bucket
  name   = %[1]q
}

data "aws_s3_bucket" "test" {
  bucket = aws_s3_access_point.test.alias
}
`, rName)
}

func testAccBucketDataSourceConfig_crossRegion(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigMultipleRegionProvider(2), fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  provider = "awsalternate"

  bucket = %[1]q
}

data "aws_s3_bucket" "test" {
  bucket = aws_s3_bucket.test.id
}
`, rName))
}
