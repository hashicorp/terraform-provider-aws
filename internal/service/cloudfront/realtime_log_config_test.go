package cloudfront_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudfront"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudfront "github.com/hashicorp/terraform-provider-aws/internal/service/cloudfront"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccCloudFrontRealtimeLogConfig_basic(t *testing.T) {
	var v cloudfront.RealtimeLogConfig
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	samplingRate := sdkacctest.RandIntRange(1, 100)
	resourceName := "aws_cloudfront_realtime_log_config.test"
	roleResourceName := "aws_iam_role.test.0"
	streamResourceName := "aws_kinesis_stream.test.0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRealtimeLogConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRealtimeLogConfig(rName, samplingRate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRealtimeLogConfigExists(resourceName, &v),
					acctest.CheckResourceAttrGlobalARN(resourceName, "arn", "cloudfront", fmt.Sprintf("realtime-log-config/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "endpoint.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint.0.stream_type", "Kinesis"),
					resource.TestCheckResourceAttr(resourceName, "endpoint.0.kinesis_stream_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint.0.kinesis_stream_config.0.role_arn", roleResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint.0.kinesis_stream_config.0.stream_arn", streamResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "fields.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "fields.*", "timestamp"),
					resource.TestCheckTypeSetElemAttr(resourceName, "fields.*", "c-ip"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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
	var v cloudfront.RealtimeLogConfig
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	samplingRate := sdkacctest.RandIntRange(1, 100)
	resourceName := "aws_cloudfront_realtime_log_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRealtimeLogConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRealtimeLogConfig(rName, samplingRate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRealtimeLogConfigExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfcloudfront.ResourceRealtimeLogConfig(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCloudFrontRealtimeLogConfig_updates(t *testing.T) {
	var v cloudfront.RealtimeLogConfig
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	samplingRate1 := sdkacctest.RandIntRange(1, 100)
	samplingRate2 := sdkacctest.RandIntRange(1, 100)
	resourceName := "aws_cloudfront_realtime_log_config.test"
	role1ResourceName := "aws_iam_role.test.0"
	stream1ResourceName := "aws_kinesis_stream.test.0"
	role2ResourceName := "aws_iam_role.test.1"
	stream2ResourceName := "aws_kinesis_stream.test.1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRealtimeLogConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRealtimeLogConfig(rName, samplingRate1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRealtimeLogConfigExists(resourceName, &v),
					acctest.CheckResourceAttrGlobalARN(resourceName, "arn", "cloudfront", fmt.Sprintf("realtime-log-config/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "endpoint.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint.0.stream_type", "Kinesis"),
					resource.TestCheckResourceAttr(resourceName, "endpoint.0.kinesis_stream_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint.0.kinesis_stream_config.0.role_arn", role1ResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint.0.kinesis_stream_config.0.stream_arn", stream1ResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "fields.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "fields.*", "timestamp"),
					resource.TestCheckTypeSetElemAttr(resourceName, "fields.*", "c-ip"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "sampling_rate", strconv.Itoa(samplingRate1)),
				),
			},
			{
				Config: testAccRealtimeLogUpdatedConfig(rName, samplingRate2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRealtimeLogConfigExists(resourceName, &v),
					acctest.CheckResourceAttrGlobalARN(resourceName, "arn", "cloudfront", fmt.Sprintf("realtime-log-config/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "endpoint.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint.0.stream_type", "Kinesis"),
					resource.TestCheckResourceAttr(resourceName, "endpoint.0.kinesis_stream_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint.0.kinesis_stream_config.0.role_arn", role2ResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint.0.kinesis_stream_config.0.stream_arn", stream2ResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "fields.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "fields.*", "c-ip"),
					resource.TestCheckTypeSetElemAttr(resourceName, "fields.*", "cs-host"),
					resource.TestCheckTypeSetElemAttr(resourceName, "fields.*", "sc-status"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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

func testAccCheckRealtimeLogConfigDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudfront_realtime_log_config" {
			continue
		}

		_, err := tfcloudfront.FindRealtimeLogConfigByARN(conn, rs.Primary.ID)

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

func testAccCheckRealtimeLogConfigExists(n string, v *cloudfront.RealtimeLogConfig) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No CloudFront Real-time Log Config ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontConn

		output, err := tfcloudfront.FindRealtimeLogConfigByARN(conn, rs.Primary.ID)

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

func testAccRealtimeLogConfig(rName string, samplingRate int) string {
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

func testAccRealtimeLogUpdatedConfig(rName string, samplingRate int) string {
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
