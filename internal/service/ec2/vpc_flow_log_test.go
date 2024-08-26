// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCFlowLog_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var flowLog awstypes.FlowLog
	cloudwatchLogGroupResourceName := "aws_cloudwatch_log_group.test"
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_flow_log.test"
	vpcResourceName := "aws_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFlowLogDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCFlowLogConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFlowLogExists(ctx, resourceName, &flowLog),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`vpc-flow-log/fl-.+`)),
					resource.TestCheckResourceAttr(resourceName, "deliver_cross_account_role", ""),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrIAMRoleARN, iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "log_destination", ""),
					resource.TestCheckResourceAttr(resourceName, "log_destination_type", "cloud-watch-logs"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrLogGroupName, cloudwatchLogGroupResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "max_aggregation_interval", "600"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "traffic_type", "ALL"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrVPCID, vpcResourceName, names.AttrID),
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

func TestAccVPCFlowLog_logFormat(t *testing.T) {
	ctx := acctest.Context(t)
	var flowLog awstypes.FlowLog
	resourceName := "aws_flow_log.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	logFormat := "${version} ${vpc-id} ${subnet-id}"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFlowLogDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCFlowLogConfig_format(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowLogExists(ctx, resourceName, &flowLog),
					resource.TestCheckResourceAttr(resourceName, "log_format", logFormat),
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

func TestAccVPCFlowLog_subnetID(t *testing.T) {
	ctx := acctest.Context(t)
	var flowLog awstypes.FlowLog
	cloudwatchLogGroupResourceName := "aws_cloudwatch_log_group.test"
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_flow_log.test"
	subnetResourceName := "aws_subnet.test.0"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFlowLogDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCFlowLogConfig_subnetID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowLogExists(ctx, resourceName, &flowLog),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrIAMRoleARN, iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "log_destination", ""),
					resource.TestCheckResourceAttr(resourceName, "log_destination_type", "cloud-watch-logs"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrLogGroupName, cloudwatchLogGroupResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "max_aggregation_interval", "600"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrSubnetID, subnetResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "traffic_type", "ALL"),
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

func TestAccVPCFlowLog_transitGatewayID(t *testing.T) {
	ctx := acctest.Context(t)
	var flowLog awstypes.FlowLog
	cloudwatchLogGroupResourceName := "aws_cloudwatch_log_group.test"
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_flow_log.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFlowLogDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCFlowLogConfig_transitGatewayID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowLogExists(ctx, resourceName, &flowLog),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`vpc-flow-log/fl-.+`)),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrIAMRoleARN, iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "log_destination", ""),
					resource.TestCheckResourceAttr(resourceName, "log_destination_type", "cloud-watch-logs"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrLogGroupName, cloudwatchLogGroupResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "max_aggregation_interval", "60"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTransitGatewayID, transitGatewayResourceName, names.AttrID),
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

func TestAccVPCFlowLog_transitGatewayAttachmentID(t *testing.T) {
	ctx := acctest.Context(t)
	var flowLog awstypes.FlowLog
	cloudwatchLogGroupResourceName := "aws_cloudwatch_log_group.test"
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_flow_log.test"
	transitGatewayAttachmentResourceName := "aws_ec2_transit_gateway_vpc_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFlowLogDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCFlowLogConfig_transitGatewayAttachmentID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowLogExists(ctx, resourceName, &flowLog),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`vpc-flow-log/fl-.+`)),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrIAMRoleARN, iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "log_destination", ""),
					resource.TestCheckResourceAttr(resourceName, "log_destination_type", "cloud-watch-logs"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrLogGroupName, cloudwatchLogGroupResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "max_aggregation_interval", "60"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTransitGatewayAttachmentID, transitGatewayAttachmentResourceName, names.AttrID),
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

func TestAccVPCFlowLog_LogDestinationType_cloudWatchLogs(t *testing.T) {
	ctx := acctest.Context(t)
	var flowLog awstypes.FlowLog
	cloudwatchLogGroupResourceName := "aws_cloudwatch_log_group.test"
	resourceName := "aws_flow_log.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFlowLogDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCFlowLogConfig_destinationTypeCloudWatchLogs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowLogExists(ctx, resourceName, &flowLog),
					// We automatically trim :* from ARNs if present
					acctest.CheckResourceAttrRegionalARN(resourceName, "log_destination", "logs", fmt.Sprintf("log-group:%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "log_destination_type", "cloud-watch-logs"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrLogGroupName, cloudwatchLogGroupResourceName, names.AttrName),
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

func TestAccVPCFlowLog_LogDestinationType_kinesisFirehose(t *testing.T) {
	ctx := acctest.Context(t)
	var flowLog awstypes.FlowLog
	kinesisFirehoseResourceName := "aws_kinesis_firehose_delivery_stream.test"
	resourceName := "aws_flow_log.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFlowLogDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCFlowLogConfig_destinationTypeKinesisFirehose(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowLogExists(ctx, resourceName, &flowLog),
					resource.TestCheckResourceAttrPair(resourceName, "log_destination", kinesisFirehoseResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "log_destination_type", "kinesis-data-firehose"),
					resource.TestCheckResourceAttr(resourceName, names.AttrLogGroupName, ""),
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

func TestAccVPCFlowLog_LogDestinationType_s3(t *testing.T) {
	ctx := acctest.Context(t)
	var flowLog awstypes.FlowLog
	s3ResourceName := "aws_s3_bucket.test"
	resourceName := "aws_flow_log.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFlowLogDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCFlowLogConfig_destinationTypeS3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowLogExists(ctx, resourceName, &flowLog),
					resource.TestCheckResourceAttrPair(resourceName, "log_destination", s3ResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "log_destination_type", "s3"),
					resource.TestCheckResourceAttr(resourceName, names.AttrLogGroupName, ""),
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

func TestAccVPCFlowLog_LogDestinationTypeS3_invalid(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix("tf-acc-test-flow-log-s3-invalid")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFlowLogDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccVPCFlowLogConfig_destinationTypeS3Invalid(rName),
				ExpectError: regexache.MustCompile(`(Access Denied for LogDestination|does not exist)`),
			},
		},
	})
}

func TestAccVPCFlowLog_LogDestinationTypeS3DO_plainText(t *testing.T) {
	ctx := acctest.Context(t)
	var flowLog awstypes.FlowLog
	s3ResourceName := "aws_s3_bucket.test"
	resourceName := "aws_flow_log.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFlowLogDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCFlowLogConfig_destinationTypeS3DOPlainText(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowLogExists(ctx, resourceName, &flowLog),
					resource.TestCheckResourceAttrPair(resourceName, "log_destination", s3ResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "log_destination_type", "s3"),
					resource.TestCheckResourceAttr(resourceName, names.AttrLogGroupName, ""),
					resource.TestCheckResourceAttr(resourceName, "destination_options.0.file_format", "plain-text"),
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

func TestAccVPCFlowLog_LogDestinationTypeS3DOPlainText_hiveCompatible(t *testing.T) {
	ctx := acctest.Context(t)
	var flowLog awstypes.FlowLog
	s3ResourceName := "aws_s3_bucket.test"
	resourceName := "aws_flow_log.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFlowLogDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCFlowLogConfig_destinationTypeS3DOPlainTextHiveCompatiblePerHour(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowLogExists(ctx, resourceName, &flowLog),
					resource.TestCheckResourceAttrPair(resourceName, "log_destination", s3ResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "log_destination_type", "s3"),
					resource.TestCheckResourceAttr(resourceName, names.AttrLogGroupName, ""),
					resource.TestCheckResourceAttr(resourceName, "destination_options.0.file_format", "plain-text"),
					resource.TestCheckResourceAttr(resourceName, "destination_options.0.hive_compatible_partitions", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "destination_options.0.per_hour_partition", acctest.CtTrue),
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

func TestAccVPCFlowLog_LogDestinationTypeS3DO_parquet(t *testing.T) {
	ctx := acctest.Context(t)
	var flowLog awstypes.FlowLog
	s3ResourceName := "aws_s3_bucket.test"
	resourceName := "aws_flow_log.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFlowLogDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCFlowLogConfig_destinationTypeS3DOParquet(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowLogExists(ctx, resourceName, &flowLog),
					resource.TestCheckResourceAttrPair(resourceName, "log_destination", s3ResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "log_destination_type", "s3"),
					resource.TestCheckResourceAttr(resourceName, names.AttrLogGroupName, ""),
					resource.TestCheckResourceAttr(resourceName, "destination_options.0.file_format", "parquet"),
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

func TestAccVPCFlowLog_LogDestinationTypeS3DOParquet_hiveCompatible(t *testing.T) {
	ctx := acctest.Context(t)
	var flowLog awstypes.FlowLog
	s3ResourceName := "aws_s3_bucket.test"
	resourceName := "aws_flow_log.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFlowLogDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCFlowLogConfig_destinationTypeS3DOParquetHiveCompatible(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowLogExists(ctx, resourceName, &flowLog),
					resource.TestCheckResourceAttrPair(resourceName, "log_destination", s3ResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "log_destination_type", "s3"),
					resource.TestCheckResourceAttr(resourceName, names.AttrLogGroupName, ""),
					resource.TestCheckResourceAttr(resourceName, "destination_options.0.file_format", "parquet"),
					resource.TestCheckResourceAttr(resourceName, "destination_options.0.hive_compatible_partitions", acctest.CtTrue),
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

func TestAccVPCFlowLog_LogDestinationTypeS3DOParquetHiveCompatible_perHour(t *testing.T) {
	ctx := acctest.Context(t)
	var flowLog awstypes.FlowLog
	s3ResourceName := "aws_s3_bucket.test"
	resourceName := "aws_flow_log.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFlowLogDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCFlowLogConfig_destinationTypeS3DOParquetHiveCompatiblePerHour(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowLogExists(ctx, resourceName, &flowLog),
					resource.TestCheckResourceAttrPair(resourceName, "log_destination", s3ResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "log_destination_type", "s3"),
					resource.TestCheckResourceAttr(resourceName, names.AttrLogGroupName, ""),
					resource.TestCheckResourceAttr(resourceName, "destination_options.0.file_format", "parquet"),
					resource.TestCheckResourceAttr(resourceName, "destination_options.0.hive_compatible_partitions", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "destination_options.0.per_hour_partition", acctest.CtTrue),
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

func TestAccVPCFlowLog_LogDestinationType_maxAggregationInterval(t *testing.T) {
	ctx := acctest.Context(t)
	var flowLog awstypes.FlowLog
	resourceName := "aws_flow_log.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFlowLogDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCFlowLogConfig_maxAggregationInterval(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowLogExists(ctx, resourceName, &flowLog),
					resource.TestCheckResourceAttr(resourceName, "max_aggregation_interval", "60"),
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

func TestAccVPCFlowLog_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var flowLog awstypes.FlowLog
	resourceName := "aws_flow_log.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFlowLogDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCFlowLogConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowLogExists(ctx, resourceName, &flowLog),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCFlowLogConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowLogExists(ctx, resourceName, &flowLog),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccVPCFlowLogConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowLogExists(ctx, resourceName, &flowLog),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccVPCFlowLog_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var flowLog awstypes.FlowLog
	resourceName := "aws_flow_log.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFlowLogDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCFlowLogConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowLogExists(ctx, resourceName, &flowLog),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceFlowLog(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckFlowLogExists(ctx context.Context, n string, v *awstypes.FlowLog) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Flow Log ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindFlowLogByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckFlowLogDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_flow_log" {
				continue
			}

			_, err := tfec2.FindFlowLogByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Flow Log %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccFlowLogConfig_base(rName string) string {
	return acctest.ConfigVPCWithSubnets(rName, 1)
}

func testAccVPCFlowLogConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccFlowLogConfig_base(rName), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ec2.${data.aws_partition.current.dns_suffix}"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_flow_log" "test" {
  iam_role_arn   = aws_iam_role.test.arn
  log_group_name = aws_cloudwatch_log_group.test.name
  traffic_type   = "ALL"
  vpc_id         = aws_vpc.test.id
}
`, rName))
}

func testAccVPCFlowLogConfig_destinationTypeCloudWatchLogs(rName string) string {
	return acctest.ConfigCompose(testAccFlowLogConfig_base(rName), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ec2.${data.aws_partition.current.dns_suffix}"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_flow_log" "test" {
  iam_role_arn         = aws_iam_role.test.arn
  log_destination      = aws_cloudwatch_log_group.test.arn
  log_destination_type = "cloud-watch-logs"
  traffic_type         = "ALL"
  vpc_id               = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCFlowLogConfig_destinationTypeS3(rName string) string {
	return acctest.ConfigCompose(testAccFlowLogConfig_base(rName), fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_flow_log" "test" {
  log_destination      = aws_s3_bucket.test.arn
  log_destination_type = "s3"
  traffic_type         = "ALL"
  vpc_id               = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCFlowLogConfig_destinationTypeS3Invalid(rName string) string {
	return acctest.ConfigCompose(testAccFlowLogConfig_base(rName), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_flow_log" "test" {
  log_destination      = "arn:${data.aws_partition.current.partition}:s3:::does-not-exist"
  log_destination_type = "s3"
  traffic_type         = "ALL"
  vpc_id               = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCFlowLogConfig_destinationTypeS3DOPlainText(rName string) string {
	return acctest.ConfigCompose(testAccFlowLogConfig_base(rName), fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_flow_log" "test" {
  log_destination      = aws_s3_bucket.test.arn
  log_destination_type = "s3"
  traffic_type         = "ALL"
  vpc_id               = aws_vpc.test.id
  destination_options {
    file_format = "plain-text"
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCFlowLogConfig_destinationTypeS3DOPlainTextHiveCompatiblePerHour(rName string) string {
	return acctest.ConfigCompose(testAccFlowLogConfig_base(rName), fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_flow_log" "test" {
  log_destination      = aws_s3_bucket.test.arn
  log_destination_type = "s3"
  traffic_type         = "ALL"
  vpc_id               = aws_vpc.test.id
  destination_options {
    file_format                = "plain-text"
    hive_compatible_partitions = true
    per_hour_partition         = true
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCFlowLogConfig_destinationTypeS3DOParquet(rName string) string {
	return acctest.ConfigCompose(testAccFlowLogConfig_base(rName), fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_flow_log" "test" {
  log_destination      = aws_s3_bucket.test.arn
  log_destination_type = "s3"
  traffic_type         = "ALL"
  vpc_id               = aws_vpc.test.id
  destination_options {
    file_format = "parquet"
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCFlowLogConfig_destinationTypeS3DOParquetHiveCompatible(rName string) string {
	return acctest.ConfigCompose(testAccFlowLogConfig_base(rName), fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_flow_log" "test" {
  log_destination      = aws_s3_bucket.test.arn
  log_destination_type = "s3"
  traffic_type         = "ALL"
  vpc_id               = aws_vpc.test.id
  destination_options {
    file_format                = "parquet"
    hive_compatible_partitions = true
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCFlowLogConfig_destinationTypeS3DOParquetHiveCompatiblePerHour(rName string) string {
	return acctest.ConfigCompose(testAccFlowLogConfig_base(rName), fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_flow_log" "test" {
  log_destination      = aws_s3_bucket.test.arn
  log_destination_type = "s3"
  traffic_type         = "ALL"
  vpc_id               = aws_vpc.test.id
  destination_options {
    file_format                = "parquet"
    hive_compatible_partitions = true
    per_hour_partition         = true
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCFlowLogConfig_subnetID(rName string) string {
	return acctest.ConfigCompose(testAccFlowLogConfig_base(rName), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ec2.${data.aws_partition.current.dns_suffix}"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_flow_log" "test" {
  iam_role_arn   = aws_iam_role.test.arn
  log_group_name = aws_cloudwatch_log_group.test.name
  subnet_id      = aws_subnet.test[0].id
  traffic_type   = "ALL"

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCFlowLogConfig_format(rName string) string {
	return acctest.ConfigCompose(testAccFlowLogConfig_base(rName), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ec2.${data.aws_partition.current.dns_suffix}"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_flow_log" "test" {
  log_destination      = aws_s3_bucket.test.arn
  log_destination_type = "s3"
  traffic_type         = "ALL"
  vpc_id               = aws_vpc.test.id
  log_format           = "$${version} $${vpc-id} $${subnet-id}"

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCFlowLogConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccFlowLogConfig_base(rName), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ec2.${data.aws_partition.current.dns_suffix}"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_flow_log" "test" {
  iam_role_arn   = aws_iam_role.test.arn
  log_group_name = aws_cloudwatch_log_group.test.name
  traffic_type   = "ALL"
  vpc_id         = aws_vpc.test.id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccVPCFlowLogConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccFlowLogConfig_base(rName), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ec2.${data.aws_partition.current.dns_suffix}"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_flow_log" "test" {
  iam_role_arn   = aws_iam_role.test.arn
  log_group_name = aws_cloudwatch_log_group.test.name
  traffic_type   = "ALL"
  vpc_id         = aws_vpc.test.id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccVPCFlowLogConfig_maxAggregationInterval(rName string) string {
	return acctest.ConfigCompose(testAccFlowLogConfig_base(rName), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ec2.${data.aws_partition.current.dns_suffix}"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_flow_log" "test" {
  iam_role_arn   = aws_iam_role.test.arn
  log_group_name = aws_cloudwatch_log_group.test.name
  traffic_type   = "ALL"
  vpc_id         = aws_vpc.test.id

  max_aggregation_interval = 60

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCFlowLogConfig_transitGatewayID(rName string) string {
	return acctest.ConfigCompose(testAccFlowLogConfig_base(rName), fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ec2.${data.aws_partition.current.dns_suffix}"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_flow_log" "test" {
  iam_role_arn             = aws_iam_role.test.arn
  log_group_name           = aws_cloudwatch_log_group.test.name
  max_aggregation_interval = 60
  transit_gateway_id       = aws_ec2_transit_gateway.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCFlowLogConfig_transitGatewayAttachmentID(rName string) string {
	return acctest.ConfigCompose(testAccFlowLogConfig_base(rName), fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id
  subnet_ids         = aws_subnet.test[*].id

  tags = {
    Name = %[1]q
  }
}

data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ec2.${data.aws_partition.current.dns_suffix}"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_flow_log" "test" {
  iam_role_arn                  = aws_iam_role.test.arn
  log_group_name                = aws_cloudwatch_log_group.test.name
  max_aggregation_interval      = 60
  transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCFlowLogConfig_destinationTypeKinesisFirehose(rName string) string {
	return acctest.ConfigCompose(testAccFlowLogConfig_base(rName), fmt.Sprintf(`
resource "aws_flow_log" "test" {
  log_destination      = aws_kinesis_firehose_delivery_stream.test.arn
  log_destination_type = "kinesis-data-firehose"
  traffic_type         = "ALL"
  vpc_id               = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  name        = %[1]q
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn   = aws_iam_role.test.arn
    bucket_arn = aws_s3_bucket.bucket.arn
  }

  tags = {
    "LogDeliveryEnabled" = "true"
  }
}

resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version":"2012-10-17",
  "Statement": [
    {
      "Action":"sts:AssumeRole",
      "Principal":{
        "Service":"firehose.amazonaws.com"
      },
      "Effect":"Allow",
      "Sid":""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version":"2012-10-17",
  "Statement":[
    {
      "Action": [
        "logs:CreateLogDelivery",
        "logs:DeleteLogDelivery",
        "logs:ListLogDeliveries",
        "logs:GetLogDelivery",
        "firehose:TagDeliveryStream"
      ],
      "Effect":"Allow",
      "Resource":"*"
    }
  ]
}
EOF
}
`, rName))
}
