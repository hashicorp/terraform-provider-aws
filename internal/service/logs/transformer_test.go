// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// Copyright 2025 Twilio Inc.
package logs_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tflogs "github.com/hashicorp/terraform-provider-aws/internal/service/logs"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLogsTransformer_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var transformer cloudwatchlogs.GetTransformerOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_transformer.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransformerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransformerConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransformerExists(ctx, t, resourceName, &transformer),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.parse_json.#", "1"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "log_group_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "log_group_arn",
			},
		},
	})
}

func TestAccLogsTransformer_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var transformer cloudwatchlogs.GetTransformerOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_transformer.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransformerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransformerConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransformerExists(ctx, t, resourceName, &transformer),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tflogs.ResourceTransformer, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLogsTransformer_disappears_logGroup(t *testing.T) {
	ctx := acctest.Context(t)
	var transformer cloudwatchlogs.GetTransformerOutput
	resourceName := "aws_cloudwatch_log_transformer.test"
	logGroupResourceName := "aws_cloudwatch_log_group.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransformerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransformerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransformerExists(ctx, t, resourceName, &transformer),
					acctest.CheckSDKResourceDisappears(ctx, t, tflogs.ResourceGroup(), logGroupResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLogsTransformer_update_transformerConfig(t *testing.T) {
	ctx := acctest.Context(t)
	var transformer cloudwatchlogs.GetTransformerOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_transformer.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransformerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransformerConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransformerExists(ctx, t, resourceName, &transformer),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.parse_json.#", "1"),
				),
			},
			{
				Config: testAccTransformerConfig_parseCloudFront(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransformerExists(ctx, t, resourceName, &transformer),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.parse_cloudfront.#", "1"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "log_group_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "log_group_arn",
			},
		},
	})
}

func TestAccLogsTransformer_update_logGroupIdentifier(t *testing.T) {
	ctx := acctest.Context(t)
	var transformer cloudwatchlogs.GetTransformerOutput
	oldRName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	newRName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_transformer.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransformerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransformerConfig_basic(oldRName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransformerExists(ctx, t, resourceName, &transformer),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.parse_json.#", "1"),
				),
			},
			{
				Config: testAccTransformerConfig_basic(newRName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransformerExists(ctx, t, resourceName, &transformer),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.parse_json.#", "1"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "log_group_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "log_group_arn",
			},
		},
	})
}

func TestAccLogsTransformer_addKeys(t *testing.T) {
	ctx := acctest.Context(t)
	var transformer cloudwatchlogs.GetTransformerOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_transformer.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransformerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransformerConfig_addKeys(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransformerExists(ctx, t, resourceName, &transformer),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.parse_json.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.add_keys.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.add_keys.0.entry.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.add_keys.0.entry.0.key", acctest.CtKey1),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.add_keys.0.entry.0.value", acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.add_keys.0.entry.0.overwrite_if_exists", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.add_keys.0.entry.1.key", acctest.CtKey2),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.add_keys.0.entry.1.value", acctest.CtValue2),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.add_keys.0.entry.1.overwrite_if_exists", acctest.CtTrue),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "log_group_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "log_group_arn",
			},
		},
	})
}

func TestAccLogsTransformer_copyValue(t *testing.T) {
	ctx := acctest.Context(t)
	var transformer cloudwatchlogs.GetTransformerOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_transformer.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransformerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransformerConfig_copyValue(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransformerExists(ctx, t, resourceName, &transformer),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.parse_json.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.copy_value.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.copy_value.0.entry.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.copy_value.0.entry.0.source", "source1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.copy_value.0.entry.0.target", "target1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.copy_value.0.entry.0.overwrite_if_exists", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.copy_value.0.entry.1.source", "source2"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.copy_value.0.entry.1.target", "target2"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.copy_value.0.entry.1.overwrite_if_exists", acctest.CtTrue),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "log_group_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "log_group_arn",
			},
		},
	})
}

func TestAccLogsTransformer_csv(t *testing.T) {
	ctx := acctest.Context(t)
	var transformer cloudwatchlogs.GetTransformerOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_transformer.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransformerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransformerConfig_csv(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransformerExists(ctx, t, resourceName, &transformer),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.csv.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.csv.0.columns.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.csv.0.delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.csv.0.quote_character", "\""),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.csv.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.csv.0.columns.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.csv.0.columns.0", "example1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.csv.0.columns.1", "example2"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.csv.0.delimiter", ";"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.csv.0.quote_character", "'"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.csv.0.source", "source1"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "log_group_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "log_group_arn",
			},
		},
	})
}

func TestAccLogsTransformer_dateTimeConverter(t *testing.T) {
	ctx := acctest.Context(t)
	var transformer cloudwatchlogs.GetTransformerOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_transformer.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransformerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransformerConfig_dateTimeConverter(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransformerExists(ctx, t, resourceName, &transformer),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.parse_json.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.date_time_converter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.date_time_converter.0.match_patterns.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.date_time_converter.0.match_patterns.0", "yyyy-MM-dd'T'HH:mm:ss.SSS'Z'"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.date_time_converter.0.match_patterns.1", "yyyy-MM-dd'T'HH:mm:ss'Z'"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.date_time_converter.0.source", "source1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.date_time_converter.0.target", "target1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.2.date_time_converter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.2.date_time_converter.0.match_patterns.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.2.date_time_converter.0.match_patterns.0", "yyyy-MM-dd'T'HH:mm:ss.SSS'Z'"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.2.date_time_converter.0.match_patterns.1", "yyyy-MM-dd'T'HH:mm:ss'Z'"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.2.date_time_converter.0.source", "source2"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.2.date_time_converter.0.target", "target2"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.2.date_time_converter.0.locale", "en"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.2.date_time_converter.0.source_timezone", "Europe/Paris"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.2.date_time_converter.0.target_format", "yyyy-MM-dd'T'HH:mm:ss.SSS'Z'"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.2.date_time_converter.0.target_timezone", "America/Chicago"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "log_group_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "log_group_arn",
			},
		},
	})
}

func TestAccLogsTransformer_deleteKeys(t *testing.T) {
	ctx := acctest.Context(t)
	var transformer cloudwatchlogs.GetTransformerOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_transformer.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransformerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransformerConfig_deleteKeys(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransformerExists(ctx, t, resourceName, &transformer),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.parse_json.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.delete_keys.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.delete_keys.0.with_keys.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.delete_keys.0.with_keys.0", acctest.CtKey1),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.delete_keys.0.with_keys.1", acctest.CtKey2),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "log_group_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "log_group_arn",
			},
		},
	})
}

func TestAccLogsTransformer_grok(t *testing.T) {
	ctx := acctest.Context(t)
	var transformer cloudwatchlogs.GetTransformerOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_transformer.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransformerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransformerConfig_grok(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransformerExists(ctx, t, resourceName, &transformer),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.grok.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.grok.0.match", "pattern"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "log_group_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "log_group_arn",
			},
		},
	})
}

func TestAccLogsTransformer_grokWithSource(t *testing.T) {
	ctx := acctest.Context(t)
	var transformer cloudwatchlogs.GetTransformerOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_transformer.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransformerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransformerConfig_grokWithSource(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransformerExists(ctx, t, resourceName, &transformer),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.grok.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.grok.0.match", "pattern"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.grok.0.source", "source1"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "log_group_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "log_group_arn",
			},
		},
	})
}

func TestAccLogsTransformer_listToMap(t *testing.T) {
	ctx := acctest.Context(t)
	var transformer cloudwatchlogs.GetTransformerOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_transformer.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransformerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransformerConfig_listToMap(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransformerExists(ctx, t, resourceName, &transformer),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.parse_json.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.list_to_map.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.list_to_map.0.key", acctest.CtKey1),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.list_to_map.0.source", "source1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.list_to_map.0.flatten", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.2.list_to_map.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.2.list_to_map.0.key", acctest.CtKey2),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.2.list_to_map.0.source", "source2"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.2.list_to_map.0.flatten", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.2.list_to_map.0.flattened_element", "first"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.2.list_to_map.0.target", "target1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.2.list_to_map.0.value_key", "valueKey1"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "log_group_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "log_group_arn",
			},
		},
	})
}

func TestAccLogsTransformer_lowerCaseString(t *testing.T) {
	ctx := acctest.Context(t)
	var transformer cloudwatchlogs.GetTransformerOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_transformer.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransformerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransformerConfig_lowerCaseString(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransformerExists(ctx, t, resourceName, &transformer),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.parse_json.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.lower_case_string.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.lower_case_string.0.with_keys.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.lower_case_string.0.with_keys.0", acctest.CtKey1),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.lower_case_string.0.with_keys.1", acctest.CtKey2),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "log_group_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "log_group_arn",
			},
		},
	})
}

func TestAccLogsTransformer_moveKeys(t *testing.T) {
	ctx := acctest.Context(t)
	var transformer cloudwatchlogs.GetTransformerOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_transformer.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransformerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransformerConfig_moveKeys(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransformerExists(ctx, t, resourceName, &transformer),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.parse_json.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.move_keys.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.move_keys.0.entry.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.move_keys.0.entry.0.source", "source1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.move_keys.0.entry.0.target", "target1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.move_keys.0.entry.0.overwrite_if_exists", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.move_keys.0.entry.1.source", "source2"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.move_keys.0.entry.1.target", "target2"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.move_keys.0.entry.1.overwrite_if_exists", acctest.CtTrue),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "log_group_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "log_group_arn",
			},
		},
	})
}

func TestAccLogsTransformer_parseCloudFront(t *testing.T) {
	ctx := acctest.Context(t)
	var transformer cloudwatchlogs.GetTransformerOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_transformer.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransformerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransformerConfig_parseCloudFront(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransformerExists(ctx, t, resourceName, &transformer),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.parse_cloudfront.#", "1"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "log_group_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "log_group_arn",
			},
		},
	})
}

func TestAccLogsTransformer_parseCloudFrontWithSource(t *testing.T) {
	ctx := acctest.Context(t)
	var transformer cloudwatchlogs.GetTransformerOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_transformer.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransformerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransformerConfig_parseCloudFrontWithSource(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransformerExists(ctx, t, resourceName, &transformer),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.parse_cloudfront.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.parse_cloudfront.0.source", "@message"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "log_group_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "log_group_arn",
			},
		},
	})
}

func TestAccLogsTransformer_parseJSON(t *testing.T) {
	ctx := acctest.Context(t)
	var transformer cloudwatchlogs.GetTransformerOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_transformer.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransformerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransformerConfig_parseJSON(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransformerExists(ctx, t, resourceName, &transformer),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.parse_json.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.parse_json.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.parse_json.0.destination", "destination1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.parse_json.0.source", "source1"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "log_group_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "log_group_arn",
			},
		},
	})
}

func TestAccLogsTransformer_parseKeyValue(t *testing.T) {
	ctx := acctest.Context(t)
	var transformer cloudwatchlogs.GetTransformerOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_transformer.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransformerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransformerConfig_parseKeyValue(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransformerExists(ctx, t, resourceName, &transformer),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.parse_key_value.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.parse_key_value.0.overwrite_if_exists", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.parse_key_value.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.parse_key_value.0.destination", "destination1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.parse_key_value.0.field_delimiter", ";"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.parse_key_value.0.key_prefix", "prefix1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.parse_key_value.0.key_value_delimiter", ":"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.parse_key_value.0.non_match_value", "nonMatch1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.parse_key_value.0.overwrite_if_exists", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.parse_key_value.0.source", "source1"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "log_group_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "log_group_arn",
			},
		},
	})
}

func TestAccLogsTransformer_parsePostgres(t *testing.T) {
	ctx := acctest.Context(t)
	var transformer cloudwatchlogs.GetTransformerOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_transformer.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransformerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransformerConfig_parsePostgres(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransformerExists(ctx, t, resourceName, &transformer),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.parse_postgres.#", "1"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "log_group_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "log_group_arn",
			},
		},
	})
}

func TestAccLogsTransformer_parsePostgresWithSource(t *testing.T) {
	ctx := acctest.Context(t)
	var transformer cloudwatchlogs.GetTransformerOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_transformer.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransformerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransformerConfig_parsePostgresWithSource(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransformerExists(ctx, t, resourceName, &transformer),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.parse_postgres.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.parse_postgres.0.source", "@message"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "log_group_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "log_group_arn",
			},
		},
	})
}

func TestAccLogsTransformer_parseRoute53(t *testing.T) {
	ctx := acctest.Context(t)
	var transformer cloudwatchlogs.GetTransformerOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_transformer.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransformerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransformerConfig_parseRoute53(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransformerExists(ctx, t, resourceName, &transformer),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.parse_route53.#", "1"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "log_group_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "log_group_arn",
			},
		},
	})
}

func TestAccLogsTransformer_parseRoute53WithSource(t *testing.T) {
	ctx := acctest.Context(t)
	var transformer cloudwatchlogs.GetTransformerOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_transformer.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransformerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransformerConfig_parseRoute53WithSource(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransformerExists(ctx, t, resourceName, &transformer),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.parse_route53.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.parse_route53.0.source", "@message"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "log_group_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "log_group_arn",
			},
		},
	})
}

func TestAccLogsTransformer_parseToOCSF(t *testing.T) {
	ctx := acctest.Context(t)
	var transformer cloudwatchlogs.GetTransformerOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_transformer.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransformerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransformerConfig_parseToOCSF(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransformerExists(ctx, t, resourceName, &transformer),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.parse_to_ocsf.0.event_source", "CloudTrail"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.parse_to_ocsf.0.ocsf_version", "V1.1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.parse_to_ocsf.#", "1"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "log_group_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "log_group_arn",
			},
		},
	})
}

func TestAccLogsTransformer_parseToOCSFWithSource(t *testing.T) {
	ctx := acctest.Context(t)
	var transformer cloudwatchlogs.GetTransformerOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_transformer.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransformerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransformerConfig_parseToOCSFWithSource(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransformerExists(ctx, t, resourceName, &transformer),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.parse_to_ocsf.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.parse_to_ocsf.0.event_source", "Route53Resolver"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.parse_to_ocsf.0.ocsf_version", "V1.1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.parse_to_ocsf.0.source", "@message"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "log_group_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "log_group_arn",
			},
		},
	})
}

func TestAccLogsTransformer_parseVPC(t *testing.T) {
	ctx := acctest.Context(t)
	var transformer cloudwatchlogs.GetTransformerOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_transformer.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransformerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransformerConfig_parseVPC(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransformerExists(ctx, t, resourceName, &transformer),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.parse_vpc.#", "1"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "log_group_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "log_group_arn",
			},
		},
	})
}

func TestAccLogsTransformer_parseVPCWithSource(t *testing.T) {
	ctx := acctest.Context(t)
	var transformer cloudwatchlogs.GetTransformerOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_transformer.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransformerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransformerConfig_parseVPCWithSource(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransformerExists(ctx, t, resourceName, &transformer),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.parse_vpc.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.parse_vpc.0.source", "@message"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "log_group_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "log_group_arn",
			},
		},
	})
}

func TestAccLogsTransformer_parseWAF(t *testing.T) {
	ctx := acctest.Context(t)
	var transformer cloudwatchlogs.GetTransformerOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_transformer.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransformerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransformerConfig_parseWAF(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransformerExists(ctx, t, resourceName, &transformer),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.parse_waf.#", "1"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "log_group_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "log_group_arn",
			},
		},
	})
}

func TestAccLogsTransformer_parseWAFWithSource(t *testing.T) {
	ctx := acctest.Context(t)
	var transformer cloudwatchlogs.GetTransformerOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_transformer.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransformerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransformerConfig_parseWAFWithSource(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransformerExists(ctx, t, resourceName, &transformer),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.parse_waf.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.parse_waf.0.source", "@message"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "log_group_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "log_group_arn",
			},
		},
	})
}

func TestAccLogsTransformer_renameKeys(t *testing.T) {
	ctx := acctest.Context(t)
	var transformer cloudwatchlogs.GetTransformerOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_transformer.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransformerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransformerConfig_renameKeys(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransformerExists(ctx, t, resourceName, &transformer),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.parse_json.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.rename_keys.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.rename_keys.0.entry.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.rename_keys.0.entry.0.key", acctest.CtKey1),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.rename_keys.0.entry.0.rename_to", "renamedKey1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.rename_keys.0.entry.0.overwrite_if_exists", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.rename_keys.0.entry.1.key", acctest.CtKey2),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.rename_keys.0.entry.1.rename_to", "renamedKey2"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.rename_keys.0.entry.1.overwrite_if_exists", acctest.CtTrue),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "log_group_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "log_group_arn",
			},
		},
	})
}

func TestAccLogsTransformer_splitString(t *testing.T) {
	ctx := acctest.Context(t)
	var transformer cloudwatchlogs.GetTransformerOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_transformer.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransformerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransformerConfig_splitString(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransformerExists(ctx, t, resourceName, &transformer),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.parse_json.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.split_string.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.split_string.0.entry.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.split_string.0.entry.0.delimiter", ":"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.split_string.0.entry.0.source", "source1"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "log_group_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "log_group_arn",
			},
		},
	})
}

func TestAccLogsTransformer_substituteString(t *testing.T) {
	ctx := acctest.Context(t)
	var transformer cloudwatchlogs.GetTransformerOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_transformer.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransformerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransformerConfig_substituteString(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransformerExists(ctx, t, resourceName, &transformer),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.parse_json.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.substitute_string.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.substitute_string.0.entry.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.substitute_string.0.entry.0.from", "from1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.substitute_string.0.entry.0.to", "to1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.substitute_string.0.entry.0.source", "source1"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "log_group_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "log_group_arn",
			},
		},
	})
}

func TestAccLogsTransformer_trimString(t *testing.T) {
	ctx := acctest.Context(t)
	var transformer cloudwatchlogs.GetTransformerOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_transformer.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransformerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransformerConfig_trimString(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransformerExists(ctx, t, resourceName, &transformer),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.parse_json.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.trim_string.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.trim_string.0.with_keys.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.trim_string.0.with_keys.0", acctest.CtKey1),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.trim_string.0.with_keys.1", acctest.CtKey2),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "log_group_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "log_group_arn",
			},
		},
	})
}

func TestAccLogsTransformer_typeConverter(t *testing.T) {
	ctx := acctest.Context(t)
	var transformer cloudwatchlogs.GetTransformerOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_transformer.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransformerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransformerConfig_typeConverter(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransformerExists(ctx, t, resourceName, &transformer),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.parse_json.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.type_converter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.type_converter.0.entry.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.type_converter.0.entry.0.key", acctest.CtKey1),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.type_converter.0.entry.0.type", "boolean"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "log_group_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "log_group_arn",
			},
		},
	})
}

func TestAccLogsTransformer_upperCaseString(t *testing.T) {
	ctx := acctest.Context(t)
	var transformer cloudwatchlogs.GetTransformerOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_transformer.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransformerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransformerConfig_upperCaseString(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransformerExists(ctx, t, resourceName, &transformer),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.parse_json.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.upper_case_string.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.upper_case_string.0.with_keys.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.upper_case_string.0.with_keys.0", acctest.CtKey1),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.upper_case_string.0.with_keys.1", acctest.CtKey2),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "log_group_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "log_group_arn",
			},
		},
	})
}

func testAccCheckTransformerDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).LogsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_logs_transformer" {
				continue
			}

			_, err := tflogs.FindTransformerByLogGroupIdentifier(ctx, conn, rs.Primary.Attributes["log_group_arn"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudWatch Logs Transformer still exists: %s", rs.Primary.Attributes["log_group_arn"])
		}

		return nil
	}
}

func testAccCheckTransformerExists(ctx context.Context, t *testing.T, n string, v *cloudwatchlogs.GetTransformerOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).LogsClient(ctx)

		output, err := tflogs.FindTransformerByLogGroupIdentifier(ctx, conn, rs.Primary.Attributes["log_group_arn"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccTransformerConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_transformer" "test" {
  log_group_arn = aws_cloudwatch_log_group.test.arn

  transformer_config {
    parse_json {}
  }
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}
`, rName)
}

func testAccTransformerConfig_addKeys(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_transformer" "test" {
  log_group_arn = aws_cloudwatch_log_group.test.arn

  transformer_config {
    parse_json {}
  }

  transformer_config {
    add_keys {
      entry {
        key   = "key1"
        value = "value1"
      }
      entry {
        key                 = "key2"
        value               = "value2"
        overwrite_if_exists = true
      }
    }
  }
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}
`, rName)
}

func testAccTransformerConfig_copyValue(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_transformer" "test" {
  log_group_arn = aws_cloudwatch_log_group.test.arn

  transformer_config {
    parse_json {}
  }

  transformer_config {
    copy_value {
      entry {
        source = "source1"
        target = "target1"
      }
      entry {
        source              = "source2"
        target              = "target2"
        overwrite_if_exists = true
      }
    }
  }
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}
`, rName)
}

func testAccTransformerConfig_csv(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_transformer" "test" {
  log_group_arn = aws_cloudwatch_log_group.test.arn

  transformer_config {
    csv {}
  }

  transformer_config {
    csv {
      columns         = ["example1", "example2"]
      delimiter       = ";"
      quote_character = "'"
      source          = "source1"
    }
  }
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}
`, rName)
}

func testAccTransformerConfig_dateTimeConverter(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_transformer" "test" {
  log_group_arn = aws_cloudwatch_log_group.test.arn

  transformer_config {
    parse_json {}
  }

  transformer_config {
    date_time_converter {
      match_patterns = ["yyyy-MM-dd'T'HH:mm:ss.SSS'Z'", "yyyy-MM-dd'T'HH:mm:ss'Z'"]
      source         = "source1"
      target         = "target1"
    }
  }

  transformer_config {
    date_time_converter {
      match_patterns  = ["yyyy-MM-dd'T'HH:mm:ss.SSS'Z'", "yyyy-MM-dd'T'HH:mm:ss'Z'"]
      source          = "source2"
      target          = "target2"
      locale          = "en"
      source_timezone = "Europe/Paris"
      target_format   = "yyyy-MM-dd'T'HH:mm:ss.SSS'Z'"
      target_timezone = "America/Chicago"
    }
  }
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}
`, rName)
}

func testAccTransformerConfig_deleteKeys(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_transformer" "test" {
  log_group_arn = aws_cloudwatch_log_group.test.arn

  transformer_config {
    parse_json {}
  }

  transformer_config {
    delete_keys {
      with_keys = ["key1", "key2"]
    }
  }
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}
`, rName)
}

func testAccTransformerConfig_grok(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_transformer" "test" {
  log_group_arn = aws_cloudwatch_log_group.test.arn

  transformer_config {
    grok {
      match = "pattern"
    }
  }
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}
`, rName)
}

func testAccTransformerConfig_grokWithSource(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_transformer" "test" {
  log_group_arn = aws_cloudwatch_log_group.test.arn

  transformer_config {
    grok {
      match  = "pattern"
      source = "source1"
    }
  }
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}
`, rName)
}

func testAccTransformerConfig_listToMap(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_transformer" "test" {
  log_group_arn = aws_cloudwatch_log_group.test.arn

  transformer_config {
    parse_json {}
  }

  transformer_config {
    list_to_map {
      key    = "key1"
      source = "source1"
    }
  }

  transformer_config {
    list_to_map {
      key               = "key2"
      source            = "source2"
      flatten           = true
      flattened_element = "first"
      target            = "target1"
      value_key         = "valueKey1"
    }
  }
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}
`, rName)
}

func testAccTransformerConfig_lowerCaseString(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_transformer" "test" {
  log_group_arn = aws_cloudwatch_log_group.test.arn

  transformer_config {
    parse_json {}
  }

  transformer_config {
    lower_case_string {
      with_keys = ["key1", "key2"]
    }
  }
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}
`, rName)
}

func testAccTransformerConfig_moveKeys(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_transformer" "test" {
  log_group_arn = aws_cloudwatch_log_group.test.arn

  transformer_config {
    parse_json {}
  }

  transformer_config {
    move_keys {
      entry {
        source = "source1"
        target = "target1"
      }
      entry {
        source              = "source2"
        target              = "target2"
        overwrite_if_exists = true
      }
    }
  }
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}
`, rName)
}

func testAccTransformerConfig_parseCloudFront(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_transformer" "test" {
  log_group_arn = aws_cloudwatch_log_group.test.arn

  transformer_config {
    parse_cloudfront {}
  }
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}
`, rName)
}

func testAccTransformerConfig_parseCloudFrontWithSource(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_transformer" "test" {
  log_group_arn = aws_cloudwatch_log_group.test.arn

  transformer_config {
    parse_cloudfront {
      source = "@message"
    }
  }
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}
`, rName)
}

func testAccTransformerConfig_parseJSON(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_transformer" "test" {
  log_group_arn = aws_cloudwatch_log_group.test.arn

  transformer_config {
    parse_json {}
  }

  transformer_config {
    parse_json {
      destination = "destination1"
      source      = "source1"
    }
  }
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}
`, rName)
}

func testAccTransformerConfig_parseKeyValue(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_transformer" "test" {
  log_group_arn = aws_cloudwatch_log_group.test.arn

  transformer_config {
    parse_key_value {}
  }

  transformer_config {
    parse_key_value {
      destination         = "destination1"
      field_delimiter     = ";"
      key_prefix          = "prefix1"
      key_value_delimiter = ":"
      non_match_value     = "nonMatch1"
      overwrite_if_exists = true
      source              = "source1"
    }
  }
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}
`, rName)
}

func testAccTransformerConfig_parsePostgres(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_transformer" "test" {
  log_group_arn = aws_cloudwatch_log_group.test.arn

  transformer_config {
    parse_postgres {}
  }
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}
`, rName)
}

func testAccTransformerConfig_parsePostgresWithSource(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_transformer" "test" {
  log_group_arn = aws_cloudwatch_log_group.test.arn

  transformer_config {
    parse_postgres {
      source = "@message"
    }
  }
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}
`, rName)
}

func testAccTransformerConfig_parseRoute53(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_transformer" "test" {
  log_group_arn = aws_cloudwatch_log_group.test.arn

  transformer_config {
    parse_route53 {}
  }
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}
`, rName)
}

func testAccTransformerConfig_parseRoute53WithSource(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_transformer" "test" {
  log_group_arn = aws_cloudwatch_log_group.test.arn

  transformer_config {
    parse_route53 {
      source = "@message"
    }
  }
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}
`, rName)
}

func testAccTransformerConfig_parseToOCSF(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_transformer" "test" {
  log_group_arn = aws_cloudwatch_log_group.test.arn

  transformer_config {
    parse_to_ocsf {
      event_source = "CloudTrail"
      ocsf_version = "V1.1"
    }
  }
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}
`, rName)
}

func testAccTransformerConfig_parseToOCSFWithSource(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_transformer" "test" {
  log_group_arn = aws_cloudwatch_log_group.test.arn

  transformer_config {
    parse_to_ocsf {
      event_source = "Route53Resolver"
      ocsf_version = "V1.1"
      source       = "@message"
    }
  }
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}
`, rName)
}

func testAccTransformerConfig_parseVPC(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_transformer" "test" {
  log_group_arn = aws_cloudwatch_log_group.test.arn

  transformer_config {
    parse_vpc {}
  }
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}
`, rName)
}

func testAccTransformerConfig_parseVPCWithSource(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_transformer" "test" {
  log_group_arn = aws_cloudwatch_log_group.test.arn

  transformer_config {
    parse_vpc {
      source = "@message"
    }
  }
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}
`, rName)
}

func testAccTransformerConfig_parseWAF(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_transformer" "test" {
  log_group_arn = aws_cloudwatch_log_group.test.arn

  transformer_config {
    parse_waf {}
  }
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}
`, rName)
}

func testAccTransformerConfig_parseWAFWithSource(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_transformer" "test" {
  log_group_arn = aws_cloudwatch_log_group.test.arn

  transformer_config {
    parse_waf {
      source = "@message"
    }
  }
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}
`, rName)
}

func testAccTransformerConfig_renameKeys(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_transformer" "test" {
  log_group_arn = aws_cloudwatch_log_group.test.arn

  transformer_config {
    parse_json {}
  }

  transformer_config {
    rename_keys {
      entry {
        key       = "key1"
        rename_to = "renamedKey1"
      }
      entry {
        key                 = "key2"
        rename_to           = "renamedKey2"
        overwrite_if_exists = true
      }
    }
  }
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}
`, rName)
}

func testAccTransformerConfig_splitString(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_transformer" "test" {
  log_group_arn = aws_cloudwatch_log_group.test.arn

  transformer_config {
    parse_json {}
  }

  transformer_config {
    split_string {
      entry {
        delimiter = ":"
        source    = "source1"
      }
    }
  }
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}
`, rName)
}

func testAccTransformerConfig_substituteString(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_transformer" "test" {
  log_group_arn = aws_cloudwatch_log_group.test.arn

  transformer_config {
    parse_json {}
  }

  transformer_config {
    substitute_string {
      entry {
        from   = "from1"
        source = "source1"
        to     = "to1"
      }
    }
  }
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}
`, rName)
}

func testAccTransformerConfig_trimString(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_transformer" "test" {
  log_group_arn = aws_cloudwatch_log_group.test.arn

  transformer_config {
    parse_json {}
  }

  transformer_config {
    trim_string {
      with_keys = ["key1", "key2"]
    }
  }
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}
`, rName)
}

func testAccTransformerConfig_typeConverter(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_transformer" "test" {
  log_group_arn = aws_cloudwatch_log_group.test.arn

  transformer_config {
    parse_json {}
  }

  transformer_config {
    type_converter {
      entry {
        key  = "key1"
        type = "boolean"
      }
    }
  }
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}
`, rName)
}

func testAccTransformerConfig_upperCaseString(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_transformer" "test" {
  log_group_arn = aws_cloudwatch_log_group.test.arn

  transformer_config {
    parse_json {}
  }

  transformer_config {
    upper_case_string {
      with_keys = ["key1", "key2"]
    }
  }
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}
`, rName)
}
