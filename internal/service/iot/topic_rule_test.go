// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiot "github.com/hashicorp/terraform-provider-aws/internal/service/iot"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(names.IoTServiceID, testAccErrorCheckSkip)
}

func testAccErrorCheckSkip(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"not been supported in region",
	)
}

func TestAccIoTTopicRule_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "iot", fmt.Sprintf("rule/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sql", "SELECT * FROM 'topic/test'"),
					resource.TestCheckResourceAttr(resourceName, "sql_version", "2015-10-08"),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct0),
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

func TestAccIoTTopicRule_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfiot.ResourceTopicRule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIoTTopicRule_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
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
				Config: testAccTopicRuleConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccTopicRuleConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccIoTTopicRule_cloudWatchAlarm(t *testing.T) {
	ctx := acctest.Context(t)
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleConfig_cloudWatchAlarm(rName, "myalarm"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cloudwatch_alarm.*", map[string]string{
						"alarm_name":   "myalarm",
						"state_reason": "test",
						"state_value":  "OK",
					}),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Example rule"),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTopicRuleConfig_cloudWatchAlarm(rName, "differentName"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cloudwatch_alarm.*", map[string]string{
						"alarm_name":   "differentName",
						"state_reason": "test",
						"state_value":  "OK",
					}),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Example rule"),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccIoTTopicRule_cloudWatchLogs(t *testing.T) {
	ctx := acctest.Context(t)
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleConfig_cloudWatchLogs(rName, "mylogs1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cloudwatch_logs.*", map[string]string{
						names.AttrLogGroupName: "mylogs1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cloudwatch_logs.*", map[string]string{
						names.AttrLogGroupName: "mylogs2",
					}),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTopicRuleConfig_cloudWatchLogs(rName, "updatedlogs1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cloudwatch_logs.*", map[string]string{
						names.AttrLogGroupName: "updatedlogs1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cloudwatch_logs.*", map[string]string{
						names.AttrLogGroupName: "mylogs2",
					}),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccIoTTopicRule_cloudWatchLogs_batch_mode(t *testing.T) {
	ctx := acctest.Context(t)
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleConfig_cloudWatchLogsBatchMode(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cloudwatch_logs.*", map[string]string{
						"batch_mode":           acctest.CtFalse,
						names.AttrLogGroupName: "mylogs1",
					}),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTopicRuleConfig_cloudWatchLogsBatchMode(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cloudwatch_logs.*", map[string]string{
						"batch_mode":           acctest.CtTrue,
						names.AttrLogGroupName: "mylogs1",
					}),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccIoTTopicRule_cloudWatchMetric(t *testing.T) {
	ctx := acctest.Context(t)
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleConfig_cloudWatchMetric(rName, "TestName"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cloudwatch_metric.*", map[string]string{
						names.AttrMetricName: "TestName",
						"metric_namespace":   "TestNS",
						"metric_unit":        "s",
						"metric_value":       acctest.Ct10,
					}),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct0),
				),
			},
			{
				Config: testAccTopicRuleConfig_cloudWatchMetric(rName, "OtherName"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cloudwatch_metric.*", map[string]string{
						names.AttrMetricName: "OtherName",
						"metric_namespace":   "TestNS",
						"metric_unit":        "s",
						"metric_value":       acctest.Ct10,
					}),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct0),
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

func TestAccIoTTopicRule_dynamoDB(t *testing.T) {
	ctx := acctest.Context(t)
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleConfig_dynamoDB(rName, "tn"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Description1"),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "dynamodb.*", map[string]string{
						"hash_key_field":    "hkf",
						"hash_key_value":    "hkv",
						"payload_field":     "pf",
						names.AttrTableName: "tn",
					}),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTopicRuleConfig_dynamoDBRangeKey(rName, "tn"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Description2"),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "dynamodb.*", map[string]string{
						"hash_key_field":    "hkf",
						"hash_key_value":    "hkv",
						"operation":         "INSERT",
						"payload_field":     "pf",
						"range_key_field":   "rkf",
						"range_key_type":    "STRING",
						"range_key_value":   "rkv",
						names.AttrTableName: "tn",
					}),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccIoTTopicRule_dynamoDBv2(t *testing.T) {
	ctx := acctest.Context(t)
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleConfig_dynamoDBv2(rName, "test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "dynamodbv2.*", map[string]string{
						"put_item.#":            acctest.Ct1,
						"put_item.0.table_name": "test",
					}),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct0),
				),
			},
			{
				Config: testAccTopicRuleConfig_dynamoDBv2(rName, "updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "dynamodbv2.*", map[string]string{
						"put_item.#":            acctest.Ct1,
						"put_item.0.table_name": "updated",
					}),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccIoTTopicRule_elasticSearch(t *testing.T) {
	ctx := acctest.Context(t)
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleConfig_elasticSearch(rName, "myindex"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "elasticsearch.*", map[string]string{
						names.AttrID:   "myIdentifier",
						"index":        "myindex",
						names.AttrType: "mydocument",
					}),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTopicRuleConfig_elasticSearch(rName, "updatedindex"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "elasticsearch.*", map[string]string{
						names.AttrID:   "myIdentifier",
						"index":        "updatedindex",
						names.AttrType: "mydocument",
					}),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccIoTTopicRule_firehose(t *testing.T) {
	ctx := acctest.Context(t)
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleConfig_firehose(rName, "mystream1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct3),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firehose.*", map[string]string{
						"delivery_stream_name": "mystream1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firehose.*", map[string]string{
						"delivery_stream_name": "mystream2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firehose.*", map[string]string{
						"delivery_stream_name": "mystream3",
					}),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTopicRuleConfig_firehose(rName, "updatedstream1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct3),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firehose.*", map[string]string{
						"delivery_stream_name": "updatedstream1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firehose.*", map[string]string{
						"delivery_stream_name": "mystream2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firehose.*", map[string]string{
						"delivery_stream_name": "mystream3",
					}),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccIoTTopicRule_Firehose_separator(t *testing.T) {
	ctx := acctest.Context(t)
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleConfig_firehoseSeparator(rName, "\n"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firehose.*", map[string]string{
						"delivery_stream_name": "mystream",
						"separator":            "\n",
					}),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTopicRuleConfig_firehoseSeparator(rName, ","),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firehose.*", map[string]string{
						"delivery_stream_name": "mystream",
						"separator":            ",",
					}),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccIoTTopicRule_Firehose_batch_mode(t *testing.T) {
	ctx := acctest.Context(t)
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleConfig_firehoseBatchMode(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firehose.*", map[string]string{
						"delivery_stream_name": "mystream",
						"batch_mode":           acctest.CtFalse,
					}),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTopicRuleConfig_firehoseBatchMode(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firehose.*", map[string]string{
						"delivery_stream_name": "mystream",
						"batch_mode":           acctest.CtTrue,
					}),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccIoTTopicRule_http(t *testing.T) {
	ctx := acctest.Context(t)
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleConfig_http(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "http.*", map[string]string{
						"confirmation_url": "",
						"http_header.#":    acctest.Ct0,
						names.AttrURL:      "https://example.com/ingress",
					}),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTopicRuleConfig_httpConfirmationURL(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "http.*", map[string]string{
						"confirmation_url": "https://example.com/",
						"http_header.#":    acctest.Ct0,
						names.AttrURL:      "https://example.com/ingress",
					}),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct0),
				),
			},
			{
				Config: testAccTopicRuleConfig_httpHeaders(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "http.*", map[string]string{
						"confirmation_url":    "",
						"http_header.#":       acctest.Ct2,
						"http_header.0.key":   "X-Header-1",
						"http_header.0.value": "v1",
						"http_header.1.key":   "X-Header-2",
						"http_header.1.value": "v2",
						names.AttrURL:         "https://example.com/ingress",
					}),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct0),
				),
			},
			{
				Config: testAccTopicRuleConfig_httpErrorAction(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.cloudwatch_logs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.http.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.http.0.url", "https://example.com/error-ingress"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.iot_analytics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.kinesis.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.timestream.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "http.*", map[string]string{
						"confirmation_url": "",
						"http_header.#":    acctest.Ct0,
						names.AttrURL:      "https://example.com/ingress",
					}),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccIoTTopicRule_IoT_analytics(t *testing.T) {
	ctx := acctest.Context(t)
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleConfig_analytics(rName, "fakedata"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "iot_analytics.*", map[string]string{
						"channel_name": "fakedata",
					}),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct0),
				),
			},
			{
				Config: testAccTopicRuleConfig_analytics(rName, "differentdata"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "iot_analytics.*", map[string]string{
						"channel_name": "differentdata",
					}),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccIoTTopicRule_IoT_analytics_batch_mode(t *testing.T) {
	ctx := acctest.Context(t)
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleConfig_analytics(rName, "fakedata"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "iot_analytics.*", map[string]string{
						"channel_name": "fakedata",
						"batch_mode":   acctest.CtFalse,
					}),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct0),
				),
			},
			{
				Config: testAccTopicRuleConfig_analyticsBatchMode(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "iot_analytics.*", map[string]string{
						"channel_name": "fakedata",
						"batch_mode":   acctest.CtTrue,
					}),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccIoTTopicRule_IoT_events(t *testing.T) {
	ctx := acctest.Context(t)
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleConfig_events(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "iot_events.*", map[string]string{
						"input_name": "fake_input_name",
						"message_id": "fake_message_id",
					}),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccIoTTopicRule_IoT_events_batch_mode(t *testing.T) {
	ctx := acctest.Context(t)
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleConfig_events(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "iot_events.*", map[string]string{
						"input_name": "fake_input_name",
						"message_id": "fake_message_id",
						"batch_mode": acctest.CtFalse,
					}),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct0),
				),
			},
			{
				Config: testAccTopicRuleConfig_eventsBatchMode(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "iot_events.*", map[string]string{
						"input_name": "fake_input_name",
						"message_id": "fake_message_id",
						"batch_mode": acctest.CtTrue,
					}),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccIoTTopicRule_kafka(t *testing.T) {
	ctx := acctest.Context(t)

	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleConfig_kafka(rName, "fake_topic", "b-1.localhost:9094"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "kafka.*", map[string]string{
						"client_properties.%":                     "8",
						"client_properties.acks":                  acctest.Ct1,
						"client_properties.bootstrap.servers":     "b-1.localhost:9094",
						"client_properties.compression.type":      "none",
						"client_properties.key.serializer":        "org.apache.kafka.common.serialization.StringSerializer",
						"client_properties.security.protocol":     "SSL",
						"client_properties.ssl.keystore.password": names.AttrPassword,
						"client_properties.value.serializer":      "org.apache.kafka.common.serialization.ByteBufferSerializer",
						"topic":                                   "fake_topic",
						"header.#":                                acctest.Ct2,
						"header.0.key":                            "header-1",
						"header.0.value":                          "value-1",
						"header.1.key":                            "header-2",
						"header.1.value":                          "value-2",
					}),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct0),
				),
			},
			{
				Config: testAccTopicRuleConfig_kafka(rName, "different_topic", "b-2.localhost:9094"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "kafka.*", map[string]string{
						"client_properties.%":                     "8",
						"client_properties.acks":                  acctest.Ct1,
						"client_properties.bootstrap.servers":     "b-2.localhost:9094",
						"client_properties.compression.type":      "none",
						"client_properties.key.serializer":        "org.apache.kafka.common.serialization.StringSerializer",
						"client_properties.security.protocol":     "SSL",
						"client_properties.ssl.keystore.password": names.AttrPassword,
						"client_properties.value.serializer":      "org.apache.kafka.common.serialization.ByteBufferSerializer",
						"topic":                                   "different_topic",
						"header.#":                                acctest.Ct2,
						"header.0.key":                            "header-1",
						"header.0.value":                          "value-1",
						"header.1.key":                            "header-2",
						"header.1.value":                          "value-2",
					}),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct0),
				),
			},
			// Validate that updates only to a value inside the schema-less client_properties also works
			{
				Config: testAccTopicRuleConfig_kafka(rName, "different_topic", "b-3.localhost:9094"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "kafka.*", map[string]string{
						"client_properties.%":                     "8",
						"client_properties.acks":                  acctest.Ct1,
						"client_properties.bootstrap.servers":     "b-3.localhost:9094",
						"client_properties.compression.type":      "none",
						"client_properties.key.serializer":        "org.apache.kafka.common.serialization.StringSerializer",
						"client_properties.security.protocol":     "SSL",
						"client_properties.ssl.keystore.password": names.AttrPassword,
						"client_properties.value.serializer":      "org.apache.kafka.common.serialization.ByteBufferSerializer",
						"topic":                                   "different_topic",
						"header.#":                                acctest.Ct2,
						"header.0.key":                            "header-1",
						"header.0.value":                          "value-1",
						"header.1.key":                            "header-2",
						"header.1.value":                          "value-2",
					}),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Delete everything but the IAM Role assumed by the IoT service.
			{
				Config: testAccTopicRuleConfig_destinationRole(rName),
			},
		},
	})
}

func TestAccIoTTopicRule_kinesis(t *testing.T) {
	ctx := acctest.Context(t)
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleConfig_kinesis(rName, "mystream"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "kinesis.*", map[string]string{
						"stream_name": "mystream",
					}),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct0),
				),
			},
			{
				Config: testAccTopicRuleConfig_kinesis(rName, "otherstream"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "kinesis.*", map[string]string{
						"stream_name": "otherstream",
					}),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct0),
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

func TestAccIoTTopicRule_lambda(t *testing.T) {
	ctx := acctest.Context(t)
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleConfig_lambda(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct0),
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

func TestAccIoTTopicRule_republish(t *testing.T) {
	ctx := acctest.Context(t)
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleConfig_republish(rName, "mytopic"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "republish.*", map[string]string{
						"qos":   acctest.Ct0,
						"topic": "mytopic",
					}),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTopicRuleConfig_republish(rName, "othertopic"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "republish.*", map[string]string{
						"qos":   acctest.Ct0,
						"topic": "othertopic",
					}),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccIoTTopicRule_republishWithQos(t *testing.T) {
	ctx := acctest.Context(t)
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleConfig_republishQoS(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "republish.*", map[string]string{
						"qos":   acctest.Ct1,
						"topic": "mytopic",
					}),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct0),
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

func TestAccIoTTopicRule_s3(t *testing.T) {
	ctx := acctest.Context(t)
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleConfig_s3(rName, "mybucket"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "s3.*", map[string]string{
						names.AttrBucketName: "mybucket",
						"canned_acl":         "private",
						names.AttrKey:        "mykey",
					}),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct0),
				),
			},
			{
				Config: testAccTopicRuleConfig_s3(rName, "yourbucket"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "s3.*", map[string]string{
						names.AttrBucketName: "yourbucket",
						"canned_acl":         "private",
						names.AttrKey:        "mykey",
					}),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct0),
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

func TestAccIoTTopicRule_sns(t *testing.T) {
	ctx := acctest.Context(t)
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleConfig_sns(rName, "RAW"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "sns.*", map[string]string{
						"message_format": "RAW",
					}),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct0),
				),
			},
			{
				Config: testAccTopicRuleConfig_sns(rName, "JSON"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "sns.*", map[string]string{
						"message_format": "JSON",
					}),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct0),
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

func TestAccIoTTopicRule_sqs(t *testing.T) {
	ctx := acctest.Context(t)
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleConfig_sqs(rName, "fakedata"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "sqs.*", map[string]string{
						"queue_url":  "fakedata",
						"use_base64": acctest.CtFalse,
					}),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct0),
				),
			},
			{
				Config: testAccTopicRuleConfig_sqs(rName, "yourdata"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "sqs.*", map[string]string{
						"queue_url":  "yourdata",
						"use_base64": acctest.CtFalse,
					}),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct0),
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

func TestAccIoTTopicRule_Step_functions(t *testing.T) {
	ctx := acctest.Context(t)
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleConfig_stepFunctions(rName, "mystatemachine"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "step_functions.*", map[string]string{
						"execution_name_prefix": "myprefix",
						"state_machine_name":    "mystatemachine",
					}),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct0),
				),
			},
			{
				Config: testAccTopicRuleConfig_stepFunctions(rName, "yourstatemachine"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "step_functions.*", map[string]string{
						"execution_name_prefix": "myprefix",
						"state_machine_name":    "yourstatemachine",
					}),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct0),
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

func TestAccIoTTopicRule_Timestream(t *testing.T) {
	ctx := acctest.Context(t)
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleConfig_timestream(rName, "dim1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "timestream.*", map[string]string{
						names.AttrDatabaseName: "TestDB",
						"dimension.#":          acctest.Ct2,
						names.AttrTableName:    "test_table",
						"timestamp.#":          acctest.Ct1,
						"timestamp.0.unit":     "MILLISECONDS",
						"timestamp.0.value":    "${time}",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "timestream.*.dimension.*", map[string]string{
						names.AttrName:  "dim1",
						names.AttrValue: "${dim1}",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "timestream.*.dimension.*", map[string]string{
						names.AttrName:  "dim2",
						names.AttrValue: "${dim2}",
					}),
				),
			},
			{
				Config: testAccTopicRuleConfig_timestream(rName, "dim3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "timestream.*", map[string]string{
						names.AttrDatabaseName: "TestDB",
						"dimension.#":          acctest.Ct2,
						names.AttrTableName:    "test_table",
						"timestamp.#":          acctest.Ct1,
						"timestamp.0.unit":     "MILLISECONDS",
						"timestamp.0.value":    "${time}",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "timestream.*.dimension.*", map[string]string{
						names.AttrName:  "dim3",
						names.AttrValue: "${dim3}",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "timestream.*.dimension.*", map[string]string{
						names.AttrName:  "dim2",
						names.AttrValue: "${dim2}",
					}),
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

func TestAccIoTTopicRule_errorAction(t *testing.T) {
	ctx := acctest.Context(t)
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleConfig_kinesisErrorAction(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.cloudwatch_logs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.iot_analytics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.kinesis.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.kinesis.0.stream_name", "mystream2"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.timestream.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "kinesis.*", map[string]string{
						"stream_name": "mystream1",
					}),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct0),
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

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/16115
func TestAccIoTTopicRule_updateKinesisErrorAction(t *testing.T) {
	ctx := acctest.Context(t)
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleConfig_kinesis(rName, "mystream"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "kinesis.*", map[string]string{
						"stream_name": "mystream",
					}),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct0),
				),
			},
			{
				Config: testAccTopicRuleConfig_kinesisErrorAction(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.cloudwatch_alarm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.cloudwatch_logs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.cloudwatch_metric.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.dynamodb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.dynamodbv2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.elasticsearch.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.iot_analytics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.kinesis.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.kinesis.0.stream_name", "mystream2"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.timestream.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "kinesis.*", map[string]string{
						"stream_name": "mystream1",
					}),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "republish.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", acctest.Ct0),
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

func testAccCheckTopicRuleDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IoTClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iot_topic_rule" {
				continue
			}

			_, err := tfiot.FindTopicRuleByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("IoT Topic Rule %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckTopicRuleExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No IoT Topic Rule ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IoTClient(ctx)

		_, err := tfiot.FindTopicRuleByName(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccTopicRuleName() string {
	return fmt.Sprintf("tf_acc_test_%[1]s", sdkacctest.RandString(20))
}

func testAccTopicRuleConfig_destinationRole(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {
      "Service": "iot.${data.aws_partition.current.dns_suffix}"
    },
    "Action": "sts:AssumeRole"
  }]
}
EOF
}

resource "aws_iam_policy" "test" {
  name        = %[1]q
  path        = "/"
  description = "IoT Topic Rule test policy"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*"
  }]
}
EOF
}

resource "aws_iam_policy_attachment" "test" {
  name       = %[1]q
  roles      = [aws_iam_role.test.name]
  policy_arn = aws_iam_policy.test.arn
}
`, rName)
}

func testAccTopicRuleConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_iot_topic_rule" "test" {
  name        = %[1]q
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"
}
`, rName)
}

func testAccTopicRuleConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_iot_topic_rule" "test" {
  name        = %[1]q
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccTopicRuleConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_iot_topic_rule" "test" {
  name        = %[1]q
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccTopicRuleConfig_cloudWatchAlarm(rName string, alarmName string) string {
	return acctest.ConfigCompose(
		testAccTopicRuleConfig_destinationRole(rName),
		fmt.Sprintf(`
resource "aws_iot_topic_rule" "test" {
  name        = %[1]q
  description = "Example rule"
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  cloudwatch_alarm {
    alarm_name   = %[2]q
    role_arn     = aws_iam_role.test.arn
    state_reason = "test"
    state_value  = "OK"
  }
}
`, rName, alarmName))
}

func testAccTopicRuleConfig_cloudWatchLogs(rName string, logGroupName string) string {
	return acctest.ConfigCompose(
		testAccTopicRuleConfig_destinationRole(rName),
		fmt.Sprintf(`
resource "aws_iot_topic_rule" "test" {
  name        = %[1]q
  enabled     = false
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  cloudwatch_logs {
    log_group_name = %[2]q
    role_arn       = aws_iam_role.test.arn
  }

  cloudwatch_logs {
    log_group_name = "mylogs2"
    role_arn       = aws_iam_role.test.arn
  }
}
`, rName, logGroupName))
}

func testAccTopicRuleConfig_cloudWatchLogsBatchMode(rName string, batchMode bool) string {
	return acctest.ConfigCompose(
		testAccTopicRuleConfig_destinationRole(rName),
		fmt.Sprintf(`
resource "aws_iot_topic_rule" "test" {
  name        = %[1]q
  enabled     = false
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  cloudwatch_logs {
    batch_mode     = %[2]t
    log_group_name = "mylogs1"
    role_arn       = aws_iam_role.test.arn
  }
}
`, rName, batchMode))
}

func testAccTopicRuleConfig_cloudWatchMetric(rName string, metricName string) string {
	return acctest.ConfigCompose(
		testAccTopicRuleConfig_destinationRole(rName),
		fmt.Sprintf(`
resource "aws_iot_topic_rule" "test" {
  name        = %[1]q
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  cloudwatch_metric {
    metric_name      = %[2]q
    metric_namespace = "TestNS"
    metric_value     = "10"
    metric_unit      = "s"
    role_arn         = aws_iam_role.test.arn
  }
}
`, rName, metricName))
}

func testAccTopicRuleConfig_dynamoDB(rName string, tableName string) string {
	return acctest.ConfigCompose(
		testAccTopicRuleConfig_destinationRole(rName),
		fmt.Sprintf(`
resource "aws_iot_topic_rule" "test" {
  name        = %[1]q
  description = "Description1"
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  dynamodb {
    hash_key_field = "hkf"
    hash_key_value = "hkv"
    payload_field  = "pf"
    role_arn       = aws_iam_role.test.arn
    table_name     = %[2]q
  }
}
`, rName, tableName))
}

func testAccTopicRuleConfig_dynamoDBRangeKey(rName string, tableName string) string {
	return acctest.ConfigCompose(
		testAccTopicRuleConfig_destinationRole(rName),
		fmt.Sprintf(`
resource "aws_iot_topic_rule" "test" {
  name        = %[1]q
  description = "Description2"
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  dynamodb {
    hash_key_field  = "hkf"
    hash_key_value  = "hkv"
    payload_field   = "pf"
    range_key_field = "rkf"
    range_key_value = "rkv"
    range_key_type  = "STRING"
    role_arn        = aws_iam_role.test.arn
    table_name      = %[2]q
    operation       = "INSERT"
  }
}
`, rName, tableName))
}

func testAccTopicRuleConfig_dynamoDBv2(rName string, tableName string) string {
	return acctest.ConfigCompose(
		testAccTopicRuleConfig_destinationRole(rName),
		fmt.Sprintf(`
resource "aws_iot_topic_rule" "test" {
  name        = %[1]q
  enabled     = true
  sql         = "SELECT field as column_name FROM 'topic/test'"
  sql_version = "2015-10-08"

  dynamodbv2 {
    put_item {
      table_name = %[2]q
    }

    role_arn = aws_iam_role.test.arn
  }
}
`, rName, tableName))
}

func testAccTopicRuleConfig_elasticSearch(rName string, index string) string {
	return acctest.ConfigCompose(
		testAccTopicRuleConfig_destinationRole(rName),
		fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_iot_topic_rule" "test" {
  name        = %[1]q
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  elasticsearch {
    endpoint = "https://domain.${data.aws_region.current.name}.es.${data.aws_partition.current.dns_suffix}"
    id       = "myIdentifier"
    index    = %[2]q
    type     = "mydocument"
    role_arn = aws_iam_role.test.arn
  }
}
`, rName, index))
}

func testAccTopicRuleConfig_firehose(rName string, streamName string) string {
	return acctest.ConfigCompose(
		testAccTopicRuleConfig_destinationRole(rName),
		fmt.Sprintf(`
resource "aws_iot_topic_rule" "test" {
  name        = %[1]q
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  firehose {
    delivery_stream_name = %[2]q
    role_arn             = aws_iam_role.test.arn
  }

  firehose {
    delivery_stream_name = "mystream2"
    role_arn             = aws_iam_role.test.arn
  }

  firehose {
    delivery_stream_name = "mystream3"
    role_arn             = aws_iam_role.test.arn
  }
}
`, rName, streamName))
}

func testAccTopicRuleConfig_firehoseSeparator(rName, separator string) string {
	return acctest.ConfigCompose(
		testAccTopicRuleConfig_destinationRole(rName),
		fmt.Sprintf(`
resource "aws_iot_topic_rule" "test" {
  name        = %[1]q
  description = "Example rule"
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  firehose {
    delivery_stream_name = "mystream"
    role_arn             = aws_iam_role.test.arn
    separator            = %[2]q
  }
}
`, rName, separator))
}

func testAccTopicRuleConfig_firehoseBatchMode(rName string, batchMode bool) string {
	return acctest.ConfigCompose(
		testAccTopicRuleConfig_destinationRole(rName),
		fmt.Sprintf(`
resource "aws_iot_topic_rule" "test" {
  name        = %[1]q
  description = "Example rule"
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  firehose {
    delivery_stream_name = "mystream"
    role_arn             = aws_iam_role.test.arn
    batch_mode           = %[2]t
  }
}
`, rName, batchMode))
}

func testAccTopicRuleConfig_http(rName string) string {
	return fmt.Sprintf(`
resource "aws_iot_topic_rule" "test" {
  name        = %[1]q
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  http {
    url = "https://example.com/ingress"
  }
}
`, rName)
}

func testAccTopicRuleConfig_httpConfirmationURL(rName string) string {
	return fmt.Sprintf(`
resource "aws_iot_topic_rule" "test" {
  name        = %[1]q
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  http {
    url              = "https://example.com/ingress"
    confirmation_url = "https://example.com/"
  }
}
`, rName)
}

func testAccTopicRuleConfig_httpHeaders(rName string) string {
	return fmt.Sprintf(`
resource "aws_iot_topic_rule" "test" {
  name        = %[1]q
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  http {
    url = "https://example.com/ingress"

    http_header {
      key   = "X-Header-1"
      value = "v1"
    }

    http_header {
      key   = "X-Header-2"
      value = "v2"
    }
  }
}
`, rName)
}

func testAccTopicRuleConfig_httpErrorAction(rName string) string {
	return fmt.Sprintf(`
resource "aws_iot_topic_rule" "test" {
  name        = %[1]q
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  http {
    url = "https://example.com/ingress"
  }

  error_action {
    http {
      url = "https://example.com/error-ingress"
    }
  }
}
`, rName)
}

func testAccTopicRuleConfig_analytics(rName string, channelName string) string {
	return acctest.ConfigCompose(
		testAccTopicRuleConfig_destinationRole(rName),
		fmt.Sprintf(`
resource "aws_iot_topic_rule" "test" {
  name        = %[1]q
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  iot_analytics {
    channel_name = %[2]q
    role_arn     = aws_iam_role.test.arn
  }
}
`, rName, channelName))
}

func testAccTopicRuleConfig_analyticsBatchMode(rName string, batchMode bool) string {
	return acctest.ConfigCompose(
		testAccTopicRuleConfig_destinationRole(rName),
		fmt.Sprintf(`
resource "aws_iot_topic_rule" "test" {
  name        = %[1]q
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  iot_analytics {
    channel_name = "fakedata"
    role_arn     = aws_iam_role.test.arn
    batch_mode   = %[2]t
  }
}
`, rName, batchMode))
}

func testAccTopicRuleConfig_events(rName string) string {
	return acctest.ConfigCompose(
		testAccTopicRuleConfig_destinationRole(rName),
		fmt.Sprintf(`
resource "aws_iot_topic_rule" "test" {
  name        = %[1]q
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  iot_events {
    input_name = "fake_input_name"
    role_arn   = aws_iam_role.test.arn
    message_id = "fake_message_id"
  }
}
`, rName))
}

func testAccTopicRuleConfig_eventsBatchMode(rName string, batchMode bool) string {
	return acctest.ConfigCompose(
		testAccTopicRuleConfig_destinationRole(rName),
		fmt.Sprintf(`
resource "aws_iot_topic_rule" "test" {
  name        = %[1]q
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  iot_events {
    input_name = "fake_input_name"
    role_arn   = aws_iam_role.test.arn
    message_id = "fake_message_id"
    batch_mode = %[2]t
  }
}
`, rName, batchMode))
}

func testAccTopicRuleConfig_kafka(rName string, topic string, broker string) string {
	// Making a topic rule destination takes several minutes, as it requires creating many networking resources.
	// It's far faster to simply use a properly-formatted but nonexistent ARN for the destination.
	return acctest.ConfigCompose(
		testAccTopicRuleConfig_destinationRole(rName),
		fmt.Sprintf(`
data "aws_region" "current" {}
data "aws_caller_identity" "current" {}

resource "aws_iot_topic_rule" "test" {
  name        = %[1]q
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  kafka {
    destination_arn = "arn:${data.aws_partition.current.partition}:iot:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:ruledestination/vpc/pretend-this-is-a-uuid"
    topic           = %[2]q

    client_properties = {
      "acks"                  = "1"
      "bootstrap.servers"     = %[3]q
      "compression.type"      = "none"
      "key.serializer"        = "org.apache.kafka.common.serialization.StringSerializer"
      "security.protocol"     = "SSL"
      "ssl.keystore"          = "$${get_secret('secret_name', 'SecretBinary', '', '${aws_iam_role.test.arn}')}"
      "ssl.keystore.password" = "password"
      "value.serializer"      = "org.apache.kafka.common.serialization.ByteBufferSerializer"
    }

    header {
      key   = "header-1"
      value = "value-1"
    }

    header {
      key   = "header-2"
      value = "value-2"
    }
  }
}
`, rName, topic, broker))
}

func testAccTopicRuleConfig_kinesis(rName string, streamName string) string {
	return acctest.ConfigCompose(
		testAccTopicRuleConfig_destinationRole(rName),
		fmt.Sprintf(`
resource "aws_iot_topic_rule" "test" {
  name        = %[1]q
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  kinesis {
    stream_name = %[2]q
    role_arn    = aws_iam_role.test.arn
  }
}
`, rName, streamName))
}

func testAccTopicRuleConfig_lambda(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_partition" "current" {}

resource "aws_iot_topic_rule" "test" {
  name        = %[1]q
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  lambda {
    function_arn = "arn:${data.aws_partition.current.partition}:lambda:${data.aws_region.current.name}:123456789012:function:ProcessKinesisRecords"
  }
}
`, rName)
}

func testAccTopicRuleConfig_republish(rName string, topic string) string {
	return acctest.ConfigCompose(
		testAccTopicRuleConfig_destinationRole(rName),
		fmt.Sprintf(`
resource "aws_iot_topic_rule" "test" {
  name        = %[1]q
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  republish {
    role_arn = aws_iam_role.test.arn
    topic    = %[2]q
  }
}
`, rName, topic))
}

func testAccTopicRuleConfig_republishQoS(rName string) string {
	return acctest.ConfigCompose(
		testAccTopicRuleConfig_destinationRole(rName),
		fmt.Sprintf(`
resource "aws_iot_topic_rule" "test" {
  name        = %[1]q
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  republish {
    role_arn = aws_iam_role.test.arn
    topic    = "mytopic"
    qos      = 1
  }
}
`, rName))
}

func testAccTopicRuleConfig_s3(rName string, bucketName string) string {
	return acctest.ConfigCompose(
		testAccTopicRuleConfig_destinationRole(rName),
		fmt.Sprintf(`
resource "aws_iot_topic_rule" "test" {
  name        = %[1]q
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  s3 {
    bucket_name = %[2]q
    canned_acl  = "private"
    key         = "mykey"
    role_arn    = aws_iam_role.test.arn
  }
}
`, rName, bucketName))
}

func testAccTopicRuleConfig_sns(rName string, messageFormat string) string {
	return acctest.ConfigCompose(
		testAccTopicRuleConfig_destinationRole(rName),
		fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_iot_topic_rule" "test" {
  name        = %[1]q
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  sns {
    message_format = %[2]q
    role_arn       = aws_iam_role.test.arn
    target_arn     = "arn:${data.aws_partition.current.partition}:sns:${data.aws_region.current.name}:123456789012:my_corporate_topic"
  }
}
`, rName, messageFormat))
}

func testAccTopicRuleConfig_sqs(rName string, queueUrl string) string {
	return acctest.ConfigCompose(
		testAccTopicRuleConfig_destinationRole(rName),
		fmt.Sprintf(`
resource "aws_iot_topic_rule" "test" {
  name        = %[1]q
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  sqs {
    queue_url  = %[2]q
    role_arn   = aws_iam_role.test.arn
    use_base64 = false
  }
}
`, rName, queueUrl))
}

func testAccTopicRuleConfig_stepFunctions(rName string, smName string) string {
	return acctest.ConfigCompose(
		testAccTopicRuleConfig_destinationRole(rName),
		fmt.Sprintf(`
resource "aws_iot_topic_rule" "test" {
  name        = %[1]q
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  step_functions {
    execution_name_prefix = "myprefix"
    state_machine_name    = %[2]q
    role_arn              = aws_iam_role.test.arn
  }
}
`, rName, smName))
}

func testAccTopicRuleConfig_timestream(rName string, dimName string) string {
	return acctest.ConfigCompose(
		testAccTopicRuleConfig_destinationRole(rName),
		fmt.Sprintf(`
resource "aws_iot_topic_rule" "test" {
  name        = %[1]q
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  timestream {
    database_name = "TestDB"
    role_arn      = aws_iam_role.test.arn
    table_name    = "test_table"

    dimension {
      name  = %[2]q
      value = "$${%[2]s}"
    }

    dimension {
      name  = "dim2"
      value = "$${dim2}"
    }

    timestamp {
      unit  = "MILLISECONDS"
      value = "$${time}"
    }
  }
}
`, rName, dimName))
}

func testAccTopicRuleConfig_kinesisErrorAction(rName string) string {
	return acctest.ConfigCompose(
		testAccTopicRuleConfig_destinationRole(rName),
		fmt.Sprintf(`
resource "aws_iot_topic_rule" "test" {
  name        = %[1]q
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  kinesis {
    stream_name = "mystream1"
    role_arn    = aws_iam_role.test.arn
  }

  error_action {
    kinesis {
      stream_name = "mystream2"
      role_arn    = aws_iam_role.test.arn
    }
  }
}
`, rName))
}
