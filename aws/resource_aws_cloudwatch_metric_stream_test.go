package aws

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSCloudWatchMetricStream_basic(t *testing.T) {
	var metricStream cloudwatch.GetMetricStreamOutput
	resourceName := "aws_cloudwatch_metric_stream.test"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudwatch.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSCloudWatchMetricStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchMetricStreamConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchMetricStreamExists(resourceName, &metricStream),
					resource.TestCheckResourceAttr(resourceName, "name", testAccAWSCloudWatchName(rInt)),
					resource.TestCheckResourceAttr(resourceName, "output_format", "json"),
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

func TestAccAWSCloudWatchMetricStream_noName(t *testing.T) {
	var metricStream cloudwatch.GetMetricStreamOutput
	resourceName := "aws_cloudwatch_metric_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudwatch.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSCloudWatchMetricStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchMetricStreamConfigNoName(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchMetricStreamExists(resourceName, &metricStream),
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

func TestAccAWSCloudWatchMetricStream_namePrefix(t *testing.T) {
	var metricStream cloudwatch.GetMetricStreamOutput
	resourceName := "aws_cloudwatch_metric_stream.test"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudwatch.EndpointsID),
		IDRefreshName:     resourceName,
		IDRefreshIgnore:   []string{"name_prefix"},
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSCloudWatchMetricStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchMetricStreamConfigNamePrefix(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchMetricStreamExists(resourceName, &metricStream),
					testAccCheckCloudWatchMetricStreamGeneratedNamePrefix(resourceName, "test-stream"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
		},
	})
}

func TestAccAWSCloudWatchMetricStream_includeFilters(t *testing.T) {
	var metricStream cloudwatch.GetMetricStreamOutput
	resourceName := "aws_cloudwatch_metric_stream.test"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudwatch.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSCloudWatchMetricStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchMetricStreamConfigIncludeFilters(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchMetricStreamExists(resourceName, &metricStream),
					resource.TestCheckResourceAttr(resourceName, "name", testAccAWSCloudWatchName(rInt)),
					resource.TestCheckResourceAttr(resourceName, "output_format", "json"),
					resource.TestCheckResourceAttr(resourceName, "include_filter.#", "2"),
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

func TestAccAWSCloudWatchMetricStream_excludeFilters(t *testing.T) {
	var metricStream cloudwatch.GetMetricStreamOutput
	resourceName := "aws_cloudwatch_metric_stream.test"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudwatch.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSCloudWatchMetricStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchMetricStreamConfigExcludeFilters(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchMetricStreamExists(resourceName, &metricStream),
					resource.TestCheckResourceAttr(resourceName, "name", testAccAWSCloudWatchName(rInt)),
					resource.TestCheckResourceAttr(resourceName, "output_format", "json"),
					resource.TestCheckResourceAttr(resourceName, "exclude_filter.#", "2"),
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

func TestAccAWSCloudWatchMetricStream_update(t *testing.T) {
	var metricStream cloudwatch.GetMetricStreamOutput
	resourceName := "aws_cloudwatch_metric_stream.test"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudwatch.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSCloudWatchMetricStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchMetricStreamConfigUpdateArn(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchMetricStreamExists(resourceName, &metricStream),
					resource.TestCheckResourceAttr(resourceName, "name", testAccAWSCloudWatchName(rInt)),
					resource.TestCheckResourceAttr(resourceName, "output_format", "json"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCloudWatchMetricStreamConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchMetricStreamExists(resourceName, &metricStream),
					resource.TestCheckResourceAttr(resourceName, "name", testAccAWSCloudWatchName(rInt)),
					resource.TestCheckResourceAttr(resourceName, "output_format", "json"),
				),
			},
		},
	})
}

func TestAccAWSCloudWatchMetricStream_updateName(t *testing.T) {
	var metricStream cloudwatch.GetMetricStreamOutput
	resourceName := "aws_cloudwatch_metric_stream.test"
	rInt := acctest.RandInt()
	rInt2 := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudwatch.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSCloudWatchMetricStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchMetricStreamConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchMetricStreamExists(resourceName, &metricStream),
					resource.TestCheckResourceAttr(resourceName, "name", testAccAWSCloudWatchName(rInt)),
				),
			},
			{
				Config: testAccAWSCloudWatchMetricStreamConfig(rInt2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchMetricStreamExists(resourceName, &metricStream),
					resource.TestCheckResourceAttr(resourceName, "name", testAccAWSCloudWatchName(rInt2)),
					testAccCheckAWSCloudWatchMetricStreamDestroyPrevious(testAccAWSCloudWatchName(rInt)),
				),
			},
		},
	})
}

func testAccCheckCloudWatchMetricStreamGeneratedNamePrefix(resource, prefix string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Resource not found")
		}
		name, ok := r.Primary.Attributes["name"]
		if !ok {
			return fmt.Errorf("Name attr not found: %#v", r.Primary.Attributes)
		}
		if !strings.HasPrefix(name, prefix) {
			return fmt.Errorf("Name: %q, does not have prefix: %q", name, prefix)
		}
		return nil
	}
}

func testAccCheckCloudWatchMetricStreamExists(n string, metricStream *cloudwatch.GetMetricStreamOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).cloudwatchconn
		params := cloudwatch.GetMetricStreamInput{
			Name: aws.String(rs.Primary.ID),
		}

		resp, err := conn.GetMetricStream(&params)
		if err != nil {
			return err
		}

		*metricStream = *resp

		return nil
	}
}

func testAccCheckAWSCloudWatchMetricStreamDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cloudwatchconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudwatch_metric_stream" {
			continue
		}

		params := cloudwatch.GetMetricStreamInput{
			Name: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetMetricStream(&params)
		if err == nil {
			return fmt.Errorf("MetricStream still exists: %s", rs.Primary.ID)
		}
		if !isAWSErr(err, cloudwatch.ErrCodeResourceNotFoundException, "") {
			return err
		}
	}

	return nil
}

func testAccCheckAWSCloudWatchMetricStreamDestroyPrevious(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).cloudwatchconn

		params := cloudwatch.GetMetricStreamInput{
			Name: aws.String(name),
		}

		_, err := conn.GetMetricStream(&params)

		if err == nil {
			return fmt.Errorf("MetricStream still exists: %s", name)
		}

		if !isAWSErr(err, cloudwatch.ErrCodeResourceNotFoundException, "") {
			return err
		}

		return nil
	}
}

func testAccAWSCloudWatchName(rInt int) string {
	return fmt.Sprintf("terraform-test-metric-stream-%d", rInt)
}

func testAccAWSCloudWatchMetricStreamConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_stream" "test" {
  name          = "terraform-test-metric-stream-%d"
  role_arn      = aws_iam_role.metric_stream_to_firehose.arn
  firehose_arn  = aws_kinesis_firehose_delivery_stream.s3_stream.arn
  output_format = "json"
}

resource "aws_iam_role" "metric_stream_to_firehose" {
  name = "metric_stream_to_firehose_role-%d"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "streams.metrics.cloudwatch.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "metric_stream_to_firehose" {
  name = "default"
  role = aws_iam_role.metric_stream_to_firehose.id

  policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "firehose:PutRecord",
                "firehose:PutRecordBatch"
            ],
            "Resource": "${aws_kinesis_firehose_delivery_stream.s3_stream.arn}"
        }
    ]
}
EOF
}

resource "aws_s3_bucket" "bucket" {
  bucket = "metric-stream-test-bucket-%d"
  acl    = "private"
}

resource "aws_iam_role" "firehose_to_s3" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "firehose.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "firehose_to_s3" {
  name = "default"
  role = aws_iam_role.firehose_to_s3.id

  policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "s3:AbortMultipartUpload",
                "s3:GetBucketLocation",
                "s3:GetObject",
                "s3:ListBucket",
                "s3:ListBucketMultipartUploads",
                "s3:PutObject"
            ],      
            "Resource": [        
                "${aws_s3_bucket.bucket.arn}",
                "${aws_s3_bucket.bucket.arn}/*"		    
            ]
        }
    ]
}
EOF
}

resource "aws_kinesis_firehose_delivery_stream" "s3_stream" {
  name        = "metric-stream-test-stream-%d"
  destination = "s3"

  s3_configuration {
    role_arn   = aws_iam_role.firehose_to_s3.arn
    bucket_arn = aws_s3_bucket.bucket.arn
  }
}
`, rInt, rInt, rInt, rInt)
}

func testAccAWSCloudWatchMetricStreamConfigUpdateArn(rInt int) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}
data "aws_region" "current" {}
data "aws_caller_identity" "current" {}

resource "aws_cloudwatch_metric_stream" "test" {
  name          = "terraform-test-metric-stream-%d"
  role_arn      = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:role/MyOtherRole"
  firehose_arn  = "arn:${data.aws_partition.current.partition}:firehose:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:deliverystream/MyOtherFirehose"
  output_format = "json"
}
`, rInt)
}

func testAccAWSCloudWatchMetricStreamConfigIncludeFilters(rInt int) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}
data "aws_region" "current" {}
data "aws_caller_identity" "current" {}

resource "aws_cloudwatch_metric_stream" "test" {
  name          = "terraform-test-metric-stream-%d"
  role_arn      = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:role/MyRole"
  firehose_arn  = "arn:${data.aws_partition.current.partition}:firehose:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:deliverystream/MyFirehose"
  output_format = "json"

  include_filter {
    namespace = "AWS/EC2"
  }

  include_filter {
    namespace = "AWS/EBS"
  }
}
`, rInt)
}

func testAccAWSCloudWatchMetricStreamConfigNoName() string {
	return `
data "aws_partition" "current" {}
data "aws_region" "current" {}
data "aws_caller_identity" "current" {}

resource "aws_cloudwatch_metric_stream" "test" {
  role_arn      = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:role/MyRole"
  firehose_arn  = "arn:${data.aws_partition.current.partition}:firehose:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:deliverystream/MyFirehose"
  output_format = "json"
}
`
}

func testAccAWSCloudWatchMetricStreamConfigNamePrefix(rInt int) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}
data "aws_region" "current" {}
data "aws_caller_identity" "current" {}

resource "aws_cloudwatch_metric_stream" "test" {
  name_prefix   = "test-stream-%d"
  role_arn      = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:role/MyRole"
  firehose_arn  = "arn:${data.aws_partition.current.partition}:firehose:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:deliverystream/MyFirehose"
  output_format = "json"
}
`, rInt)
}

func testAccAWSCloudWatchMetricStreamConfigExcludeFilters(rInt int) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}
data "aws_region" "current" {}
data "aws_caller_identity" "current" {}

resource "aws_cloudwatch_metric_stream" "test" {
  name          = "terraform-test-metric-stream-%d"
  role_arn      = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:role/MyRole"
  firehose_arn  = "arn:${data.aws_partition.current.partition}:firehose:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:deliverystream/MyFirehose"
  output_format = "json"

  exclude_filter {
    namespace = "AWS/EC2"
  }

  exclude_filter {
    namespace = "AWS/EBS"
  }
}
`, rInt)
}
