package cloudwatch_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudwatch "github.com/hashicorp/terraform-provider-aws/internal/service/cloudwatch"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(cloudwatch.EndpointsID, testAccErrorCheckSkipCloudwatch)
}

func testAccErrorCheckSkipCloudwatch(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"context deadline exceeded", // tests never fail in GovCloud, they just timeout
	)
}

func TestAccCloudWatchMetricStream_basic(t *testing.T) {
	resourceName := "aws_cloudwatch_metric_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMetricStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMetricStreamConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricStreamExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "output_format", "json"),
					resource.TestCheckResourceAttr(resourceName, "state", tfcloudwatch.StateRunning),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.metric_stream_to_firehose", "arn"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "cloudwatch", fmt.Sprintf("metric-stream/%s", rName)),
					acctest.CheckResourceAttrRFC3339(resourceName, "creation_date"),
					acctest.CheckResourceAttrRFC3339(resourceName, "last_update_date"),
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

func TestAccCloudWatchMetricStream_noName(t *testing.T) {
	resourceName := "aws_cloudwatch_metric_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMetricStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMetricStreamNoNameConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricStreamExists(resourceName),
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

func TestAccCloudWatchMetricStream_namePrefix(t *testing.T) {
	resourceName := "aws_cloudwatch_metric_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudwatch.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMetricStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMetricStreamNamePrefixConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricStreamExists(resourceName),
					testAccCheckMetricStreamGeneratedNamePrefix(resourceName, acctest.ResourcePrefix),
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

func TestAccCloudWatchMetricStream_includeFilters(t *testing.T) {
	resourceName := "aws_cloudwatch_metric_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMetricStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMetricStreamIncludeFiltersConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricStreamExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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

func TestAccCloudWatchMetricStream_excludeFilters(t *testing.T) {
	resourceName := "aws_cloudwatch_metric_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMetricStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMetricStreamExcludeFiltersConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricStreamExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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

func TestAccCloudWatchMetricStream_update(t *testing.T) {
	resourceName := "aws_cloudwatch_metric_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMetricStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMetricStreamUpdateARNConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricStreamExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "output_format", "json"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMetricStreamConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricStreamExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "output_format", "json"),
				),
			},
		},
	})
}

func TestAccCloudWatchMetricStream_updateName(t *testing.T) {
	resourceName := "aws_cloudwatch_metric_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMetricStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMetricStreamConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricStreamExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				Config: testAccMetricStreamConfig(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricStreamExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					testAccCheckMetricStreamDestroyPrevious(rName),
				),
			},
		},
	})
}

func TestAccCloudWatchMetricStream_tags(t *testing.T) {
	resourceName := "aws_cloudwatch_metric_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMetricStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMetricStreamTagsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricStreamExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
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

func testAccCheckMetricStreamGeneratedNamePrefix(resource, prefix string) resource.TestCheckFunc {
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

func testAccCheckMetricStreamExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudWatchConn
		params := cloudwatch.GetMetricStreamInput{
			Name: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetMetricStream(&params)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckMetricStreamDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CloudWatchConn

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
		if !tfawserr.ErrCodeEquals(err, cloudwatch.ErrCodeResourceNotFoundException) {
			return err
		}
	}

	return nil
}

func testAccCheckMetricStreamDestroyPrevious(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudWatchConn

		params := cloudwatch.GetMetricStreamInput{
			Name: aws.String(name),
		}

		_, err := conn.GetMetricStream(&params)

		if err == nil {
			return fmt.Errorf("MetricStream still exists: %s", name)
		}

		if !tfawserr.ErrCodeEquals(err, cloudwatch.ErrCodeResourceNotFoundException) {
			return err
		}

		return nil
	}
}

func testAccMetricStreamConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_cloudwatch_metric_stream" "test" {
  name          = %[1]q
  role_arn      = aws_iam_role.metric_stream_to_firehose.arn
  firehose_arn  = aws_kinesis_firehose_delivery_stream.s3_stream.arn
  output_format = "json"
}

resource "aws_iam_role" "metric_stream_to_firehose" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "streams.metrics.cloudwatch.${data.aws_partition.current.dns_suffix}"
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
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "bucket_acl" {
  bucket = aws_s3_bucket.bucket.id
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
        "Service": "firehose.${data.aws_partition.current.dns_suffix}"
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
  name        = %[1]q
  destination = "s3"

  s3_configuration {
    role_arn   = aws_iam_role.firehose_to_s3.arn
    bucket_arn = aws_s3_bucket.bucket.arn
  }
}
`, rName)
}

func testAccMetricStreamUpdateARNConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}
data "aws_region" "current" {}
data "aws_caller_identity" "current" {}

resource "aws_cloudwatch_metric_stream" "test" {
  name          = %[1]q
  role_arn      = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:role/MyOtherRole"
  firehose_arn  = "arn:${data.aws_partition.current.partition}:firehose:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:deliverystream/MyOtherFirehose"
  output_format = "json"
}
`, rName)
}

func testAccMetricStreamIncludeFiltersConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}
data "aws_region" "current" {}
data "aws_caller_identity" "current" {}

resource "aws_cloudwatch_metric_stream" "test" {
  name          = %[1]q
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
`, rName)
}

func testAccMetricStreamNoNameConfig() string {
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

func testAccMetricStreamNamePrefixConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}
data "aws_region" "current" {}
data "aws_caller_identity" "current" {}

resource "aws_cloudwatch_metric_stream" "test" {
  name_prefix   = %[1]q
  role_arn      = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:role/MyRole"
  firehose_arn  = "arn:${data.aws_partition.current.partition}:firehose:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:deliverystream/MyFirehose"
  output_format = "json"
}
`, rName)
}

func testAccMetricStreamExcludeFiltersConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}
data "aws_region" "current" {}
data "aws_caller_identity" "current" {}

resource "aws_cloudwatch_metric_stream" "test" {
  name          = %[1]q
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
`, rName)
}

func testAccMetricStreamTagsConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}
data "aws_region" "current" {}
data "aws_caller_identity" "current" {}

resource "aws_cloudwatch_metric_stream" "test" {
  name          = %[1]q
  role_arn      = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:role/MyRole"
  firehose_arn  = "arn:${data.aws_partition.current.partition}:firehose:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:deliverystream/MyFirehose"
  output_format = "json"

  tags = {
    Name     = %[1]q
    Mercedes = "Toto"
  }
}
`, rName)
}
