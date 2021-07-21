package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"reflect"
	"regexp"
	"sort"
	"testing"
	"time"
)

func TestAccAWSS3BucketIntelligentTieringConfiguration_basic(t *testing.T) {
	var itc s3.IntelligentTieringConfiguration
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_s3_bucket_intelligent_tiering_configuration.test"
	accessTier := "DEEP_ARCHIVE_ACCESS"
	accessTierDays := "180"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, s3.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketIntelligentTieringConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketIntelligentTieringConfigurationOneTier(rName, rName, accessTier, accessTierDays),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketIntelligentTieringConfigurationExists(resourceName, &itc),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "aws_s3_bucket.test", "bucket"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tier.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tier.0.access_tier", accessTier),
					resource.TestCheckResourceAttr(resourceName, "tier.0.days", accessTierDays),
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

func TestAccAWSS3BucketIntelligentTieringConfiguration_removed(t *testing.T) {
	var itc s3.IntelligentTieringConfiguration
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_s3_bucket_intelligent_tiering_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, s3.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketIntelligentTieringConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketIntelligentTieringConfiguration(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketIntelligentTieringConfigurationExists(resourceName, &itc),
				),
			},
			{
				Config: testAccAWSS3BucketIntelligentTieringConfiguration_removed(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketIntelligentTieringConfigurationRemoved(rName, rName),
				),
			},
		},
	})
}

func TestAccAWSS3BucketIntelligentTieringConfiguration_updateBasic(t *testing.T) {
	var itc s3.IntelligentTieringConfiguration
	originalITCName := acctest.RandomWithPrefix("tf-acc-test")
	originalBucketName := acctest.RandomWithPrefix("tf-acc-test")
	updatedITCName := acctest.RandomWithPrefix("tf-acc-test")
	updatedBucketName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_s3_bucket_intelligent_tiering_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, s3.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketIntelligentTieringConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketIntelligentTieringConfiguration(originalITCName, originalBucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketIntelligentTieringConfigurationExists(resourceName, &itc),
					resource.TestCheckResourceAttr(resourceName, "name", originalITCName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "aws_s3_bucket.test", "bucket"),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tier.#", "1"),
				),
			},
			{
				Config: testAccAWSS3BucketIntelligentTieringConfiguration(updatedITCName, originalBucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketIntelligentTieringConfigurationExists(resourceName, &itc),
					testAccCheckAWSS3BucketIntelligentTieringConfigurationRemoved(originalITCName, originalBucketName),
					resource.TestCheckResourceAttr(resourceName, "name", updatedITCName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "aws_s3_bucket.test", "bucket"),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tier.#", "1"),
				),
			},
			{
				Config: testAccAWSS3BucketIntelligentTieringConfigurationUpdateBucket(updatedITCName, originalBucketName, updatedBucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketIntelligentTieringConfigurationExists(resourceName, &itc),
					testAccCheckAWSS3BucketIntelligentTieringConfigurationRemoved(updatedITCName, originalBucketName),
					resource.TestCheckResourceAttr(resourceName, "name", updatedITCName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "aws_s3_bucket.test_2", "bucket"),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tier.#", "1"),
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

func TestAccAWSS3BucketIntelligentTieringConfiguration_WithFilter_Empty(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, s3.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketIntelligentTieringConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSS3BucketIntelligentTieringConfigurationWithEmptyFilter(rName, rName),
				ExpectError: regexp.MustCompile(`one of .* must be specified`),
			},
		},
	})
}

func TestAccAWSS3BucketIntelligentTieringConfiguration_WithFilter_Prefix(t *testing.T) {
	var itc s3.IntelligentTieringConfiguration
	rInt := acctest.RandInt()
	resourceName := "aws_s3_bucket_intelligent_tiering_configuration.test"

	rName := fmt.Sprintf("tf-acc-test-%d", rInt)
	prefix := fmt.Sprintf("prefix-%d/", rInt)
	prefixUpdate := fmt.Sprintf("prefix-update-%d/", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, s3.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketIntelligentTieringConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketIntelligentTieringConfigurationWithFilterPrefix(rName, rName, prefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketIntelligentTieringConfigurationExists(resourceName, &itc),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", prefix),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", "0"),
				),
			},
			{
				Config: testAccAWSS3BucketIntelligentTieringConfigurationWithFilterPrefix(rName, rName, prefixUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketIntelligentTieringConfigurationExists(resourceName, &itc),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", prefixUpdate),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", "0"),
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

func TestAccAWSS3BucketIntelligentTieringConfiguration_WithFilter_SingleTag(t *testing.T) {
	var itc s3.IntelligentTieringConfiguration
	rInt := acctest.RandInt()
	resourceName := "aws_s3_bucket_intelligent_tiering_configuration.test"

	rName := fmt.Sprintf("tf-acc-test-%d", rInt)
	tag1 := fmt.Sprintf("tag-%d", rInt)
	tag1Update := fmt.Sprintf("tag-update-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, s3.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketIntelligentTieringConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketIntelligentTieringConfigurationWithFilterSingleTag(rName, rName, tag1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketIntelligentTieringConfigurationExists(resourceName, &itc),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.tag1", tag1),
				),
			},
			{
				Config: testAccAWSS3BucketIntelligentTieringConfigurationWithFilterSingleTag(rName, rName, tag1Update),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketIntelligentTieringConfigurationExists(resourceName, &itc),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", "1"),
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

func TestAccAWSS3BucketIntelligentTieringConfiguration_WithFilter_MultipleTags(t *testing.T) {
	var itc s3.IntelligentTieringConfiguration
	rInt := acctest.RandInt()
	resourceName := "aws_s3_bucket_intelligent_tiering_configuration.test"

	rName := fmt.Sprintf("tf-acc-test-%d", rInt)
	tag1 := fmt.Sprintf("tag1-%d", rInt)
	tag1Update := fmt.Sprintf("tag1-update-%d", rInt)
	tag2 := fmt.Sprintf("tag2-%d", rInt)
	tag2Update := fmt.Sprintf("tag2-update-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, s3.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketIntelligentTieringConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketIntelligentTieringConfigurationWithFilterMultipleTags(rName, rName, tag1, tag2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketIntelligentTieringConfigurationExists(resourceName, &itc),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.tag1", tag1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.tag2", tag2),
				),
			},
			{
				Config: testAccAWSS3BucketIntelligentTieringConfigurationWithFilterMultipleTags(rName, rName, tag1Update, tag2Update),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketIntelligentTieringConfigurationExists(resourceName, &itc),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", "2"),
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

func TestAccAWSS3BucketIntelligentTieringConfiguration_WithFilter_PrefixAndTags(t *testing.T) {
	var itc s3.IntelligentTieringConfiguration
	rInt := acctest.RandInt()
	resourceName := "aws_s3_bucket_intelligent_tiering_configuration.test"

	rName := fmt.Sprintf("tf-acc-test-%d", rInt)
	prefix := fmt.Sprintf("prefix-%d/", rInt)
	prefixUpdate := fmt.Sprintf("prefix-update-%d/", rInt)
	tag1 := fmt.Sprintf("tag1-%d", rInt)
	tag1Update := fmt.Sprintf("tag1-update-%d", rInt)
	tag2 := fmt.Sprintf("tag2-%d", rInt)
	tag2Update := fmt.Sprintf("tag2-update-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, s3.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketIntelligentTieringConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketIntelligentTieringConfigurationWithFilterPrefixAndTags(rName, rName, prefix, tag1, tag2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketIntelligentTieringConfigurationExists(resourceName, &itc),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", prefix),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.tag1", tag1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.tag2", tag2),
				),
			},
			{
				Config: testAccAWSS3BucketIntelligentTieringConfigurationWithFilterPrefixAndTags(rName, rName, prefixUpdate, tag1Update, tag2Update),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketIntelligentTieringConfigurationExists(resourceName, &itc),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", prefixUpdate),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", "2"),
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

func TestAccAWSS3BucketIntelligentTieringConfiguration_WithFilter_Remove(t *testing.T) {
	var itc s3.IntelligentTieringConfiguration
	rInt := acctest.RandInt()
	resourceName := "aws_s3_bucket_intelligent_tiering_configuration.test"

	rName := fmt.Sprintf("tf-acc-test-%d", rInt)
	prefix := fmt.Sprintf("prefix-%d/", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, s3.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketIntelligentTieringConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketIntelligentTieringConfigurationWithFilterPrefix(rName, rName, prefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketIntelligentTieringConfigurationExists(resourceName, &itc),
				),
			},
			{
				Config: testAccAWSS3BucketIntelligentTieringConfiguration(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketIntelligentTieringConfigurationExists(resourceName, &itc),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "0"),
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

func TestAccAWSS3BucketIntelligentTieringConfiguration_WithTier_Empty(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, s3.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketIntelligentTieringConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSS3BucketIntelligentTieringConfigurationWithEmptyTier(rName, rName),
				ExpectError: regexp.MustCompile(`The argument "access_tier" is required, but no definition was found.`),
			},
		},
	})
}

func TestAccAWSS3BucketIntelligentTieringConfiguration_WithTwoTiers_Default(t *testing.T) {
	var itc s3.IntelligentTieringConfiguration
	resourceName := "aws_s3_bucket_intelligent_tiering_configuration.test"

	accessTierOne := "ARCHIVE_ACCESS"
	accessTierDaysOne := "120"

	accessTierTwo := "DEEP_ARCHIVE_ACCESS"
	accessTierDaysTwo := "240"

	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, s3.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketIntelligentTieringConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketIntelligentTieringConfigurationTwoTiers(rName, rName, accessTierOne, accessTierDaysOne, accessTierTwo, accessTierDaysTwo),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketIntelligentTieringConfigurationExists(resourceName, &itc),
					resource.TestCheckResourceAttr(resourceName, "tier.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tier.0.access_tier", accessTierOne),
					resource.TestCheckResourceAttr(resourceName, "tier.0.days", accessTierDaysOne),
					resource.TestCheckResourceAttr(resourceName, "tier.1.access_tier", accessTierTwo),
					resource.TestCheckResourceAttr(resourceName, "tier.1.days", accessTierDaysTwo),
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

func TestAccAWSS3BucketIntelligentTieringConfiguration_WithTwoTiers_RemoveOne(t *testing.T) {
	var itc s3.IntelligentTieringConfiguration
	resourceName := "aws_s3_bucket_intelligent_tiering_configuration.test"

	accessTierOne := "ARCHIVE_ACCESS"
	accessTierDaysOne := "120"

	accessTierTwo := "DEEP_ARCHIVE_ACCESS"
	accessTierDaysTwo := "240"

	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, s3.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketIntelligentTieringConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketIntelligentTieringConfigurationTwoTiers(rName, rName, accessTierOne, accessTierDaysOne, accessTierTwo, accessTierDaysTwo),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketIntelligentTieringConfigurationExists(resourceName, &itc),
					resource.TestCheckResourceAttr(resourceName, "tier.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tier.0.access_tier", accessTierOne),
					resource.TestCheckResourceAttr(resourceName, "tier.0.days", accessTierDaysOne),
					resource.TestCheckResourceAttr(resourceName, "tier.1.access_tier", accessTierTwo),
					resource.TestCheckResourceAttr(resourceName, "tier.1.days", accessTierDaysTwo),
				),
			},
			{
				Config: testAccAWSS3BucketIntelligentTieringConfigurationOneTier(rName, rName, accessTierOne, accessTierDaysOne),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketIntelligentTieringConfigurationExists(resourceName, &itc),
					resource.TestCheckResourceAttr(resourceName, "tier.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tier.0.access_tier", accessTierOne),
					resource.TestCheckResourceAttr(resourceName, "tier.0.days", accessTierDaysOne),
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

func testAccAWSS3BucketIntelligentTieringConfiguration(name, bucket string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_intelligent_tiering_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = "%s"
  tier {
	access_tier = "DEEP_ARCHIVE_ACCESS"
	days        = 180
	}
}

resource "aws_s3_bucket" "test" {
  bucket = "%s"
}
`, name, bucket)
}

func testAccAWSS3BucketIntelligentTieringConfigurationOneTier(name, bucket, accessTier, accessTierDays string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_intelligent_tiering_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = "%s"
  tier {
    access_tier = "%s"
    days        = %s
  }
}

resource "aws_s3_bucket" "test" {
  bucket = "%s"
}
`, name, accessTier, accessTierDays, bucket)
}

func testAccAWSS3BucketIntelligentTieringConfigurationTwoTiers(name, bucket, accessTierOne, accessTierDaysOne, accessTierTwo, accessTierDaysTwo string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_intelligent_tiering_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = "%s"
  tier {
    access_tier = "%s"
    days        = %s
  }
  tier {
    access_tier = "%s"
    days        = %s
  }
}

resource "aws_s3_bucket" "test" {
  bucket = "%s"
}
`, name, accessTierOne, accessTierDaysOne, accessTierTwo, accessTierDaysTwo, bucket)
}

func testAccAWSS3BucketIntelligentTieringConfiguration_removed(bucket string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = "%s"
}
`, bucket)
}

func testAccAWSS3BucketIntelligentTieringConfigurationUpdateBucket(name, originalBucket, updatedBucket string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_intelligent_tiering_configuration" "test" {
	bucket = aws_s3_bucket.test_2.bucket
	name   = "%s"
	tier {
		access_tier = "ARCHIVE_ACCESS"
		days        = 90
	}
}

resource "aws_s3_bucket" "test" {
  bucket = "%s"
}

resource "aws_s3_bucket" "test_2" {
  bucket = "%s"
}
`, name, originalBucket, updatedBucket)
}

func testAccAWSS3BucketIntelligentTieringConfigurationWithEmptyFilter(name, bucket string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_intelligent_tiering_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = "%s"

  tier {
	access_tier = "DEEP_ARCHIVE_ACCESS"
	days        = 180
	}

  filter {
  }
}

resource "aws_s3_bucket" "test" {
  bucket = "%s"
}
`, name, bucket)
}

func testAccAWSS3BucketIntelligentTieringConfigurationWithFilterPrefix(name, bucket, prefix string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_intelligent_tiering_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = "%s"
  tier {
	access_tier = "ARCHIVE_ACCESS"
	days        = 90
	}

  filter {
    prefix = "%s"
  }
}

resource "aws_s3_bucket" "test" {
  bucket = "%s"
}
`, name, prefix, bucket)
}

func testAccAWSS3BucketIntelligentTieringConfigurationWithFilterSingleTag(name, bucket, tag string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_intelligent_tiering_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = "%s"

  tier {
	access_tier = "ARCHIVE_ACCESS"
	days        = 90
	}

  filter {
    tags = {
      "tag1" = "%s"
    }
  }
}

resource "aws_s3_bucket" "test" {
  bucket = "%s"
}
`, name, tag, bucket)
}

func testAccAWSS3BucketIntelligentTieringConfigurationWithFilterMultipleTags(name, bucket, tag1, tag2 string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_intelligent_tiering_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = "%s"

  tier {
	access_tier = "ARCHIVE_ACCESS"
	days        = 90
	}

  filter {
    tags = {
      "tag1" = "%s"
      "tag2" = "%s"
    }
  }
}

resource "aws_s3_bucket" "test" {
  bucket = "%s"
}
`, name, tag1, tag2, bucket)
}

func testAccAWSS3BucketIntelligentTieringConfigurationWithFilterPrefixAndTags(name, bucket, prefix, tag1, tag2 string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_intelligent_tiering_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = "%s"

  tier {
	access_tier = "ARCHIVE_ACCESS"
	days        = 90
	}

  filter {
    prefix = "%s"

    tags = {
      "tag1" = "%s"
      "tag2" = "%s"
    }
  }
}

resource "aws_s3_bucket" "test" {
  bucket = "%s"
}
`, name, prefix, tag1, tag2, bucket)
}

func testAccAWSS3BucketIntelligentTieringConfigurationWithEmptyTier(name, bucket string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_intelligent_tiering_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = "%s"
  tier {
  }
}

resource "aws_s3_bucket" "test" {
  bucket = "%s"
}
`, name, bucket)
}

func testAccCheckAWSS3BucketIntelligentTieringConfigurationExists(n string, itc *s3.IntelligentTieringConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).s3conn
		output, err := conn.GetBucketIntelligentTieringConfiguration(&s3.GetBucketIntelligentTieringConfigurationInput{
			Bucket: aws.String(rs.Primary.Attributes["bucket"]),
			Id:     aws.String(rs.Primary.Attributes["name"]),
		})

		if err != nil {
			return err
		}

		if output == nil || output.IntelligentTieringConfiguration == nil {
			return fmt.Errorf("error reading S3 Bucket Analytics Configuration %q: empty response", rs.Primary.ID)
		}

		*itc = *output.IntelligentTieringConfiguration

		return nil
	}
}

func testAccCheckAWSS3BucketIntelligentTieringConfigurationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).s3conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_s3_bucket_intelligent_tiering_configuration" {
			continue
		}
		bucket, name, err := resourceAwsS3BucketIntelligentTieringConfigurationParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		return waitForDeleteS3BucketIntelligentTieringConfiguration(conn, bucket, name, 1*time.Minute)
	}

	return nil

}

func testAccCheckAWSS3BucketIntelligentTieringConfigurationRemoved(name, bucket string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).s3conn
		return waitForDeleteS3BucketIntelligentTieringConfiguration(conn, bucket, name, 1*time.Minute)
	}
}

func TestExpandS3IntelligentTieringFilter(t *testing.T) {
	testCases := map[string]struct {
		Input    []interface{}
		Expected *s3.IntelligentTieringFilter
	}{
		"nil input": {
			Input:    nil,
			Expected: nil,
		},
		"empty input": {
			Input:    []interface{}{},
			Expected: nil,
		},
		"prefix only": {
			Input: []interface{}{
				map[string]interface{}{
					"prefix": "prefix/",
				},
			},
			Expected: &s3.IntelligentTieringFilter{
				Prefix: aws.String("prefix/"),
			},
		},
		"prefix and single tag": {
			Input: []interface{}{
				map[string]interface{}{
					"prefix": "prefix/",
					"tags": map[string]interface{}{
						"tag1key": "tag1value",
					},
				},
			},
			Expected: &s3.IntelligentTieringFilter{
				And: &s3.IntelligentTieringAndOperator{
					Prefix: aws.String("prefix/"),
					Tags: []*s3.Tag{
						{
							Key:   aws.String("tag1key"),
							Value: aws.String("tag1value"),
						},
					},
				},
			},
		},
		"prefix and multiple tags": {
			Input: []interface{}{map[string]interface{}{
				"prefix": "prefix/",
				"tags": map[string]interface{}{
					"tag1key": "tag1value",
					"tag2key": "tag2value",
				},
			},
			},
			Expected: &s3.IntelligentTieringFilter{
				And: &s3.IntelligentTieringAndOperator{
					Prefix: aws.String("prefix/"),
					Tags: []*s3.Tag{
						{
							Key:   aws.String("tag1key"),
							Value: aws.String("tag1value"),
						},
						{
							Key:   aws.String("tag2key"),
							Value: aws.String("tag2value"),
						},
					},
				},
			},
		},
		"single tag only": {
			Input: []interface{}{
				map[string]interface{}{
					"tags": map[string]interface{}{
						"tag1key": "tag1value",
					},
				},
			},
			Expected: &s3.IntelligentTieringFilter{
				Tag: &s3.Tag{
					Key:   aws.String("tag1key"),
					Value: aws.String("tag1value"),
				},
			},
		},
		"multiple tags only": {
			Input: []interface{}{
				map[string]interface{}{
					"tags": map[string]interface{}{
						"tag1key": "tag1value",
						"tag2key": "tag2value",
					},
				},
			},
			Expected: &s3.IntelligentTieringFilter{
				And: &s3.IntelligentTieringAndOperator{
					Tags: []*s3.Tag{
						{
							Key:   aws.String("tag1key"),
							Value: aws.String("tag1value"),
						},
						{
							Key:   aws.String("tag2key"),
							Value: aws.String("tag2value"),
						},
					},
				},
			},
		},
	}

	for k, tc := range testCases {
		value := expandS3IntelligentTieringFilter(tc.Input)

		if value == nil {
			if tc.Expected == nil {
				continue
			} else {
				t.Errorf("Case %q: Got nil\nExpected:\n%v", k, tc.Expected)
			}
		}

		if tc.Expected == nil {
			t.Errorf("Case %q: Got: %v\nExpected: nil", k, value)
		}

		// Sort tags by key for consistency
		if value.And != nil && value.And.Tags != nil {
			sort.Slice(value.And.Tags, func(i, j int) bool {
				return *value.And.Tags[i].Key < *value.And.Tags[j].Key
			})
		}

		// Convert to strings to avoid dealing with pointers
		valueS := fmt.Sprintf("%v", value)
		expectedValueS := fmt.Sprintf("%v", tc.Expected)

		if valueS != expectedValueS {
			t.Errorf("Case %q: Given:\n%s\n\nExpected:\n%s", k, valueS, expectedValueS)
		}
	}
}

func TestExpandS3IntelligentTieringConfigurations(t *testing.T) {
	testCases := map[string]struct {
		Input    []interface{}
		Expected []*s3.Tiering
	}{
		"nil input": {
			Input:    nil,
			Expected: nil,
		},
		"empty input": {
			Input:    []interface{}{map[string]interface{}{}},
			Expected: nil,
		},
		"empty tier": {
			Input: []interface{}{
				map[string]interface{}{},
			},
			Expected: []*s3.Tiering{},
		},
		"one tier": {
			Input: []interface{}{
				map[string]interface{}{
					"access_tier": "test",
					"days":        55,
				},
			},
			Expected: []*s3.Tiering{
				{
					AccessTier: aws.String("test"),
					Days:       aws.Int64(int64(55)),
				},
			},
		},
		"two tiers": {
			Input: []interface{}{
				map[string]interface{}{
					"access_tier": "test",
					"days":        55,
				},
				map[string]interface{}{
					"access_tier": "test2",
					"days":        56,
				},
			},
			Expected: []*s3.Tiering{
				{
					AccessTier: aws.String("test"),
					Days:       aws.Int64(int64(55)),
				},
				{
					AccessTier: aws.String("test2"),
					Days:       aws.Int64(int64(56)),
				},
			},
		},
	}

	for k, tc := range testCases {
		value := expandS3IntelligentTieringConfigurations(tc.Input)

		if !reflect.DeepEqual(value, tc.Expected) {
			t.Errorf("Case %q:\nGot:\n%v\nExpected:\n%v", k, value, tc.Expected)
		}
	}
}
