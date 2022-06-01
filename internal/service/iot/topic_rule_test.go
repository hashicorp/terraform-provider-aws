package iot_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/iot"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiot "github.com/hashicorp/terraform-provider-aws/internal/service/iot"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(iot.EndpointsID, testAccErrorCheckSkip)
}

func testAccErrorCheckSkip(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"not been supported in region",
	)
}

func TestAccIoTTopicRule_basic(t *testing.T) {
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "iot", fmt.Sprintf("rule/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "http.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "republish.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "s3.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sns.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sql", "SELECT * FROM 'topic/test'"),
					resource.TestCheckResourceAttr(resourceName, "sql_version", "2015-10-08"),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", "0"),
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
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfiot.ResourceTopicRule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIoTTopicRule_tags(t *testing.T) {
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(resourceName),
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
				Config: testAccTopicRuleConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccTopicRuleConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccIoTTopicRule_cloudWatchAlarm(t *testing.T) {
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleCloudWatchAlarmConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cloudwatch_alarm.*", map[string]string{
						"alarm_name":   "myalarm",
						"state_reason": "test",
						"state_value":  "OK",
					}),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", "Example rule"),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "http.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "republish.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "s3.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sns.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", "0"),
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

func TestAccIoTTopicRule_cloudWatchLogs(t *testing.T) {
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleCloudWatchLogsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cloudwatch_logs.*", map[string]string{
						"log_group_name": "mylogs1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cloudwatch_logs.*", map[string]string{
						"log_group_name": "mylogs2",
					}),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "http.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "republish.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "s3.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sns.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", "0"),
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

func TestAccIoTTopicRule_cloudWatchMetric(t *testing.T) {
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleCloudWatchMetricConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cloudwatch_metric.*", map[string]string{
						"metric_name":      "TestName",
						"metric_namespace": "TestNS",
						"metric_unit":      "s",
						"metric_value":     "10",
					}),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "http.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "republish.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "s3.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sns.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", "0"),
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
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleDynamoDBConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", "Description1"),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "dynamodb.*", map[string]string{
						"hash_key_field": "hkf",
						"hash_key_value": "hkv",
						"payload_field":  "pf",
						"table_name":     "tn",
					}),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "http.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "republish.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "s3.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sns.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTopicRuleDynamoDBRangeKeyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", "Description2"),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "dynamodb.*", map[string]string{
						"hash_key_field":  "hkf",
						"hash_key_value":  "hkv",
						"operation":       "INSERT",
						"payload_field":   "pf",
						"range_key_field": "rkf",
						"range_key_type":  "STRING",
						"range_key_value": "rkv",
						"table_name":      "tn",
					}),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "http.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "republish.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "s3.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sns.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", "0"),
				),
			},
		},
	})
}

func TestAccIoTTopicRule_dynamoDBv2(t *testing.T) {
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleDynamoDBv2Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "dynamodbv2.*", map[string]string{
						"put_item.#":            "1",
						"put_item.0.table_name": "test",
					}),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "http.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "republish.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "s3.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sns.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", "0"),
				),
			},
		},
	})
}

func TestAccIoTTopicRule_elasticSearch(t *testing.T) {
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleElasticsearchConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "elasticsearch.*", map[string]string{
						"id":    "myIdentifier",
						"index": "myindex",
						"type":  "mydocument",
					}),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "http.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "republish.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "s3.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sns.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", "0"),
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

func TestAccIoTTopicRule_firehose(t *testing.T) {
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleFirehoseConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firehose.*", map[string]string{
						"delivery_stream_name": "mystream1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firehose.*", map[string]string{
						"delivery_stream_name": "mystream2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firehose.*", map[string]string{
						"delivery_stream_name": "mystream3",
					}),
					resource.TestCheckResourceAttr(resourceName, "http.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "republish.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "s3.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sns.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", "0"),
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

func TestAccIoTTopicRule_Firehose_separator(t *testing.T) {
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleFirehoseSeparatorConfig(rName, "\n"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firehose.*", map[string]string{
						"delivery_stream_name": "mystream",
						"separator":            "\n",
					}),
					resource.TestCheckResourceAttr(resourceName, "http.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "republish.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "s3.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sns.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTopicRuleFirehoseSeparatorConfig(rName, ","),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firehose.*", map[string]string{
						"delivery_stream_name": "mystream",
						"separator":            ",",
					}),
					resource.TestCheckResourceAttr(resourceName, "http.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "republish.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "s3.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sns.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", "0"),
				),
			},
		},
	})
}

func TestAccIoTTopicRule_http(t *testing.T) {
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleHTTPConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "http.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "http.*", map[string]string{
						"confirmation_url": "",
						"http_header.#":    "0",
						"url":              "https://example.com/ingress",
					}),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "republish.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "s3.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sns.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTopicRuleHTTPConfirmationURLConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "http.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "http.*", map[string]string{
						"confirmation_url": "https://example.com/",
						"http_header.#":    "0",
						"url":              "https://example.com/ingress",
					}),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "republish.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "s3.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sns.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", "0"),
				),
			},
			{
				Config: testAccTopicRuleHTTPHeadersConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "http.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "http.*", map[string]string{
						"confirmation_url":    "",
						"http_header.#":       "2",
						"http_header.0.key":   "X-Header-1",
						"http_header.0.value": "v1",
						"http_header.1.key":   "X-Header-2",
						"http_header.1.value": "v2",
						"url":                 "https://example.com/ingress",
					}),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "republish.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "s3.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sns.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", "0"),
				),
			},
			{
				Config: testAccTopicRuleHTTPErrorActionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.cloudwatch_alarm.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.cloudwatch_logs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.cloudwatch_metric.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.dynamodb.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.dynamodbv2.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.elasticsearch.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.firehose.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.http.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.http.0.url", "https://example.com/error-ingress"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.iot_analytics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.iot_events.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.kafka.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.kinesis.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.lambda.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.republish.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.s3.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.sns.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.sqs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.step_functions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.timestream.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "http.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "http.*", map[string]string{
						"confirmation_url": "",
						"http_header.#":    "0",
						"url":              "https://example.com/ingress",
					}),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "republish.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "s3.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sns.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", "0"),
				),
			},
		},
	})
}

func TestAccIoTTopicRule_IoT_analytics(t *testing.T) {
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleConfig_analytics(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "http.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "iot_analytics.*", map[string]string{
						"channel_name": "fakedata",
					}),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "republish.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "s3.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sns.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", "0"),
				),
			},
		},
	})
}

func TestAccIoTTopicRule_IoT_events(t *testing.T) {
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleConfig_events(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "http.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "iot_events.*", map[string]string{
						"input_name": "fake_input_name",
						"message_id": "fake_message_id",
					}),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "republish.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "s3.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sns.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", "0"),
				),
			},
		},
	})
}

func TestAccIoTTopicRule_kafka(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleKafkaConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "http.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "kafka.*", map[string]string{
						"client_properties.%":                     "8",
						"client_properties.acks":                  "1",
						"client_properties.bootstrap.servers":     "b-1.localhost:9094",
						"client_properties.compression.type":      "none",
						"client_properties.key.serializer":        "org.apache.kafka.common.serialization.StringSerializer",
						"client_properties.security.protocol":     "SSL",
						"client_properties.ssl.keystore.password": "password",
						"client_properties.value.serializer":      "org.apache.kafka.common.serialization.ByteBufferSerializer",
						"topic":                                   "fake_topic",
					}),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "republish.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "s3.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sns.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Delete everything but the IAM Role assumed by the IoT service.
			{
				Config: testAccTopicRuleRoleConfig(rName),
			},
		},
	})
}

func TestAccIoTTopicRule_kinesis(t *testing.T) {
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleKinesisConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "http.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "kinesis.*", map[string]string{
						"stream_name": "mystream",
					}),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "republish.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "s3.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sns.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", "0"),
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
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleLambdaConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "http.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "republish.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "s3.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sns.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", "0"),
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
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleRepublishConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "http.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "republish.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "republish.*", map[string]string{
						"qos":   "0",
						"topic": "mytopic",
					}),
					resource.TestCheckResourceAttr(resourceName, "s3.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sns.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", "0"),
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

func TestAccIoTTopicRule_republishWithQos(t *testing.T) {
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleRepublishWithQoSConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "http.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "republish.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "republish.*", map[string]string{
						"qos":   "1",
						"topic": "mytopic",
					}),
					resource.TestCheckResourceAttr(resourceName, "s3.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sns.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", "0"),
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
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleS3Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "http.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "republish.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "s3.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "s3.*", map[string]string{
						"bucket_name": "mybucket",
						"canned_acl":  "private",
						"key":         "mykey",
					}),
					resource.TestCheckResourceAttr(resourceName, "sns.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", "0"),
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
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleSNSConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "http.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "republish.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "s3.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", "0"),
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
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleSQSConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "http.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "republish.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "s3.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sns.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "sqs.*", map[string]string{
						"queue_url":  "fakedata",
						"use_base64": "false",
					}),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", "0"),
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
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleStepFunctionsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "http.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "republish.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "s3.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sns.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "step_functions.*", map[string]string{
						"execution_name_prefix": "myprefix",
						"state_machine_name":    "mystatemachine",
					}),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", "0"),
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
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleTimestreamConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "http.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "republish.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "s3.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sns.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "timestream.*", map[string]string{
						"database_name":     "TestDB",
						"dimension.#":       "1",
						"table_name":        "test_table",
						"timestamp.#":       "1",
						"timestamp.0.unit":  "MILLISECONDS",
						"timestamp.0.value": "${time}",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "timestream.*.dimension.*", map[string]string{
						"name":  "dim",
						"value": "${dim}",
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
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleErrorActionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.cloudwatch_alarm.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.cloudwatch_logs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.cloudwatch_metric.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.dynamodb.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.dynamodbv2.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.elasticsearch.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.firehose.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.http.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.iot_analytics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.iot_events.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.kafka.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.kinesis.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.kinesis.0.stream_name", "mystream2"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.lambda.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.republish.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.s3.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.sns.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.sqs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.step_functions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.timestream.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "http.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "kinesis.*", map[string]string{
						"stream_name": "mystream1",
					}),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "republish.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "s3.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sns.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", "0"),
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
	rName := testAccTopicRuleName()
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleKinesisConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "http.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "kinesis.*", map[string]string{
						"stream_name": "mystream",
					}),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "republish.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "s3.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sns.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", "0"),
				),
			},
			{
				Config: testAccTopicRuleErrorActionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_metric.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dynamodb.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dynamodbv2.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.cloudwatch_alarm.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.cloudwatch_logs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.cloudwatch_metric.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.dynamodb.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.dynamodbv2.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.elasticsearch.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.firehose.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.http.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.iot_analytics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.iot_events.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.kafka.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.kinesis.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.kinesis.0.stream_name", "mystream2"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.lambda.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.republish.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.s3.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.sns.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.sqs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.step_functions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "error_action.0.timestream.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "firehose.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "http.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_analytics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iot_events.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kafka.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kinesis.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "kinesis.*", map[string]string{
						"stream_name": "mystream1",
					}),
					resource.TestCheckResourceAttr(resourceName, "lambda.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "republish.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "s3.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sns.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sqs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "step_functions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "timestream.#", "0"),
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

func testAccCheckTopicRuleDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).IoTConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iot_topic_rule" {
			continue
		}

		_, err := tfiot.FindTopicRuleByName(conn, rs.Primary.ID)

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

func testAccCheckTopicRuleExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No IoT Topic Rule ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IoTConn

		_, err := tfiot.FindTopicRuleByName(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccTopicRuleName() string {
	return fmt.Sprintf("tf_acc_test_%[1]s", sdkacctest.RandString(20))
}

func testAccTopicRuleRoleConfig(rName string) string {
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

func testAccTopicRuleConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_iot_topic_rule" "test" {
  name        = %[1]q
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"
}
`, rName)
}

func testAccTopicRuleConfigTags1(rName, tagKey1, tagValue1 string) string {
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

func testAccTopicRuleConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
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

func testAccTopicRuleCloudWatchAlarmConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccTopicRuleRoleConfig(rName),
		fmt.Sprintf(`
resource "aws_iot_topic_rule" "test" {
  name        = %[1]q
  description = "Example rule"
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  cloudwatch_alarm {
    alarm_name   = "myalarm"
    role_arn     = aws_iam_role.test.arn
    state_reason = "test"
    state_value  = "OK"
  }
}
`, rName))
}

func testAccTopicRuleCloudWatchLogsConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccTopicRuleRoleConfig(rName),
		fmt.Sprintf(`
resource "aws_iot_topic_rule" "test" {
  name        = %[1]q
  enabled     = false
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  cloudwatch_logs {
    log_group_name = "mylogs1"
    role_arn       = aws_iam_role.test.arn
  }

  cloudwatch_logs {
    log_group_name = "mylogs2"
    role_arn       = aws_iam_role.test.arn
  }
}
`, rName))
}

func testAccTopicRuleCloudWatchMetricConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccTopicRuleRoleConfig(rName),
		fmt.Sprintf(`
resource "aws_iot_topic_rule" "test" {
  name        = %[1]q
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  cloudwatch_metric {
    metric_name      = "TestName"
    metric_namespace = "TestNS"
    metric_value     = "10"
    metric_unit      = "s"
    role_arn         = aws_iam_role.test.arn
  }
}
`, rName))
}

func testAccTopicRuleDynamoDBConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccTopicRuleRoleConfig(rName),
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
    table_name     = "tn"
  }
}
`, rName))
}

func testAccTopicRuleDynamoDBRangeKeyConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccTopicRuleRoleConfig(rName),
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
    table_name      = "tn"
    operation       = "INSERT"
  }
}
`, rName))
}

func testAccTopicRuleDynamoDBv2Config(rName string) string {
	return acctest.ConfigCompose(
		testAccTopicRuleRoleConfig(rName),
		fmt.Sprintf(`
resource "aws_iot_topic_rule" "test" {
  name        = %[1]q
  enabled     = true
  sql         = "SELECT field as column_name FROM 'topic/test'"
  sql_version = "2015-10-08"

  dynamodbv2 {
    put_item {
      table_name = "test"
    }

    role_arn = aws_iam_role.test.arn
  }
}
`, rName))
}

func testAccTopicRuleElasticsearchConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccTopicRuleRoleConfig(rName),
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
    index    = "myindex"
    type     = "mydocument"
    role_arn = aws_iam_role.test.arn
  }
}
`, rName))
}

func testAccTopicRuleFirehoseConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccTopicRuleRoleConfig(rName),
		fmt.Sprintf(`
resource "aws_iot_topic_rule" "test" {
  name        = %[1]q
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  firehose {
    delivery_stream_name = "mystream1"
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
`, rName))
}

func testAccTopicRuleFirehoseSeparatorConfig(rName, separator string) string {
	return acctest.ConfigCompose(
		testAccTopicRuleRoleConfig(rName),
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

func testAccTopicRuleHTTPConfig(rName string) string {
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

func testAccTopicRuleHTTPConfirmationURLConfig(rName string) string {
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

func testAccTopicRuleHTTPHeadersConfig(rName string) string {
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

func testAccTopicRuleHTTPErrorActionConfig(rName string) string {
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

func testAccTopicRuleConfig_analytics(rName string) string {
	return acctest.ConfigCompose(
		testAccTopicRuleRoleConfig(rName),
		fmt.Sprintf(`
resource "aws_iot_topic_rule" "test" {
  name        = %[1]q
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  iot_analytics {
    channel_name = "fakedata"
    role_arn     = aws_iam_role.test.arn
  }
}
`, rName))
}

func testAccTopicRuleConfig_events(rName string) string {
	return acctest.ConfigCompose(
		testAccTopicRuleRoleConfig(rName),
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

func testAccTopicRuleKafkaConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccTopicRuleDestinationConfig(rName),
		fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_iot_topic_rule" "test" {
  name        = %[1]q
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  kafka {
    destination_arn = aws_iot_topic_rule_destination.test.arn
    topic           = "fake_topic"

    client_properties = {
      "acks"                  = "1"
      "bootstrap.servers"     = "b-1.localhost:9094"
      "compression.type"      = "none"
      "key.serializer"        = "org.apache.kafka.common.serialization.StringSerializer"
      "security.protocol"     = "SSL"
      "ssl.keystore"          = "$${get_secret('secret_name', 'SecretBinary', '', '${aws_iam_role.test.arn}')}"
      "ssl.keystore.password" = "password"
      "value.serializer"      = "org.apache.kafka.common.serialization.ByteBufferSerializer"
    }
  }
}
`, rName))
}

func testAccTopicRuleKinesisConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccTopicRuleRoleConfig(rName),
		fmt.Sprintf(`
resource "aws_iot_topic_rule" "test" {
  name        = %[1]q
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  kinesis {
    stream_name = "mystream"
    role_arn    = aws_iam_role.test.arn
  }
}
`, rName))
}

func testAccTopicRuleLambdaConfig(rName string) string {
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

func testAccTopicRuleRepublishConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccTopicRuleRoleConfig(rName),
		fmt.Sprintf(`
resource "aws_iot_topic_rule" "test" {
  name        = %[1]q
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  republish {
    role_arn = aws_iam_role.test.arn
    topic    = "mytopic"
  }
}
`, rName))
}

func testAccTopicRuleRepublishWithQoSConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccTopicRuleRoleConfig(rName),
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

func testAccTopicRuleS3Config(rName string) string {
	return acctest.ConfigCompose(
		testAccTopicRuleRoleConfig(rName),
		fmt.Sprintf(`
resource "aws_iot_topic_rule" "test" {
  name        = %[1]q
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  s3 {
    bucket_name = "mybucket"
    canned_acl  = "private"
    key         = "mykey"
    role_arn    = aws_iam_role.test.arn
  }
}
`, rName))
}

func testAccTopicRuleSNSConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccTopicRuleRoleConfig(rName),
		fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_iot_topic_rule" "test" {
  name        = %[1]q
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  sns {
    role_arn   = aws_iam_role.test.arn
    target_arn = "arn:${data.aws_partition.current.partition}:sns:${data.aws_region.current.name}:123456789012:my_corporate_topic"
  }
}
`, rName))
}

func testAccTopicRuleSQSConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccTopicRuleRoleConfig(rName),
		fmt.Sprintf(`
resource "aws_iot_topic_rule" "test" {
  name        = %[1]q
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  sqs {
    queue_url  = "fakedata"
    role_arn   = aws_iam_role.test.arn
    use_base64 = false
  }
}
`, rName))
}

func testAccTopicRuleStepFunctionsConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccTopicRuleRoleConfig(rName),
		fmt.Sprintf(`
resource "aws_iot_topic_rule" "test" {
  name        = %[1]q
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  step_functions {
    execution_name_prefix = "myprefix"
    state_machine_name    = "mystatemachine"
    role_arn              = aws_iam_role.test.arn
  }
}
`, rName))
}

func testAccTopicRuleTimestreamConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccTopicRuleRoleConfig(rName),
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
      name  = "dim"
      value = "$${dim}"
    }

    timestamp {
      unit  = "MILLISECONDS"
      value = "$${time}"
    }
  }
}
`, rName))
}

func testAccTopicRuleErrorActionConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccTopicRuleRoleConfig(rName),
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
