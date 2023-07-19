// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudwatch_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudwatch "github.com/hashicorp/terraform-provider-aws/internal/service/cloudwatch"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(cloudwatch.EndpointsID, testAccErrorCheckSkip)
}

func testAccErrorCheckSkip(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"context deadline exceeded", // tests never fail in GovCloud, they just timeout
	)
}

func TestAccCloudWatchMetricStream_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cloudwatch_metric_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMetricStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMetricStreamConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMetricStreamExists(ctx, resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "cloudwatch", fmt.Sprintf("metric-stream/%s", rName)),
					acctest.CheckResourceAttrRFC3339(resourceName, "creation_date"),
					resource.TestCheckResourceAttr(resourceName, "exclude_filter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "include_filter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "include_linked_accounts_metrics", "false"),
					acctest.CheckResourceAttrRFC3339(resourceName, "last_update_date"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "output_format", "json"),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.metric_stream_to_firehose", "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					resource.TestCheckResourceAttr(resourceName, "statistics_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccCloudWatchMetricStream_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cloudwatch_metric_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMetricStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMetricStreamConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMetricStreamExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcloudwatch.ResourceMetricStream(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCloudWatchMetricStream_nameGenerated(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cloudwatch_metric_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMetricStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMetricStreamConfig_nameGenerated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricStreamExists(ctx, resourceName),
					acctest.CheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", id.UniqueIdPrefix),
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
	ctx := acctest.Context(t)
	resourceName := "aws_cloudwatch_metric_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, cloudwatch.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMetricStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMetricStreamConfig_namePrefix(rName, "tf-acc-test-prefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricStreamExists(ctx, resourceName),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, "name", "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "tf-acc-test-prefix-"),
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
	ctx := acctest.Context(t)
	resourceName := "aws_cloudwatch_metric_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMetricStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMetricStreamConfig_includeFilters(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricStreamExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "output_format", "json"),
					resource.TestCheckResourceAttr(resourceName, "include_filter.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "include_filter.0.metric_names.#", "0"),
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

func TestAccCloudWatchMetricStream_includeFiltersWithMetricNames(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cloudwatch_metric_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMetricStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMetricStreamConfig_includeFiltersWithMetricNames(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricStreamExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "output_format", "json"),
					resource.TestCheckResourceAttr(resourceName, "include_filter.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "include_filter.0.metric_names.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "include_filter.0.metric_names.0", "CPUUtilization"),
					resource.TestCheckResourceAttr(resourceName, "include_filter.1.metric_names.#", "0"),
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
	ctx := acctest.Context(t)
	resourceName := "aws_cloudwatch_metric_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMetricStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMetricStreamConfig_excludeFilters(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricStreamExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "output_format", "json"),
					resource.TestCheckResourceAttr(resourceName, "exclude_filter.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "exclude_filter.0.metric_names.#", "0")),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCloudWatchMetricStream_excludeFiltersWithMetricNames(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cloudwatch_metric_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMetricStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMetricStreamConfig_excludeFiltersWithMetricNames(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricStreamExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "output_format", "json"),
					resource.TestCheckResourceAttr(resourceName, "exclude_filter.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "exclude_filter.0.metric_names.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "exclude_filter.0.metric_names.0", "CPUUtilization"),
					resource.TestCheckResourceAttr(resourceName, "exclude_filter.1.metric_names.#", "0"),
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
	ctx := acctest.Context(t)
	resourceName := "aws_cloudwatch_metric_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMetricStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMetricStreamConfig_arns(rName, "S1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricStreamExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "firehose_arn", "firehose", regexp.MustCompile(`deliverystream/S1$`)),
					acctest.MatchResourceAttrGlobalARN(resourceName, "role_arn", "iam", regexp.MustCompile(`role/S1$`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMetricStreamConfig_arns(rName, "S2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricStreamExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "firehose_arn", "firehose", regexp.MustCompile(`deliverystream/S2$`)),
					acctest.MatchResourceAttrGlobalARN(resourceName, "role_arn", "iam", regexp.MustCompile(`role/S2$`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccMetricStreamConfig_arnsWithTag(rName, "S3", "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricStreamExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "firehose_arn", "firehose", regexp.MustCompile(`deliverystream/S3$`)),
					acctest.MatchResourceAttrGlobalARN(resourceName, "role_arn", "iam", regexp.MustCompile(`role/S3$`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccMetricStreamConfig_arnsWithTag(rName, "S4", "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricStreamExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "firehose_arn", "firehose", regexp.MustCompile(`deliverystream/S4$`)),
					acctest.MatchResourceAttrGlobalARN(resourceName, "role_arn", "iam", regexp.MustCompile(`role/S4$`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccMetricStreamConfig_arnsWithTag(rName, "S4", "key1", "value1updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricStreamExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "firehose_arn", "firehose", regexp.MustCompile(`deliverystream/S4$`)),
					acctest.MatchResourceAttrGlobalARN(resourceName, "role_arn", "iam", regexp.MustCompile(`role/S4$`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
				),
			},
		},
	})
}

func TestAccCloudWatchMetricStream_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cloudwatch_metric_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMetricStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMetricStreamConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricStreamExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMetricStreamConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricStreamExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccMetricStreamConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricStreamExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccCloudWatchMetricStream_additional_statistics(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cloudwatch_metric_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMetricStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccMetricStreamConfig_additionalStatistics(rName, "p0"),
				ExpectError: regexp.MustCompile(`invalid statistic, see: https:\/\/docs\.aws\.amazon\.com\/.*`),
			},
			{
				Config:      testAccMetricStreamConfig_additionalStatistics(rName, "p100"),
				ExpectError: regexp.MustCompile(`invalid statistic, see: https:\/\/docs\.aws\.amazon\.com\/.*`),
			},
			{
				Config:      testAccMetricStreamConfig_additionalStatistics(rName, "p"),
				ExpectError: regexp.MustCompile(`invalid statistic, see: https:\/\/docs\.aws\.amazon\.com\/.*`),
			},
			{
				Config:      testAccMetricStreamConfig_additionalStatistics(rName, "tm"),
				ExpectError: regexp.MustCompile(`invalid statistic, see: https:\/\/docs\.aws\.amazon\.com\/.*`),
			},
			{
				Config:      testAccMetricStreamConfig_additionalStatistics(rName, "tc()"),
				ExpectError: regexp.MustCompile(`invalid statistic, see: https:\/\/docs\.aws\.amazon\.com\/.*`),
			},
			{
				Config:      testAccMetricStreamConfig_additionalStatistics(rName, "p99.12345678901"),
				ExpectError: regexp.MustCompile(`invalid statistic, see: https:\/\/docs\.aws\.amazon\.com\/.*`),
			},
			{
				Config: testAccMetricStreamConfig_additionalStatistics(rName, "IQM"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricStreamExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "statistics_configuration.#", "2"),
				),
			},
			{
				Config: testAccMetricStreamConfig_additionalStatistics(rName, "PR(:50)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricStreamExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "statistics_configuration.#", "2"),
				),
			},
			{
				Config: testAccMetricStreamConfig_additionalStatistics(rName, "TS(50.5:)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricStreamExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "statistics_configuration.#", "2"),
				),
			},
			{
				Config: testAccMetricStreamConfig_additionalStatistics(rName, "TC(1:100)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricStreamExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "statistics_configuration.#", "2"),
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

func TestAccCloudWatchMetricStream_includeLinkedAccountsMetrics(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cloudwatch_metric_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMetricStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMetricStreamConfig_includeLinkedAccountsMetrics(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricStreamExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "include_linked_accounts_metrics", "false"),
				),
			},
			{
				Config: testAccMetricStreamConfig_includeLinkedAccountsMetrics(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricStreamExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "include_linked_accounts_metrics", "true"),
				),
			},
		},
	})
}

func testAccCheckMetricStreamExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No CloudWatch Metric Stream ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudWatchConn(ctx)

		_, err := tfcloudwatch.FindMetricStreamByName(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckMetricStreamDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudWatchConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudwatch_metric_stream" {
				continue
			}

			_, err := tfcloudwatch.FindMetricStreamByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudWatch Metric Stream %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccMetricStreamConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

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
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn   = aws_iam_role.firehose_to_s3.arn
    bucket_arn = aws_s3_bucket.bucket.arn
  }
}
`, rName)
}

func testAccMetricStreamConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccMetricStreamConfig_base(rName), fmt.Sprintf(`
resource "aws_cloudwatch_metric_stream" "test" {
  name          = %[1]q
  role_arn      = aws_iam_role.metric_stream_to_firehose.arn
  firehose_arn  = aws_kinesis_firehose_delivery_stream.s3_stream.arn
  output_format = "json"
}
`, rName))
}

func testAccMetricStreamConfig_nameGenerated(rName string) string {
	return acctest.ConfigCompose(testAccMetricStreamConfig_base(rName), `
resource "aws_cloudwatch_metric_stream" "test" {
  role_arn      = aws_iam_role.metric_stream_to_firehose.arn
  firehose_arn  = aws_kinesis_firehose_delivery_stream.s3_stream.arn
  output_format = "json"
}
`)
}

func testAccMetricStreamConfig_namePrefix(rName, namePrefix string) string {
	return acctest.ConfigCompose(testAccMetricStreamConfig_base(rName), fmt.Sprintf(`
resource "aws_cloudwatch_metric_stream" "test" {
  name_prefix   = %[1]q
  role_arn      = aws_iam_role.metric_stream_to_firehose.arn
  firehose_arn  = aws_kinesis_firehose_delivery_stream.s3_stream.arn
  output_format = "json"
}
`, namePrefix))
}

func testAccMetricStreamConfig_arns(rName, arnSuffix string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}
data "aws_region" "current" {}
data "aws_caller_identity" "current" {}

resource "aws_cloudwatch_metric_stream" "test" {
  name          = %[1]q
  role_arn      = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:role/%[2]s"
  firehose_arn  = "arn:${data.aws_partition.current.partition}:firehose:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:deliverystream/%[2]s"
  output_format = "json"
}
`, rName, arnSuffix)
}

func testAccMetricStreamConfig_arnsWithTag(rName, arnSuffix, tagKey, tagValue string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}
data "aws_region" "current" {}
data "aws_caller_identity" "current" {}

resource "aws_cloudwatch_metric_stream" "test" {
  name          = %[1]q
  role_arn      = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:role/%[2]s"
  firehose_arn  = "arn:${data.aws_partition.current.partition}:firehose:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:deliverystream/%[2]s"
  output_format = "json"

  tags = {
    %[3]q = %[4]q
  }
}
`, rName, arnSuffix, tagKey, tagValue)
}

func testAccMetricStreamConfig_includeFilters(rName string) string {
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

func testAccMetricStreamConfig_includeFiltersWithMetricNames(rName string) string {
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
    namespace    = "AWS/EC2"
    metric_names = ["CPUUtilization", "NetworkOut"]
  }

  include_filter {
    namespace    = "AWS/EBS"
    metric_names = []
  }
}
`, rName)
}

func testAccMetricStreamConfig_excludeFilters(rName string) string {
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

func testAccMetricStreamConfig_excludeFiltersWithMetricNames(rName string) string {
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
    namespace    = "AWS/EC2"
    metric_names = ["CPUUtilization", "NetworkOut"]
  }

  exclude_filter {
    namespace    = "AWS/EBS"
    metric_names = []
  }
}
`, rName)
}

func testAccMetricStreamConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccMetricStreamConfig_base(rName), fmt.Sprintf(`
resource "aws_cloudwatch_metric_stream" "test" {
  name          = %[1]q
  role_arn      = aws_iam_role.metric_stream_to_firehose.arn
  firehose_arn  = aws_kinesis_firehose_delivery_stream.s3_stream.arn
  output_format = "json"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccMetricStreamConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccMetricStreamConfig_base(rName), fmt.Sprintf(`
resource "aws_cloudwatch_metric_stream" "test" {
  name          = %[1]q
  role_arn      = aws_iam_role.metric_stream_to_firehose.arn
  firehose_arn  = aws_kinesis_firehose_delivery_stream.s3_stream.arn
  output_format = "json"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccMetricStreamConfig_additionalStatistics(rName string, stat string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}
data "aws_region" "current" {}
data "aws_caller_identity" "current" {}

resource "aws_cloudwatch_metric_stream" "test" {
  name          = %[1]q
  role_arn      = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:role/MyRole"
  firehose_arn  = "arn:${data.aws_partition.current.partition}:firehose:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:deliverystream/MyFirehose"
  output_format = "json"

  statistics_configuration {
    additional_statistics = [
      "p1", "tm99"
    ]

    include_metric {
      metric_name = "CPUUtilization"
      namespace   = "AWS/EC2"
    }
  }

  statistics_configuration {
    additional_statistics = [
	  %[2]q
    ]

    include_metric {
      metric_name = "CPUUtilization"
      namespace   = "AWS/EC2"
    }
  }
}
`, rName, stat)
}

func testAccMetricStreamConfig_includeLinkedAccountsMetrics(rName string, include bool) string {
	return acctest.ConfigCompose(testAccMetricStreamConfig_base(rName), fmt.Sprintf(`
resource "aws_cloudwatch_metric_stream" "test" {
  name                            = %[1]q
  role_arn                        = aws_iam_role.metric_stream_to_firehose.arn
  firehose_arn                    = aws_kinesis_firehose_delivery_stream.s3_stream.arn
  output_format                   = "json"
  include_linked_accounts_metrics = %[2]t
}
`, rName, include))
}
