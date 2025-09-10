// Copyright (c) HashiCorp, Inc.
// Copyright 2025 Twilio Inc.
// SPDX-License-Identifier: MPL-2.0

package logs_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tflogs "github.com/hashicorp/terraform-provider-aws/internal/service/logs"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// TIP: File Structure. The basic outline for all test files should be as
// follows. Improve this resource's maintainability by following this
// outline.
//
// 1. Package declaration (add "_test" since this is a test file)
// 2. Imports
// 3. Unit tests
// 4. Basic test
// 5. Disappears test
// 6. All the other tests
// 7. Helper functions (exists, destroy, check, etc.)
// 8. Functions that return Terraform configurations

func TestAccLogsTransformer_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var transformer cloudwatchlogs.GetTransformerOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_transformer.test"
	logGroupResourceName := "aws_cloudwatch_log_group.test"

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
					resource.TestCheckResourceAttrPair(resourceName, "log_group_identifier", logGroupResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.parse_json.#", "1"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccTransformerImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "log_group_identifier",
			},
		},
	})
}

func TestAccLogsTransformer_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var transformer cloudwatchlogs.GetTransformerOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_transformer.test"

	resource.ParallelTest(t, resource.TestCase{
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
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tflogs.ResourceTransformer, resourceName),
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
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tflogs.ResourceGroup(), logGroupResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLogsTransformer_addKeys(t *testing.T) {
	ctx := acctest.Context(t)

	var transformer cloudwatchlogs.GetTransformerOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_transformer.test"
	logGroupResourceName := "aws_cloudwatch_log_group.test"

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
					resource.TestCheckResourceAttrPair(resourceName, "log_group_identifier", logGroupResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.parse_json.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.add_keys.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.add_keys.0.entries.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.add_keys.0.entries.0.key", "key1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.add_keys.0.entries.0.value", "value1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.add_keys.0.entries.0.overwrite_if_exists", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.add_keys.0.entries.1.key", "key2"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.add_keys.0.entries.1.value", "value2"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.add_keys.0.entries.1.overwrite_if_exists", acctest.CtTrue),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccTransformerImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "log_group_identifier",
			},
		},
	})
}

func TestAccLogsTransformer_copyValue(t *testing.T) {
	ctx := acctest.Context(t)

	var transformer cloudwatchlogs.GetTransformerOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_transformer.test"
	logGroupResourceName := "aws_cloudwatch_log_group.test"

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
					resource.TestCheckResourceAttrPair(resourceName, "log_group_identifier", logGroupResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.parse_json.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.copy_value.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.copy_value.0.entries.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.copy_value.0.entries.0.source", "source1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.copy_value.0.entries.0.target", "target1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.copy_value.0.entries.0.overwrite_if_exists", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.copy_value.0.entries.1.source", "source2"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.copy_value.0.entries.1.target", "target2"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.copy_value.0.entries.1.overwrite_if_exists", acctest.CtTrue),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccTransformerImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "log_group_identifier",
			},
		},
	})
}

func TestAccLogsTransformer_csv(t *testing.T) {
	ctx := acctest.Context(t)

	var transformer cloudwatchlogs.GetTransformerOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_transformer.test"
	logGroupResourceName := "aws_cloudwatch_log_group.test"

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
					resource.TestCheckResourceAttrPair(resourceName, "log_group_identifier", logGroupResourceName, names.AttrName),
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
				ImportStateIdFunc:                    testAccTransformerImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "log_group_identifier",
			},
		},
	})
}

func TestAccLogsTransformer_dateTimeConverter(t *testing.T) {
	ctx := acctest.Context(t)

	var transformer cloudwatchlogs.GetTransformerOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_transformer.test"
	logGroupResourceName := "aws_cloudwatch_log_group.test"

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
					resource.TestCheckResourceAttrPair(resourceName, "log_group_identifier", logGroupResourceName, names.AttrName),
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
				ImportStateIdFunc:                    testAccTransformerImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "log_group_identifier",
			},
		},
	})
}

func TestAccLogsTransformer_deleteKeys(t *testing.T) {
	ctx := acctest.Context(t)

	var transformer cloudwatchlogs.GetTransformerOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_transformer.test"
	logGroupResourceName := "aws_cloudwatch_log_group.test"

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
					resource.TestCheckResourceAttrPair(resourceName, "log_group_identifier", logGroupResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.parse_json.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.delete_keys.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.delete_keys.0.with_keys.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.delete_keys.0.with_keys.0", "key1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.delete_keys.0.with_keys.1", "key2"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccTransformerImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "log_group_identifier",
			},
		},
	})
}

func TestAccLogsTransformer_grok(t *testing.T) {
	ctx := acctest.Context(t)

	var transformer cloudwatchlogs.GetTransformerOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_transformer.test"
	logGroupResourceName := "aws_cloudwatch_log_group.test"

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
					resource.TestCheckResourceAttrPair(resourceName, "log_group_identifier", logGroupResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.grok.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.grok.0.match", "pattern"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccTransformerImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "log_group_identifier",
			},
		},
	})
}

func TestAccLogsTransformer_grokWithSource(t *testing.T) {
	ctx := acctest.Context(t)

	var transformer cloudwatchlogs.GetTransformerOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_transformer.test"
	logGroupResourceName := "aws_cloudwatch_log_group.test"

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
					resource.TestCheckResourceAttrPair(resourceName, "log_group_identifier", logGroupResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.grok.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.grok.0.match", "pattern"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.grok.0.source", "source1"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccTransformerImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "log_group_identifier",
			},
		},
	})
}

func TestAccLogsTransformer_listToMap(t *testing.T) {
	ctx := acctest.Context(t)

	var transformer cloudwatchlogs.GetTransformerOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_transformer.test"
	logGroupResourceName := "aws_cloudwatch_log_group.test"

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
					resource.TestCheckResourceAttrPair(resourceName, "log_group_identifier", logGroupResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.parse_json.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.list_to_map.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.list_to_map.0.key", "key1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.list_to_map.0.source", "source1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.list_to_map.0.flatten", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.2.list_to_map.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.2.list_to_map.0.key", "key2"),
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
				ImportStateIdFunc:                    testAccTransformerImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "log_group_identifier",
			},
		},
	})
}

func TestAccLogsTransformer_lowerCaseString(t *testing.T) {
	ctx := acctest.Context(t)

	var transformer cloudwatchlogs.GetTransformerOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_transformer.test"
	logGroupResourceName := "aws_cloudwatch_log_group.test"

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
					resource.TestCheckResourceAttrPair(resourceName, "log_group_identifier", logGroupResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.parse_json.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.lower_case_string.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.lower_case_string.0.with_keys.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.lower_case_string.0.with_keys.0", "key1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.lower_case_string.0.with_keys.1", "key2"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccTransformerImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "log_group_identifier",
			},
		},
	})
}

func TestAccLogsTransformer_moveKeys(t *testing.T) {
	ctx := acctest.Context(t)

	var transformer cloudwatchlogs.GetTransformerOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_transformer.test"
	logGroupResourceName := "aws_cloudwatch_log_group.test"

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
					resource.TestCheckResourceAttrPair(resourceName, "log_group_identifier", logGroupResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.parse_json.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.move_keys.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.move_keys.0.entries.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.move_keys.0.entries.0.source", "source1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.move_keys.0.entries.0.target", "target1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.move_keys.0.entries.0.overwrite_if_exists", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.move_keys.0.entries.1.source", "source2"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.move_keys.0.entries.1.target", "target2"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.1.move_keys.0.entries.1.overwrite_if_exists", acctest.CtTrue),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccTransformerImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "log_group_identifier",
			},
		},
	})
}

func testAccTransformerImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.Attributes["log_group_identifier"], nil
	}
}

func testAccCheckTransformerDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).LogsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_logs_transformer" {
				continue
			}

			_, err := tflogs.FindTransformerByLogGroupIdentifier(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.Logs, create.ErrActionCheckingDestroyed, tflogs.ResNameTransformer, rs.Primary.ID, err)
			}

			return create.Error(names.Logs, create.ErrActionCheckingDestroyed, tflogs.ResNameTransformer, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckTransformerExists(ctx context.Context, t *testing.T, name string, transformer *cloudwatchlogs.GetTransformerOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Logs, create.ErrActionCheckingExistence, tflogs.ResNameTransformer, name, errors.New("not found"))
		}

		conn := acctest.ProviderMeta(ctx, t).LogsClient(ctx)

		resp, err := tflogs.FindTransformerByLogGroupIdentifier(ctx, conn, rs.Primary.Attributes["log_group_identifier"])
		if err != nil {
			return create.Error(names.Logs, create.ErrActionCheckingExistence, tflogs.ResNameTransformer, rs.Primary.Attributes["log_group_identifier"], err)
		}

		*transformer = *resp

		return nil
	}
}

func testAccTransformerConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_log_transformer" "test" {
  log_group_identifier = aws_cloudwatch_log_group.test.name

  transformer_config {
    parse_json {}
  }
}
`, rName)
}

func testAccTransformerConfig_addKeys(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_log_transformer" "test" {
  log_group_identifier = aws_cloudwatch_log_group.test.name

  transformer_config {
    parse_json {}
  }
  
  transformer_config {
    add_keys {
		entries {
			key   = "key1"
			value = "value1"
		}
		entries {
			key                 = "key2"
			value               = "value2"
			overwrite_if_exists = true
		}
	}
  }
}
`, rName)
}

func testAccTransformerConfig_copyValue(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_log_transformer" "test" {
  log_group_identifier = aws_cloudwatch_log_group.test.name

  transformer_config {
    parse_json {}
  }
  
  transformer_config {
    copy_value {
		entries {
			source   = "source1"
			target   = "target1"
		}
		entries {
			source                 = "source2"
			target                 = "target2"
			overwrite_if_exists    = true
		}
	}
  }
}
`, rName)
}

func testAccTransformerConfig_csv(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_log_transformer" "test" {
  log_group_identifier = aws_cloudwatch_log_group.test.name

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
`, rName)
}

func testAccTransformerConfig_dateTimeConverter(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_log_transformer" "test" {
  log_group_identifier = aws_cloudwatch_log_group.test.name

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
`, rName)
}

func testAccTransformerConfig_deleteKeys(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_log_transformer" "test" {
  log_group_identifier = aws_cloudwatch_log_group.test.name

  transformer_config {
	parse_json {}
  }

  transformer_config {
    delete_keys {
		with_keys = ["key1", "key2"]
	}
  }
}
`, rName)
}

func testAccTransformerConfig_grok(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_log_transformer" "test" {
  log_group_identifier = aws_cloudwatch_log_group.test.name

  transformer_config {
	grok {
		match = "pattern"
	}
  }
}
`, rName)
}

func testAccTransformerConfig_grokWithSource(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_log_transformer" "test" {
  log_group_identifier = aws_cloudwatch_log_group.test.name

  transformer_config {
	grok {
		match  = "pattern"
		source = "source1"
	}
  }
}
`, rName)
}

func testAccTransformerConfig_listToMap(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}
  
resource "aws_cloudwatch_log_transformer" "test" {
  log_group_identifier = aws_cloudwatch_log_group.test.name

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
`, rName)
}

func testAccTransformerConfig_lowerCaseString(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_log_transformer" "test" {
  log_group_identifier = aws_cloudwatch_log_group.test.name

  transformer_config {
	parse_json {}
  }

  transformer_config {
    lower_case_string {
		with_keys = ["key1", "key2"]
	}
  }
}
`, rName)
}

func testAccTransformerConfig_moveKeys(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_log_transformer" "test" {
  log_group_identifier = aws_cloudwatch_log_group.test.name

  transformer_config {
    parse_json {}
  }
  
  transformer_config {
    move_keys {
		entries {
			source   = "source1"
			target   = "target1"
		}
		entries {
			source                 = "source2"
			target                 = "target2"
			overwrite_if_exists    = true
		}
	}
  }
}
`, rName)
}
