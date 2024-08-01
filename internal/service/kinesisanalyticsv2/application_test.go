// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kinesisanalyticsv2_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/kinesisanalyticsv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kinesisanalyticsv2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfkinesisanalyticsv2 "github.com/hashicorp/terraform-provider-aws/internal/service/kinesisanalyticsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKinesisAnalyticsV2Application_basicFlinkApplication(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_basicFlink(rName, "FLINK-1_6"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.0.snapshots_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpointing_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpoint_interval", "60000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.min_pause_between_checkpoints", "5000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.log_level", "INFO"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.metrics_level", "APPLICATION"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.auto_scaling_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism_per_kpu", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "FLINK-1_6"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct1),
				),
			},
			{
				Config: testAccApplicationConfig_basicFlink(rName, "FLINK-1_8"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.0.snapshots_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpointing_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpoint_interval", "60000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.min_pause_between_checkpoints", "5000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.log_level", "INFO"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.metrics_level", "APPLICATION"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.auto_scaling_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism_per_kpu", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "FLINK-1_8"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct1),
				),
			},
			{
				Config: testAccApplicationConfig_basicFlink(rName, "FLINK-1_11"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.0.snapshots_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpointing_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpoint_interval", "60000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.min_pause_between_checkpoints", "5000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.log_level", "INFO"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.metrics_level", "APPLICATION"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.auto_scaling_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism_per_kpu", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "FLINK-1_11"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct1),
				),
			},
			{
				Config: testAccApplicationConfig_basicFlink(rName, "FLINK-1_13"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.0.snapshots_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpointing_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpoint_interval", "60000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.min_pause_between_checkpoints", "5000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.log_level", "INFO"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.metrics_level", "APPLICATION"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.auto_scaling_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism_per_kpu", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "FLINK-1_13"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct1),
				),
			},
			{
				Config: testAccApplicationConfig_basicFlink(rName, "FLINK-1_15"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.0.snapshots_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpointing_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpoint_interval", "60000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.min_pause_between_checkpoints", "5000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.log_level", "INFO"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.metrics_level", "APPLICATION"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.auto_scaling_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism_per_kpu", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "FLINK-1_15"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct1),
				),
			},
			{
				Config: testAccApplicationConfig_basicFlink(rName, "FLINK-1_18"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.0.snapshots_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpointing_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpoint_interval", "60000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.min_pause_between_checkpoints", "5000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.log_level", "INFO"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.metrics_level", "APPLICATION"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.auto_scaling_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism_per_kpu", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "FLINK-1_18"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct1),
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

func TestAccKinesisAnalyticsV2Application_basicSQLApplication(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_basicSQL(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct1),
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

func TestAccKinesisAnalyticsV2Application_applicationMode(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_applicationMode(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_mode", "STREAMING"),
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

func TestAccKinesisAnalyticsV2Application_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_basicSQL(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfkinesisanalyticsv2.ResourceApplication(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccKinesisAnalyticsV2Application_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct1),
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
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct1),
				),
			},
			{
				Config: testAccApplicationConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccKinesisAnalyticsV2Application_ApplicationCode_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_applicationCodeConfiguration(rName, "SELECT 1;\n"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "SELECT 1;\n"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct1),
				),
			},
			{
				Config: testAccApplicationConfig_applicationCodeConfiguration(rName, "SELECT 2;\n"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "SELECT 2;\n"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct2),
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

func TestAccKinesisAnalyticsV2Application_CloudWatchLoggingOptions_add(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	cloudWatchLogStreamResourceName := "aws_cloudwatch_log_stream.test.0"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_basicSQL(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct1),
				),
			},
			{
				Config: testAccApplicationConfig_cloudWatchLoggingOptions(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_logging_options.0.log_stream_arn", cloudWatchLogStreamResourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct2),
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

func TestAccKinesisAnalyticsV2Application_CloudWatchLoggingOptions_delete(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	cloudWatchLogStreamResourceName := "aws_cloudwatch_log_stream.test.0"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_cloudWatchLoggingOptions(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_logging_options.0.log_stream_arn", cloudWatchLogStreamResourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct1),
				),
			},
			{
				Config: testAccApplicationConfig_basicSQL(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct2),
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

func TestAccKinesisAnalyticsV2Application_CloudWatchLoggingOptions_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	cloudWatchLogStream1ResourceName := "aws_cloudwatch_log_stream.test.0"
	cloudWatchLogStream2ResourceName := "aws_cloudwatch_log_stream.test.1"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_cloudWatchLoggingOptions(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_logging_options.0.log_stream_arn", cloudWatchLogStream1ResourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct1),
				),
			},
			{
				Config: testAccApplicationConfig_cloudWatchLoggingOptions(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_logging_options.0.log_stream_arn", cloudWatchLogStream2ResourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct2),
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

func TestAccKinesisAnalyticsV2Application_EnvironmentProperties_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	s3BucketResourceName := "aws_s3_bucket.test"
	s3ObjectResourceName := "aws_s3_object.test.0"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_environmentProperties(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.bucket_arn", s3BucketResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.file_key", s3ObjectResourceName, names.AttrKey),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.object_version", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "ZIPFILE"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.0.snapshots_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.environment_properties.0.property_group.*", map[string]string{
						"property_group_id": "PROPERTY-GROUP-ID1",
						"property_map.%":    acctest.Ct2,
						"property_map.Key9": "Value1",
						"property_map.Key8": "Value2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.environment_properties.0.property_group.*", map[string]string{
						"property_group_id": "PROPERTY-GROUP-ID2",
						"property_map.%":    acctest.Ct3,
						"property_map.KeyA": "ValueZ",
						"property_map.KeyB": "ValueY",
						"property_map.KeyC": "ValueX",
					}),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpointing_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpoint_interval", "60000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.min_pause_between_checkpoints", "5000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.log_level", "INFO"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.metrics_level", "APPLICATION"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.auto_scaling_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism_per_kpu", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "FLINK-1_8"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct1),
				),
			},
			{
				Config: testAccApplicationConfig_environmentPropertiesUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.bucket_arn", s3BucketResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.file_key", s3ObjectResourceName, names.AttrKey),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.object_version", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "ZIPFILE"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.0.snapshots_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.environment_properties.0.property_group.*", map[string]string{
						"property_group_id": "PROPERTY-GROUP-ID2",
						"property_map.%":    acctest.Ct2,
						"property_map.KeyA": "ValueZ",
						"property_map.KeyC": "ValueW",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.environment_properties.0.property_group.*", map[string]string{
						"property_group_id": "PROPERTY-GROUP-ID3",
						"property_map.%":    acctest.Ct1,
						"property_map.Key":  "Value",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.environment_properties.0.property_group.*", map[string]string{
						"property_group_id":     "PROPERTY-GROUP-ID4",
						"property_map.%":        acctest.Ct1,
						"property_map.KeyAlpha": "ValueOmega",
					}),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpointing_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpoint_interval", "60000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.min_pause_between_checkpoints", "5000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.log_level", "INFO"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.metrics_level", "APPLICATION"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.auto_scaling_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism_per_kpu", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "FLINK-1_8"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccApplicationConfig_environmentPropertiesNotSpecified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.bucket_arn", s3BucketResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.file_key", s3ObjectResourceName, names.AttrKey),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.object_version", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "ZIPFILE"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.0.snapshots_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpointing_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpoint_interval", "60000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.min_pause_between_checkpoints", "5000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.log_level", "INFO"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.metrics_level", "APPLICATION"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.auto_scaling_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism_per_kpu", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "FLINK-1_8"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct3),
				),
			},
		},
	})
}

func TestAccKinesisAnalyticsV2Application_FlinkApplication_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	s3BucketResourceName := "aws_s3_bucket.test"
	s3Object1ResourceName := "aws_s3_object.test.0"
	s3Object2ResourceName := "aws_s3_object.test.1"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_flinkConfiguration(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.bucket_arn", s3BucketResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.file_key", s3Object1ResourceName, names.AttrKey),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.object_version", s3Object1ResourceName, "version_id"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "ZIPFILE"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.0.snapshots_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpointing_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpoint_interval", "60000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.min_pause_between_checkpoints", "5000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.configuration_type", "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.log_level", "DEBUG"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.metrics_level", "TASK"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.auto_scaling_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.configuration_type", "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism_per_kpu", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "FLINK-1_8"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct1),
				),
			},
			{
				Config: testAccApplicationConfig_flinkConfigurationUpdated(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.bucket_arn", s3BucketResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.file_key", s3Object2ResourceName, names.AttrKey),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.object_version", s3Object2ResourceName, "version_id"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "ZIPFILE"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.0.snapshots_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpointing_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpoint_interval", "55000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.configuration_type", "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.min_pause_between_checkpoints", "5500"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.configuration_type", "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.log_level", "ERROR"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.metrics_level", "OPERATOR"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.auto_scaling_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism_per_kpu", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "FLINK-1_8"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct2),
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

func TestAccKinesisAnalyticsV2Application_FlinkApplicationEnvironmentProperties_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRole1ResourceName := "aws_iam_role.test.0"
	iamRole2ResourceName := "aws_iam_role.test.1"
	s3BucketResourceName := "aws_s3_bucket.test"
	s3Object1ResourceName := "aws_s3_object.test.0"
	s3Object2ResourceName := "aws_s3_object.test.1"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_flinkConfigurationEnvironmentProperties(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.bucket_arn", s3BucketResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.file_key", s3Object1ResourceName, names.AttrKey),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.object_version", s3Object1ResourceName, "version_id"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "ZIPFILE"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.0.snapshots_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.environment_properties.0.property_group.*", map[string]string{
						"property_group_id": "PROPERTY-GROUP-ID1",
						"property_map.%":    acctest.Ct2,
						"property_map.Key9": "Value1",
						"property_map.Key8": "Value2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.environment_properties.0.property_group.*", map[string]string{
						"property_group_id": "PROPERTY-GROUP-ID2",
						"property_map.%":    acctest.Ct3,
						"property_map.KeyA": "ValueZ",
						"property_map.KeyB": "ValueY",
						"property_map.KeyC": "ValueX",
					}),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpointing_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpoint_interval", "60000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.min_pause_between_checkpoints", "5000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.configuration_type", "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.log_level", "DEBUG"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.metrics_level", "TASK"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.auto_scaling_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.configuration_type", "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism_per_kpu", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "FLINK-1_8"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRole1ResourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct1),
				),
			},
			{
				Config: testAccApplicationConfig_flinkConfigurationEnvironmentPropertiesUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.bucket_arn", s3BucketResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.file_key", s3Object2ResourceName, names.AttrKey),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.object_version", s3Object2ResourceName, "version_id"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "ZIPFILE"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.0.snapshots_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.environment_properties.0.property_group.*", map[string]string{
						"property_group_id": "PROPERTY-GROUP-ID2",
						"property_map.%":    acctest.Ct2,
						"property_map.KeyA": "ValueZ",
						"property_map.KeyC": "ValueW",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.environment_properties.0.property_group.*", map[string]string{
						"property_group_id": "PROPERTY-GROUP-ID3",
						"property_map.%":    acctest.Ct1,
						"property_map.Key":  "Value",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.environment_properties.0.property_group.*", map[string]string{
						"property_group_id":     "PROPERTY-GROUP-ID4",
						"property_map.%":        acctest.Ct1,
						"property_map.KeyAlpha": "ValueOmega",
					}),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpointing_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpoint_interval", "55000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.configuration_type", "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.min_pause_between_checkpoints", "5500"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.configuration_type", "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.log_level", "ERROR"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.metrics_level", "OPERATOR"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.auto_scaling_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism_per_kpu", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "FLINK-1_8"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRole2ResourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "Value3"),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct2),
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

func TestAccKinesisAnalyticsV2Application_FlinkApplication_restoreFromSnapshot(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	s3BucketResourceName := "aws_s3_bucket.test"
	s3ObjectResourceName := "aws_s3_object.test"
	snapshotResourceName := "aws_kinesisanalyticsv2_application_snapshot.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_startSnapshotableFlink(rName, "RESTORE_FROM_LATEST_SNAPSHOT", "", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.bucket_arn", s3BucketResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.file_key", s3ObjectResourceName, names.AttrKey),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.object_version", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "ZIPFILE"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.0.snapshots_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.environment_properties.0.property_group.*", map[string]string{
						"property_group_id":                      "ConsumerConfigProperties",
						"property_map.%":                         acctest.Ct3,
						"property_map.flink.inputstream.initpos": "LATEST",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.environment_properties.0.property_group.*", map[string]string{
						"property_group_id":               "ProducerConfigProperties",
						"property_map.%":                  acctest.Ct3,
						"property_map.AggregationEnabled": acctest.CtFalse,
					}),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpointing_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpoint_interval", "60000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.min_pause_between_checkpoints", "5000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.log_level", "INFO"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.metrics_level", "APPLICATION"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.auto_scaling_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism_per_kpu", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.0.application_restore_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.0.application_restore_configuration.0.application_restore_type", "RESTORE_FROM_LATEST_SNAPSHOT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.0.application_restore_configuration.0.snapshot_name", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.0.flink_run_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.0.flink_run_configuration.0.allow_non_restored_state", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "FLINK-1_11"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "start_application", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct1),
				),
			},
			{
				Config: testAccApplicationConfig_stopSnapshotableFlink(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.bucket_arn", s3BucketResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.file_key", s3ObjectResourceName, names.AttrKey),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.object_version", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "ZIPFILE"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.0.snapshots_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.environment_properties.0.property_group.*", map[string]string{
						"property_group_id":                      "ConsumerConfigProperties",
						"property_map.%":                         acctest.Ct3,
						"property_map.flink.inputstream.initpos": "LATEST",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.environment_properties.0.property_group.*", map[string]string{
						"property_group_id":               "ProducerConfigProperties",
						"property_map.%":                  acctest.Ct3,
						"property_map.AggregationEnabled": acctest.CtFalse,
					}),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpointing_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpoint_interval", "60000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.min_pause_between_checkpoints", "5000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.log_level", "INFO"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.metrics_level", "APPLICATION"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.auto_scaling_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism_per_kpu", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "force_stop", acctest.CtFalse),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "FLINK-1_11"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "start_application", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_stop", "start_application"},
			},
			{
				Config: testAccApplicationConfig_startSnapshotableFlink(rName, "RESTORE_FROM_CUSTOM_SNAPSHOT", rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.bucket_arn", s3BucketResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.file_key", s3ObjectResourceName, names.AttrKey),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.object_version", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "ZIPFILE"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.0.snapshots_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.environment_properties.0.property_group.*", map[string]string{
						"property_group_id":                      "ConsumerConfigProperties",
						"property_map.%":                         acctest.Ct3,
						"property_map.flink.inputstream.initpos": "LATEST",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.environment_properties.0.property_group.*", map[string]string{
						"property_group_id":               "ProducerConfigProperties",
						"property_map.%":                  acctest.Ct3,
						"property_map.AggregationEnabled": acctest.CtFalse,
					}),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpointing_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpoint_interval", "60000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.min_pause_between_checkpoints", "5000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.log_level", "INFO"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.metrics_level", "APPLICATION"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.auto_scaling_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism_per_kpu", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.0.application_restore_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.0.application_restore_configuration.0.application_restore_type", "RESTORE_FROM_CUSTOM_SNAPSHOT"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.run_configuration.0.application_restore_configuration.0.snapshot_name", snapshotResourceName, "snapshot_name"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.0.flink_run_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.0.flink_run_configuration.0.allow_non_restored_state", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "force_stop", acctest.CtFalse),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "FLINK-1_11"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "start_application", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct1),
				),
			},
			{
				Config: testAccApplicationConfig_stopSnapshotableFlink(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.bucket_arn", s3BucketResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.file_key", s3ObjectResourceName, names.AttrKey),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.object_version", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "ZIPFILE"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.0.snapshots_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.environment_properties.0.property_group.*", map[string]string{
						"property_group_id":                      "ConsumerConfigProperties",
						"property_map.%":                         acctest.Ct3,
						"property_map.flink.inputstream.initpos": "LATEST",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.environment_properties.0.property_group.*", map[string]string{
						"property_group_id":               "ProducerConfigProperties",
						"property_map.%":                  acctest.Ct3,
						"property_map.AggregationEnabled": acctest.CtFalse,
					}),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpointing_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpoint_interval", "60000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.min_pause_between_checkpoints", "5000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.log_level", "INFO"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.metrics_level", "APPLICATION"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.auto_scaling_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism_per_kpu", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "force_stop", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "FLINK-1_11"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "start_application", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccKinesisAnalyticsV2Application_FlinkApplicationStartApplication_onCreate(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	s3BucketResourceName := "aws_s3_bucket.test"
	s3ObjectResourceName := "aws_s3_object.test.0"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_flinkConfiguration(rName, acctest.CtTrue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.bucket_arn", s3BucketResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.file_key", s3ObjectResourceName, names.AttrKey),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.object_version", s3ObjectResourceName, "version_id"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "ZIPFILE"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.0.snapshots_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpointing_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpoint_interval", "60000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.min_pause_between_checkpoints", "5000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.configuration_type", "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.log_level", "DEBUG"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.metrics_level", "TASK"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.auto_scaling_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.configuration_type", "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism_per_kpu", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.0.application_restore_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.0.application_restore_configuration.0.application_restore_type", "RESTORE_FROM_LATEST_SNAPSHOT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.0.flink_run_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.0.flink_run_configuration.0.allow_non_restored_state", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "FLINK-1_8"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "start_application", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct1),
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

func TestAccKinesisAnalyticsV2Application_FlinkApplicationStartApplication_onUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	s3BucketResourceName := "aws_s3_bucket.test"
	s3ObjectResourceName := "aws_s3_object.test.0"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_flinkConfiguration(rName, acctest.CtFalse),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.bucket_arn", s3BucketResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.file_key", s3ObjectResourceName, names.AttrKey),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.object_version", s3ObjectResourceName, "version_id"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "ZIPFILE"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.0.snapshots_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpointing_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpoint_interval", "60000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.min_pause_between_checkpoints", "5000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.configuration_type", "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.log_level", "DEBUG"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.metrics_level", "TASK"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.auto_scaling_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.configuration_type", "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism_per_kpu", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "FLINK-1_8"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "start_application", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"start_application"},
			},
			{
				Config: testAccApplicationConfig_flinkConfiguration(rName, acctest.CtTrue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.bucket_arn", s3BucketResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.file_key", s3ObjectResourceName, names.AttrKey),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.object_version", s3ObjectResourceName, "version_id"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "ZIPFILE"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.0.snapshots_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpointing_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpoint_interval", "60000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.min_pause_between_checkpoints", "5000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.configuration_type", "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.log_level", "DEBUG"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.metrics_level", "TASK"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.auto_scaling_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.configuration_type", "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism_per_kpu", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.0.application_restore_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.0.application_restore_configuration.0.application_restore_type", "RESTORE_FROM_LATEST_SNAPSHOT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.0.flink_run_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.0.flink_run_configuration.0.allow_non_restored_state", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "FLINK-1_8"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "start_application", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct1),
				),
			},
			{
				Config: testAccApplicationConfig_flinkConfiguration(rName, acctest.CtFalse),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.bucket_arn", s3BucketResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.file_key", s3ObjectResourceName, names.AttrKey),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.object_version", s3ObjectResourceName, "version_id"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "ZIPFILE"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.0.snapshots_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpointing_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpoint_interval", "60000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.min_pause_between_checkpoints", "5000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.configuration_type", "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.log_level", "DEBUG"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.metrics_level", "TASK"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.auto_scaling_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.configuration_type", "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism_per_kpu", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "FLINK-1_8"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "start_application", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccKinesisAnalyticsV2Application_FlinkApplication_updateRunning(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	s3BucketResourceName := "aws_s3_bucket.test"
	s3Object1ResourceName := "aws_s3_object.test.0"
	s3Object2ResourceName := "aws_s3_object.test.1"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_flinkConfiguration(rName, acctest.CtTrue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.bucket_arn", s3BucketResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.file_key", s3Object1ResourceName, names.AttrKey),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.object_version", s3Object1ResourceName, "version_id"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "ZIPFILE"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.0.snapshots_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpointing_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpoint_interval", "60000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.min_pause_between_checkpoints", "5000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.configuration_type", "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.log_level", "DEBUG"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.metrics_level", "TASK"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.auto_scaling_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.configuration_type", "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism_per_kpu", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.0.application_restore_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.0.application_restore_configuration.0.application_restore_type", "RESTORE_FROM_LATEST_SNAPSHOT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.0.flink_run_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.0.flink_run_configuration.0.allow_non_restored_state", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "FLINK-1_8"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "start_application", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"start_application"},
			},
			{
				Config: testAccApplicationConfig_flinkConfigurationUpdated(rName, acctest.CtTrue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.bucket_arn", s3BucketResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.file_key", s3Object2ResourceName, names.AttrKey),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.object_version", s3Object2ResourceName, "version_id"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "ZIPFILE"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.0.snapshots_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpointing_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpoint_interval", "55000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.configuration_type", "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.min_pause_between_checkpoints", "5500"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.configuration_type", "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.log_level", "ERROR"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.metrics_level", "OPERATOR"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.auto_scaling_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism_per_kpu", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.0.application_restore_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.0.application_restore_configuration.0.application_restore_type", "RESTORE_FROM_LATEST_SNAPSHOT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.0.flink_run_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.0.flink_run_configuration.0.allow_non_restored_state", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "FLINK-1_8"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "start_application", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccKinesisAnalyticsV2Application_ServiceExecutionRole_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRole1ResourceName := "aws_iam_role.test.0"
	iamRole2ResourceName := "aws_iam_role.test.1"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_basicSQLPlusDescription(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Testing"),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRole1ResourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct1),
				),
			},
			{
				Config: testAccApplicationConfig_basicSQLServiceExecutionRoleUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Testing"),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRole2ResourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct2),
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

func TestAccKinesisAnalyticsV2Application_SQLApplicationInput_add(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	firehoseResourceName := "aws_kinesis_firehose_delivery_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_sqlConfigurationNotSpecified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "SELECT 1;\n"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct1),
				),
			},
			{
				Config: testAccApplicationConfig_sqlConfigurationInput(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "SELECT 1;\n"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.in_app_stream_names.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_id"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.json_mapping_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.name_prefix", "NAME_PREFIX_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.0.count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input.0.resource_arn", firehoseResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_streams_input.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_starting_position_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_starting_position_configuration.0.input_starting_position", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.output.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct2),
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

func TestAccKinesisAnalyticsV2Application_SQLApplicationInput_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	firehoseResourceName := "aws_kinesis_firehose_delivery_stream.test"
	streamsResourceName := "aws_kinesis_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_sqlConfigurationInput(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "SELECT 1;\n"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.in_app_stream_names.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_id"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.json_mapping_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.name_prefix", "NAME_PREFIX_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.0.count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input.0.resource_arn", firehoseResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_streams_input.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_starting_position_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_starting_position_configuration.0.input_starting_position", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.output.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct1),
				),
			},
			{
				Config: testAccApplicationConfig_sqlConfigurationInputUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "SELECT 1;\n"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.in_app_stream_names.#", "42"),
					resource.TestCheckResourceAttrSet(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_id"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.mapping", "MAPPING-2"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.name", "COLUMN_2"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.sql_type", "VARCHAR(8)"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.1.mapping", "MAPPING-3"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.1.name", "COLUMN_3"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.1.sql_type", "DOUBLE"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_encoding", "UTF-8"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.json_mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.json_mapping_parameters.0.record_row_path", "$path.to.record"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.record_format_type", "JSON"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.name_prefix", "NAME_PREFIX_2"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.0.count", "42"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_streams_input.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_streams_input.0.resource_arn", streamsResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_starting_position_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_starting_position_configuration.0.input_starting_position", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.output.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct2),
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

func TestAccKinesisAnalyticsV2Application_SQLApplicationInputProcessing_add(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	lambdaResourceName := "aws_lambda_function.test.0"
	firehoseResourceName := "aws_kinesis_firehose_delivery_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_sqlConfigurationInput(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "SELECT 1;\n"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.in_app_stream_names.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_id"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.json_mapping_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.name_prefix", "NAME_PREFIX_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.0.count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input.0.resource_arn", firehoseResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_streams_input.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_starting_position_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_starting_position_configuration.0.input_starting_position", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.output.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct1),
				),
			},
			{
				Config: testAccApplicationConfig_sqlConfigurationInputProcessingConfiguration(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "SELECT 1;\n"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.in_app_stream_names.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_id"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.json_mapping_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.name_prefix", "NAME_PREFIX_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.0.count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration.0.input_lambda_processor.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration.0.input_lambda_processor.0.resource_arn", lambdaResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input.0.resource_arn", firehoseResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_streams_input.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_starting_position_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_starting_position_configuration.0.input_starting_position", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.output.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct3), // Add input processing configuration + update input.
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

func TestAccKinesisAnalyticsV2Application_SQLApplicationInputProcessing_delete(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	lambdaResourceName := "aws_lambda_function.test.0"
	firehoseResourceName := "aws_kinesis_firehose_delivery_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_sqlConfigurationInputProcessingConfiguration(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "SELECT 1;\n"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.in_app_stream_names.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_id"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.json_mapping_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.name_prefix", "NAME_PREFIX_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.0.count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration.0.input_lambda_processor.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration.0.input_lambda_processor.0.resource_arn", lambdaResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input.0.resource_arn", firehoseResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_streams_input.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_starting_position_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_starting_position_configuration.0.input_starting_position", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.output.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct1),
				),
			},
			{
				Config: testAccApplicationConfig_sqlConfigurationInput(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "SELECT 1;\n"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.in_app_stream_names.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_id"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.json_mapping_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.name_prefix", "NAME_PREFIX_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.0.count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input.0.resource_arn", firehoseResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_streams_input.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_starting_position_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_starting_position_configuration.0.input_starting_position", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.output.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct3), // Delete input processing configuration + update input.
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

func TestAccKinesisAnalyticsV2Application_SQLApplicationInputProcessing_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	lambda1ResourceName := "aws_lambda_function.test.0"
	lambda2ResourceName := "aws_lambda_function.test.1"
	firehoseResourceName := "aws_kinesis_firehose_delivery_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_sqlConfigurationInputProcessingConfiguration(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "SELECT 1;\n"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.in_app_stream_names.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_id"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.json_mapping_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.name_prefix", "NAME_PREFIX_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.0.count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration.0.input_lambda_processor.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration.0.input_lambda_processor.0.resource_arn", lambda1ResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input.0.resource_arn", firehoseResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_streams_input.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_starting_position_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_starting_position_configuration.0.input_starting_position", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.output.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct1),
				),
			},
			{
				Config: testAccApplicationConfig_sqlConfigurationInputProcessingConfiguration(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "SELECT 1;\n"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.in_app_stream_names.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_id"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.json_mapping_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.name_prefix", "NAME_PREFIX_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.0.count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration.0.input_lambda_processor.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration.0.input_lambda_processor.0.resource_arn", lambda2ResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input.0.resource_arn", firehoseResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_streams_input.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_starting_position_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_starting_position_configuration.0.input_starting_position", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.output.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct2),
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

func TestAccKinesisAnalyticsV2Application_SQLApplicationMultiple_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
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
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_sqlConfigurationMultiple(rName, "", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "SELECT 1;\n"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.in_app_stream_names.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_id"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.json_mapping_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.name_prefix", "NAME_PREFIX_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.0.count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration.0.input_lambda_processor.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration.0.input_lambda_processor.0.resource_arn", lambdaResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input.0.resource_arn", firehoseResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_streams_input.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_starting_position_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_starting_position_configuration.0.input_starting_position", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.output.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.sql_application_configuration.0.output.*", map[string]string{
						names.AttrName:                            "OUTPUT_1",
						"destination_schema.#":                    acctest.Ct1,
						"destination_schema.0.record_format_type": "CSV",
						"kinesis_firehose_output.#":               acctest.Ct1,
						"kinesis_streams_output.#":                acctest.Ct0,
						"lambda_output.#":                         acctest.Ct0,
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.output.*.kinesis_firehose_output.0.resource_arn", firehoseResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_logging_options.0.log_stream_arn", cloudWatchLogStreamResourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRole1ResourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct1),
				),
			},
			{
				Config: testAccApplicationConfig_sqlConfigurationMultipleUpdated(rName, "", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "SELECT 2;\n"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.in_app_stream_names.#", "42"),
					resource.TestCheckResourceAttrSet(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_id"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.mapping", "MAPPING-2"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.name", "COLUMN_2"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.sql_type", "VARCHAR(8)"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.1.mapping", "MAPPING-3"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.1.name", "COLUMN_3"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.1.sql_type", "DOUBLE"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_encoding", "UTF-8"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.json_mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.json_mapping_parameters.0.record_row_path", "$path.to.record"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.record_format_type", "JSON"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.name_prefix", "NAME_PREFIX_2"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.0.count", "42"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_streams_input.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_streams_input.0.resource_arn", streamsResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_starting_position_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_starting_position_configuration.0.input_starting_position", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.output.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.sql_application_configuration.0.output.*", map[string]string{
						names.AttrName:                            "OUTPUT_2",
						"destination_schema.#":                    acctest.Ct1,
						"destination_schema.0.record_format_type": "JSON",
						"kinesis_firehose_output.#":               acctest.Ct0,
						"kinesis_streams_output.#":                acctest.Ct1,
						"lambda_output.#":                         acctest.Ct0,
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.output.*.kinesis_streams_output.0.resource_arn", streamsResourceName, names.AttrARN),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.sql_application_configuration.0.output.*", map[string]string{
						names.AttrName:                            "OUTPUT_3",
						"destination_schema.#":                    acctest.Ct1,
						"destination_schema.0.record_format_type": "CSV",
						"kinesis_firehose_output.#":               acctest.Ct0,
						"kinesis_streams_output.#":                acctest.Ct0,
						"lambda_output.#":                         acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.output.*.lambda_output.0.resource_arn", lambdaResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_column.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_column.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_column.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.0.json_mapping_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.s3_reference_data_source.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.s3_reference_data_source.0.bucket_arn", s3BucketResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.s3_reference_data_source.0.file_key", "KEY-1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.table_name", "TABLE-1"),
					resource.TestCheckResourceAttrSet(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_id"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRole2ResourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "Value3"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "8"), // Delete CloudWatch logging options + add reference data source + delete input processing configuration+ update application + delete output + 2 * add output.
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

func TestAccKinesisAnalyticsV2Application_SQLApplicationOutput_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	lambdaResourceName := "aws_lambda_function.test.0"
	firehoseResourceName := "aws_kinesis_firehose_delivery_stream.test"
	streamsResourceName := "aws_kinesis_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_sqlConfigurationOutput(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "SELECT 1;\n"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.output.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.sql_application_configuration.0.output.*", map[string]string{
						names.AttrName:                            "OUTPUT_1",
						"destination_schema.#":                    acctest.Ct1,
						"destination_schema.0.record_format_type": "CSV",
						"kinesis_firehose_output.#":               acctest.Ct1,
						"kinesis_streams_output.#":                acctest.Ct0,
						"lambda_output.#":                         acctest.Ct0,
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.output.*.kinesis_firehose_output.0.resource_arn", firehoseResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct1),
				),
			},
			{
				Config: testAccApplicationConfig_sqlConfigurationOutputUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "SELECT 1;\n"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.output.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.sql_application_configuration.0.output.*", map[string]string{
						names.AttrName:                            "OUTPUT_2",
						"destination_schema.#":                    acctest.Ct1,
						"destination_schema.0.record_format_type": "JSON",
						"kinesis_firehose_output.#":               acctest.Ct0,
						"kinesis_streams_output.#":                acctest.Ct1,
						"lambda_output.#":                         acctest.Ct0,
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.output.*.kinesis_streams_output.0.resource_arn", streamsResourceName, names.AttrARN),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.sql_application_configuration.0.output.*", map[string]string{
						names.AttrName:                            "OUTPUT_3",
						"destination_schema.#":                    acctest.Ct1,
						"destination_schema.0.record_format_type": "CSV",
						"kinesis_firehose_output.#":               acctest.Ct0,
						"kinesis_streams_output.#":                acctest.Ct0,
						"lambda_output.#":                         acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.output.*.lambda_output.0.resource_arn", lambdaResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct4), // 1 * output deletion + 2 * output addition.
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccApplicationConfig_sqlConfigurationNotSpecified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "SELECT 1;\n"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", "6"), // 2 * output deletion.
				),
			},
		},
	})
}

func TestAccKinesisAnalyticsV2Application_SQLApplicationReferenceDataSource_add(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	s3BucketResourceName := "aws_s3_bucket.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_sqlConfigurationNotSpecified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "SELECT 1;\n"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct1),
				),
			},
			{
				Config: testAccApplicationConfig_sqlConfigurationReferenceDataSource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "SELECT 1;\n"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.output.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_column.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_column.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_column.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.0.json_mapping_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.s3_reference_data_source.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.s3_reference_data_source.0.bucket_arn", s3BucketResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.s3_reference_data_source.0.file_key", "KEY-1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.table_name", "TABLE-1"),
					resource.TestCheckResourceAttrSet(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_id"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct2),
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

func TestAccKinesisAnalyticsV2Application_SQLApplicationReferenceDataSource_delete(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	s3BucketResourceName := "aws_s3_bucket.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_sqlConfigurationReferenceDataSource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "SELECT 1;\n"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.output.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_column.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_column.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_column.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.0.json_mapping_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.s3_reference_data_source.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.s3_reference_data_source.0.bucket_arn", s3BucketResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.s3_reference_data_source.0.file_key", "KEY-1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.table_name", "TABLE-1"),
					resource.TestCheckResourceAttrSet(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_id"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct1),
				),
			},
			{
				Config: testAccApplicationConfig_sqlConfigurationNotSpecified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "SELECT 1;\n"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct2),
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

func TestAccKinesisAnalyticsV2Application_SQLApplicationReferenceDataSource_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	s3BucketResourceName := "aws_s3_bucket.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_sqlConfigurationReferenceDataSource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "SELECT 1;\n"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.output.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_column.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_column.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_column.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.0.json_mapping_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.s3_reference_data_source.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.s3_reference_data_source.0.bucket_arn", s3BucketResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.s3_reference_data_source.0.file_key", "KEY-1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.table_name", "TABLE-1"),
					resource.TestCheckResourceAttrSet(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_id"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct1),
				),
			},
			{
				Config: testAccApplicationConfig_sqlConfigurationReferenceDataSourceUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "SELECT 1;\n"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.output.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_column.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_column.0.mapping", "MAPPING-2"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_column.0.name", "COLUMN_2"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_column.0.sql_type", "VARCHAR(8)"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_column.1.mapping", "MAPPING-3"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_column.1.name", "COLUMN_3"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_column.1.sql_type", "DOUBLE"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_encoding", "UTF-8"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.0.json_mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.0.json_mapping_parameters.0.record_row_path", "$path.to.record"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.record_format_type", "JSON"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.s3_reference_data_source.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.s3_reference_data_source.0.bucket_arn", s3BucketResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.s3_reference_data_source.0.file_key", "KEY-2"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.table_name", "TABLE-2"),
					resource.TestCheckResourceAttrSet(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_id"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct2),
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

func TestAccKinesisAnalyticsV2Application_SQLApplicationStartApplication_onCreate(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	firehoseResourceName := "aws_kinesis_firehose_delivery_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_sqlConfigurationStart(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "SELECT 1;\n"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.in_app_stream_names.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_id"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.json_mapping_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.name_prefix", "NAME_PREFIX_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.0.count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input.0.resource_arn", firehoseResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_streams_input.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_starting_position_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_starting_position_configuration.0.input_starting_position", "NOW"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.output.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "start_application", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct1),
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

func TestAccKinesisAnalyticsV2Application_SQLApplicationStartApplication_onUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	firehoseResourceName := "aws_kinesis_firehose_delivery_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_sqlConfigurationStart(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "SELECT 1;\n"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.in_app_stream_names.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_id"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.json_mapping_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.name_prefix", "NAME_PREFIX_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.0.count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input.0.resource_arn", firehoseResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_streams_input.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_starting_position_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_starting_position_configuration.0.input_starting_position", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.output.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "start_application", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"start_application"},
			},
			{
				Config: testAccApplicationConfig_sqlConfigurationStart(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "SELECT 1;\n"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.in_app_stream_names.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_id"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.json_mapping_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.name_prefix", "NAME_PREFIX_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.0.count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input.0.resource_arn", firehoseResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_streams_input.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_starting_position_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_starting_position_configuration.0.input_starting_position", "NOW"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.output.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "start_application", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct2),
				),
			},
			{
				Config: testAccApplicationConfig_sqlConfigurationStart(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "SELECT 1;\n"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.in_app_stream_names.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_id"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.json_mapping_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.name_prefix", "NAME_PREFIX_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.0.count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input.0.resource_arn", firehoseResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_streams_input.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_starting_position_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_starting_position_configuration.0.input_starting_position", "NOW"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.output.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "start_application", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccKinesisAnalyticsV2Application_SQLApplication_updateRunning(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
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
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_sqlConfigurationMultiple(rName, acctest.CtTrue, "LAST_STOPPED_POINT"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "SELECT 1;\n"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.in_app_stream_names.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_id"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.json_mapping_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.name_prefix", "NAME_PREFIX_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.0.count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration.0.input_lambda_processor.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration.0.input_lambda_processor.0.resource_arn", lambdaResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input.0.resource_arn", firehoseResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_streams_input.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_starting_position_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_starting_position_configuration.0.input_starting_position", "LAST_STOPPED_POINT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.output.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.sql_application_configuration.0.output.*", map[string]string{
						names.AttrName:                            "OUTPUT_1",
						"destination_schema.#":                    acctest.Ct1,
						"destination_schema.0.record_format_type": "CSV",
						"kinesis_firehose_output.#":               acctest.Ct1,
						"kinesis_streams_output.#":                acctest.Ct0,
						"lambda_output.#":                         acctest.Ct0,
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.output.*.kinesis_firehose_output.0.resource_arn", firehoseResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_logging_options.0.log_stream_arn", cloudWatchLogStreamResourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRole1ResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "start_application", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct1),
				),
			},
			{
				Config: testAccApplicationConfig_sqlConfigurationMultipleUpdated(rName, acctest.CtTrue, "LAST_STOPPED_POINT"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "SELECT 2;\n"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.in_app_stream_names.#", "42"),
					resource.TestCheckResourceAttrSet(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_id"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.mapping", "MAPPING-2"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.name", "COLUMN_2"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.sql_type", "VARCHAR(8)"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.1.mapping", "MAPPING-3"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.1.name", "COLUMN_3"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.1.sql_type", "DOUBLE"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_encoding", "UTF-8"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.json_mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.json_mapping_parameters.0.record_row_path", "$path.to.record"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.record_format_type", "JSON"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.name_prefix", "NAME_PREFIX_2"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.0.count", "42"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_streams_input.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_streams_input.0.resource_arn", streamsResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_starting_position_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_starting_position_configuration.0.input_starting_position", "LAST_STOPPED_POINT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.output.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.sql_application_configuration.0.output.*", map[string]string{
						names.AttrName:                            "OUTPUT_2",
						"destination_schema.#":                    acctest.Ct1,
						"destination_schema.0.record_format_type": "JSON",
						"kinesis_firehose_output.#":               acctest.Ct0,
						"kinesis_streams_output.#":                acctest.Ct1,
						"lambda_output.#":                         acctest.Ct0,
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.output.*.kinesis_streams_output.0.resource_arn", streamsResourceName, names.AttrARN),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.sql_application_configuration.0.output.*", map[string]string{
						names.AttrName:                            "OUTPUT_3",
						"destination_schema.#":                    acctest.Ct1,
						"destination_schema.0.record_format_type": "CSV",
						"kinesis_firehose_output.#":               acctest.Ct0,
						"kinesis_streams_output.#":                acctest.Ct0,
						"lambda_output.#":                         acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.output.*.lambda_output.0.resource_arn", lambdaResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_column.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_column.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_column.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.0.json_mapping_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.s3_reference_data_source.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.s3_reference_data_source.0.bucket_arn", s3BucketResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.s3_reference_data_source.0.file_key", "KEY-1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.table_name", "TABLE-1"),
					resource.TestCheckResourceAttrSet(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_id"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRole2ResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "start_application", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "Value3"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "8"), // Delete CloudWatch logging options + add reference data source + delete input processing configuration+ update application + delete output + 2 * add output.
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

func TestAccKinesisAnalyticsV2Application_SQLApplicationVPC_add(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	s3BucketResourceName := "aws_s3_bucket.test"
	s3ObjectResourceName := "aws_s3_object.test.0"
	vpcResourceName := "aws_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_vpcConfigurationNotSpecified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.bucket_arn", s3BucketResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.file_key", s3ObjectResourceName, names.AttrKey),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.object_version", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "ZIPFILE"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.0.snapshots_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpointing_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpoint_interval", "60000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.min_pause_between_checkpoints", "5000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.log_level", "INFO"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.metrics_level", "APPLICATION"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.auto_scaling_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism_per_kpu", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "FLINK-1_8"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct1),
				),
			},
			{
				Config: testAccApplicationConfig_vpcConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.bucket_arn", s3BucketResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.file_key", s3ObjectResourceName, names.AttrKey),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.object_version", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "ZIPFILE"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.0.snapshots_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpointing_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpoint_interval", "60000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.min_pause_between_checkpoints", "5000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.log_level", "INFO"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.metrics_level", "APPLICATION"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.auto_scaling_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism_per_kpu", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.0.security_group_ids.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.0.subnet_ids.#", acctest.Ct2),
					resource.TestCheckResourceAttrSet(resourceName, "application_configuration.0.vpc_configuration.0.vpc_configuration_id"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.vpc_configuration.0.vpc_id", vpcResourceName, names.AttrID),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "FLINK-1_8"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct2),
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

func TestAccKinesisAnalyticsV2Application_SQLApplicationVPC_delete(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	s3BucketResourceName := "aws_s3_bucket.test"
	s3ObjectResourceName := "aws_s3_object.test.0"
	vpcResourceName := "aws_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_vpcConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.bucket_arn", s3BucketResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.file_key", s3ObjectResourceName, names.AttrKey),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.object_version", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "ZIPFILE"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.0.snapshots_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpointing_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpoint_interval", "60000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.min_pause_between_checkpoints", "5000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.log_level", "INFO"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.metrics_level", "APPLICATION"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.auto_scaling_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism_per_kpu", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.0.security_group_ids.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.0.subnet_ids.#", acctest.Ct2),
					resource.TestCheckResourceAttrSet(resourceName, "application_configuration.0.vpc_configuration.0.vpc_configuration_id"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.vpc_configuration.0.vpc_id", vpcResourceName, names.AttrID),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "FLINK-1_8"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct1),
				),
			},
			{
				Config: testAccApplicationConfig_vpcConfigurationNotSpecified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.bucket_arn", s3BucketResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.file_key", s3ObjectResourceName, names.AttrKey),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.object_version", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "ZIPFILE"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.0.snapshots_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpointing_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpoint_interval", "60000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.min_pause_between_checkpoints", "5000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.log_level", "INFO"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.metrics_level", "APPLICATION"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.auto_scaling_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism_per_kpu", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "FLINK-1_8"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct2),
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

func TestAccKinesisAnalyticsV2Application_SQLApplicationVPC_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	s3BucketResourceName := "aws_s3_bucket.test"
	s3ObjectResourceName := "aws_s3_object.test.0"
	vpcResourceName := "aws_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_vpcConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.bucket_arn", s3BucketResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.file_key", s3ObjectResourceName, names.AttrKey),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.object_version", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "ZIPFILE"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.0.snapshots_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpointing_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpoint_interval", "60000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.min_pause_between_checkpoints", "5000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.log_level", "INFO"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.metrics_level", "APPLICATION"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.auto_scaling_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism_per_kpu", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.0.security_group_ids.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.0.subnet_ids.#", acctest.Ct2),
					resource.TestCheckResourceAttrSet(resourceName, "application_configuration.0.vpc_configuration.0.vpc_configuration_id"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.vpc_configuration.0.vpc_id", vpcResourceName, names.AttrID),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "FLINK-1_8"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct1),
				),
			},
			{
				Config: testAccApplicationConfig_vpcConfigurationUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.bucket_arn", s3BucketResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.file_key", s3ObjectResourceName, names.AttrKey),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.object_version", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "ZIPFILE"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.0.snapshots_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpointing_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpoint_interval", "60000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.min_pause_between_checkpoints", "5000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.log_level", "INFO"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.metrics_level", "APPLICATION"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.auto_scaling_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism_per_kpu", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.0.security_group_ids.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.0.subnet_ids.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "application_configuration.0.vpc_configuration.0.vpc_configuration_id"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.vpc_configuration.0.vpc_id", vpcResourceName, names.AttrID),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "FLINK-1_8"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "start_application"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct2),
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

func TestAccKinesisAnalyticsV2Application_RunConfiguration_Update(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	s3BucketResourceName := "aws_s3_bucket.test"
	s3ObjectResourceName := "aws_s3_object.test"
	snapshotResourceName := "aws_kinesisanalyticsv2_application_snapshot.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_startSnapshotableFlink(rName, "RESTORE_FROM_LATEST_SNAPSHOT", "", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.bucket_arn", s3BucketResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.file_key", s3ObjectResourceName, names.AttrKey),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.object_version", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "ZIPFILE"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.0.snapshots_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.environment_properties.0.property_group.*", map[string]string{
						"property_group_id":                      "ConsumerConfigProperties",
						"property_map.%":                         acctest.Ct3,
						"property_map.flink.inputstream.initpos": "LATEST",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.environment_properties.0.property_group.*", map[string]string{
						"property_group_id":               "ProducerConfigProperties",
						"property_map.%":                  acctest.Ct3,
						"property_map.AggregationEnabled": acctest.CtFalse,
					}),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpointing_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpoint_interval", "60000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.min_pause_between_checkpoints", "5000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.log_level", "INFO"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.metrics_level", "APPLICATION"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.auto_scaling_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism_per_kpu", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.0.application_restore_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.0.application_restore_configuration.0.application_restore_type", "RESTORE_FROM_LATEST_SNAPSHOT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.0.flink_run_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.0.flink_run_configuration.0.allow_non_restored_state", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "FLINK-1_11"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "start_application", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct1),
				),
			},
			{
				Config: testAccApplicationConfig_startSnapshotableFlink(rName, "RESTORE_FROM_CUSTOM_SNAPSHOT", rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.bucket_arn", s3BucketResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.file_key", s3ObjectResourceName, names.AttrKey),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.object_version", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "ZIPFILE"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.0.snapshots_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.environment_properties.0.property_group.*", map[string]string{
						"property_group_id":                      "ConsumerConfigProperties",
						"property_map.%":                         acctest.Ct3,
						"property_map.flink.inputstream.initpos": "LATEST",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.environment_properties.0.property_group.*", map[string]string{
						"property_group_id":               "ProducerConfigProperties",
						"property_map.%":                  acctest.Ct3,
						"property_map.AggregationEnabled": acctest.CtFalse,
					}),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpointing_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpoint_interval", "60000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.min_pause_between_checkpoints", "5000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.log_level", "INFO"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.metrics_level", "APPLICATION"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.auto_scaling_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism_per_kpu", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.0.application_restore_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.0.application_restore_configuration.0.application_restore_type", "RESTORE_FROM_CUSTOM_SNAPSHOT"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.run_configuration.0.application_restore_configuration.0.snapshot_name", snapshotResourceName, "snapshot_name"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.0.flink_run_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.run_configuration.0.flink_run_configuration.0.allow_non_restored_state", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "force_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "FLINK-1_11"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "start_application", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "version_id", acctest.Ct2),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_stop", "start_application"},
			},
		},
	})
}

func testAccCheckApplicationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).KinesisAnalyticsV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_kinesisanalyticsv2_application" {
				continue
			}

			_, err := tfkinesisanalyticsv2.FindApplicationDetailByName(ctx, conn, rs.Primary.Attributes[names.AttrName])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Kinesis Analytics v2 Application %s still exists", rs.Primary.ID)
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

		conn := acctest.Provider.Meta().(*conns.AWSClient).KinesisAnalyticsV2Client(ctx)

		output, err := tfkinesisanalyticsv2.FindApplicationDetailByName(ctx, conn, rs.Primary.Attributes[names.AttrName])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).KinesisAnalyticsV2Client(ctx)

	input := &kinesisanalyticsv2.ListApplicationsInput{}

	_, err := conn.ListApplications(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccApplicationConfig_baseServiceExecutionIAMRole(rName string) string {
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
      "Action": ["s3:*"],
      "Resource": ["*"]
    },
    {
      "Effect": "Allow",
      "Action": ["kinesis:*"],
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

func testAccApplicationConfigBaseFlinkApplication(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object" "test" {
  count = 2

  bucket = aws_s3_bucket.test.bucket
  key    = "%[1]s.${count.index}"
  source = "test-fixtures/flink-app.jar"
}
`, rName)
}

func testAccApplicationConfigBaseSQLApplication(rName string) string {
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

func testAccApplicationConfigBaseVPC(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 2

  vpc_id            = aws_vpc.test.id
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 2, count.index)
  availability_zone = data.aws_availability_zones.available.names[count.index]

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  count = 2

  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccApplicationConfig_basicFlink(rName, runtimeEnvironment string) string {
	return acctest.ConfigCompose(
		testAccApplicationConfig_baseServiceExecutionIAMRole(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = %[2]q
  service_execution_role = aws_iam_role.test[0].arn
}
`, rName, runtimeEnvironment))
}

func testAccApplicationConfig_basicSQL(rName string) string {
	return acctest.ConfigCompose(
		testAccApplicationConfig_baseServiceExecutionIAMRole(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "SQL-1_0"
  service_execution_role = aws_iam_role.test[0].arn
}
`, rName))
}

func testAccApplicationConfig_applicationMode(rName string) string {
	return acctest.ConfigCompose(
		testAccApplicationConfig_baseServiceExecutionIAMRole(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  application_mode       = "STREAMING"
  runtime_environment    = "FLINK-1_13"
  service_execution_role = aws_iam_role.test[0].arn
}
`, rName))
}

func testAccApplicationConfig_basicSQLPlusDescription(rName string) string {
	return acctest.ConfigCompose(
		testAccApplicationConfig_baseServiceExecutionIAMRole(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "SQL-1_0"
  service_execution_role = aws_iam_role.test[0].arn
  description            = "Testing"
}
`, rName))
}

func testAccApplicationConfig_basicSQLServiceExecutionRoleUpdated(rName string) string {
	return acctest.ConfigCompose(
		testAccApplicationConfig_baseServiceExecutionIAMRole(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "SQL-1_0"
  service_execution_role = aws_iam_role.test[1].arn
  description            = "Testing"
}
`, rName))
}

func testAccApplicationConfig_applicationCodeConfiguration(rName, textContent string) string {
	return acctest.ConfigCompose(
		testAccApplicationConfig_baseServiceExecutionIAMRole(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "SQL-1_0"
  service_execution_role = aws_iam_role.test[0].arn

  application_configuration {
    application_code_configuration {
      code_content {
        text_content = %[2]q
      }

      code_content_type = "PLAINTEXT"
    }
  }
}
`, rName, textContent))
}

func testAccApplicationConfig_cloudWatchLoggingOptions(rName string, streamIndex int) string {
	return acctest.ConfigCompose(
		testAccApplicationConfig_baseServiceExecutionIAMRole(rName),
		fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_log_stream" "test" {
  count = 2

  name           = "%[1]s.${count.index}"
  log_group_name = aws_cloudwatch_log_group.test.name
}

resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "SQL-1_0"
  service_execution_role = aws_iam_role.test[0].arn

  cloudwatch_logging_options {
    log_stream_arn = aws_cloudwatch_log_stream.test.%[2]d.arn
  }
}
`, rName, streamIndex))
}

func testAccApplicationConfig_environmentProperties(rName string) string {
	return acctest.ConfigCompose(
		testAccApplicationConfig_baseServiceExecutionIAMRole(rName),
		testAccApplicationConfigBaseFlinkApplication(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "FLINK-1_8"
  service_execution_role = aws_iam_role.test[0].arn

  application_configuration {
    application_code_configuration {
      code_content {
        s3_content_location {
          bucket_arn = aws_s3_bucket.test.arn
          file_key   = aws_s3_object.test[0].key
        }
      }

      code_content_type = "ZIPFILE"
    }

    environment_properties {
      property_group {
        property_group_id = "PROPERTY-GROUP-ID1"

        property_map = {
          Key9 = "Value1"
          Key8 = "Value2"
        }
      }

      property_group {
        property_group_id = "PROPERTY-GROUP-ID2"

        property_map = {
          KeyA = "ValueZ"
          KeyB = "ValueY"
          KeyC = "ValueX"
        }
      }
    }
  }
}
`, rName))
}

func testAccApplicationConfig_environmentPropertiesUpdated(rName string) string {
	return acctest.ConfigCompose(
		testAccApplicationConfig_baseServiceExecutionIAMRole(rName),
		testAccApplicationConfigBaseFlinkApplication(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "FLINK-1_8"
  service_execution_role = aws_iam_role.test[0].arn

  application_configuration {
    application_code_configuration {
      code_content {
        s3_content_location {
          bucket_arn = aws_s3_bucket.test.arn
          file_key   = aws_s3_object.test[0].key
        }
      }

      code_content_type = "ZIPFILE"
    }

    environment_properties {
      property_group {
        property_group_id = "PROPERTY-GROUP-ID3"

        property_map = {
          Key = "Value"
        }
      }

      property_group {
        property_group_id = "PROPERTY-GROUP-ID2"

        property_map = {
          KeyA = "ValueZ"
          KeyC = "ValueW"
        }
      }

      property_group {
        property_group_id = "PROPERTY-GROUP-ID4"

        property_map = {
          KeyAlpha = "ValueOmega"
        }
      }
    }
  }
}
`, rName))
}

func testAccApplicationConfig_environmentPropertiesNotSpecified(rName string) string {
	return acctest.ConfigCompose(
		testAccApplicationConfig_baseServiceExecutionIAMRole(rName),
		testAccApplicationConfigBaseFlinkApplication(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "FLINK-1_8"
  service_execution_role = aws_iam_role.test[0].arn

  application_configuration {
    application_code_configuration {
      code_content {
        s3_content_location {
          bucket_arn = aws_s3_bucket.test.arn
          file_key   = aws_s3_object.test[0].key
        }
      }

      code_content_type = "ZIPFILE"
    }
  }
}
`, rName))
}

func testAccApplicationConfig_flinkConfiguration(rName, startApplication string) string {
	if startApplication == "" {
		startApplication = "null"
	}

	return acctest.ConfigCompose(
		testAccApplicationConfig_baseServiceExecutionIAMRole(rName),
		testAccApplicationConfigBaseFlinkApplication(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "FLINK-1_8"
  service_execution_role = aws_iam_role.test[0].arn

  application_configuration {
    application_code_configuration {
      code_content {
        s3_content_location {
          bucket_arn     = aws_s3_bucket.test.arn
          file_key       = aws_s3_object.test[0].key
          object_version = aws_s3_object.test[0].version_id
        }
      }

      code_content_type = "ZIPFILE"
    }

    application_snapshot_configuration {
      snapshots_enabled = false
    }

    flink_application_configuration {
      checkpoint_configuration {
        configuration_type = "DEFAULT"
      }

      monitoring_configuration {
        configuration_type = "CUSTOM"
        log_level          = "DEBUG"
        metrics_level      = "TASK"
      }

      parallelism_configuration {
        auto_scaling_enabled = true
        configuration_type   = "CUSTOM"
        parallelism          = 10
        parallelism_per_kpu  = 4
      }
    }
  }

  start_application = %[2]s
}
`, rName, startApplication))
}

func testAccApplicationConfig_flinkConfigurationUpdated(rName, startApplication string) string {
	if startApplication == "" {
		startApplication = "null"
	}

	return acctest.ConfigCompose(
		testAccApplicationConfig_baseServiceExecutionIAMRole(rName),
		testAccApplicationConfigBaseFlinkApplication(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "FLINK-1_8"
  service_execution_role = aws_iam_role.test[0].arn

  application_configuration {
    application_code_configuration {
      code_content {
        s3_content_location {
          bucket_arn     = aws_s3_bucket.test.arn
          file_key       = aws_s3_object.test[1].key
          object_version = aws_s3_object.test[1].version_id
        }
      }

      code_content_type = "ZIPFILE"
    }

    application_snapshot_configuration {
      snapshots_enabled = false
    }

    flink_application_configuration {
      checkpoint_configuration {
        checkpointing_enabled         = true
        configuration_type            = "CUSTOM"
        checkpoint_interval           = 55000
        min_pause_between_checkpoints = 5500
      }

      monitoring_configuration {
        configuration_type = "CUSTOM"
        log_level          = "ERROR"
        metrics_level      = "OPERATOR"
      }

      parallelism_configuration {
        configuration_type = "DEFAULT"
      }
    }
  }

  start_application = %[2]s
}
`, rName, startApplication))
}

func testAccApplicationConfig_flinkConfigurationEnvironmentProperties(rName string) string {
	return acctest.ConfigCompose(
		testAccApplicationConfig_baseServiceExecutionIAMRole(rName),
		testAccApplicationConfigBaseFlinkApplication(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "FLINK-1_8"
  service_execution_role = aws_iam_role.test[0].arn

  application_configuration {
    application_code_configuration {
      code_content {
        s3_content_location {
          bucket_arn     = aws_s3_bucket.test.arn
          file_key       = aws_s3_object.test[0].key
          object_version = aws_s3_object.test[0].version_id
        }
      }

      code_content_type = "ZIPFILE"
    }

    application_snapshot_configuration {
      snapshots_enabled = false
    }

    environment_properties {
      property_group {
        property_group_id = "PROPERTY-GROUP-ID1"

        property_map = {
          Key9 = "Value1"
          Key8 = "Value2"
        }
      }

      property_group {
        property_group_id = "PROPERTY-GROUP-ID2"

        property_map = {
          KeyA = "ValueZ"
          KeyB = "ValueY"
          KeyC = "ValueX"
        }
      }
    }

    flink_application_configuration {
      checkpoint_configuration {
        configuration_type = "DEFAULT"
      }

      monitoring_configuration {
        configuration_type = "CUSTOM"
        log_level          = "DEBUG"
        metrics_level      = "TASK"
      }

      parallelism_configuration {
        auto_scaling_enabled = true
        configuration_type   = "CUSTOM"
        parallelism          = 10
        parallelism_per_kpu  = 4
      }
    }
  }

  tags = {
    Key1 = "Value1"
  }
}
`, rName))
}

func testAccApplicationConfig_flinkConfigurationEnvironmentPropertiesUpdated(rName string) string {
	return acctest.ConfigCompose(
		testAccApplicationConfig_baseServiceExecutionIAMRole(rName),
		testAccApplicationConfigBaseFlinkApplication(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "FLINK-1_8"
  service_execution_role = aws_iam_role.test[1].arn

  application_configuration {
    application_code_configuration {
      code_content {
        s3_content_location {
          bucket_arn     = aws_s3_bucket.test.arn
          file_key       = aws_s3_object.test[1].key
          object_version = aws_s3_object.test[1].version_id
        }
      }

      code_content_type = "ZIPFILE"
    }

    application_snapshot_configuration {
      snapshots_enabled = true
    }

    environment_properties {
      property_group {
        property_group_id = "PROPERTY-GROUP-ID3"

        property_map = {
          Key = "Value"
        }
      }

      property_group {
        property_group_id = "PROPERTY-GROUP-ID2"

        property_map = {
          KeyA = "ValueZ"
          KeyC = "ValueW"
        }
      }

      property_group {
        property_group_id = "PROPERTY-GROUP-ID4"

        property_map = {
          KeyAlpha = "ValueOmega"
        }
      }
    }

    flink_application_configuration {
      checkpoint_configuration {
        checkpointing_enabled         = true
        configuration_type            = "CUSTOM"
        checkpoint_interval           = 55000
        min_pause_between_checkpoints = 5500
      }

      monitoring_configuration {
        configuration_type = "CUSTOM"
        log_level          = "ERROR"
        metrics_level      = "OPERATOR"
      }

      parallelism_configuration {
        configuration_type = "DEFAULT"
      }
    }
  }

  tags = {
    Key2 = "Value2"
    Key3 = "Value3"
  }
}
`, rName))
}

func testAccApplicationConfig_startSnapshotableFlink(rName, applicationRestoreType, snapshotName string, allowNonRestoredState bool) string {
	if snapshotName == "" {
		snapshotName = "null"
	} else {
		snapshotName = strconv.Quote(snapshotName)
	}

	return acctest.ConfigCompose(
		testAccApplicationConfig_baseServiceExecutionIAMRole(rName),
		fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.bucket
  key    = "flink-app.jar"
  source = "test-fixtures/flink-app.jar"
}

resource "aws_kinesis_stream" "input" {
  name        = "%[1]s-input"
  shard_count = 1
}

resource "aws_kinesis_stream" "output" {
  name        = "%[1]s-output"
  shard_count = 1
}

resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "FLINK-1_11"
  service_execution_role = aws_iam_role.test[0].arn

  application_configuration {
    application_code_configuration {
      code_content {
        s3_content_location {
          bucket_arn = aws_s3_bucket.test.arn
          file_key   = aws_s3_object.test.key
        }
      }

      code_content_type = "ZIPFILE"
    }

    application_snapshot_configuration {
      snapshots_enabled = true
    }

    environment_properties {
      property_group {
        property_group_id = "ConsumerConfigProperties"

        property_map = {
          "flink.inputstream.initpos" = "LATEST"
          "aws.region"                = data.aws_region.current.name
          "InputStreamName"           = aws_kinesis_stream.input.name
        }
      }

      property_group {
        property_group_id = "ProducerConfigProperties"

        property_map = {
          "aws.region"         = data.aws_region.current.name
          "AggregationEnabled" = "false"
          "OutputStreamName"   = aws_kinesis_stream.output.name
        }
      }
    }

    run_configuration {
      application_restore_configuration {
        application_restore_type = %[2]q
        snapshot_name            = %[3]s
      }
      flink_run_configuration {
        allow_non_restored_state = %[4]t
      }
    }
  }

  start_application = true
}

resource "aws_kinesisanalyticsv2_application_snapshot" "test" {
  application_name = aws_kinesisanalyticsv2_application.test.name
  snapshot_name    = %[1]q
}
`, rName, applicationRestoreType, snapshotName, allowNonRestoredState))
}

func testAccApplicationConfig_stopSnapshotableFlink(rName string, forceStop bool) string {
	return acctest.ConfigCompose(
		testAccApplicationConfig_baseServiceExecutionIAMRole(rName),
		fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.bucket
  key    = "flink-app.jar"
  source = "test-fixtures/flink-app.jar"
}

resource "aws_kinesis_stream" "input" {
  name        = "%[1]s-input"
  shard_count = 1
}

resource "aws_kinesis_stream" "output" {
  name        = "%[1]s-output"
  shard_count = 1
}

resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "FLINK-1_11"
  service_execution_role = aws_iam_role.test[0].arn

  application_configuration {
    application_code_configuration {
      code_content {
        s3_content_location {
          bucket_arn = aws_s3_bucket.test.arn
          file_key   = aws_s3_object.test.key
        }
      }

      code_content_type = "ZIPFILE"
    }

    application_snapshot_configuration {
      snapshots_enabled = true
    }

    environment_properties {
      property_group {
        property_group_id = "ConsumerConfigProperties"

        property_map = {
          "flink.inputstream.initpos" = "LATEST"
          "aws.region"                = data.aws_region.current.name
          "InputStreamName"           = aws_kinesis_stream.input.name
        }
      }

      property_group {
        property_group_id = "ProducerConfigProperties"

        property_map = {
          "aws.region"         = data.aws_region.current.name
          "AggregationEnabled" = "false"
          "OutputStreamName"   = aws_kinesis_stream.output.name
        }
      }
    }
  }

  start_application = false
  force_stop        = %[2]t
}

resource "aws_kinesisanalyticsv2_application_snapshot" "test" {
  application_name = aws_kinesisanalyticsv2_application.test.name
  snapshot_name    = %[1]q
}
`, rName, forceStop))
}

func testAccApplicationConfig_sqlConfigurationNotSpecified(rName string) string {
	return acctest.ConfigCompose(
		testAccApplicationConfig_baseServiceExecutionIAMRole(rName),
		testAccApplicationConfigBaseSQLApplication(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "SQL-1_0"
  service_execution_role = aws_iam_role.test[0].arn

  application_configuration {
    application_code_configuration {
      code_content {
        text_content = "SELECT 1;\n"
      }

      code_content_type = "PLAINTEXT"
    }
  }
}
`, rName))
}

func testAccApplicationConfig_sqlConfigurationInput(rName string) string {
	return acctest.ConfigCompose(
		testAccApplicationConfig_baseServiceExecutionIAMRole(rName),
		testAccApplicationConfigBaseSQLApplication(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "SQL-1_0"
  service_execution_role = aws_iam_role.test[0].arn

  application_configuration {
    application_code_configuration {
      code_content {
        text_content = "SELECT 1;\n"
      }

      code_content_type = "PLAINTEXT"
    }

    sql_application_configuration {
      input {
        name_prefix = "NAME_PREFIX_1"

        input_schema {
          record_column {
            name     = "COLUMN_1"
            sql_type = "INTEGER"
          }

          record_format {
            record_format_type = "CSV"

            mapping_parameters {
              csv_mapping_parameters {
                record_column_delimiter = ","
                record_row_delimiter    = "|"
              }
            }
          }
        }

        kinesis_firehose_input {
          resource_arn = aws_kinesis_firehose_delivery_stream.test.arn
        }
      }
    }
  }
}
`, rName))
}

func testAccApplicationConfig_sqlConfigurationInputUpdated(rName string) string {
	return acctest.ConfigCompose(
		testAccApplicationConfig_baseServiceExecutionIAMRole(rName),
		testAccApplicationConfigBaseSQLApplication(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "SQL-1_0"
  service_execution_role = aws_iam_role.test[0].arn

  application_configuration {
    application_code_configuration {
      code_content {
        text_content = "SELECT 1;\n"
      }

      code_content_type = "PLAINTEXT"
    }

    sql_application_configuration {
      input {
        name_prefix = "NAME_PREFIX_2"

        input_parallelism {
          count = 42
        }

        input_schema {
          record_column {
            name     = "COLUMN_2"
            sql_type = "VARCHAR(8)"
            mapping  = "MAPPING-2"
          }

          record_column {
            name     = "COLUMN_3"
            sql_type = "DOUBLE"
            mapping  = "MAPPING-3"
          }

          record_encoding = "UTF-8"

          record_format {
            record_format_type = "JSON"

            mapping_parameters {
              json_mapping_parameters {
                record_row_path = "$path.to.record"
              }
            }
          }
        }

        kinesis_streams_input {
          resource_arn = aws_kinesis_stream.test.arn
        }
      }
    }
  }
}
`, rName))
}

func testAccApplicationConfig_sqlConfigurationInputProcessingConfiguration(rName string, lambdaIndex int) string {
	return acctest.ConfigCompose(
		testAccApplicationConfig_baseServiceExecutionIAMRole(rName),
		testAccApplicationConfigBaseSQLApplication(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "SQL-1_0"
  service_execution_role = aws_iam_role.test[0].arn

  application_configuration {
    application_code_configuration {
      code_content {
        text_content = "SELECT 1;\n"
      }

      code_content_type = "PLAINTEXT"
    }

    sql_application_configuration {
      input {
        name_prefix = "NAME_PREFIX_1"

        input_schema {
          record_column {
            name     = "COLUMN_1"
            sql_type = "INTEGER"
          }

          record_format {
            record_format_type = "CSV"

            mapping_parameters {
              csv_mapping_parameters {
                record_column_delimiter = ","
                record_row_delimiter    = "|"
              }
            }
          }
        }

        input_processing_configuration {
          input_lambda_processor {
            resource_arn = aws_lambda_function.test.%[2]d.arn
          }
        }

        kinesis_firehose_input {
          resource_arn = aws_kinesis_firehose_delivery_stream.test.arn
        }
      }
    }
  }
}
`, rName, lambdaIndex))
}

func testAccApplicationConfig_sqlConfigurationMultiple(rName, startApplication, startingPosition string) string {
	if startApplication == "" {
		startApplication = "null"
	}
	if startingPosition == "" {
		startingPosition = "null"
	} else {
		startingPosition = strconv.Quote(startingPosition)
	}

	return acctest.ConfigCompose(
		testAccApplicationConfig_baseServiceExecutionIAMRole(rName),
		testAccApplicationConfigBaseSQLApplication(rName),
		fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_log_stream" "test" {
  name           = %[1]q
  log_group_name = aws_cloudwatch_log_group.test.name
}

resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "SQL-1_0"
  service_execution_role = aws_iam_role.test[0].arn

  application_configuration {
    application_code_configuration {
      code_content {
        text_content = "SELECT 1;\n"
      }

      code_content_type = "PLAINTEXT"
    }

    sql_application_configuration {
      input {
        name_prefix = "NAME_PREFIX_1"

        input_schema {
          record_column {
            name     = "COLUMN_1"
            sql_type = "INTEGER"
          }

          record_format {
            record_format_type = "CSV"

            mapping_parameters {
              csv_mapping_parameters {
                record_column_delimiter = ","
                record_row_delimiter    = "|"
              }
            }
          }
        }

        input_processing_configuration {
          input_lambda_processor {
            resource_arn = aws_lambda_function.test[0].arn
          }
        }

        kinesis_firehose_input {
          resource_arn = aws_kinesis_firehose_delivery_stream.test.arn
        }

        input_starting_position_configuration {
          input_starting_position = %[3]s
        }
      }

      output {
        name = "OUTPUT_1"

        destination_schema {
          record_format_type = "CSV"
        }

        kinesis_firehose_output {
          resource_arn = aws_kinesis_firehose_delivery_stream.test.arn
        }
      }
    }
  }

  cloudwatch_logging_options {
    log_stream_arn = aws_cloudwatch_log_stream.test.arn
  }

  tags = {
    Key1 = "Value1"
  }

  start_application = %[2]s
}
`, rName, startApplication, startingPosition))
}

func testAccApplicationConfig_sqlConfigurationMultipleUpdated(rName, startApplication, startingPosition string) string {
	if startApplication == "" {
		startApplication = "null"
	}
	if startingPosition == "" {
		startingPosition = "null"
	} else {
		startingPosition = strconv.Quote(startingPosition)
	}

	return acctest.ConfigCompose(
		testAccApplicationConfig_baseServiceExecutionIAMRole(rName),
		testAccApplicationConfigBaseSQLApplication(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "SQL-1_0"
  service_execution_role = aws_iam_role.test[1].arn

  application_configuration {
    application_code_configuration {
      code_content {
        text_content = "SELECT 2;\n"
      }

      code_content_type = "PLAINTEXT"
    }

    sql_application_configuration {
      input {
        name_prefix = "NAME_PREFIX_2"

        input_parallelism {
          count = 42
        }

        input_schema {
          record_column {
            name     = "COLUMN_2"
            sql_type = "VARCHAR(8)"
            mapping  = "MAPPING-2"
          }

          record_column {
            name     = "COLUMN_3"
            sql_type = "DOUBLE"
            mapping  = "MAPPING-3"
          }

          record_encoding = "UTF-8"

          record_format {
            record_format_type = "JSON"

            mapping_parameters {
              json_mapping_parameters {
                record_row_path = "$path.to.record"
              }
            }
          }
        }

        kinesis_streams_input {
          resource_arn = aws_kinesis_stream.test.arn
        }

        input_starting_position_configuration {
          input_starting_position = %[3]s
        }
      }

      output {
        name = "OUTPUT_2"

        destination_schema {
          record_format_type = "JSON"
        }

        kinesis_streams_output {
          resource_arn = aws_kinesis_stream.test.arn
        }
      }

      output {
        name = "OUTPUT_3"

        destination_schema {
          record_format_type = "CSV"
        }

        lambda_output {
          resource_arn = aws_lambda_function.test[0].arn
        }
      }

      reference_data_source {
        table_name = "TABLE-1"

        reference_schema {
          record_column {
            name     = "COLUMN_1"
            sql_type = "INTEGER"
          }

          record_format {
            record_format_type = "CSV"

            mapping_parameters {
              csv_mapping_parameters {
                record_column_delimiter = ","
                record_row_delimiter    = "|"
              }
            }
          }
        }

        s3_reference_data_source {
          bucket_arn = aws_s3_bucket.test.arn
          file_key   = "KEY-1"
        }
      }
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

func testAccApplicationConfig_sqlConfigurationOutput(rName string) string {
	return acctest.ConfigCompose(
		testAccApplicationConfig_baseServiceExecutionIAMRole(rName),
		testAccApplicationConfigBaseSQLApplication(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "SQL-1_0"
  service_execution_role = aws_iam_role.test[0].arn

  application_configuration {
    application_code_configuration {
      code_content {
        text_content = "SELECT 1;\n"
      }

      code_content_type = "PLAINTEXT"
    }

    sql_application_configuration {
      output {
        name = "OUTPUT_1"

        destination_schema {
          record_format_type = "CSV"
        }

        kinesis_firehose_output {
          resource_arn = aws_kinesis_firehose_delivery_stream.test.arn
        }
      }
    }
  }
}
`, rName))
}

func testAccApplicationConfig_sqlConfigurationOutputUpdated(rName string) string {
	return acctest.ConfigCompose(
		testAccApplicationConfig_baseServiceExecutionIAMRole(rName),
		testAccApplicationConfigBaseSQLApplication(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "SQL-1_0"
  service_execution_role = aws_iam_role.test[0].arn

  application_configuration {
    application_code_configuration {
      code_content {
        text_content = "SELECT 1;\n"
      }

      code_content_type = "PLAINTEXT"
    }

    sql_application_configuration {
      output {
        name = "OUTPUT_2"

        destination_schema {
          record_format_type = "JSON"
        }

        kinesis_streams_output {
          resource_arn = aws_kinesis_stream.test.arn
        }
      }

      output {
        name = "OUTPUT_3"

        destination_schema {
          record_format_type = "CSV"
        }

        lambda_output {
          resource_arn = aws_lambda_function.test[0].arn
        }
      }
    }
  }
}
`, rName))
}

func testAccApplicationConfig_sqlConfigurationReferenceDataSource(rName string) string {
	return acctest.ConfigCompose(
		testAccApplicationConfig_baseServiceExecutionIAMRole(rName),
		testAccApplicationConfigBaseSQLApplication(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "SQL-1_0"
  service_execution_role = aws_iam_role.test[0].arn

  application_configuration {
    application_code_configuration {
      code_content {
        text_content = "SELECT 1;\n"
      }

      code_content_type = "PLAINTEXT"
    }

    sql_application_configuration {
      reference_data_source {
        table_name = "TABLE-1"

        reference_schema {
          record_column {
            name     = "COLUMN_1"
            sql_type = "INTEGER"
          }

          record_format {
            record_format_type = "CSV"

            mapping_parameters {
              csv_mapping_parameters {
                record_column_delimiter = ","
                record_row_delimiter    = "|"
              }
            }
          }
        }

        s3_reference_data_source {
          bucket_arn = aws_s3_bucket.test.arn
          file_key   = "KEY-1"
        }
      }
    }
  }
}
`, rName))
}

func testAccApplicationConfig_sqlConfigurationReferenceDataSourceUpdated(rName string) string {
	return acctest.ConfigCompose(
		testAccApplicationConfig_baseServiceExecutionIAMRole(rName),
		testAccApplicationConfigBaseSQLApplication(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "SQL-1_0"
  service_execution_role = aws_iam_role.test[0].arn

  application_configuration {
    application_code_configuration {
      code_content {
        text_content = "SELECT 1;\n"
      }

      code_content_type = "PLAINTEXT"
    }

    sql_application_configuration {
      reference_data_source {
        table_name = "TABLE-2"

        reference_schema {
          record_column {
            name     = "COLUMN_2"
            sql_type = "VARCHAR(8)"
            mapping  = "MAPPING-2"
          }

          record_column {
            name     = "COLUMN_3"
            sql_type = "DOUBLE"
            mapping  = "MAPPING-3"
          }

          record_encoding = "UTF-8"

          record_format {
            record_format_type = "JSON"

            mapping_parameters {
              json_mapping_parameters {
                record_row_path = "$path.to.record"
              }
            }
          }
        }

        s3_reference_data_source {
          bucket_arn = aws_s3_bucket.test.arn
          file_key   = "KEY-2"
        }
      }
    }
  }
}
`, rName))
}

func testAccApplicationConfig_sqlConfigurationStart(rName string, start bool) string {
	return acctest.ConfigCompose(
		testAccApplicationConfig_baseServiceExecutionIAMRole(rName),
		testAccApplicationConfigBaseSQLApplication(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "SQL-1_0"
  service_execution_role = aws_iam_role.test[0].arn

  application_configuration {
    application_code_configuration {
      code_content {
        text_content = "SELECT 1;\n"
      }

      code_content_type = "PLAINTEXT"
    }

    sql_application_configuration {
      input {
        name_prefix = "NAME_PREFIX_1"

        input_schema {
          record_column {
            name     = "COLUMN_1"
            sql_type = "INTEGER"
          }

          record_format {
            record_format_type = "CSV"

            mapping_parameters {
              csv_mapping_parameters {
                record_column_delimiter = ","
                record_row_delimiter    = "|"
              }
            }
          }
        }

        kinesis_firehose_input {
          resource_arn = aws_kinesis_firehose_delivery_stream.test.arn
        }

        input_starting_position_configuration {
          input_starting_position = (%[2]t ? "NOW" : null)
        }
      }
    }
  }

  start_application = %[2]t
}
`, rName, start))
}

func testAccApplicationConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccApplicationConfig_baseServiceExecutionIAMRole(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "SQL-1_0"
  service_execution_role = aws_iam_role.test[0].arn

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccApplicationConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccApplicationConfig_baseServiceExecutionIAMRole(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "SQL-1_0"
  service_execution_role = aws_iam_role.test[0].arn

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccApplicationConfig_vpcConfiguration(rName string) string {
	return acctest.ConfigCompose(
		testAccApplicationConfig_baseServiceExecutionIAMRole(rName),
		testAccApplicationConfigBaseVPC(rName),
		testAccApplicationConfigBaseFlinkApplication(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "FLINK-1_8"
  service_execution_role = aws_iam_role.test[0].arn

  application_configuration {
    application_code_configuration {
      code_content {
        s3_content_location {
          bucket_arn = aws_s3_bucket.test.arn
          file_key   = aws_s3_object.test[0].key
        }
      }

      code_content_type = "ZIPFILE"
    }

    vpc_configuration {
      security_group_ids = aws_security_group.test[*].id
      subnet_ids         = aws_subnet.test[*].id
    }
  }
}
`, rName))
}

func testAccApplicationConfig_vpcConfigurationUpdated(rName string) string {
	return acctest.ConfigCompose(
		testAccApplicationConfig_baseServiceExecutionIAMRole(rName),
		testAccApplicationConfigBaseVPC(rName),
		testAccApplicationConfigBaseFlinkApplication(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "FLINK-1_8"
  service_execution_role = aws_iam_role.test[0].arn

  application_configuration {
    application_code_configuration {
      code_content {
        s3_content_location {
          bucket_arn = aws_s3_bucket.test.arn
          file_key   = aws_s3_object.test[0].key
        }
      }

      code_content_type = "ZIPFILE"
    }

    vpc_configuration {
      security_group_ids = [aws_security_group.test[0].id]
      subnet_ids         = [aws_subnet.test[0].id]
    }
  }
}
`, rName))
}

func testAccApplicationConfig_vpcConfigurationNotSpecified(rName string) string {
	return acctest.ConfigCompose(
		testAccApplicationConfig_baseServiceExecutionIAMRole(rName),
		testAccApplicationConfigBaseVPC(rName),
		testAccApplicationConfigBaseFlinkApplication(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "FLINK-1_8"
  service_execution_role = aws_iam_role.test[0].arn

  application_configuration {
    application_code_configuration {
      code_content {
        s3_content_location {
          bucket_arn = aws_s3_bucket.test.arn
          file_key   = aws_s3_object.test[0].key
        }
      }

      code_content_type = "ZIPFILE"
    }
  }
}
`, rName))
}
