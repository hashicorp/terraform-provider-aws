package s3_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccS3BucketIntelligentTieringConfiguration_basic(t *testing.T) {
	var itc s3.IntelligentTieringConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_intelligent_tiering_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketIntelligentTieringConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketIntelligentTieringConfiguration(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckS3BucketIntelligentTieringConfigurationExists(resourceName, &itc),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "aws_s3_bucket.test", "bucket"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
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

func TestAccS3BucketIntelligentTieringConfiguration_disableBasic(t *testing.T) {
	var itc s3.IntelligentTieringConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_intelligent_tiering_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketIntelligentTieringConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketIntelligentTieringConfiguration(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckS3BucketIntelligentTieringConfigurationExists(resourceName, &itc),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "aws_s3_bucket.test", "bucket"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tier.#", "1"),
				),
			},
			{
				Config: testAccBucketIntelligentTieringConfigurationDisabled(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckS3BucketIntelligentTieringConfigurationExists(resourceName, &itc),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "aws_s3_bucket.test", "bucket"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
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

func TestAccS3BucketIntelligentTieringConfiguration_updateBasic(t *testing.T) {
	var itc s3.IntelligentTieringConfiguration
	originalITCName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	originalBucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	updatedITCName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	updatedBucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_intelligent_tiering_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketIntelligentTieringConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketIntelligentTieringConfiguration(originalITCName, originalBucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckS3BucketIntelligentTieringConfigurationExists(resourceName, &itc),
					resource.TestCheckResourceAttr(resourceName, "name", originalITCName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "aws_s3_bucket.test", "bucket"),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tier.#", "1"),
				),
			},
			{
				Config: testAccBucketIntelligentTieringConfiguration(updatedITCName, originalBucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckS3BucketIntelligentTieringConfigurationExists(resourceName, &itc),
					resource.TestCheckResourceAttr(resourceName, "name", updatedITCName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "aws_s3_bucket.test", "bucket"),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tier.#", "1"),
				),
			},
			{
				Config: testAccBucketIntelligentTieringConfigurationUpdateBucket(updatedITCName, originalBucketName, updatedBucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckS3BucketIntelligentTieringConfigurationExists(resourceName, &itc),
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

func TestAccS3BucketIntelligentTieringConfiguration_WithFilter_Empty(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketIntelligentTieringConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccBucketIntelligentTieringConfigurationWithEmptyFilter(rName, rName),
				ExpectError: regexp.MustCompile(`one of .* must be specified`),
			},
		},
	})
}

func TestAccS3BucketIntelligentTieringConfiguration_WithFilter_Prefix(t *testing.T) {
	var itc s3.IntelligentTieringConfiguration
	rInt := sdkacctest.RandInt()
	resourceName := "aws_s3_bucket_intelligent_tiering_configuration.test"

	rName := fmt.Sprintf("tf-acc-test-%d", rInt)
	prefix := fmt.Sprintf("prefix-%d/", rInt)
	prefixUpdate := fmt.Sprintf("prefix-update-%d/", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketIntelligentTieringConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketIntelligentTieringConfigurationWithFilterPrefix(rName, rName, prefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckS3BucketIntelligentTieringConfigurationExists(resourceName, &itc),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", prefix),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", "0"),
				),
			},
			{
				Config: testAccBucketIntelligentTieringConfigurationWithFilterPrefix(rName, rName, prefixUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckS3BucketIntelligentTieringConfigurationExists(resourceName, &itc),
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

func TestAccS3BucketIntelligentTieringConfiguration_WithFilter_SingleTag(t *testing.T) {
	var itc s3.IntelligentTieringConfiguration
	rInt := sdkacctest.RandInt()
	resourceName := "aws_s3_bucket_intelligent_tiering_configuration.test"

	rName := fmt.Sprintf("tf-acc-test-%d", rInt)
	tag1 := fmt.Sprintf("tag-%d", rInt)
	tag1Update := fmt.Sprintf("tag-update-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketIntelligentTieringConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketIntelligentTieringConfigurationWithFilterSingleTag(rName, rName, tag1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckS3BucketIntelligentTieringConfigurationExists(resourceName, &itc),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.tag1", tag1),
				),
			},
			{
				Config: testAccBucketIntelligentTieringConfigurationWithFilterSingleTag(rName, rName, tag1Update),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckS3BucketIntelligentTieringConfigurationExists(resourceName, &itc),
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

func TestAccS3BucketIntelligentTieringConfiguration_WithFilter_MultipleTags(t *testing.T) {
	var itc s3.IntelligentTieringConfiguration
	rInt := sdkacctest.RandInt()
	resourceName := "aws_s3_bucket_intelligent_tiering_configuration.test"

	rName := fmt.Sprintf("tf-acc-test-%d", rInt)
	tag1 := fmt.Sprintf("tag1-%d", rInt)
	tag1Update := fmt.Sprintf("tag1-update-%d", rInt)
	tag2 := fmt.Sprintf("tag2-%d", rInt)
	tag2Update := fmt.Sprintf("tag2-update-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketIntelligentTieringConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketIntelligentTieringConfigurationWithFilterMultipleTags(rName, rName, tag1, tag2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckS3BucketIntelligentTieringConfigurationExists(resourceName, &itc),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.tag1", tag1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.tag2", tag2),
				),
			},
			{
				Config: testAccBucketIntelligentTieringConfigurationWithFilterMultipleTags(rName, rName, tag1Update, tag2Update),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckS3BucketIntelligentTieringConfigurationExists(resourceName, &itc),
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

func TestAccS3BucketIntelligentTieringConfiguration_WithFilter_PrefixAndTags(t *testing.T) {
	var itc s3.IntelligentTieringConfiguration
	rInt := sdkacctest.RandInt()
	resourceName := "aws_s3_bucket_intelligent_tiering_configuration.test"

	rName := fmt.Sprintf("tf-acc-test-%d", rInt)
	prefix := fmt.Sprintf("prefix-%d/", rInt)
	prefixUpdate := fmt.Sprintf("prefix-update-%d/", rInt)
	tag1 := fmt.Sprintf("tag1-%d", rInt)
	tag1Update := fmt.Sprintf("tag1-update-%d", rInt)
	tag2 := fmt.Sprintf("tag2-%d", rInt)
	tag2Update := fmt.Sprintf("tag2-update-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketIntelligentTieringConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketIntelligentTieringConfigurationWithFilterPrefixAndTags(rName, rName, prefix, tag1, tag2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckS3BucketIntelligentTieringConfigurationExists(resourceName, &itc),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", prefix),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.tag1", tag1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.tag2", tag2),
				),
			},
			{
				Config: testAccBucketIntelligentTieringConfigurationWithFilterPrefixAndTags(rName, rName, prefixUpdate, tag1Update, tag2Update),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckS3BucketIntelligentTieringConfigurationExists(resourceName, &itc),
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

func TestAccS3BucketIntelligentTieringConfiguration_WithFilter_Remove(t *testing.T) {
	var itc s3.IntelligentTieringConfiguration
	rInt := sdkacctest.RandInt()
	resourceName := "aws_s3_bucket_intelligent_tiering_configuration.test"

	rName := fmt.Sprintf("tf-acc-test-%d", rInt)
	prefix := fmt.Sprintf("prefix-%d/", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketIntelligentTieringConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketIntelligentTieringConfigurationWithFilterPrefix(rName, rName, prefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckS3BucketIntelligentTieringConfigurationExists(resourceName, &itc),
				),
			},
			{
				Config: testAccBucketIntelligentTieringConfiguration(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckS3BucketIntelligentTieringConfigurationExists(resourceName, &itc),
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

func TestAccS3BucketIntelligentTieringConfiguration_WithTier_Empty(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketIntelligentTieringConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccBucketIntelligentTieringConfigurationWithEmptyTier(rName, rName),
				ExpectError: regexp.MustCompile(`The argument "access_tier" is required, but no definition was found.`),
			},
		},
	})
}

func testAccBucketIntelligentTieringConfiguration_WithTwoTiers_Default(t *testing.T) {
	var itc s3.IntelligentTieringConfiguration
	resourceName := "aws_s3_bucket_intelligent_tiering_configuration.test"

	accessTierOne := "ARCHIVE_ACCESS"
	accessTierDaysOne := "120"

	accessTierTwo := "DEEP_ARCHIVE_ACCESS"
	accessTierDaysTwo := "240"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketIntelligentTieringConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketIntelligentTieringConfigurationTwoTiers(rName, rName, accessTierOne, accessTierDaysOne, accessTierTwo, accessTierDaysTwo),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckS3BucketIntelligentTieringConfigurationExists(resourceName, &itc),
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

func testAccBucketIntelligentTieringConfiguration_WithOneTier_UpdateDays(t *testing.T) {
	var itc s3.IntelligentTieringConfiguration
	resourceName := "aws_s3_bucket_intelligent_tiering_configuration.test"

	accessTier := "ARCHIVE_ACCESS"
	accessTierDays := "120"

	accessTierDaysNew := "240"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketIntelligentTieringConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketIntelligentTieringConfigurationOneTier(rName, rName, accessTier, accessTierDays),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckS3BucketIntelligentTieringConfigurationExists(resourceName, &itc),
					resource.TestCheckResourceAttr(resourceName, "tier.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tier.0.access_tier", accessTier),
					resource.TestCheckResourceAttr(resourceName, "tier.0.days", accessTierDays),
				),
			},
			{
				Config: testAccBucketIntelligentTieringConfigurationOneTier(rName, rName, accessTier, accessTierDaysNew),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckS3BucketIntelligentTieringConfigurationExists(resourceName, &itc),
					resource.TestCheckResourceAttr(resourceName, "tier.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tier.0.access_tier", accessTier),
					resource.TestCheckResourceAttr(resourceName, "tier.0.days", accessTierDaysNew),
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

func testAccBucketIntelligentTieringConfiguration_WithTwoTiers_RemoveOne(t *testing.T) {
	var itc s3.IntelligentTieringConfiguration
	resourceName := "aws_s3_bucket_intelligent_tiering_configuration.test"

	accessTierOne := "ARCHIVE_ACCESS"
	accessTierDaysOne := "120"

	accessTierTwo := "DEEP_ARCHIVE_ACCESS"
	accessTierDaysTwo := "240"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketIntelligentTieringConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketIntelligentTieringConfigurationTwoTiers(rName, rName, accessTierOne, accessTierDaysOne, accessTierTwo, accessTierDaysTwo),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckS3BucketIntelligentTieringConfigurationExists(resourceName, &itc),
					resource.TestCheckResourceAttr(resourceName, "tier.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tier.0.access_tier", accessTierOne),
					resource.TestCheckResourceAttr(resourceName, "tier.0.days", accessTierDaysOne),
					resource.TestCheckResourceAttr(resourceName, "tier.1.access_tier", accessTierTwo),
					resource.TestCheckResourceAttr(resourceName, "tier.1.days", accessTierDaysTwo),
				),
			},
			{
				Config: testAccBucketIntelligentTieringConfigurationOneTier(rName, rName, accessTierOne, accessTierDaysOne),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckS3BucketIntelligentTieringConfigurationExists(resourceName, &itc),
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

func testAccBucketIntelligentTieringConfiguration(name, bucket string) string {
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

func testAccBucketIntelligentTieringConfigurationDisabled(name, bucket string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_intelligent_tiering_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = "%s"
  enabled = false
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

func testAccBucketIntelligentTieringConfigurationOneTier(name, bucket, accessTier, accessTierDays string) string {
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

func testAccBucketIntelligentTieringConfigurationTwoTiers(name, bucket, accessTierOne, accessTierDaysOne, accessTierTwo, accessTierDaysTwo string) string {
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

func testAccBucketIntelligentTieringConfigurationUpdateBucket(name, originalBucket, updatedBucket string) string {
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

func testAccBucketIntelligentTieringConfigurationWithEmptyFilter(name, bucket string) string {
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

func testAccBucketIntelligentTieringConfigurationWithFilterPrefix(name, bucket, prefix string) string {
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

func testAccBucketIntelligentTieringConfigurationWithFilterSingleTag(name, bucket, tag string) string {
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

func testAccBucketIntelligentTieringConfigurationWithFilterMultipleTags(name, bucket, tag1, tag2 string) string {
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

func testAccBucketIntelligentTieringConfigurationWithFilterPrefixAndTags(name, bucket, prefix, tag1, tag2 string) string {
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

func testAccBucketIntelligentTieringConfigurationWithEmptyTier(name, bucket string) string {
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

func testAccCheckS3BucketIntelligentTieringConfigurationExists(n string, itc *s3.IntelligentTieringConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn

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

func testAccCheckBucketIntelligentTieringConfigurationDestroy(s *terraform.State) error {
	// conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_s3_bucket_intelligent_tiering_configuration" {
			continue
		}

		// TODO
		// bucket, name, err := resourceAwsS3BucketIntelligentTieringConfigurationParseID(rs.Primary.ID)

		// if err != nil {
		// 	return err
		// }

		// return waitForDeleteS3BucketIntelligentTieringConfiguration(conn, bucket, name, 1*time.Minute)
	}

	return nil

}
