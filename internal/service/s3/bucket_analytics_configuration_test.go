package s3_test

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
)

func TestAccS3BucketAnalyticsConfiguration_basic(t *testing.T) {
	var ac s3.AnalyticsConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_analytics_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketAnalyticsConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketAnalyticsConfiguration(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAnalyticsConfigurationExists(resourceName, &ac),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "aws_s3_bucket.test", "bucket"),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.#", "0"),
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

func TestAccS3BucketAnalyticsConfiguration_removed(t *testing.T) {
	var ac s3.AnalyticsConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_analytics_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketAnalyticsConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketAnalyticsConfiguration(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAnalyticsConfigurationExists(resourceName, &ac),
				),
			},
			{
				Config: testAccBucketAnalyticsConfiguration_removed(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAnalyticsConfigurationRemoved(rName, rName),
				),
			},
		},
	})
}

func TestAccS3BucketAnalyticsConfiguration_updateBasic(t *testing.T) {
	var ac s3.AnalyticsConfiguration
	originalACName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	originalBucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	updatedACName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	updatedBucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_analytics_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketAnalyticsConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketAnalyticsConfiguration(originalACName, originalBucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAnalyticsConfigurationExists(resourceName, &ac),
					resource.TestCheckResourceAttr(resourceName, "name", originalACName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "aws_s3_bucket.test", "bucket"),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.#", "0"),
				),
			},
			{
				Config: testAccBucketAnalyticsConfiguration(updatedACName, originalBucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAnalyticsConfigurationExists(resourceName, &ac),
					testAccCheckBucketAnalyticsConfigurationRemoved(originalACName, originalBucketName),
					resource.TestCheckResourceAttr(resourceName, "name", updatedACName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "aws_s3_bucket.test", "bucket"),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.#", "0"),
				),
			},
			{
				Config: testAccBucketAnalyticsConfigurationUpdateBucket(updatedACName, originalBucketName, updatedBucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAnalyticsConfigurationExists(resourceName, &ac),
					testAccCheckBucketAnalyticsConfigurationRemoved(updatedACName, originalBucketName),
					resource.TestCheckResourceAttr(resourceName, "name", updatedACName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "aws_s3_bucket.test_2", "bucket"),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.#", "0"),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketAnalyticsConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccBucketAnalyticsConfigurationWithEmptyFilter(rName, rName),
				ExpectError: regexp.MustCompile(`one of .* must be specified`),
			},
		},
	})
}

func TestAccS3BucketAnalyticsConfiguration_WithFilter_prefix(t *testing.T) {
	var ac s3.AnalyticsConfiguration
	rInt := sdkacctest.RandInt()
	resourceName := "aws_s3_bucket_analytics_configuration.test"

	rName := fmt.Sprintf("tf-acc-test-%d", rInt)
	prefix := fmt.Sprintf("prefix-%d/", rInt)
	prefixUpdate := fmt.Sprintf("prefix-update-%d/", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketAnalyticsConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketAnalyticsConfigurationWithFilterPrefix(rName, rName, prefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAnalyticsConfigurationExists(resourceName, &ac),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", prefix),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", "0"),
				),
			},
			{
				Config: testAccBucketAnalyticsConfigurationWithFilterPrefix(rName, rName, prefixUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAnalyticsConfigurationExists(resourceName, &ac),
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

func TestAccS3BucketAnalyticsConfiguration_WithFilter_singleTag(t *testing.T) {
	var ac s3.AnalyticsConfiguration
	rInt := sdkacctest.RandInt()
	resourceName := "aws_s3_bucket_analytics_configuration.test"

	rName := fmt.Sprintf("tf-acc-test-%d", rInt)
	tag1 := fmt.Sprintf("tag-%d", rInt)
	tag1Update := fmt.Sprintf("tag-update-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketAnalyticsConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketAnalyticsConfigurationWithFilterSingleTag(rName, rName, tag1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAnalyticsConfigurationExists(resourceName, &ac),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.tag1", tag1),
				),
			},
			{
				Config: testAccBucketAnalyticsConfigurationWithFilterSingleTag(rName, rName, tag1Update),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAnalyticsConfigurationExists(resourceName, &ac),
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

func TestAccS3BucketAnalyticsConfiguration_WithFilter_multipleTags(t *testing.T) {
	var ac s3.AnalyticsConfiguration
	rInt := sdkacctest.RandInt()
	resourceName := "aws_s3_bucket_analytics_configuration.test"

	rName := fmt.Sprintf("tf-acc-test-%d", rInt)
	tag1 := fmt.Sprintf("tag1-%d", rInt)
	tag1Update := fmt.Sprintf("tag1-update-%d", rInt)
	tag2 := fmt.Sprintf("tag2-%d", rInt)
	tag2Update := fmt.Sprintf("tag2-update-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketAnalyticsConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketAnalyticsConfigurationWithFilterMultipleTags(rName, rName, tag1, tag2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAnalyticsConfigurationExists(resourceName, &ac),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.tag1", tag1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.tag2", tag2),
				),
			},
			{
				Config: testAccBucketAnalyticsConfigurationWithFilterMultipleTags(rName, rName, tag1Update, tag2Update),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAnalyticsConfigurationExists(resourceName, &ac),
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

func TestAccS3BucketAnalyticsConfiguration_WithFilter_prefixAndTags(t *testing.T) {
	var ac s3.AnalyticsConfiguration
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
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketAnalyticsConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketAnalyticsConfigurationWithFilterPrefixAndTags(rName, rName, prefix, tag1, tag2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAnalyticsConfigurationExists(resourceName, &ac),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", prefix),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.tag1", tag1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.tag2", tag2),
				),
			},
			{
				Config: testAccBucketAnalyticsConfigurationWithFilterPrefixAndTags(rName, rName, prefixUpdate, tag1Update, tag2Update),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAnalyticsConfigurationExists(resourceName, &ac),
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

func TestAccS3BucketAnalyticsConfiguration_WithFilter_remove(t *testing.T) {
	var ac s3.AnalyticsConfiguration
	rInt := sdkacctest.RandInt()
	resourceName := "aws_s3_bucket_analytics_configuration.test"

	rName := fmt.Sprintf("tf-acc-test-%d", rInt)
	prefix := fmt.Sprintf("prefix-%d/", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketAnalyticsConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketAnalyticsConfigurationWithFilterPrefix(rName, rName, prefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAnalyticsConfigurationExists(resourceName, &ac),
				),
			},
			{
				Config: testAccBucketAnalyticsConfiguration(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAnalyticsConfigurationExists(resourceName, &ac),
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

func TestAccS3BucketAnalyticsConfiguration_WithStorageClassAnalysis_empty(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketAnalyticsConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccBucketAnalyticsConfigurationWithEmptyStorageClassAnalysis(rName, rName),
				ExpectError: regexp.MustCompile(`running pre-apply refresh`),
			},
		},
	})
}

func TestAccS3BucketAnalyticsConfiguration_WithStorageClassAnalysis_default(t *testing.T) {
	var ac s3.AnalyticsConfiguration
	resourceName := "aws_s3_bucket_analytics_configuration.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketAnalyticsConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketAnalyticsConfigurationWithDefaultStorageClassAnalysis(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAnalyticsConfigurationExists(resourceName, &ac),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.0.data_export.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.0.data_export.0.output_schema_version", "V_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.0.data_export.0.destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.0.data_export.0.destination.0.s3_bucket_destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.0.data_export.0.destination.0.s3_bucket_destination.0.format", "CSV"),
					resource.TestCheckResourceAttrPair(resourceName, "storage_class_analysis.0.data_export.0.destination.0.s3_bucket_destination.0.bucket_arn", "aws_s3_bucket.destination", "arn"),
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
	var ac s3.AnalyticsConfiguration
	resourceName := "aws_s3_bucket_analytics_configuration.test"

	rInt := sdkacctest.RandInt()
	rName := fmt.Sprintf("tf-acc-test-%d", rInt)
	prefix := fmt.Sprintf("prefix-%d/", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketAnalyticsConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketAnalyticsConfigurationWithFullStorageClassAnalysis(rName, rName, prefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAnalyticsConfigurationExists(resourceName, &ac),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.0.data_export.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.0.data_export.0.output_schema_version", "V_1"),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.0.data_export.0.destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.0.data_export.0.destination.0.s3_bucket_destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.0.data_export.0.destination.0.s3_bucket_destination.0.format", "CSV"),
					resource.TestCheckResourceAttrPair(resourceName, "storage_class_analysis.0.data_export.0.destination.0.s3_bucket_destination.0.bucket_arn", "aws_s3_bucket.destination", "arn"),
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

func testAccCheckBucketAnalyticsConfigurationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_s3_bucket_analytics_configuration" {
			continue
		}

		bucket, name, err := tfs3.BucketAnalyticsConfigurationParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		return tfs3.WaitForDeleteBucketAnalyticsConfiguration(conn, bucket, name, 1*time.Minute)

	}
	return nil
}

func testAccCheckBucketAnalyticsConfigurationExists(n string, ac *s3.AnalyticsConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn
		output, err := conn.GetBucketAnalyticsConfiguration(&s3.GetBucketAnalyticsConfigurationInput{
			Bucket: aws.String(rs.Primary.Attributes["bucket"]),
			Id:     aws.String(rs.Primary.Attributes["name"]),
		})

		if err != nil {
			return err
		}

		if output == nil || output.AnalyticsConfiguration == nil {
			return fmt.Errorf("error reading S3 Bucket Analytics Configuration %q: empty response", rs.Primary.ID)
		}

		*ac = *output.AnalyticsConfiguration

		return nil
	}
}

func testAccCheckBucketAnalyticsConfigurationRemoved(name, bucket string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn
		return tfs3.WaitForDeleteBucketAnalyticsConfiguration(conn, bucket, name, 1*time.Minute)
	}
}

func testAccBucketAnalyticsConfiguration(name, bucket string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_analytics_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = "%s"
}

resource "aws_s3_bucket" "test" {
  bucket = "%s"
}
`, name, bucket)
}

func testAccBucketAnalyticsConfiguration_removed(bucket string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = "%s"
}
`, bucket)
}

func testAccBucketAnalyticsConfigurationUpdateBucket(name, originalBucket, updatedBucket string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_analytics_configuration" "test" {
  bucket = aws_s3_bucket.test_2.bucket
  name   = "%s"
}

resource "aws_s3_bucket" "test" {
  bucket = "%s"
}

resource "aws_s3_bucket" "test_2" {
  bucket = "%s"
}
`, name, originalBucket, updatedBucket)
}

func testAccBucketAnalyticsConfigurationWithEmptyFilter(name, bucket string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_analytics_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = "%s"

  filter {
  }
}

resource "aws_s3_bucket" "test" {
  bucket = "%s"
}
`, name, bucket)
}

func testAccBucketAnalyticsConfigurationWithFilterPrefix(name, bucket, prefix string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_analytics_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = "%s"

  filter {
    prefix = "%s"
  }
}

resource "aws_s3_bucket" "test" {
  bucket = "%s"
}
`, name, prefix, bucket)
}

func testAccBucketAnalyticsConfigurationWithFilterSingleTag(name, bucket, tag string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_analytics_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = "%s"

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

func testAccBucketAnalyticsConfigurationWithFilterMultipleTags(name, bucket, tag1, tag2 string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_analytics_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = "%s"

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

func testAccBucketAnalyticsConfigurationWithFilterPrefixAndTags(name, bucket, prefix, tag1, tag2 string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_analytics_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = "%s"

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

func testAccBucketAnalyticsConfigurationWithEmptyStorageClassAnalysis(name, bucket string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_analytics_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = "%s"

  storage_class_analysis {
  }
}

resource "aws_s3_bucket" "test" {
  bucket = "%s"
}
`, name, bucket)
}

func testAccBucketAnalyticsConfigurationWithDefaultStorageClassAnalysis(name, bucket string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_analytics_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = "%s"

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
  bucket = "%[2]s"
}

resource "aws_s3_bucket" "destination" {
  bucket = "%[2]s-destination"
}
`, name, bucket)
}

func testAccBucketAnalyticsConfigurationWithFullStorageClassAnalysis(name, bucket, prefix string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_analytics_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = "%s"

  storage_class_analysis {
    data_export {
      output_schema_version = "V_1"

      destination {
        s3_bucket_destination {
          format     = "CSV"
          bucket_arn = aws_s3_bucket.destination.arn
          prefix     = "%s"
        }
      }
    }
  }
}

resource "aws_s3_bucket" "test" {
  bucket = "%[3]s"
}

resource "aws_s3_bucket" "destination" {
  bucket = "%[3]s-destination"
}
`, name, prefix, bucket)
}

func TestExpandAnalyticsFilter(t *testing.T) {
	testCases := map[string]struct {
		Input    []interface{}
		Expected *s3.AnalyticsFilter
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
			Expected: &s3.AnalyticsFilter{
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
			Expected: &s3.AnalyticsFilter{
				And: &s3.AnalyticsAndOperator{
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
			Expected: &s3.AnalyticsFilter{
				And: &s3.AnalyticsAndOperator{
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
			Expected: &s3.AnalyticsFilter{
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
			Expected: &s3.AnalyticsFilter{
				And: &s3.AnalyticsAndOperator{
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
		value := tfs3.ExpandAnalyticsFilter(tc.Input)

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

func TestExpandStorageClassAnalysis(t *testing.T) {
	testCases := map[string]struct {
		Input    []interface{}
		Expected *s3.StorageClassAnalysis
	}{
		"nil input": {
			Input:    nil,
			Expected: &s3.StorageClassAnalysis{},
		},
		"empty input": {
			Input:    []interface{}{},
			Expected: &s3.StorageClassAnalysis{},
		},
		"nil array": {
			Input: []interface{}{
				nil,
			},
			Expected: &s3.StorageClassAnalysis{},
		},
		"empty data_export": {
			Input: []interface{}{
				map[string]interface{}{
					"data_export": []interface{}{},
				},
			},
			Expected: &s3.StorageClassAnalysis{
				DataExport: &s3.StorageClassAnalysisDataExport{},
			},
		},
		"data_export complete": {
			Input: []interface{}{
				map[string]interface{}{
					"data_export": []interface{}{
						map[string]interface{}{
							"output_schema_version": s3.StorageClassAnalysisSchemaVersionV1,
							"destination":           []interface{}{},
						},
					},
				},
			},
			Expected: &s3.StorageClassAnalysis{
				DataExport: &s3.StorageClassAnalysisDataExport{
					OutputSchemaVersion: aws.String(s3.StorageClassAnalysisSchemaVersionV1),
					Destination:         &s3.AnalyticsExportDestination{},
				},
			},
		},
		"empty s3_bucket_destination": {
			Input: []interface{}{
				map[string]interface{}{
					"data_export": []interface{}{
						map[string]interface{}{
							"destination": []interface{}{
								map[string]interface{}{
									"s3_bucket_destination": []interface{}{},
								},
							},
						},
					},
				},
			},
			Expected: &s3.StorageClassAnalysis{
				DataExport: &s3.StorageClassAnalysisDataExport{
					Destination: &s3.AnalyticsExportDestination{
						S3BucketDestination: &s3.AnalyticsS3BucketDestination{},
					},
				},
			},
		},
		"s3_bucket_destination complete": {
			Input: []interface{}{
				map[string]interface{}{
					"data_export": []interface{}{
						map[string]interface{}{
							"destination": []interface{}{
								map[string]interface{}{
									"s3_bucket_destination": []interface{}{
										map[string]interface{}{
											"bucket_arn":        "arn:aws:s3", //lintignore:AWSAT005
											"bucket_account_id": "1234567890",
											"format":            s3.AnalyticsS3ExportFileFormatCsv,
											"prefix":            "prefix/",
										},
									},
								},
							},
						},
					},
				},
			},
			Expected: &s3.StorageClassAnalysis{
				DataExport: &s3.StorageClassAnalysisDataExport{
					Destination: &s3.AnalyticsExportDestination{
						S3BucketDestination: &s3.AnalyticsS3BucketDestination{
							Bucket:          aws.String("arn:aws:s3"), //lintignore:AWSAT005
							BucketAccountId: aws.String("1234567890"),
							Format:          aws.String(s3.AnalyticsS3ExportFileFormatCsv),
							Prefix:          aws.String("prefix/"),
						},
					},
				},
			},
		},
		"s3_bucket_destination required": {
			Input: []interface{}{
				map[string]interface{}{
					"data_export": []interface{}{
						map[string]interface{}{
							"destination": []interface{}{
								map[string]interface{}{
									"s3_bucket_destination": []interface{}{
										map[string]interface{}{
											"bucket_arn": "arn:aws:s3", //lintignore:AWSAT005
											"format":     s3.AnalyticsS3ExportFileFormatCsv,
										},
									},
								},
							},
						},
					},
				},
			},
			Expected: &s3.StorageClassAnalysis{
				DataExport: &s3.StorageClassAnalysisDataExport{
					Destination: &s3.AnalyticsExportDestination{
						S3BucketDestination: &s3.AnalyticsS3BucketDestination{
							Bucket:          aws.String("arn:aws:s3"), //lintignore:AWSAT005
							BucketAccountId: nil,
							Format:          aws.String(s3.AnalyticsS3ExportFileFormatCsv),
							Prefix:          nil,
						},
					},
				},
			},
		},
	}

	for k, tc := range testCases {
		value := tfs3.ExpandStorageClassAnalysis(tc.Input)

		if !reflect.DeepEqual(value, tc.Expected) {
			t.Errorf("Case %q:\nGot:\n%v\nExpected:\n%v", k, value, tc.Expected)
		}
	}
}

func TestFlattenAnalyticsFilter(t *testing.T) {
	testCases := map[string]struct {
		Input    *s3.AnalyticsFilter
		Expected []map[string]interface{}
	}{
		"nil input": {
			Input:    nil,
			Expected: nil,
		},
		"empty input": {
			Input:    &s3.AnalyticsFilter{},
			Expected: nil,
		},
		"prefix only": {
			Input: &s3.AnalyticsFilter{
				Prefix: aws.String("prefix/"),
			},
			Expected: []map[string]interface{}{
				{
					"prefix": "prefix/",
				},
			},
		},
		"prefix and single tag": {
			Input: &s3.AnalyticsFilter{
				And: &s3.AnalyticsAndOperator{
					Prefix: aws.String("prefix/"),
					Tags: []*s3.Tag{
						{
							Key:   aws.String("tag1key"),
							Value: aws.String("tag1value"),
						},
					},
				},
			},
			Expected: []map[string]interface{}{
				{
					"prefix": "prefix/",
					"tags": map[string]string{
						"tag1key": "tag1value",
					},
				},
			},
		},
		"prefix and multiple tags": {
			Input: &s3.AnalyticsFilter{
				And: &s3.AnalyticsAndOperator{
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
			Expected: []map[string]interface{}{
				{
					"prefix": "prefix/",
					"tags": map[string]string{
						"tag1key": "tag1value",
						"tag2key": "tag2value",
					},
				},
			},
		},
		"single tag only": {
			Input: &s3.AnalyticsFilter{
				Tag: &s3.Tag{
					Key:   aws.String("tag1key"),
					Value: aws.String("tag1value"),
				},
			},
			Expected: []map[string]interface{}{
				{
					"tags": map[string]string{
						"tag1key": "tag1value",
					},
				},
			},
		},
		"multiple tags only": {
			Input: &s3.AnalyticsFilter{
				And: &s3.AnalyticsAndOperator{
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
			Expected: []map[string]interface{}{
				{
					"tags": map[string]string{
						"tag1key": "tag1value",
						"tag2key": "tag2value",
					},
				},
			},
		},
	}

	for k, tc := range testCases {
		value := tfs3.FlattenAnalyticsFilter(tc.Input)

		if !reflect.DeepEqual(value, tc.Expected) {
			t.Errorf("Case %q: Got:\n%v\n\nExpected:\n%v", k, value, tc.Expected)
		}
	}
}

func TestFlattenStorageClassAnalysis(t *testing.T) {
	testCases := map[string]struct {
		Input    *s3.StorageClassAnalysis
		Expected []map[string]interface{}
	}{
		"nil value": {
			Input:    nil,
			Expected: []map[string]interface{}{},
		},
		"empty root": {
			Input:    &s3.StorageClassAnalysis{},
			Expected: []map[string]interface{}{},
		},
		"empty data_export": {
			Input: &s3.StorageClassAnalysis{
				DataExport: &s3.StorageClassAnalysisDataExport{},
			},
			Expected: []map[string]interface{}{
				{
					"data_export": []interface{}{
						map[string]interface{}{},
					},
				},
			},
		},
		"data_export complete": {
			Input: &s3.StorageClassAnalysis{
				DataExport: &s3.StorageClassAnalysisDataExport{
					OutputSchemaVersion: aws.String(s3.StorageClassAnalysisSchemaVersionV1),
					Destination:         &s3.AnalyticsExportDestination{},
				},
			},
			Expected: []map[string]interface{}{
				{
					"data_export": []interface{}{
						map[string]interface{}{
							"output_schema_version": s3.StorageClassAnalysisSchemaVersionV1,
							"destination":           []interface{}{},
						},
					},
				},
			},
		},
		"s3_bucket_destination required": {
			Input: &s3.StorageClassAnalysis{
				DataExport: &s3.StorageClassAnalysisDataExport{
					Destination: &s3.AnalyticsExportDestination{
						S3BucketDestination: &s3.AnalyticsS3BucketDestination{
							Bucket: aws.String("arn:aws:s3"), //lintignore:AWSAT005
							Format: aws.String(s3.AnalyticsS3ExportFileFormatCsv),
						},
					},
				},
			},
			Expected: []map[string]interface{}{
				{
					"data_export": []interface{}{
						map[string]interface{}{
							"destination": []interface{}{
								map[string]interface{}{
									"s3_bucket_destination": []interface{}{
										map[string]interface{}{
											"bucket_arn": "arn:aws:s3", //lintignore:AWSAT005
											"format":     s3.AnalyticsS3ExportFileFormatCsv,
										},
									},
								},
							},
						},
					},
				},
			},
		},
		"s3_bucket_destination complete": {
			Input: &s3.StorageClassAnalysis{
				DataExport: &s3.StorageClassAnalysisDataExport{
					Destination: &s3.AnalyticsExportDestination{
						S3BucketDestination: &s3.AnalyticsS3BucketDestination{
							Bucket:          aws.String("arn:aws:s3"), //lintignore:AWSAT005
							BucketAccountId: aws.String("1234567890"),
							Format:          aws.String(s3.AnalyticsS3ExportFileFormatCsv),
							Prefix:          aws.String("prefix/"),
						},
					},
				},
			},
			Expected: []map[string]interface{}{
				{
					"data_export": []interface{}{
						map[string]interface{}{
							"destination": []interface{}{
								map[string]interface{}{
									"s3_bucket_destination": []interface{}{
										map[string]interface{}{
											"bucket_arn":        "arn:aws:s3", //lintignore:AWSAT005
											"bucket_account_id": "1234567890",
											"format":            s3.AnalyticsS3ExportFileFormatCsv,
											"prefix":            "prefix/",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for k, tc := range testCases {
		value := tfs3.FlattenStorageClassAnalysis(tc.Input)

		if !reflect.DeepEqual(value, tc.Expected) {
			t.Errorf("Case %q:\nGot:\n%v\nExpected:\n%v", k, value, tc.Expected)
		}
	}
}
