package aws

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSS3BucketAnalyticsConfiguration_basic(t *testing.T) {
	var ac s3.AnalyticsConfiguration
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_s3_bucket_analytics_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketAnalyticsConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketAnalyticsConfiguration(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketAnalyticsConfigurationExists(resourceName, &ac),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "aws_s3_bucket.test", "bucket"),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.#", "0"),
				),
			},
		},
	})
}

func TestAccAWSS3BucketAnalyticsConfiguration_removed(t *testing.T) {
	var ac s3.AnalyticsConfiguration
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_s3_bucket_analytics_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketAnalyticsConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketAnalyticsConfiguration(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketAnalyticsConfigurationExists(resourceName, &ac),
				),
			},
			{
				Config: testAccAWSS3BucketAnalyticsConfiguration_removed(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketAnalyticsConfigurationRemoved(rName, rName),
				),
			},
		},
	})
}

func TestAccAWSS3BucketAnalyticsConfiguration_updateBasic(t *testing.T) {
	var ac s3.AnalyticsConfiguration
	originalACName := acctest.RandomWithPrefix("tf-acc-test")
	originalBucketName := acctest.RandomWithPrefix("tf-acc-test")
	updatedACName := acctest.RandomWithPrefix("tf-acc-test")
	updatedBucketName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_s3_bucket_analytics_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketAnalyticsConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketAnalyticsConfiguration(originalACName, originalBucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketAnalyticsConfigurationExists(resourceName, &ac),
					resource.TestCheckResourceAttr(resourceName, "name", originalACName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "aws_s3_bucket.test", "bucket"),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.#", "0"),
				),
			},
			{
				Config: testAccAWSS3BucketAnalyticsConfiguration(updatedACName, originalBucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketAnalyticsConfigurationExists(resourceName, &ac),
					testAccCheckAWSS3BucketAnalyticsConfigurationRemoved(originalACName, originalBucketName),
					resource.TestCheckResourceAttr(resourceName, "name", updatedACName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "aws_s3_bucket.test", "bucket"),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.#", "0"),
				),
			},
			{
				Config: testAccAWSS3BucketAnalyticsConfigurationUpdateBucket(updatedACName, originalBucketName, updatedBucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketAnalyticsConfigurationExists(resourceName, &ac),
					testAccCheckAWSS3BucketAnalyticsConfigurationRemoved(updatedACName, originalBucketName),
					resource.TestCheckResourceAttr(resourceName, "name", updatedACName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "aws_s3_bucket.test_2", "bucket"),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.#", "0"),
				),
			},
		},
	})
}

func TestAccAWSS3BucketAnalyticsConfiguration_WithEmptyFilter(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketMetricDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSS3BucketAnalyticsConfigurationWithEmptyFilter(rName, rName),
				ExpectError: regexp.MustCompile(`config is invalid: 2 problems:`),
			},
		},
	})
}

func TestAccAWSS3BucketAnalyticsConfiguration_WithFilterPrefix(t *testing.T) {
	var ac s3.AnalyticsConfiguration
	rInt := acctest.RandInt()
	resourceName := "aws_s3_bucket_analytics_configuration.test"

	rName := fmt.Sprintf("tf-acc-test-%d", rInt)
	prefix := fmt.Sprintf("prefix-%d/", rInt)
	prefixUpdate := fmt.Sprintf("prefix-update-%d/", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketMetricDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketAnalyticsConfigurationWithFilterPrefix(rName, rName, prefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketAnalyticsConfigurationExists(resourceName, &ac),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", prefix),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", "0"),
				),
			},
			{
				Config: testAccAWSS3BucketAnalyticsConfigurationWithFilterPrefix(rName, rName, prefixUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketAnalyticsConfigurationExists(resourceName, &ac),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", prefixUpdate),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", "0"),
				),
			},
		},
	})
}

func testAccCheckAWSS3BucketAnalyticsConfigurationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).s3conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_s3_bucket_analytics_configuration" {
			continue
		}

		bucket, name, err := resourceAwsS3BucketAnalyticsConfigurationParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		return waitForDeleteS3BucketAnalyticsConfiguration(conn, bucket, name, 1*time.Minute)

	}
	return nil
}

func testAccCheckAWSS3BucketAnalyticsConfigurationExists(n string, ac *s3.AnalyticsConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).s3conn
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

func testAccCheckAWSS3BucketAnalyticsConfigurationRemoved(name, bucket string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).s3conn
		return waitForDeleteS3BucketAnalyticsConfiguration(conn, bucket, name, 1*time.Minute)
	}
}

func testAccAWSS3BucketAnalyticsConfiguration(name, bucket string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_analytics_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %q
}

resource "aws_s3_bucket" "test" {
  bucket = %q
}
`, name, bucket)
}

func testAccAWSS3BucketAnalyticsConfiguration_removed(bucket string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %q
}
`, bucket)
}

func testAccAWSS3BucketAnalyticsConfigurationUpdateBucket(name, originalBucket, updatedBucket string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_analytics_configuration" "test" {
  bucket = aws_s3_bucket.test_2.bucket
  name   = %q
}

resource "aws_s3_bucket" "test" {
  bucket = %q
}

resource "aws_s3_bucket" "test_2" {
  bucket = %q
}
`, name, originalBucket, updatedBucket)
}

func testAccAWSS3BucketAnalyticsConfigurationWithEmptyFilter(name, bucket string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_analytics_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %q

  filter {
  }
}

resource "aws_s3_bucket" "test" {
  bucket = %q
}
`, name, bucket)
}

func testAccAWSS3BucketAnalyticsConfigurationWithFilterPrefix(name, bucket, prefix string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_analytics_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %q

  filter {
    prefix = "%s"
  }
}

resource "aws_s3_bucket" "test" {
	bucket = %q
  }
  `, name, prefix, bucket)
}

func TestExpandS3AnalyticsFilter(t *testing.T) {
	testCases := []struct {
		Config          []interface{}
		AnalyticsFilter *s3.AnalyticsFilter
	}{
		{
			Config:          nil,
			AnalyticsFilter: nil,
		},
		{
			Config:          []interface{}{},
			AnalyticsFilter: nil,
		},
		{
			Config: []interface{}{
				map[string]interface{}{
					"prefix": "prefix/",
				},
			},
			AnalyticsFilter: &s3.AnalyticsFilter{
				Prefix: aws.String("prefix/"),
			},
		},
		{
			Config: []interface{}{
				map[string]interface{}{
					"prefix": "prefix/",
					"tags": map[string]interface{}{
						"tag1key": "tag1value",
					},
				},
			},
			AnalyticsFilter: &s3.AnalyticsFilter{
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
		{
			Config: []interface{}{map[string]interface{}{
				"prefix": "prefix/",
				"tags": map[string]interface{}{
					"tag1key": "tag1value",
					"tag2key": "tag2value",
				},
			},
			},
			AnalyticsFilter: &s3.AnalyticsFilter{
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
		{
			Config: []interface{}{
				map[string]interface{}{
					"tags": map[string]interface{}{
						"tag1key": "tag1value",
					},
				},
			},
			AnalyticsFilter: &s3.AnalyticsFilter{
				Tag: &s3.Tag{
					Key:   aws.String("tag1key"),
					Value: aws.String("tag1value"),
				},
			},
		},
		{
			Config: []interface{}{
				map[string]interface{}{
					"tags": map[string]interface{}{
						"tag1key": "tag1value",
						"tag2key": "tag2value",
					},
				},
			},
			AnalyticsFilter: &s3.AnalyticsFilter{
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

	for i, tc := range testCases {
		value := expandS3AnalyticsFilter(tc.Config)

		if value == nil {
			if tc.AnalyticsFilter == nil {
				continue
			} else {
				t.Errorf("Case #%d: Got nil\nExpected:\n%v", i, tc.AnalyticsFilter)
			}
		}

		if tc.AnalyticsFilter == nil {
			t.Errorf("Case #%d: Got: %v\nExpected: nil", i, value)
		}

		// Sort tags by key for consistency
		if value.And != nil && value.And.Tags != nil {
			sort.Slice(value.And.Tags, func(i, j int) bool {
				return *value.And.Tags[i].Key < *value.And.Tags[j].Key
			})
		}

		// Convert to strings to avoid dealing with pointers
		valueS := fmt.Sprintf("%v", value)
		expectedValueS := fmt.Sprintf("%v", tc.AnalyticsFilter)

		if valueS != expectedValueS {
			t.Errorf("Case #%d: Given:\n%s\n\nExpected:\n%s", i, valueS, expectedValueS)
		}
	}
}

func TestFlattenS3AnalyticsFilter(t *testing.T) {
	testCases := []struct {
		AnalyticsFilter *s3.AnalyticsFilter
		ExpectedFilter  []map[string]interface{}
	}{
		{
			AnalyticsFilter: nil,
			ExpectedFilter:  nil,
		},
		{
			AnalyticsFilter: &s3.AnalyticsFilter{},
			ExpectedFilter:  nil,
		},
		{
			AnalyticsFilter: &s3.AnalyticsFilter{
				Prefix: aws.String("prefix/"),
			},
			ExpectedFilter: []map[string]interface{}{
				{
					"prefix": "prefix/",
				},
			},
		},
		{
			AnalyticsFilter: &s3.AnalyticsFilter{
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
			ExpectedFilter: []map[string]interface{}{
				{
					"prefix": "prefix/",
					"tags": map[string]string{
						"tag1key": "tag1value",
					},
				},
			},
		},
		{
			AnalyticsFilter: &s3.AnalyticsFilter{
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
			ExpectedFilter: []map[string]interface{}{
				{
					"prefix": "prefix/",
					"tags": map[string]string{
						"tag1key": "tag1value",
						"tag2key": "tag2value",
					},
				},
			},
		},
		{
			AnalyticsFilter: &s3.AnalyticsFilter{
				Tag: &s3.Tag{
					Key:   aws.String("tag1key"),
					Value: aws.String("tag1value"),
				},
			},
			ExpectedFilter: []map[string]interface{}{
				{
					"tags": map[string]string{
						"tag1key": "tag1value",
					},
				},
			},
		},
		{
			AnalyticsFilter: &s3.AnalyticsFilter{
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
			ExpectedFilter: []map[string]interface{}{
				{
					"tags": map[string]string{
						"tag1key": "tag1value",
						"tag2key": "tag2value",
					},
				},
			},
		},
	}

	for i, tc := range testCases {
		value := flattenS3AnalyticsFilter(tc.AnalyticsFilter)

		if !reflect.DeepEqual(value, tc.ExpectedFilter) {
			t.Errorf("Case #%d: Got:\n%v\n\nExpected:\n%v", i, value, tc.ExpectedFilter)
		}
	}
}
