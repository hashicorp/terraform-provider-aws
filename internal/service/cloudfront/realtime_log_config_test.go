// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudfront "github.com/hashicorp/terraform-provider-aws/internal/service/cloudfront"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudFrontRealtimeLogConfig_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.RealtimeLogConfig
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	samplingRate := sdkacctest.RandIntRange(1, 100)
	resourceName := "aws_cloudfront_realtime_log_config.test"
	roleResourceName := "aws_iam_role.test.0"
	streamResourceName := "aws_kinesis_stream.test.0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRealtimeLogConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRealtimeLogConfigConfig_basic(rName, samplingRate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRealtimeLogConfigExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrGlobalARN(resourceName, names.AttrARN, "cloudfront", fmt.Sprintf("realtime-log-config/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "endpoint.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "endpoint.0.stream_type", "Kinesis"),
					resource.TestCheckResourceAttr(resourceName, "endpoint.0.kinesis_stream_config.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint.0.kinesis_stream_config.0.role_arn", roleResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint.0.kinesis_stream_config.0.stream_arn", streamResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "fields.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "fields.*", "timestamp"),
					resource.TestCheckTypeSetElemAttr(resourceName, "fields.*", "c-ip"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "sampling_rate", strconv.Itoa(samplingRate)),
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

func TestAccCloudFrontRealtimeLogConfig_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.RealtimeLogConfig
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	samplingRate := sdkacctest.RandIntRange(1, 100)
	resourceName := "aws_cloudfront_realtime_log_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRealtimeLogConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRealtimeLogConfigConfig_basic(rName, samplingRate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRealtimeLogConfigExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcloudfront.ResourceRealtimeLogConfig(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCloudFrontRealtimeLogConfig_updates(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.RealtimeLogConfig
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	samplingRate1 := sdkacctest.RandIntRange(1, 100)
	samplingRate2 := sdkacctest.RandIntRange(1, 100)
	resourceName := "aws_cloudfront_realtime_log_config.test"
	role1ResourceName := "aws_iam_role.test.0"
	stream1ResourceName := "aws_kinesis_stream.test.0"
	role2ResourceName := "aws_iam_role.test.1"
	stream2ResourceName := "aws_kinesis_stream.test.1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRealtimeLogConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRealtimeLogConfigConfig_basic(rName, samplingRate1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRealtimeLogConfigExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrGlobalARN(resourceName, names.AttrARN, "cloudfront", fmt.Sprintf("realtime-log-config/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "endpoint.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "endpoint.0.stream_type", "Kinesis"),
					resource.TestCheckResourceAttr(resourceName, "endpoint.0.kinesis_stream_config.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint.0.kinesis_stream_config.0.role_arn", role1ResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint.0.kinesis_stream_config.0.stream_arn", stream1ResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "fields.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "fields.*", "timestamp"),
					resource.TestCheckTypeSetElemAttr(resourceName, "fields.*", "c-ip"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "sampling_rate", strconv.Itoa(samplingRate1)),
				),
			},
			{
				Config: testAccRealtimeLogConfigConfig_updated(rName, samplingRate2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRealtimeLogConfigExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrGlobalARN(resourceName, names.AttrARN, "cloudfront", fmt.Sprintf("realtime-log-config/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "endpoint.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "endpoint.0.stream_type", "Kinesis"),
					resource.TestCheckResourceAttr(resourceName, "endpoint.0.kinesis_stream_config.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint.0.kinesis_stream_config.0.role_arn", role2ResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint.0.kinesis_stream_config.0.stream_arn", stream2ResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "fields.#", acctest.Ct3),
					resource.TestCheckTypeSetElemAttr(resourceName, "fields.*", "c-ip"),
					resource.TestCheckTypeSetElemAttr(resourceName, "fields.*", "cs-host"),
					resource.TestCheckTypeSetElemAttr(resourceName, "fields.*", "sc-status"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "sampling_rate", strconv.Itoa(samplingRate2)),
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

func testAccCheckRealtimeLogConfigDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudfront_realtime_log_config" {
				continue
			}

			_, err := tfcloudfront.FindRealtimeLogConfigByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudFront Real-time Log Config %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckRealtimeLogConfigExists(ctx context.Context, n string, v *awstypes.RealtimeLogConfig) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontClient(ctx)

		output, err := tfcloudfront.FindRealtimeLogConfigByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccRealtimeLogBaseConfig(rName string, count int) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  count = %[2]d

  name        = format("%%s-%%d", %[1]q, count.index)
  shard_count = 2
}

resource "aws_iam_role" "test" {
  count = %[2]d

  name = format("%%s-%%d", %[1]q, count.index)

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Principal": {
      "Service": "cloudfront.amazonaws.com"
    },
    "Effect": "Allow"
  }]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  count = %[2]d

  name = format("%%s-%%d", %[1]q, count.index)
  role = aws_iam_role.test[count.index].id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": [
      "kinesis:DescribeStreamSummary",
      "kinesis:DescribeStream",
      "kinesis:PutRecord",
      "kinesis:PutRecords"
    ],
    "Resource": "${aws_kinesis_stream.test[count.index].arn}"
  }]
}
EOF
}
`, rName, count)
}

func testAccRealtimeLogConfigConfig_basic(rName string, samplingRate int) string {
	return acctest.ConfigCompose(
		testAccRealtimeLogBaseConfig(rName, 1),
		fmt.Sprintf(`
resource "aws_cloudfront_realtime_log_config" "test" {
  name          = %[1]q
  sampling_rate = %[2]d
  fields        = ["timestamp", "c-ip"]

  endpoint {
    stream_type = "Kinesis"

    kinesis_stream_config {
      role_arn   = aws_iam_role.test[0].arn
      stream_arn = aws_kinesis_stream.test[0].arn
    }
  }

  depends_on = [aws_iam_role_policy.test[0]]
}
`, rName, samplingRate))
}

func testAccRealtimeLogConfigConfig_updated(rName string, samplingRate int) string {
	return acctest.ConfigCompose(
		testAccRealtimeLogBaseConfig(rName, 2),
		fmt.Sprintf(`
resource "aws_cloudfront_realtime_log_config" "test" {
  name          = %[1]q
  sampling_rate = %[2]d
  fields        = ["c-ip", "cs-host", "sc-status"]

  endpoint {
    stream_type = "Kinesis"

    kinesis_stream_config {
      role_arn   = aws_iam_role.test[1].arn
      stream_arn = aws_kinesis_stream.test[1].arn
    }
  }

  depends_on = [aws_iam_role_policy.test[1]]
}
`, rName, samplingRate))
}
