// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsync "github.com/hashicorp/terraform-provider-aws/internal/experimental/sync"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccVerifiedAccessInstanceLoggingConfiguration_accessLogsIncludeTrustContext(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v types.VerifiedAccessInstanceLoggingConfiguration
	resourceName := "aws_verifiedaccess_instance_logging_configuration.test"
	instanceResourceName := "aws_verifiedaccess_instance.test"
	include_trust_context_original := true
	include_trust_context_updated := false

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckVerifiedAccessSynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckVerifiedAccessInstanceLoggingConfiguration(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVerifiedAccessInstanceLoggingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoggingConfigurationConfig_basic_accessLogsIncludeTrustContext(include_trust_context_original),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessInstanceLoggingConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.include_trust_context", strconv.FormatBool(include_trust_context_original)),
					resource.TestCheckResourceAttrPair(resourceName, "verifiedaccess_instance_id", instanceResourceName, names.AttrID),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			{
				Config: testAccLoggingConfigurationConfig_basic_accessLogsIncludeTrustContext(include_trust_context_updated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessInstanceLoggingConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.include_trust_context", strconv.FormatBool(include_trust_context_updated)),
					resource.TestCheckResourceAttrPair(resourceName, "verifiedaccess_instance_id", instanceResourceName, names.AttrID),
				),
			},
		},
	})
}

func testAccVerifiedAccessInstanceLoggingConfiguration_accessLogsLogVersion(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v types.VerifiedAccessInstanceLoggingConfiguration
	resourceName := "aws_verifiedaccess_instance_logging_configuration.test"
	instanceResourceName := "aws_verifiedaccess_instance.test"
	log_version_original := "ocsf-0.1"
	log_version_updated := "ocsf-1.0.0-rc.2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckVerifiedAccessSynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckVerifiedAccessInstanceLoggingConfiguration(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVerifiedAccessInstanceLoggingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoggingConfigurationConfig_basic_accessLogsLogVersion(log_version_original),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessInstanceLoggingConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.log_version", log_version_original),
					resource.TestCheckResourceAttrPair(resourceName, "verifiedaccess_instance_id", instanceResourceName, names.AttrID),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			{
				Config: testAccLoggingConfigurationConfig_basic_accessLogsLogVersion(log_version_updated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessInstanceLoggingConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.log_version", log_version_updated),
					resource.TestCheckResourceAttrPair(resourceName, "verifiedaccess_instance_id", instanceResourceName, names.AttrID),
				),
			},
		},
	})
}

func testAccVerifiedAccessInstanceLoggingConfiguration_accessLogsCloudWatchLogs(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v types.VerifiedAccessInstanceLoggingConfiguration
	resourceName := "aws_verifiedaccess_instance_logging_configuration.test"
	instanceResourceName := "aws_verifiedaccess_instance.test"
	logGroupName := "aws_cloudwatch_log_group.test"
	logGroupName2 := "aws_cloudwatch_log_group.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckVerifiedAccessSynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckVerifiedAccessInstanceLoggingConfiguration(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVerifiedAccessInstanceLoggingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoggingConfigurationConfig_basic_accessLogsCloudWatchLogs("first"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessInstanceLoggingConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.cloudwatch_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.cloudwatch_logs.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "access_logs.0.cloudwatch_logs.0.log_group", logGroupName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "verifiedaccess_instance_id", instanceResourceName, names.AttrID),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			{
				Config: testAccLoggingConfigurationConfig_basic_accessLogsCloudWatchLogs("second"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessInstanceLoggingConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.cloudwatch_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.cloudwatch_logs.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "access_logs.0.cloudwatch_logs.0.log_group", logGroupName2, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "verifiedaccess_instance_id", instanceResourceName, names.AttrID),
				),
			},
		},
	})
}

func testAccVerifiedAccessInstanceLoggingConfiguration_accessLogsKinesisDataFirehose(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v types.VerifiedAccessInstanceLoggingConfiguration
	resourceName := "aws_verifiedaccess_instance_logging_configuration.test"
	instanceResourceName := "aws_verifiedaccess_instance.test"
	kinesisStreamName := "aws_kinesis_firehose_delivery_stream.test"
	kinesisStreamName2 := "aws_kinesis_firehose_delivery_stream.test2"
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName3 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckVerifiedAccessSynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckVerifiedAccessInstanceLoggingConfiguration(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVerifiedAccessInstanceLoggingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoggingConfigurationConfig_basic_accessLogsKinesisDataFirehose(rName1, rName2, rName3, "first"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessInstanceLoggingConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.kinesis_data_firehose.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.kinesis_data_firehose.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "access_logs.0.kinesis_data_firehose.0.delivery_stream", kinesisStreamName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, "verifiedaccess_instance_id", instanceResourceName, names.AttrID),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			{
				Config: testAccLoggingConfigurationConfig_basic_accessLogsKinesisDataFirehose(rName1, rName2, rName3, "second"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessInstanceLoggingConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.kinesis_data_firehose.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.kinesis_data_firehose.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "access_logs.0.kinesis_data_firehose.0.delivery_stream", kinesisStreamName2, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, "verifiedaccess_instance_id", instanceResourceName, names.AttrID),
				),
			},
		},
	})
}

func testAccVerifiedAccessInstanceLoggingConfiguration_accessLogsS3(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v types.VerifiedAccessInstanceLoggingConfiguration
	resourceName := "aws_verifiedaccess_instance_logging_configuration.test"
	instanceResourceName := "aws_verifiedaccess_instance.test"
	bucketName := "aws_s3_bucket.test"
	bucketName2 := "aws_s3_bucket.test2"
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	prefix_original := "prefix-original"
	prefix_updated := "prefix-updated"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckVerifiedAccessSynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckVerifiedAccessInstanceLoggingConfiguration(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVerifiedAccessInstanceLoggingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoggingConfigurationConfig_basic_accessLogsS3(rName1, rName2, "first", prefix_original),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessInstanceLoggingConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.s3.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.s3.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "access_logs.0.s3.0.bucket_name", bucketName, names.AttrID),
					acctest.CheckResourceAttrAccountID(resourceName, "access_logs.0.s3.0.bucket_owner"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.s3.0.prefix", prefix_original),
					resource.TestCheckResourceAttrPair(resourceName, "verifiedaccess_instance_id", instanceResourceName, names.AttrID),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			{
				Config: testAccLoggingConfigurationConfig_basic_accessLogsS3(rName1, rName2, "second", prefix_updated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessInstanceLoggingConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.s3.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.s3.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "access_logs.0.s3.0.bucket_name", bucketName2, names.AttrID),
					acctest.CheckResourceAttrAccountID(resourceName, "access_logs.0.s3.0.bucket_owner"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.s3.0.prefix", prefix_updated),
					resource.TestCheckResourceAttrPair(resourceName, "verifiedaccess_instance_id", instanceResourceName, names.AttrID),
				),
			},
		},
	})
}

func testAccVerifiedAccessInstanceLoggingConfiguration_accessLogsCloudWatchLogsKinesisDataFirehoseS3(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v types.VerifiedAccessInstanceLoggingConfiguration
	resourceName := "aws_verifiedaccess_instance_logging_configuration.test"
	instanceResourceName := "aws_verifiedaccess_instance.test"
	logGroupName := "aws_cloudwatch_log_group.test"
	kinesisStreamName := "aws_kinesis_firehose_delivery_stream.test"
	bucketName := "aws_s3_bucket.test"
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName3 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckVerifiedAccessSynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckVerifiedAccessInstanceLoggingConfiguration(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVerifiedAccessInstanceLoggingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				// Test all 3 logging configurations together - CloudWatch, Kinesis Data Firehose, S3
				Config: testAccLoggingConfigurationConfig_basic_accessLogsCloudWatchLogsKinesisDataFirehoseS3(rName1, rName2, rName3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessInstanceLoggingConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.cloudwatch_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.cloudwatch_logs.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "access_logs.0.cloudwatch_logs.0.log_group", logGroupName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.kinesis_data_firehose.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.kinesis_data_firehose.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "access_logs.0.kinesis_data_firehose.0.delivery_stream", kinesisStreamName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.s3.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.s3.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "access_logs.0.s3.0.bucket_name", bucketName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "verifiedaccess_instance_id", instanceResourceName, names.AttrID),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			{
				// Test 2 logging configurations together - CloudWatch, Kinesis Data Firehose
				Config: testAccLoggingConfigurationConfig_basic_accessLogsCloudWatchLogsKinesisDataFirehose(rName1, rName2, rName3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessInstanceLoggingConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.cloudwatch_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.cloudwatch_logs.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "access_logs.0.cloudwatch_logs.0.log_group", logGroupName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.kinesis_data_firehose.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.kinesis_data_firehose.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "access_logs.0.kinesis_data_firehose.0.delivery_stream", kinesisStreamName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.s3.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.s3.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, "verifiedaccess_instance_id", instanceResourceName, names.AttrID),
				),
			},
			{
				// Test 2 logging configurations together - CloudWatch, S3
				Config: testAccLoggingConfigurationConfig_basic_accessLogsCloudWatchLogsS3(rName1, rName2, rName3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessInstanceLoggingConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.cloudwatch_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.cloudwatch_logs.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "access_logs.0.cloudwatch_logs.0.log_group", logGroupName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.kinesis_data_firehose.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.kinesis_data_firehose.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.s3.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.s3.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "access_logs.0.s3.0.bucket_name", bucketName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "verifiedaccess_instance_id", instanceResourceName, names.AttrID),
				),
			},
			{
				// Test 2 logging configurations together - Kinesis Data Firehose, S3
				Config: testAccLoggingConfigurationConfig_basic_accessLogsKinesisDataFirehoseS3(rName1, rName2, rName3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessInstanceLoggingConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.cloudwatch_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.cloudwatch_logs.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.kinesis_data_firehose.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.kinesis_data_firehose.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "access_logs.0.kinesis_data_firehose.0.delivery_stream", kinesisStreamName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.s3.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.s3.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "access_logs.0.s3.0.bucket_name", bucketName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "verifiedaccess_instance_id", instanceResourceName, names.AttrID),
				),
			},
		},
	})
}

func testAccVerifiedAccessInstanceLoggingConfiguration_disappears(t *testing.T, semaphore tfsync.Semaphore) {
	// note: disappears test does not test the logging configuration since the instance is deleted
	// the logging configuration cannot be deleted, rather, the boolean flags and logging version are reset to the default values
	ctx := acctest.Context(t)
	var v types.VerifiedAccessInstanceLoggingConfiguration
	resourceName := "aws_verifiedaccess_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckVerifiedAccessSynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckVerifiedAccessInstanceLoggingConfiguration(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVerifiedAccessInstanceLoggingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoggingConfigurationConfig_basic_accessLogsIncludeTrustContext(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessInstanceLoggingConfigurationExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceVerifiedAccessInstanceLoggingConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckVerifiedAccessInstanceLoggingConfigurationExists(ctx context.Context, n string, v *types.VerifiedAccessInstanceLoggingConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindVerifiedAccessInstanceLoggingConfigurationByInstanceID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckVerifiedAccessInstanceLoggingConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_verifiedaccess_instance_logging_configuration" {
				continue
			}

			_, err := tfec2.FindVerifiedAccessInstanceLoggingConfigurationByInstanceID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Verified Access Instance Logging Configuration %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccPreCheckVerifiedAccessInstanceLoggingConfiguration(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DescribeVerifiedAccessInstanceLoggingConfigurationsInput{}
	_, err := conn.DescribeVerifiedAccessInstanceLoggingConfigurations(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccVerifiedAccessInstanceLoggingConfigurationConfig_instance() string {
	return `
resource "aws_verifiedaccess_instance" "test" {}
`
}

func testAccVerifiedAccessInstanceLoggingConfigurationConfig_cloudwatch() string {
	return `
resource "aws_cloudwatch_log_group" "test" {}
`
}

func testAccVerifiedAccessInstanceLoggingConfigurationConfig_cloudwatchTwoLogGroups() string {
	return `
resource "aws_cloudwatch_log_group" "test" {}

resource "aws_cloudwatch_log_group" "test2" {}
`
}

func testAccInstanceStorageDeliveryStreamConfig_Base(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_iam_role" "firehose" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "firehose.amazonaws.com"
      },
      "Action": "sts:AssumeRole",
      "Condition": {
        "StringEquals": {
          "sts:ExternalId": "${data.aws_caller_identity.current.account_id}"
        }
      }
    }
  ]
}
EOF
}

resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
}

resource "aws_iam_role_policy" "firehose" {
  name = %[1]q
  role = aws_iam_role.firehose.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
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
    },
    {
      "Sid": "GlueAccess",
      "Effect": "Allow",
      "Action": [
        "glue:GetTable",
        "glue:GetTableVersion",
        "glue:GetTableVersions"
      ],
      "Resource": [
        "*"
      ]
    },
    {
      "Sid": "LakeFormationDataAccess",
      "Effect": "Allow",
      "Action": [
        "lakeformation:GetDataAccess"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}
`, rName)
}

func testAccVerifiedAccessInstanceLoggingConfigurationConfig_firehose(rName, rName2 string) string {
	return acctest.ConfigCompose(
		testAccInstanceStorageDeliveryStreamConfig_Base(rName),
		fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on  = [aws_iam_role_policy.firehose]
  name        = %[1]q
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn   = aws_iam_role.firehose.arn
    bucket_arn = aws_s3_bucket.bucket.arn
  }

  tags = {
    LogDeliveryEnabled = "true"
  }
}
`, rName2))
}

func testAccVerifiedAccessInstanceLoggingConfigurationConfig_firehoseTwoStreams(rName, rName2, rName3 string) string {
	return acctest.ConfigCompose(
		testAccInstanceStorageDeliveryStreamConfig_Base(rName),
		fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on  = [aws_iam_role_policy.firehose]
  name        = %[1]q
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn   = aws_iam_role.firehose.arn
    bucket_arn = aws_s3_bucket.bucket.arn
  }

  tags = {
    LogDeliveryEnabled = "true"
  }
}

resource "aws_kinesis_firehose_delivery_stream" "test2" {
  depends_on  = [aws_iam_role_policy.firehose]
  name        = %[2]q
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn   = aws_iam_role.firehose.arn
    bucket_arn = aws_s3_bucket.bucket.arn
  }

  tags = {
    LogDeliveryEnabled = "true"
  }
}
`, rName2, rName3))
}

func testAccVerifiedAccessInstanceLoggingConfigurationConfig_s3(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}
`, rName)
}

func testAccVerifiedAccessInstanceLoggingConfigurationConfig_s3TwoBuckets(rName, rName2 string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_bucket" "test2" {
  bucket        = %[2]q
  force_destroy = true
}
`, rName, rName2)
}

func testAccLoggingConfigurationConfig_basic_accessLogsIncludeTrustContext(includeTrustContext bool) string {
	return acctest.ConfigCompose(
		testAccVerifiedAccessInstanceLoggingConfigurationConfig_instance(),
		fmt.Sprintf(`
resource "aws_verifiedaccess_instance_logging_configuration" "test" {
  access_logs {
    include_trust_context = %[1]t
  }

  verifiedaccess_instance_id = aws_verifiedaccess_instance.test.id
}
`, includeTrustContext))
}

func testAccLoggingConfigurationConfig_basic_accessLogsLogVersion(logVersion string) string {
	return acctest.ConfigCompose(
		testAccVerifiedAccessInstanceLoggingConfigurationConfig_instance(),
		fmt.Sprintf(`
resource "aws_verifiedaccess_instance_logging_configuration" "test" {
  access_logs {
    log_version = %[1]q
  }

  verifiedaccess_instance_id = aws_verifiedaccess_instance.test.id
}
`, logVersion))
}

func testAccLoggingConfigurationConfig_basic_accessLogsCloudWatchLogs(selectLogGroup string) string {
	return acctest.ConfigCompose(
		testAccVerifiedAccessInstanceLoggingConfigurationConfig_instance(),
		testAccVerifiedAccessInstanceLoggingConfigurationConfig_cloudwatchTwoLogGroups(),
		fmt.Sprintf(`
locals {
  select_log_group = %[1]q
}

resource "aws_verifiedaccess_instance_logging_configuration" "test" {
  access_logs {
    cloudwatch_logs {
      enabled   = true
      log_group = local.select_log_group == "first" ? aws_cloudwatch_log_group.test.id : aws_cloudwatch_log_group.test2.id
    }
  }

  verifiedaccess_instance_id = aws_verifiedaccess_instance.test.id
}
`, selectLogGroup))
}

func testAccLoggingConfigurationConfig_basic_accessLogsKinesisDataFirehose(rName, rName2, rName3, selectStream string) string {
	return acctest.ConfigCompose(
		testAccVerifiedAccessInstanceLoggingConfigurationConfig_instance(),
		testAccVerifiedAccessInstanceLoggingConfigurationConfig_firehoseTwoStreams(rName, rName2, rName3),
		fmt.Sprintf(`
locals {
  select_stream = %[1]q
}

resource "aws_verifiedaccess_instance_logging_configuration" "test" {
  access_logs {
    kinesis_data_firehose {
      delivery_stream = local.select_stream == "first" ? aws_kinesis_firehose_delivery_stream.test.name : aws_kinesis_firehose_delivery_stream.test2.name
      enabled         = true
    }
  }

  verifiedaccess_instance_id = aws_verifiedaccess_instance.test.id
}
`, selectStream))
}

func testAccLoggingConfigurationConfig_basic_accessLogsS3(rName, rName2, selectBucket, prefix string) string {
	return acctest.ConfigCompose(
		testAccVerifiedAccessInstanceLoggingConfigurationConfig_instance(),
		testAccVerifiedAccessInstanceLoggingConfigurationConfig_s3TwoBuckets(rName, rName2),
		fmt.Sprintf(`
locals {
  select_bucket = %[1]q
}

data "aws_caller_identity" "test" {}

resource "aws_verifiedaccess_instance_logging_configuration" "test" {
  access_logs {
    s3 {
      enabled      = true
      bucket_name  = local.select_bucket == "first" ? aws_s3_bucket.test.id : aws_s3_bucket.test2.id
      bucket_owner = data.aws_caller_identity.test.account_id
      prefix       = %[2]q
    }
  }

  verifiedaccess_instance_id = aws_verifiedaccess_instance.test.id
}
`, selectBucket, prefix))
}

func testAccLoggingConfigurationConfig_basic_accessLogsCloudWatchLogsKinesisDataFirehoseS3(rName, rName2, rName3 string) string {
	return acctest.ConfigCompose(
		testAccVerifiedAccessInstanceLoggingConfigurationConfig_instance(),
		testAccVerifiedAccessInstanceLoggingConfigurationConfig_cloudwatch(),
		testAccVerifiedAccessInstanceLoggingConfigurationConfig_firehose(rName, rName2),
		testAccVerifiedAccessInstanceLoggingConfigurationConfig_s3(rName3),
		`
resource "aws_verifiedaccess_instance_logging_configuration" "test" {
  access_logs {
    cloudwatch_logs {
      enabled   = true
      log_group = aws_cloudwatch_log_group.test.id
    }

    kinesis_data_firehose {
      delivery_stream = aws_kinesis_firehose_delivery_stream.test.name
      enabled         = true
    }

    s3 {
      enabled     = true
      bucket_name = aws_s3_bucket.test.id
    }
  }

  verifiedaccess_instance_id = aws_verifiedaccess_instance.test.id
}
`)
}

func testAccLoggingConfigurationConfig_basic_accessLogsCloudWatchLogsKinesisDataFirehose(rName, rName2, rName3 string) string {
	return acctest.ConfigCompose(
		testAccVerifiedAccessInstanceLoggingConfigurationConfig_instance(),
		testAccVerifiedAccessInstanceLoggingConfigurationConfig_cloudwatch(),
		testAccVerifiedAccessInstanceLoggingConfigurationConfig_firehose(rName, rName2),
		testAccVerifiedAccessInstanceLoggingConfigurationConfig_s3(rName3),
		`
resource "aws_verifiedaccess_instance_logging_configuration" "test" {
  access_logs {
    cloudwatch_logs {
      enabled   = true
      log_group = aws_cloudwatch_log_group.test.id
    }

    kinesis_data_firehose {
      delivery_stream = aws_kinesis_firehose_delivery_stream.test.name
      enabled         = true
    }
  }

  verifiedaccess_instance_id = aws_verifiedaccess_instance.test.id
}
`)
}

func testAccLoggingConfigurationConfig_basic_accessLogsCloudWatchLogsS3(rName, rName2, rName3 string) string {
	return acctest.ConfigCompose(
		testAccVerifiedAccessInstanceLoggingConfigurationConfig_instance(),
		testAccVerifiedAccessInstanceLoggingConfigurationConfig_cloudwatch(),
		testAccVerifiedAccessInstanceLoggingConfigurationConfig_firehose(rName, rName2),
		testAccVerifiedAccessInstanceLoggingConfigurationConfig_s3(rName3),
		`
resource "aws_verifiedaccess_instance_logging_configuration" "test" {
  access_logs {
    cloudwatch_logs {
      enabled   = true
      log_group = aws_cloudwatch_log_group.test.id
    }

    s3 {
      enabled     = true
      bucket_name = aws_s3_bucket.test.id
    }
  }

  verifiedaccess_instance_id = aws_verifiedaccess_instance.test.id
}
`)
}

func testAccLoggingConfigurationConfig_basic_accessLogsKinesisDataFirehoseS3(rName, rName2, rName3 string) string {
	return acctest.ConfigCompose(
		testAccVerifiedAccessInstanceLoggingConfigurationConfig_instance(),
		testAccVerifiedAccessInstanceLoggingConfigurationConfig_cloudwatch(),
		testAccVerifiedAccessInstanceLoggingConfigurationConfig_firehose(rName, rName2),
		testAccVerifiedAccessInstanceLoggingConfigurationConfig_s3(rName3),
		`
resource "aws_verifiedaccess_instance_logging_configuration" "test" {
  access_logs {
    kinesis_data_firehose {
      delivery_stream = aws_kinesis_firehose_delivery_stream.test.name
      enabled         = true
    }

    s3 {
      enabled     = true
      bucket_name = aws_s3_bucket.test.id
    }
  }

  verifiedaccess_instance_id = aws_verifiedaccess_instance.test.id
}
`)
}
