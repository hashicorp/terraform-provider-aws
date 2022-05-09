package s3_test

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
	"sort"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
)

func TestExpandS3MetricsFilter(t *testing.T) {
	testCases := []struct {
		Config                  map[string]interface{}
		ExpectedS3MetricsFilter *s3.MetricsFilter
	}{
		{
			Config: map[string]interface{}{
				"prefix": "prefix/",
			},
			ExpectedS3MetricsFilter: &s3.MetricsFilter{
				Prefix: aws.String("prefix/"),
			},
		},
		{
			Config: map[string]interface{}{
				"prefix": "prefix/",
				"tags": map[string]interface{}{
					"tag1key": "tag1value",
				},
			},
			ExpectedS3MetricsFilter: &s3.MetricsFilter{
				And: &s3.MetricsAndOperator{
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
			Config: map[string]interface{}{
				"prefix": "prefix/",
				"tags": map[string]interface{}{
					"tag1key": "tag1value",
					"tag2key": "tag2value",
				},
			},
			ExpectedS3MetricsFilter: &s3.MetricsFilter{
				And: &s3.MetricsAndOperator{
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
			Config: map[string]interface{}{
				"tags": map[string]interface{}{
					"tag1key": "tag1value",
				},
			},
			ExpectedS3MetricsFilter: &s3.MetricsFilter{
				Tag: &s3.Tag{
					Key:   aws.String("tag1key"),
					Value: aws.String("tag1value"),
				},
			},
		},
		{
			Config: map[string]interface{}{
				"tags": map[string]interface{}{
					"tag1key": "tag1value",
					"tag2key": "tag2value",
				},
			},
			ExpectedS3MetricsFilter: &s3.MetricsFilter{
				And: &s3.MetricsAndOperator{
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
		value := tfs3.ExpandMetricsFilter(tc.Config)

		// Sort tags by key for consistency
		if value.And != nil && value.And.Tags != nil {
			sort.Slice(value.And.Tags, func(i, j int) bool {
				return *value.And.Tags[i].Key < *value.And.Tags[j].Key
			})
		}

		// Convert to strings to avoid dealing with pointers
		valueS := fmt.Sprintf("%v", value)
		expectedValueS := fmt.Sprintf("%v", tc.ExpectedS3MetricsFilter)

		if valueS != expectedValueS {
			t.Fatalf("Case #%d: Given:\n%s\n\nExpected:\n%s", i, valueS, expectedValueS)
		}
	}
}

func TestFlattenS3MetricsFilter(t *testing.T) {
	testCases := []struct {
		S3MetricsFilter *s3.MetricsFilter
		ExpectedConfig  map[string]interface{}
	}{
		{
			S3MetricsFilter: &s3.MetricsFilter{
				Prefix: aws.String("prefix/"),
			},
			ExpectedConfig: map[string]interface{}{
				"prefix": "prefix/",
			},
		},
		{
			S3MetricsFilter: &s3.MetricsFilter{
				And: &s3.MetricsAndOperator{
					Prefix: aws.String("prefix/"),
					Tags: []*s3.Tag{
						{
							Key:   aws.String("tag1key"),
							Value: aws.String("tag1value"),
						},
					},
				},
			},
			ExpectedConfig: map[string]interface{}{
				"prefix": "prefix/",
				"tags": map[string]string{
					"tag1key": "tag1value",
				},
			},
		},
		{
			S3MetricsFilter: &s3.MetricsFilter{
				And: &s3.MetricsAndOperator{
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
			ExpectedConfig: map[string]interface{}{
				"prefix": "prefix/",
				"tags": map[string]string{
					"tag1key": "tag1value",
					"tag2key": "tag2value",
				},
			},
		},
		{
			S3MetricsFilter: &s3.MetricsFilter{
				Tag: &s3.Tag{
					Key:   aws.String("tag1key"),
					Value: aws.String("tag1value"),
				},
			},
			ExpectedConfig: map[string]interface{}{
				"tags": map[string]string{
					"tag1key": "tag1value",
				},
			},
		},
		{
			S3MetricsFilter: &s3.MetricsFilter{
				And: &s3.MetricsAndOperator{
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
			ExpectedConfig: map[string]interface{}{
				"tags": map[string]string{
					"tag1key": "tag1value",
					"tag2key": "tag2value",
				},
			},
		},
	}

	for i, tc := range testCases {
		value := tfs3.FlattenMetricsFilter(tc.S3MetricsFilter)

		if !reflect.DeepEqual(value, tc.ExpectedConfig) {
			t.Fatalf("Case #%d: Given:\n%s\n\nExpected:\n%s", i, value, tc.ExpectedConfig)
		}
	}
}

func TestBucketMetricParseID(t *testing.T) {
	validIds := []string{
		"foo:bar",
		"my-bucket:entire-bucket",
	}

	for _, s := range validIds {
		_, _, err := tfs3.BucketMetricParseID(s)
		if err != nil {
			t.Fatalf("%s should be a valid S3 bucket metrics configuration id: %s", s, err)
		}
	}

	invalidIds := []string{
		"",
		"foo",
		"foo:bar:",
		"foo:bar:baz",
		"foo::bar",
		"foo.bar",
	}

	for _, s := range invalidIds {
		_, _, err := tfs3.BucketMetricParseID(s)
		if err == nil {
			t.Fatalf("%s should not be a valid S3 bucket metrics configuration id", s)
		}
	}
}

func TestAccS3BucketMetric_basic(t *testing.T) {
	var conf s3.MetricsConfiguration
	rInt := sdkacctest.RandInt()
	resourceName := "aws_s3_bucket_metric.test"

	bucketName := fmt.Sprintf("tf-acc-%d", rInt)
	metricName := t.Name()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketMetricDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketMetricsWithoutFilterConfig(bucketName, metricName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketMetricsExistsConfig(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", metricName),
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

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/11813
// Disallow Empty filter block
func TestAccS3BucketMetric_withEmptyFilter(t *testing.T) {
	var conf s3.MetricsConfiguration
	rInt := sdkacctest.RandInt()
	resourceName := "aws_s3_bucket_metric.test"

	bucketName := fmt.Sprintf("tf-acc-%d", rInt)
	metricName := t.Name()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketMetricDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketMetricsWithEmptyFilterConfig(bucketName, metricName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketMetricsExistsConfig(resourceName, &conf),
				),
				ExpectError: regexp.MustCompile(`one of .* must be specified`),
			},
		},
	})
}

func TestAccS3BucketMetric_withFilterPrefix(t *testing.T) {
	var conf s3.MetricsConfiguration
	rInt := sdkacctest.RandInt()
	resourceName := "aws_s3_bucket_metric.test"

	bucketName := fmt.Sprintf("tf-acc-%d", rInt)
	metricName := t.Name()
	prefix := fmt.Sprintf("prefix-%d/", rInt)
	prefixUpdate := fmt.Sprintf("prefix-update-%d/", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketMetricDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketMetricsWithFilterPrefixConfig(bucketName, metricName, prefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketMetricsExistsConfig(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", prefix),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", "0"),
				),
			},
			{
				Config: testAccBucketMetricsWithFilterPrefixConfig(bucketName, metricName, prefixUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketMetricsExistsConfig(resourceName, &conf),
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

func TestAccS3BucketMetric_withFilterPrefixAndMultipleTags(t *testing.T) {
	var conf s3.MetricsConfiguration
	rInt := sdkacctest.RandInt()
	resourceName := "aws_s3_bucket_metric.test"

	bucketName := fmt.Sprintf("tf-acc-%d", rInt)
	metricName := t.Name()
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
		CheckDestroy:      testAccCheckBucketMetricDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketMetricsWithFilterPrefixAndMultipleTagsConfig(bucketName, metricName, prefix, tag1, tag2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketMetricsExistsConfig(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", prefix),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.tag1", tag1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.tag2", tag2),
				),
			},
			{
				Config: testAccBucketMetricsWithFilterPrefixAndMultipleTagsConfig(bucketName, metricName, prefixUpdate, tag1Update, tag2Update),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketMetricsExistsConfig(resourceName, &conf),
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

func TestAccS3BucketMetric_withFilterPrefixAndSingleTag(t *testing.T) {
	var conf s3.MetricsConfiguration
	rInt := sdkacctest.RandInt()
	resourceName := "aws_s3_bucket_metric.test"

	bucketName := fmt.Sprintf("tf-acc-%d", rInt)
	metricName := t.Name()
	prefix := fmt.Sprintf("prefix-%d/", rInt)
	prefixUpdate := fmt.Sprintf("prefix-update-%d/", rInt)
	tag1 := fmt.Sprintf("tag-%d", rInt)
	tag1Update := fmt.Sprintf("tag-update-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketMetricDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketMetricsWithFilterPrefixAndSingleTagConfig(bucketName, metricName, prefix, tag1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketMetricsExistsConfig(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", prefix),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.tag1", tag1),
				),
			},
			{
				Config: testAccBucketMetricsWithFilterPrefixAndSingleTagConfig(bucketName, metricName, prefixUpdate, tag1Update),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketMetricsExistsConfig(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", prefixUpdate),
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

func TestAccS3BucketMetric_withFilterMultipleTags(t *testing.T) {
	var conf s3.MetricsConfiguration
	rInt := sdkacctest.RandInt()
	resourceName := "aws_s3_bucket_metric.test"

	bucketName := fmt.Sprintf("tf-acc-%d", rInt)
	metricName := t.Name()
	tag1 := fmt.Sprintf("tag1-%d", rInt)
	tag1Update := fmt.Sprintf("tag1-update-%d", rInt)
	tag2 := fmt.Sprintf("tag2-%d", rInt)
	tag2Update := fmt.Sprintf("tag2-update-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketMetricDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketMetricsWithFilterMultipleTagsConfig(bucketName, metricName, tag1, tag2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketMetricsExistsConfig(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.tag1", tag1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.tag2", tag2),
				),
			},
			{
				Config: testAccBucketMetricsWithFilterMultipleTagsConfig(bucketName, metricName, tag1Update, tag2Update),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketMetricsExistsConfig(resourceName, &conf),
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

func TestAccS3BucketMetric_withFilterSingleTag(t *testing.T) {
	var conf s3.MetricsConfiguration
	rInt := sdkacctest.RandInt()
	resourceName := "aws_s3_bucket_metric.test"

	bucketName := fmt.Sprintf("tf-acc-%d", rInt)
	metricName := t.Name()
	tag1 := fmt.Sprintf("tag-%d", rInt)
	tag1Update := fmt.Sprintf("tag-update-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketMetricDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketMetricsWithFilterSingleTagConfig(bucketName, metricName, tag1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketMetricsExistsConfig(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.tag1", tag1),
				),
			},
			{
				Config: testAccBucketMetricsWithFilterSingleTagConfig(bucketName, metricName, tag1Update),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketMetricsExistsConfig(resourceName, &conf),
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

func testAccCheckBucketMetricDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_s3_bucket_metric" {
			continue
		}

		bucket, name, err := tfs3.BucketMetricParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		err = resource.Retry(1*time.Minute, func() *resource.RetryError {
			input := &s3.GetBucketMetricsConfigurationInput{
				Bucket: aws.String(bucket),
				Id:     aws.String(name),
			}
			log.Printf("[DEBUG] Reading S3 bucket metrics configuration: %s", input)
			output, err := conn.GetBucketMetricsConfiguration(input)
			if err != nil {
				if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) || tfawserr.ErrMessageContains(err, "NoSuchConfiguration", "The specified configuration does not exist.") {
					return nil
				}
				return resource.NonRetryableError(err)
			}
			if output.MetricsConfiguration != nil {
				return resource.RetryableError(fmt.Errorf("S3 bucket metrics configuration exists: %v", output))
			}

			return nil
		})

		if err != nil {
			return err
		}
	}
	return nil
}

func testAccCheckBucketMetricsExistsConfig(n string, res *s3.MetricsConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No S3 bucket metrics configuration ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn
		bucket, name, err := tfs3.BucketMetricParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &s3.GetBucketMetricsConfigurationInput{
			Bucket: aws.String(bucket),
			Id:     aws.String(name),
		}
		log.Printf("[DEBUG] Reading S3 bucket metrics configuration: %s", input)
		output, err := conn.GetBucketMetricsConfiguration(input)
		if err != nil {
			return err
		}

		*res = *output.MetricsConfiguration

		return nil
	}
}

func testAccBucketMetricsBucketConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = "%s"
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.bucket.id
  acl    = "public-read"
}
`, name)
}

func testAccBucketMetricsWithEmptyFilterConfig(bucketName, metricName string) string {
	return fmt.Sprintf(`
%s

resource "aws_s3_bucket_metric" "test" {
  bucket = aws_s3_bucket.bucket.id
  name   = "%s"

  filter {}
}
`, testAccBucketMetricsBucketConfig(bucketName), metricName)
}

func testAccBucketMetricsWithFilterPrefixConfig(bucketName, metricName, prefix string) string {
	return fmt.Sprintf(`
%s

resource "aws_s3_bucket_metric" "test" {
  bucket = aws_s3_bucket.bucket.id
  name   = "%s"

  filter {
    prefix = "%s"
  }
}
`, testAccBucketMetricsBucketConfig(bucketName), metricName, prefix)
}

func testAccBucketMetricsWithFilterPrefixAndMultipleTagsConfig(bucketName, metricName, prefix, tag1, tag2 string) string {
	return fmt.Sprintf(`
%s

resource "aws_s3_bucket_metric" "test" {
  bucket = aws_s3_bucket.bucket.id
  name   = "%s"

  filter {
    prefix = "%s"

    tags = {
      "tag1" = "%s"
      "tag2" = "%s"
    }
  }
}
`, testAccBucketMetricsBucketConfig(bucketName), metricName, prefix, tag1, tag2)
}

func testAccBucketMetricsWithFilterPrefixAndSingleTagConfig(bucketName, metricName, prefix, tag string) string {
	return fmt.Sprintf(`
%s

resource "aws_s3_bucket_metric" "test" {
  bucket = aws_s3_bucket.bucket.id
  name   = "%s"

  filter {
    prefix = "%s"

    tags = {
      "tag1" = "%s"
    }
  }
}
`, testAccBucketMetricsBucketConfig(bucketName), metricName, prefix, tag)
}

func testAccBucketMetricsWithFilterMultipleTagsConfig(bucketName, metricName, tag1, tag2 string) string {
	return fmt.Sprintf(`
%s

resource "aws_s3_bucket_metric" "test" {
  bucket = aws_s3_bucket.bucket.id
  name   = "%s"

  filter {
    tags = {
      "tag1" = "%s"
      "tag2" = "%s"
    }
  }
}
`, testAccBucketMetricsBucketConfig(bucketName), metricName, tag1, tag2)
}

func testAccBucketMetricsWithFilterSingleTagConfig(bucketName, metricName, tag string) string {
	return fmt.Sprintf(`
%s

resource "aws_s3_bucket_metric" "test" {
  bucket = aws_s3_bucket.bucket.id
  name   = "%s"

  filter {
    tags = {
      "tag1" = "%s"
    }
  }
}
`, testAccBucketMetricsBucketConfig(bucketName), metricName, tag)
}

func testAccBucketMetricsWithoutFilterConfig(bucketName, metricName string) string {
	return fmt.Sprintf(`
%s

resource "aws_s3_bucket_metric" "test" {
  bucket = aws_s3_bucket.bucket.id
  name   = "%s"
}
`, testAccBucketMetricsBucketConfig(bucketName), metricName)
}
