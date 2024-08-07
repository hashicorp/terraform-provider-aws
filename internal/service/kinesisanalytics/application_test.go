// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kinesisanalytics_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/kinesisanalytics"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kinesisanalytics/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfkinesisanalytics "github.com/hashicorp/terraform-provider-aws/internal/service/kinesisanalytics"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKinesisAnalyticsApplication_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesis_analytics_application.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", acctest.Ct0),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct1),
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

func TestAccKinesisAnalyticsApplication_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesis_analytics_application.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfkinesisanalytics.ResourceApplication(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccKinesisAnalyticsApplication_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesis_analytics_application.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccApplicationConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct1),
				),
			},
			{
				Config: testAccApplicationConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct1),
				),
			},
		},
	})
}

func TestAccKinesisAnalyticsApplication_Code_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesis_analytics_application.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_code(rName, "SELECT 1;\n"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "code", "SELECT 1;\n"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", acctest.Ct0),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct1),
				),
			},
			{
				Config: testAccApplicationConfig_code(rName, "SELECT 2;\n"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "code", "SELECT 2;\n"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", acctest.Ct0),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct2),
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

func TestAccKinesisAnalyticsApplication_CloudWatchLoggingOptions_add(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesis_analytics_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	cloudWatchLogStreamResourceName := "aws_cloudwatch_log_stream.test.0"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", acctest.Ct0),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct1),
				),
			},
			{
				Config: testAccApplicationConfig_cloudWatchLoggingOptions(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_logging_options.0.log_stream_arn", cloudWatchLogStreamResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_logging_options.0.role_arn", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", acctest.Ct0),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct2),
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

func TestAccKinesisAnalyticsApplication_CloudWatchLoggingOptions_delete(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesis_analytics_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	cloudWatchLogStreamResourceName := "aws_cloudwatch_log_stream.test.0"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_cloudWatchLoggingOptions(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_logging_options.0.log_stream_arn", cloudWatchLogStreamResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_logging_options.0.role_arn", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", acctest.Ct0),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct1),
				),
			},
			{
				Config: testAccApplicationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", acctest.Ct0),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct2),
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

func TestAccKinesisAnalyticsApplication_CloudWatchLoggingOptions_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesis_analytics_application.test"
	iamRole1ResourceName := "aws_iam_role.test.0"
	iamRole2ResourceName := "aws_iam_role.test.1"
	cloudWatchLogStream1ResourceName := "aws_cloudwatch_log_stream.test.0"
	cloudWatchLogStream2ResourceName := "aws_cloudwatch_log_stream.test.1"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_cloudWatchLoggingOptions(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_logging_options.0.log_stream_arn", cloudWatchLogStream1ResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_logging_options.0.role_arn", iamRole1ResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", acctest.Ct0),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct1),
				),
			},
			{
				Config: testAccApplicationConfig_cloudWatchLoggingOptions(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_logging_options.0.log_stream_arn", cloudWatchLogStream2ResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_logging_options.0.role_arn", iamRole2ResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", acctest.Ct0),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct2),
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

func TestAccKinesisAnalyticsApplication_Input_add(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesis_analytics_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	firehoseResourceName := "aws_kinesis_firehose_delivery_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", acctest.Ct0),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct1),
				),
			},
			{
				Config: testAccApplicationConfig_input(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.stream_names.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "inputs.0.id"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.json.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.name_prefix", "NAME_PREFIX_1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.0.count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.processing_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_firehose.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_firehose.0.resource_arn", firehoseResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_firehose.0.role_arn", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_stream.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.starting_position_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.starting_position_configuration.0.starting_position", ""),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", acctest.Ct0),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct2),
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

func TestAccKinesisAnalyticsApplication_Input_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesis_analytics_application.test"
	iamRole1ResourceName := "aws_iam_role.test.0"
	iamRole2ResourceName := "aws_iam_role.test.1"
	firehoseResourceName := "aws_kinesis_firehose_delivery_stream.test"
	streamsResourceName := "aws_kinesis_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_input(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.stream_names.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "inputs.0.id"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.json.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.name_prefix", "NAME_PREFIX_1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.0.count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.processing_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_firehose.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_firehose.0.resource_arn", firehoseResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_firehose.0.role_arn", iamRole1ResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_stream.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.starting_position_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.starting_position_configuration.0.starting_position", ""),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", acctest.Ct0),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct1),
				),
			},
			{
				Config: testAccApplicationConfig_inputUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.stream_names.#", "42"),
					resource.TestCheckResourceAttrSet(resourceName, "inputs.0.id"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.mapping", "MAPPING-2"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.name", "COLUMN_2"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.sql_type", "VARCHAR(8)"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.1.mapping", "MAPPING-3"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.1.name", "COLUMN_3"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.1.sql_type", "DOUBLE"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_encoding", "UTF-8"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.json.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.json.0.record_row_path", "$path.to.record"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.record_format_type", "JSON"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.name_prefix", "NAME_PREFIX_2"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.0.count", "42"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.processing_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_stream.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_stream.0.resource_arn", streamsResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_stream.0.role_arn", iamRole2ResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.starting_position_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.starting_position_configuration.0.starting_position", ""),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", acctest.Ct0),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct2),
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

func TestAccKinesisAnalyticsApplication_InputProcessing_add(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesis_analytics_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	lambdaResourceName := "aws_lambda_function.test.0"
	firehoseResourceName := "aws_kinesis_firehose_delivery_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_input(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.stream_names.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "inputs.0.id"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.json.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.name_prefix", "NAME_PREFIX_1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.0.count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.processing_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_firehose.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_firehose.0.resource_arn", firehoseResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_firehose.0.role_arn", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_stream.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.starting_position_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.starting_position_configuration.0.starting_position", ""),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", acctest.Ct0),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct1),
				),
			},
			{
				Config: testAccApplicationConfig_inputProcessing(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.stream_names.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "inputs.0.id"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.json.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.name_prefix", "NAME_PREFIX_1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.0.count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.processing_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.processing_configuration.0.lambda.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.processing_configuration.0.lambda.0.resource_arn", lambdaResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.processing_configuration.0.lambda.0.role_arn", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_firehose.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_firehose.0.resource_arn", firehoseResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_firehose.0.role_arn", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_stream.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.starting_position_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.starting_position_configuration.0.starting_position", ""),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", acctest.Ct0),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct3), // Add input processing configuration + update input.
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

func TestAccKinesisAnalyticsApplication_InputProcessing_delete(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesis_analytics_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	lambdaResourceName := "aws_lambda_function.test.0"
	firehoseResourceName := "aws_kinesis_firehose_delivery_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_inputProcessing(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.stream_names.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "inputs.0.id"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.json.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.name_prefix", "NAME_PREFIX_1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.0.count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.processing_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.processing_configuration.0.lambda.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.processing_configuration.0.lambda.0.resource_arn", lambdaResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.processing_configuration.0.lambda.0.role_arn", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_firehose.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_firehose.0.resource_arn", firehoseResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_firehose.0.role_arn", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_stream.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.starting_position_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.starting_position_configuration.0.starting_position", ""),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", acctest.Ct0),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct1),
				),
			},
			{
				Config: testAccApplicationConfig_input(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.stream_names.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "inputs.0.id"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.json.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.name_prefix", "NAME_PREFIX_1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.0.count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.processing_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_firehose.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_firehose.0.resource_arn", firehoseResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_firehose.0.role_arn", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_stream.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.starting_position_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.starting_position_configuration.0.starting_position", ""),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", acctest.Ct0),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct3), // Delete input processing configuration + update input.
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

func TestAccKinesisAnalyticsApplication_InputProcessing_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesis_analytics_application.test"
	iamRole1ResourceName := "aws_iam_role.test.0"
	iamRole2ResourceName := "aws_iam_role.test.1"
	lambda1ResourceName := "aws_lambda_function.test.0"
	lambda2ResourceName := "aws_lambda_function.test.1"
	firehoseResourceName := "aws_kinesis_firehose_delivery_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_inputProcessing(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.stream_names.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "inputs.0.id"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.json.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.name_prefix", "NAME_PREFIX_1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.0.count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.processing_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.processing_configuration.0.lambda.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.processing_configuration.0.lambda.0.resource_arn", lambda1ResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.processing_configuration.0.lambda.0.role_arn", iamRole1ResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_firehose.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_firehose.0.resource_arn", firehoseResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_firehose.0.role_arn", iamRole1ResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_stream.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.starting_position_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.starting_position_configuration.0.starting_position", ""),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", acctest.Ct0),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct1),
				),
			},
			{
				Config: testAccApplicationConfig_inputProcessing(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.stream_names.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "inputs.0.id"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.json.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.name_prefix", "NAME_PREFIX_1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.0.count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.processing_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.processing_configuration.0.lambda.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.processing_configuration.0.lambda.0.resource_arn", lambda2ResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.processing_configuration.0.lambda.0.role_arn", iamRole2ResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_firehose.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_firehose.0.resource_arn", firehoseResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_firehose.0.role_arn", iamRole1ResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_stream.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.starting_position_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.starting_position_configuration.0.starting_position", ""),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", acctest.Ct0),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct2),
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

func TestAccKinesisAnalyticsApplication_Multiple_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesis_analytics_application.test"
	iamRole1ResourceName := "aws_iam_role.test.0"
	iamRole2ResourceName := "aws_iam_role.test.1"
	cloudWatchLogStreamResourceName := "aws_cloudwatch_log_stream.test"
	lambdaResourceName := "aws_lambda_function.test.0"
	firehoseResourceName := "aws_kinesis_firehose_delivery_stream.test"
	streamsResourceName := "aws_kinesis_stream.test"
	s3BucketResourceName := "aws_s3_bucket.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_multiple(rName, "", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_logging_options.0.log_stream_arn", cloudWatchLogStreamResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_logging_options.0.role_arn", iamRole2ResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.stream_names.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "inputs.0.id"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.json.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.name_prefix", "NAME_PREFIX_1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.0.count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.processing_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.processing_configuration.0.lambda.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.processing_configuration.0.lambda.0.resource_arn", lambdaResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.processing_configuration.0.lambda.0.role_arn", iamRole1ResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_firehose.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_firehose.0.resource_arn", firehoseResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_firehose.0.role_arn", iamRole1ResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_stream.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.starting_position_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.starting_position_configuration.0.starting_position", ""),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "outputs.*", map[string]string{
						names.AttrName:                "OUTPUT_1",
						"schema.#":                    acctest.Ct1,
						"schema.0.record_format_type": "CSV",
						"kinesis_firehose.#":          acctest.Ct1,
						"kinesis_stream.#":            acctest.Ct0,
						"lambda.#":                    acctest.Ct0,
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "outputs.*.kinesis_firehose.0.resource_arn", firehoseResourceName, names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "outputs.*.kinesis_firehose.0.role_arn", iamRole2ResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct1),
				),
			},
			{
				Config: testAccApplicationConfig_multipleUpdated(rName, "", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.stream_names.#", "42"),
					resource.TestCheckResourceAttrSet(resourceName, "inputs.0.id"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.mapping", "MAPPING-2"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.name", "COLUMN_2"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.sql_type", "VARCHAR(8)"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.1.mapping", "MAPPING-3"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.1.name", "COLUMN_3"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.1.sql_type", "DOUBLE"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_encoding", "UTF-8"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.json.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.json.0.record_row_path", "$path.to.record"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.record_format_type", "JSON"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.name_prefix", "NAME_PREFIX_2"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.0.count", "42"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.processing_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_stream.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_stream.0.resource_arn", streamsResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_stream.0.role_arn", iamRole2ResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.starting_position_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.starting_position_configuration.0.starting_position", ""),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "outputs.*", map[string]string{
						names.AttrName:                "OUTPUT_2",
						"schema.#":                    acctest.Ct1,
						"schema.0.record_format_type": "JSON",
						"kinesis_firehose.#":          acctest.Ct0,
						"kinesis_stream.#":            acctest.Ct1,
						"lambda.#":                    acctest.Ct0,
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "outputs.*.kinesis_stream.0.resource_arn", streamsResourceName, names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "outputs.*.kinesis_stream.0.role_arn", iamRole2ResourceName, names.AttrARN),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "outputs.*", map[string]string{
						names.AttrName:                "OUTPUT_3",
						"schema.#":                    acctest.Ct1,
						"schema.0.record_format_type": "CSV",
						"kinesis_firehose.#":          acctest.Ct0,
						"kinesis_stream.#":            acctest.Ct0,
						"lambda.#":                    acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "outputs.*.lambda.0.resource_arn", lambdaResourceName, names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "outputs.*.lambda.0.role_arn", iamRole1ResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_columns.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_columns.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.0.csv.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.0.json.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.s3.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "reference_data_sources.0.s3.0.bucket_arn", s3BucketResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.s3.0.file_key", "KEY-1"),
					resource.TestCheckResourceAttrPair(resourceName, "reference_data_sources.0.s3.0.role_arn", iamRole2ResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.table_name", "TABLE-1"),
					resource.TestCheckResourceAttrSet(resourceName, "reference_data_sources.0.id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "Value3"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "8"), // Delete CloudWatch logging options + add reference data source + delete input processing configuration+ update application + delete output + 2 * add output.
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

func TestAccKinesisAnalyticsApplication_Output_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesis_analytics_application.test"
	iamRole1ResourceName := "aws_iam_role.test.0"
	iamRole2ResourceName := "aws_iam_role.test.1"
	lambdaResourceName := "aws_lambda_function.test.0"
	firehoseResourceName := "aws_kinesis_firehose_delivery_stream.test"
	streamsResourceName := "aws_kinesis_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_output(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "outputs.*", map[string]string{
						names.AttrName:                "OUTPUT_1",
						"schema.#":                    acctest.Ct1,
						"schema.0.record_format_type": "CSV",
						"kinesis_firehose.#":          acctest.Ct1,
						"kinesis_stream.#":            acctest.Ct0,
						"lambda.#":                    acctest.Ct0,
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "outputs.*.kinesis_firehose.0.resource_arn", firehoseResourceName, names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "outputs.*.kinesis_firehose.0.role_arn", iamRole1ResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", acctest.Ct0),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct1),
				),
			},
			{
				Config: testAccApplicationConfig_outputUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "outputs.*", map[string]string{
						names.AttrName:                "OUTPUT_2",
						"schema.#":                    acctest.Ct1,
						"schema.0.record_format_type": "JSON",
						"kinesis_firehose.#":          acctest.Ct0,
						"kinesis_stream.#":            acctest.Ct1,
						"lambda.#":                    acctest.Ct0,
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "outputs.*.kinesis_stream.0.resource_arn", streamsResourceName, names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "outputs.*.kinesis_stream.0.role_arn", iamRole2ResourceName, names.AttrARN),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "outputs.*", map[string]string{
						names.AttrName:                "OUTPUT_3",
						"schema.#":                    acctest.Ct1,
						"schema.0.record_format_type": "CSV",
						"kinesis_firehose.#":          acctest.Ct0,
						"kinesis_stream.#":            acctest.Ct0,
						"lambda.#":                    acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "outputs.*.lambda.0.resource_arn", lambdaResourceName, names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "outputs.*.lambda.0.role_arn", iamRole1ResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", acctest.Ct0),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct4), // 1 * output deletion + 2 * output addition.
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccApplicationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", acctest.Ct0),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "6"), // 2 * output deletion.
				),
			},
		},
	})
}

func TestAccKinesisAnalyticsApplication_ReferenceDataSource_add(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesis_analytics_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	s3BucketResourceName := "aws_s3_bucket.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", acctest.Ct0),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct1),
				),
			},
			{
				Config: testAccApplicationConfig_referenceDataSource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_columns.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_columns.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.0.csv.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.0.json.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.s3.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "reference_data_sources.0.s3.0.bucket_arn", s3BucketResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.s3.0.file_key", "KEY-1"),
					resource.TestCheckResourceAttrPair(resourceName, "reference_data_sources.0.s3.0.role_arn", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.table_name", "TABLE-1"),
					resource.TestCheckResourceAttrSet(resourceName, "reference_data_sources.0.id"),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct2),
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

func TestAccKinesisAnalyticsApplication_ReferenceDataSource_delete(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesis_analytics_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	s3BucketResourceName := "aws_s3_bucket.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_referenceDataSource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_columns.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_columns.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.0.csv.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.0.json.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.s3.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "reference_data_sources.0.s3.0.bucket_arn", s3BucketResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.s3.0.file_key", "KEY-1"),
					resource.TestCheckResourceAttrPair(resourceName, "reference_data_sources.0.s3.0.role_arn", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.table_name", "TABLE-1"),
					resource.TestCheckResourceAttrSet(resourceName, "reference_data_sources.0.id"),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct2),
				),
			},
			{
				Config: testAccApplicationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", acctest.Ct0),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct3),
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

func TestAccKinesisAnalyticsApplication_ReferenceDataSource_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesis_analytics_application.test"
	iamRole1ResourceName := "aws_iam_role.test.0"
	iamRole2ResourceName := "aws_iam_role.test.1"
	s3BucketResourceName := "aws_s3_bucket.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_referenceDataSource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_columns.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_columns.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.0.csv.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.0.json.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.s3.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "reference_data_sources.0.s3.0.bucket_arn", s3BucketResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.s3.0.file_key", "KEY-1"),
					resource.TestCheckResourceAttrPair(resourceName, "reference_data_sources.0.s3.0.role_arn", iamRole1ResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.table_name", "TABLE-1"),
					resource.TestCheckResourceAttrSet(resourceName, "reference_data_sources.0.id"),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct2),
				),
			},
			{
				Config: testAccApplicationConfig_referenceDataSourceUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_columns.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_columns.0.mapping", "MAPPING-2"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_columns.0.name", "COLUMN_2"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_columns.0.sql_type", "VARCHAR(8)"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_columns.1.mapping", "MAPPING-3"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_columns.1.name", "COLUMN_3"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_columns.1.sql_type", "DOUBLE"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_encoding", "UTF-8"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.0.csv.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.0.json.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.0.json.0.record_row_path", "$path.to.record"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.record_format_type", "JSON"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.s3.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "reference_data_sources.0.s3.0.bucket_arn", s3BucketResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.s3.0.file_key", "KEY-2"),
					resource.TestCheckResourceAttrPair(resourceName, "reference_data_sources.0.s3.0.role_arn", iamRole2ResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.table_name", "TABLE-2"),
					resource.TestCheckResourceAttrSet(resourceName, "reference_data_sources.0.id"),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct3),
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

func TestAccKinesisAnalyticsApplication_StartApplication_onCreate(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesis_analytics_application.test"
	iamRole1ResourceName := "aws_iam_role.test.0"
	firehoseResourceName := "aws_kinesis_firehose_delivery_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_start(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.stream_names.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "inputs.0.id"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.json.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.name_prefix", "NAME_PREFIX_1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.0.count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.processing_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_firehose.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_firehose.0.resource_arn", firehoseResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_firehose.0.role_arn", iamRole1ResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_stream.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.starting_position_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.starting_position_configuration.0.starting_position", "NOW"),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "start_application", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"start_application"},
			},
		},
	})
}

func TestAccKinesisAnalyticsApplication_StartApplication_onUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesis_analytics_application.test"
	iamRole1ResourceName := "aws_iam_role.test.0"
	firehoseResourceName := "aws_kinesis_firehose_delivery_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_start(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.stream_names.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "inputs.0.id"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.json.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.name_prefix", "NAME_PREFIX_1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.0.count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.processing_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_firehose.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_firehose.0.resource_arn", firehoseResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_firehose.0.role_arn", iamRole1ResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_stream.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.starting_position_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.starting_position_configuration.0.starting_position", ""),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "start_application", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"start_application"},
			},
			{
				Config: testAccApplicationConfig_start(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.stream_names.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "inputs.0.id"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.json.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.name_prefix", "NAME_PREFIX_1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.0.count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.processing_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_firehose.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_firehose.0.resource_arn", firehoseResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_firehose.0.role_arn", iamRole1ResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_stream.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.starting_position_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.starting_position_configuration.0.starting_position", "NOW"),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "start_application", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct2),
				),
			},
			{
				Config: testAccApplicationConfig_start(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.stream_names.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "inputs.0.id"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.json.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.name_prefix", "NAME_PREFIX_1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.0.count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.processing_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_firehose.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_firehose.0.resource_arn", firehoseResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_firehose.0.role_arn", iamRole1ResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_stream.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.starting_position_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.starting_position_configuration.0.starting_position", "NOW"),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "start_application", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct2),
				),
			},
		},
	})
}

func TestAccKinesisAnalyticsApplication_StartApplication_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesis_analytics_application.test"
	iamRole1ResourceName := "aws_iam_role.test.0"
	iamRole2ResourceName := "aws_iam_role.test.1"
	cloudWatchLogStreamResourceName := "aws_cloudwatch_log_stream.test"
	lambdaResourceName := "aws_lambda_function.test.0"
	firehoseResourceName := "aws_kinesis_firehose_delivery_stream.test"
	streamsResourceName := "aws_kinesis_stream.test"
	s3BucketResourceName := "aws_s3_bucket.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_multiple(rName, acctest.CtTrue, "LAST_STOPPED_POINT"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_logging_options.0.log_stream_arn", cloudWatchLogStreamResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_logging_options.0.role_arn", iamRole2ResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.stream_names.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "inputs.0.id"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.json.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.name_prefix", "NAME_PREFIX_1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.0.count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.processing_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.processing_configuration.0.lambda.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.processing_configuration.0.lambda.0.resource_arn", lambdaResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.processing_configuration.0.lambda.0.role_arn", iamRole1ResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_firehose.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_firehose.0.resource_arn", firehoseResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_firehose.0.role_arn", iamRole1ResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_stream.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.starting_position_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.starting_position_configuration.0.starting_position", "LAST_STOPPED_POINT"),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "outputs.*", map[string]string{
						names.AttrName:                "OUTPUT_1",
						"schema.#":                    acctest.Ct1,
						"schema.0.record_format_type": "CSV",
						"kinesis_firehose.#":          acctest.Ct1,
						"kinesis_stream.#":            acctest.Ct0,
						"lambda.#":                    acctest.Ct0,
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "outputs.*.kinesis_firehose.0.resource_arn", firehoseResourceName, names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "outputs.*.kinesis_firehose.0.role_arn", iamRole2ResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "start_application", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct1),
				),
			},
			{
				Config: testAccApplicationConfig_multipleUpdated(rName, acctest.CtTrue, "LAST_STOPPED_POINT"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.stream_names.#", "42"),
					resource.TestCheckResourceAttrSet(resourceName, "inputs.0.id"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.mapping", "MAPPING-2"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.name", "COLUMN_2"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.sql_type", "VARCHAR(8)"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.1.mapping", "MAPPING-3"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.1.name", "COLUMN_3"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.1.sql_type", "DOUBLE"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_encoding", "UTF-8"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.json.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.json.0.record_row_path", "$path.to.record"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.record_format_type", "JSON"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.name_prefix", "NAME_PREFIX_2"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.0.count", "42"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.processing_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_stream.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_stream.0.resource_arn", streamsResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_stream.0.role_arn", iamRole2ResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.starting_position_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.starting_position_configuration.0.starting_position", "LAST_STOPPED_POINT"),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "outputs.*", map[string]string{
						names.AttrName:                "OUTPUT_2",
						"schema.#":                    acctest.Ct1,
						"schema.0.record_format_type": "JSON",
						"kinesis_firehose.#":          acctest.Ct0,
						"kinesis_stream.#":            acctest.Ct1,
						"lambda.#":                    acctest.Ct0,
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "outputs.*.kinesis_stream.0.resource_arn", streamsResourceName, names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "outputs.*.kinesis_stream.0.role_arn", iamRole2ResourceName, names.AttrARN),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "outputs.*", map[string]string{
						names.AttrName:                "OUTPUT_3",
						"schema.#":                    acctest.Ct1,
						"schema.0.record_format_type": "CSV",
						"kinesis_firehose.#":          acctest.Ct0,
						"kinesis_stream.#":            acctest.Ct0,
						"lambda.#":                    acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "outputs.*.lambda.0.resource_arn", lambdaResourceName, names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "outputs.*.lambda.0.role_arn", iamRole1ResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_columns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_columns.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_columns.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.0.csv.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.0.json.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.s3.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "reference_data_sources.0.s3.0.bucket_arn", s3BucketResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.s3.0.file_key", "KEY-1"),
					resource.TestCheckResourceAttrPair(resourceName, "reference_data_sources.0.s3.0.role_arn", iamRole2ResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.table_name", "TABLE-1"),
					resource.TestCheckResourceAttrSet(resourceName, "reference_data_sources.0.id"),
					resource.TestCheckResourceAttr(resourceName, "start_application", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "Value3"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "8"), // Delete CloudWatch logging options + add reference data source + delete input processing configuration+ update application + delete output + 2 * add output.
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"start_application"},
			},
		},
	})
}

func testAccCheckApplicationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).KinesisAnalyticsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_kinesis_analytics_application" {
				continue
			}

			_, err := tfkinesisanalytics.FindApplicationDetailByName(ctx, conn, rs.Primary.Attributes[names.AttrName])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Kinesis Analytics Application %s still exists", rs.Primary.ID)
		}
		return nil
	}
}

func testAccCheckApplicationExists(ctx context.Context, n string, v *awstypes.ApplicationDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).KinesisAnalyticsClient(ctx)

		application, err := tfkinesisanalytics.FindApplicationDetailByName(ctx, conn, rs.Primary.Attributes[names.AttrName])

		if err != nil {
			return err
		}

		*v = *application

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).KinesisAnalyticsClient(ctx)

	input := &kinesisanalytics.ListApplicationsInput{}

	_, err := conn.ListApplications(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccApplicationConfigBaseIAMRole(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  count = 2

  name               = "%[1]s.${count.index}"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": ["sts:AssumeRole"],
      "Principal": {"Service": "firehose.amazonaws.com"}
    },
    {
      "Effect": "Allow",
      "Action": ["sts:AssumeRole"],
      "Principal": {"Service": "kinesisanalytics.amazonaws.com"}
    },
    {
      "Effect": "Allow",
      "Action": ["sts:AssumeRole"],
      "Principal": {"Service": "lambda.amazonaws.com"}
    }
  ]
}
EOF
}

resource "aws_iam_policy" "test" {
  name   = %[1]q
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": ["ec2:*"],
      "Resource": ["*"]
    },
    {
      "Effect": "Allow",
      "Action": ["firehose:*"],
      "Resource": ["*"]
    },
    {
      "Effect": "Allow",
      "Action": ["lambda:*"],
      "Resource": ["*"]
    },
    {
      "Effect": "Allow",
      "Action": ["logs:*"],
      "Resource": ["*"]
    },
    {
      "Effect": "Allow",
      "Action": ["s3:*"],
      "Resource": ["*"]
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "test" {
  count = 2

  role       = aws_iam_role.test[count.index].name
  policy_arn = aws_iam_policy.test.arn
}
`, rName)
}

func testAccApplicationConfigBaseInputOutput(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_lambda_function" "test" {
  count = 2

  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%[1]s_${count.index}"
  handler       = "exports.example"
  role          = aws_iam_role.test[0].arn
  runtime       = "nodejs16.x"
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  name        = %[1]q
  destination = "extended_s3"

  extended_s3_configuration {
    bucket_arn = aws_s3_bucket.test.arn
    role_arn   = aws_iam_role.test[0].arn
  }
}

resource "aws_kinesis_stream" "test" {
  name        = %[1]q
  shard_count = 1
}
`, rName)
}

func testAccApplicationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_analytics_application" "test" {
  name = %[1]q
}
`, rName)
}

func testAccApplicationConfig_code(rName, code string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_analytics_application" "test" {
  name        = %[1]q
  description = "test"
  code        = %[2]q
}
`, rName, code)
}

func testAccApplicationConfig_cloudWatchLoggingOptions(rName string, streamIndex int) string {
	return acctest.ConfigCompose(
		testAccApplicationConfigBaseIAMRole(rName),
		fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_log_stream" "test" {
  count = 2

  name           = "%[1]s.${count.index}"
  log_group_name = aws_cloudwatch_log_group.test.name
}

resource "aws_kinesis_analytics_application" "test" {
  name = %[1]q

  cloudwatch_logging_options {
    log_stream_arn = aws_cloudwatch_log_stream.test.%[2]d.arn
    role_arn       = aws_iam_role.test.%[2]d.arn
  }
}
`, rName, streamIndex))
}

func testAccApplicationConfig_input(rName string) string {
	return acctest.ConfigCompose(
		testAccApplicationConfigBaseIAMRole(rName),
		testAccApplicationConfigBaseInputOutput(rName),
		fmt.Sprintf(`
resource "aws_kinesis_analytics_application" "test" {
  name = %[1]q

  inputs {
    name_prefix = "NAME_PREFIX_1"

    schema {
      record_columns {
        name     = "COLUMN_1"
        sql_type = "INTEGER"
      }

      record_format {
        mapping_parameters {
          csv {
            record_column_delimiter = ","
            record_row_delimiter    = "|"
          }
        }
      }
    }

    kinesis_firehose {
      resource_arn = aws_kinesis_firehose_delivery_stream.test.arn
      role_arn     = aws_iam_role.test[0].arn
    }
  }
}
`, rName))
}

func testAccApplicationConfig_inputUpdated(rName string) string {
	return acctest.ConfigCompose(
		testAccApplicationConfigBaseIAMRole(rName),
		testAccApplicationConfigBaseInputOutput(rName),
		fmt.Sprintf(`
resource "aws_kinesis_analytics_application" "test" {
  name = %[1]q

  inputs {
    name_prefix = "NAME_PREFIX_2"

    parallelism {
      count = 42
    }

    schema {
      record_columns {
        name     = "COLUMN_2"
        sql_type = "VARCHAR(8)"
        mapping  = "MAPPING-2"
      }

      record_columns {
        name     = "COLUMN_3"
        sql_type = "DOUBLE"
        mapping  = "MAPPING-3"
      }

      record_encoding = "UTF-8"

      record_format {
        mapping_parameters {
          json {
            record_row_path = "$path.to.record"
          }
        }
      }
    }

    kinesis_stream {
      resource_arn = aws_kinesis_stream.test.arn
      role_arn     = aws_iam_role.test[1].arn
    }
  }
}
`, rName))
}

func testAccApplicationConfig_inputProcessing(rName string, lambdaIndex int) string {
	return acctest.ConfigCompose(
		testAccApplicationConfigBaseIAMRole(rName),
		testAccApplicationConfigBaseInputOutput(rName),
		fmt.Sprintf(`
resource "aws_kinesis_analytics_application" "test" {
  name = %[1]q

  inputs {
    name_prefix = "NAME_PREFIX_1"

    schema {
      record_columns {
        name     = "COLUMN_1"
        sql_type = "INTEGER"
      }

      record_format {
        mapping_parameters {
          csv {
            record_column_delimiter = ","
            record_row_delimiter    = "|"
          }
        }
      }
    }

    processing_configuration {
      lambda {
        resource_arn = aws_lambda_function.test.%[2]d.arn
        role_arn     = aws_iam_role.test.%[2]d.arn
      }
    }

    kinesis_firehose {
      resource_arn = aws_kinesis_firehose_delivery_stream.test.arn
      role_arn     = aws_iam_role.test[0].arn
    }
  }
}
`, rName, lambdaIndex))
}

func testAccApplicationConfig_multiple(rName, startApplication, startingPosition string) string {
	if startApplication == "" {
		startApplication = "null"
	}
	if startingPosition == "" {
		startingPosition = "null"
	} else {
		startingPosition = strconv.Quote(startingPosition)
	}

	return acctest.ConfigCompose(
		testAccApplicationConfigBaseIAMRole(rName),
		testAccApplicationConfigBaseInputOutput(rName),
		fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_log_stream" "test" {
  name           = %[1]q
  log_group_name = aws_cloudwatch_log_group.test.name
}

resource "aws_kinesis_analytics_application" "test" {
  name = %[1]q

  cloudwatch_logging_options {
    log_stream_arn = aws_cloudwatch_log_stream.test.arn
    role_arn       = aws_iam_role.test[1].arn
  }

  inputs {
    name_prefix = "NAME_PREFIX_1"

    schema {
      record_columns {
        name     = "COLUMN_1"
        sql_type = "INTEGER"
      }

      record_format {
        mapping_parameters {
          csv {
            record_column_delimiter = ","
            record_row_delimiter    = "|"
          }
        }
      }
    }

    processing_configuration {
      lambda {
        resource_arn = aws_lambda_function.test[0].arn
        role_arn     = aws_iam_role.test[0].arn
      }
    }

    kinesis_firehose {
      resource_arn = aws_kinesis_firehose_delivery_stream.test.arn
      role_arn     = aws_iam_role.test[0].arn
    }

    starting_position_configuration {
      starting_position = %[3]s
    }
  }

  outputs {
    name = "OUTPUT_1"

    schema {
      record_format_type = "CSV"
    }

    kinesis_firehose {
      resource_arn = aws_kinesis_firehose_delivery_stream.test.arn
      role_arn     = aws_iam_role.test[1].arn
    }
  }

  tags = {
    Key1 = "Value1"
  }

  start_application = %[2]s
}
`, rName, startApplication, startingPosition))
}

func testAccApplicationConfig_multipleUpdated(rName, startApplication, startingPosition string) string {
	if startApplication == "" {
		startApplication = "null"
	}
	if startingPosition == "" {
		startingPosition = "null"
	} else {
		startingPosition = strconv.Quote(startingPosition)
	}

	return acctest.ConfigCompose(
		testAccApplicationConfigBaseIAMRole(rName),
		testAccApplicationConfigBaseInputOutput(rName),
		fmt.Sprintf(`
resource "aws_kinesis_analytics_application" "test" {
  name = %[1]q

  inputs {
    name_prefix = "NAME_PREFIX_2"

    parallelism {
      count = 42
    }

    schema {
      record_columns {
        name     = "COLUMN_2"
        sql_type = "VARCHAR(8)"
        mapping  = "MAPPING-2"
      }

      record_columns {
        name     = "COLUMN_3"
        sql_type = "DOUBLE"
        mapping  = "MAPPING-3"
      }

      record_encoding = "UTF-8"

      record_format {
        mapping_parameters {
          json {
            record_row_path = "$path.to.record"
          }
        }
      }
    }

    kinesis_stream {
      resource_arn = aws_kinesis_stream.test.arn
      role_arn     = aws_iam_role.test[1].arn
    }

    starting_position_configuration {
      starting_position = %[3]s
    }
  }

  outputs {
    name = "OUTPUT_2"

    schema {
      record_format_type = "JSON"
    }

    kinesis_stream {
      resource_arn = aws_kinesis_stream.test.arn
      role_arn     = aws_iam_role.test[1].arn
    }
  }

  outputs {
    name = "OUTPUT_3"

    schema {
      record_format_type = "CSV"
    }

    lambda {
      resource_arn = aws_lambda_function.test[0].arn
      role_arn     = aws_iam_role.test[0].arn
    }
  }

  reference_data_sources {
    table_name = "TABLE-1"

    schema {
      record_columns {
        name     = "COLUMN_1"
        sql_type = "INTEGER"
      }

      record_format {
        mapping_parameters {
          csv {
            record_column_delimiter = ","
            record_row_delimiter    = "|"
          }
        }
      }
    }

    s3 {
      bucket_arn = aws_s3_bucket.test.arn
      file_key   = "KEY-1"
      role_arn   = aws_iam_role.test[1].arn
    }
  }

  tags = {
    Key2 = "Value2"
    Key3 = "Value3"
  }

  start_application = %[2]s
}
`, rName, startApplication, startingPosition))
}

func testAccApplicationConfig_output(rName string) string {
	return acctest.ConfigCompose(
		testAccApplicationConfigBaseIAMRole(rName),
		testAccApplicationConfigBaseInputOutput(rName),
		fmt.Sprintf(`
resource "aws_kinesis_analytics_application" "test" {
  name = %[1]q

  outputs {
    name = "OUTPUT_1"

    schema {
      record_format_type = "CSV"
    }

    kinesis_firehose {
      resource_arn = aws_kinesis_firehose_delivery_stream.test.arn
      role_arn     = aws_iam_role.test[0].arn
    }
  }
}
`, rName))
}

func testAccApplicationConfig_outputUpdated(rName string) string {
	return acctest.ConfigCompose(
		testAccApplicationConfigBaseIAMRole(rName),
		testAccApplicationConfigBaseInputOutput(rName),
		fmt.Sprintf(`
resource "aws_kinesis_analytics_application" "test" {
  name = %[1]q

  outputs {
    name = "OUTPUT_2"

    schema {
      record_format_type = "JSON"
    }

    kinesis_stream {
      resource_arn = aws_kinesis_stream.test.arn
      role_arn     = aws_iam_role.test[1].arn
    }
  }

  outputs {
    name = "OUTPUT_3"

    schema {
      record_format_type = "CSV"
    }

    lambda {
      resource_arn = aws_lambda_function.test[0].arn
      role_arn     = aws_iam_role.test[0].arn
    }
  }
}
`, rName))
}

func testAccApplicationConfig_referenceDataSource(rName string) string {
	return acctest.ConfigCompose(
		testAccApplicationConfigBaseIAMRole(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_kinesis_analytics_application" "test" {
  name = %[1]q

  reference_data_sources {
    table_name = "TABLE-1"

    schema {
      record_columns {
        name     = "COLUMN_1"
        sql_type = "INTEGER"
      }

      record_format {
        mapping_parameters {
          csv {
            record_column_delimiter = ","
            record_row_delimiter    = "|"
          }
        }
      }
    }

    s3 {
      bucket_arn = aws_s3_bucket.test.arn
      file_key   = "KEY-1"
      role_arn   = aws_iam_role.test[0].arn
    }
  }
}
`, rName))
}

func testAccApplicationConfig_referenceDataSourceUpdated(rName string) string {
	return acctest.ConfigCompose(
		testAccApplicationConfigBaseIAMRole(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_kinesis_analytics_application" "test" {
  name = %[1]q

  reference_data_sources {
    table_name = "TABLE-2"

    schema {
      record_columns {
        name     = "COLUMN_2"
        sql_type = "VARCHAR(8)"
        mapping  = "MAPPING-2"
      }

      record_columns {
        name     = "COLUMN_3"
        sql_type = "DOUBLE"
        mapping  = "MAPPING-3"
      }

      record_encoding = "UTF-8"

      record_format {
        mapping_parameters {
          json {
            record_row_path = "$path.to.record"
          }
        }
      }
    }

    s3 {
      bucket_arn = aws_s3_bucket.test.arn
      file_key   = "KEY-2"
      role_arn   = aws_iam_role.test[1].arn
    }
  }
}
`, rName))
}

func testAccApplicationConfig_start(rName string, start bool) string {
	return acctest.ConfigCompose(
		testAccApplicationConfigBaseIAMRole(rName),
		testAccApplicationConfigBaseInputOutput(rName),
		fmt.Sprintf(`
resource "aws_kinesis_analytics_application" "test" {
  name = %[1]q

  inputs {
    name_prefix = "NAME_PREFIX_1"

    schema {
      record_columns {
        name     = "COLUMN_1"
        sql_type = "INTEGER"
      }

      record_format {
        mapping_parameters {
          csv {
            record_column_delimiter = ","
            record_row_delimiter    = "|"
          }
        }
      }
    }

    kinesis_firehose {
      resource_arn = aws_kinesis_firehose_delivery_stream.test.arn
      role_arn     = aws_iam_role.test[0].arn
    }

    starting_position_configuration {
      starting_position = (%[2]t ? "NOW" : null)
    }
  }

  start_application = %[2]t
}
`, rName, start))
}

func testAccApplicationConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_analytics_application" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccApplicationConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_analytics_application" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
