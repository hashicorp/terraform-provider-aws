// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sqs_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/service/sqs"
	awspolicy "github.com/hashicorp/awspolicyequivalence"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsqs "github.com/hashicorp/terraform-provider-aws/internal/service/sqs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(sqs.EndpointsID, testAccErrorCheckSkip)
}

func testAccErrorCheckSkip(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"Unknown Attribute RedriveAllowPolicy",
	)
}

func TestAccSQSQueue_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var queueAttributes map[string]string
	resourceName := "aws_sqs_queue.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, sqs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_name(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &queueAttributes),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "sqs", rName),
					resource.TestCheckResourceAttr(resourceName, "content_based_deduplication", "false"),
					resource.TestCheckResourceAttr(resourceName, "deduplication_scope", ""),
					resource.TestCheckResourceAttr(resourceName, "delay_seconds", strconv.Itoa(tfsqs.DefaultQueueDelaySeconds)),
					resource.TestCheckResourceAttr(resourceName, "fifo_queue", "false"),
					resource.TestCheckResourceAttr(resourceName, "fifo_throughput_limit", ""),
					resource.TestCheckResourceAttr(resourceName, "kms_data_key_reuse_period_seconds", strconv.Itoa(tfsqs.DefaultQueueKMSDataKeyReusePeriodSeconds)),
					resource.TestCheckResourceAttr(resourceName, "kms_master_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "max_message_size", strconv.Itoa(tfsqs.DefaultQueueMaximumMessageSize)),
					resource.TestCheckResourceAttr(resourceName, "message_retention_seconds", strconv.Itoa(tfsqs.DefaultQueueMessageRetentionPeriod)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "policy", ""),
					resource.TestCheckResourceAttr(resourceName, "receive_wait_time_seconds", strconv.Itoa(tfsqs.DefaultQueueReceiveMessageWaitTimeSeconds)),
					resource.TestCheckResourceAttr(resourceName, "redrive_policy", ""),
					resource.TestCheckResourceAttr(resourceName, "redrive_allow_policy", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "url", resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "visibility_timeout_seconds", strconv.Itoa(tfsqs.DefaultQueueVisibilityTimeout)),
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

func TestAccSQSQueue_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var queueAttributes map[string]string
	resourceName := "aws_sqs_queue.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, sqs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &queueAttributes),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsqs.ResourceQueue(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSQSQueue_Name_generated(t *testing.T) {
	ctx := acctest.Context(t)
	var queueAttributes map[string]string
	resourceName := "aws_sqs_queue.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ErrorCheck:               acctest.ErrorCheck(t, sqs.EndpointsID),
		CheckDestroy:             testAccCheckQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_nameGenerated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &queueAttributes),
					acctest.CheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
					resource.TestCheckResourceAttr(resourceName, "fifo_queue", "false"),
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

func TestAccSQSQueue_NameGenerated_fifoQueue(t *testing.T) {
	ctx := acctest.Context(t)
	var queueAttributes map[string]string
	resourceName := "aws_sqs_queue.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ErrorCheck:               acctest.ErrorCheck(t, sqs.EndpointsID),
		CheckDestroy:             testAccCheckQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_nameGeneratedFIFO,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &queueAttributes),
					acctest.CheckResourceAttrNameWithSuffixGenerated(resourceName, "name", tfsqs.FIFOQueueNameSuffix),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
					resource.TestCheckResourceAttr(resourceName, "fifo_queue", "true"),
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

func TestAccSQSQueue_namePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var queueAttributes map[string]string
	resourceName := "aws_sqs_queue.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, sqs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_namePrefix("tf-acc-test-prefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &queueAttributes),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, "name", "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, "fifo_queue", "false"),
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

func TestAccSQSQueue_NamePrefix_fifoQueue(t *testing.T) {
	ctx := acctest.Context(t)
	var queueAttributes map[string]string
	resourceName := "aws_sqs_queue.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, sqs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_namePrefixFIFO("tf-acc-test-prefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &queueAttributes),
					acctest.CheckResourceAttrNameWithSuffixFromPrefix(resourceName, "name", "tf-acc-test-prefix-", tfsqs.FIFOQueueNameSuffix),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, "fifo_queue", "true"),
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

func TestAccSQSQueue_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var queueAttributes map[string]string
	resourceName := "aws_sqs_queue.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, sqs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &queueAttributes),
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
				Config: testAccQueueConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &queueAttributes),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccQueueConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &queueAttributes),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccSQSQueue_update(t *testing.T) {
	ctx := acctest.Context(t)
	var queueAttributes map[string]string
	resourceName := "aws_sqs_queue.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, sqs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_name(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &queueAttributes),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "sqs", rName),
					resource.TestCheckResourceAttr(resourceName, "content_based_deduplication", "false"),
					resource.TestCheckResourceAttr(resourceName, "deduplication_scope", ""),
					resource.TestCheckResourceAttr(resourceName, "delay_seconds", strconv.Itoa(tfsqs.DefaultQueueDelaySeconds)),
					resource.TestCheckResourceAttr(resourceName, "fifo_queue", "false"),
					resource.TestCheckResourceAttr(resourceName, "fifo_throughput_limit", ""),
					resource.TestCheckResourceAttr(resourceName, "kms_data_key_reuse_period_seconds", strconv.Itoa(tfsqs.DefaultQueueKMSDataKeyReusePeriodSeconds)),
					resource.TestCheckResourceAttr(resourceName, "kms_master_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "max_message_size", strconv.Itoa(tfsqs.DefaultQueueMaximumMessageSize)),
					resource.TestCheckResourceAttr(resourceName, "message_retention_seconds", strconv.Itoa(tfsqs.DefaultQueueMessageRetentionPeriod)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "policy", ""),
					resource.TestCheckResourceAttr(resourceName, "receive_wait_time_seconds", strconv.Itoa(tfsqs.DefaultQueueReceiveMessageWaitTimeSeconds)),
					resource.TestCheckResourceAttr(resourceName, "redrive_policy", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "visibility_timeout_seconds", strconv.Itoa(tfsqs.DefaultQueueVisibilityTimeout)),
				),
			},
			{
				Config: testAccQueueConfig_updated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &queueAttributes),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "sqs", rName),
					resource.TestCheckResourceAttr(resourceName, "content_based_deduplication", "false"),
					resource.TestCheckResourceAttr(resourceName, "deduplication_scope", ""),
					resource.TestCheckResourceAttr(resourceName, "delay_seconds", "90"),
					resource.TestCheckResourceAttr(resourceName, "fifo_queue", "false"),
					resource.TestCheckResourceAttr(resourceName, "fifo_throughput_limit", ""),
					resource.TestCheckResourceAttr(resourceName, "kms_data_key_reuse_period_seconds", strconv.Itoa(tfsqs.DefaultQueueKMSDataKeyReusePeriodSeconds)),
					resource.TestCheckResourceAttr(resourceName, "kms_master_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "max_message_size", "2048"),
					resource.TestCheckResourceAttr(resourceName, "message_retention_seconds", "86400"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "policy", ""),
					resource.TestCheckResourceAttr(resourceName, "receive_wait_time_seconds", "10"),
					resource.TestCheckResourceAttr(resourceName, "redrive_policy", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "visibility_timeout_seconds", "60"),
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

func TestAccSQSQueue_Policy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var queueAttributes map[string]string
	resourceName := "aws_sqs_queue.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	expectedPolicy := `
{
  "Version": "2012-10-17",
  "Id": "sqspolicy",
  "Statement":[{
    "Sid": "Stmt1451501026839",
    "Effect": "Allow",
    "Principal":"*",
    "Action":"sqs:SendMessage",
    "Resource":"arn:%[1]s:sqs:%[2]s:%[3]s:%[4]s",
    "Condition":{
      "ArnEquals":{"aws:SourceArn":"arn:%[1]s:sns:%[2]s:%[3]s:%[4]s"}
    }
  }]
}
`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, sqs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_policy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &queueAttributes),
					testAccCheckQueuePolicyAttribute(&queueAttributes, rName, expectedPolicy),
					resource.TestCheckResourceAttr(resourceName, "delay_seconds", "90"),
					resource.TestCheckResourceAttr(resourceName, "max_message_size", "2048"),
					resource.TestCheckResourceAttr(resourceName, "message_retention_seconds", "86400"),
					resource.TestCheckResourceAttr(resourceName, "receive_wait_time_seconds", "10"),
					resource.TestCheckResourceAttr(resourceName, "visibility_timeout_seconds", "60"),
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

func TestAccSQSQueue_Policy_ignoreEquivalent(t *testing.T) {
	ctx := acctest.Context(t)
	var queueAttributes map[string]string
	resourceName := "aws_sqs_queue.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	expectedPolicy := `
{
  "Version": "2012-10-17",
  "Id": "sqspolicy",
  "Statement":[{
    "Sid": "SID1993561419",
    "Effect": "Allow",
    "Principal":"*",
    "Action":[
      "sqs:SendMessage",
      "sqs:DeleteMessage",
      "sqs:ListQueues"
    ],
    "Resource":"arn:%[1]s:sqs:%[2]s:%[3]s:%[4]s",
    "Condition":{
      "ArnEquals":{"aws:SourceArn":"arn:%[1]s:sns:%[2]s:%[3]s:%[4]s"}
    }
  }]
}
`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, sqs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_policyEquivalent(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &queueAttributes),
					testAccCheckQueuePolicyAttribute(&queueAttributes, rName, expectedPolicy),
					resource.TestCheckResourceAttr(resourceName, "delay_seconds", "90"),
					resource.TestCheckResourceAttr(resourceName, "max_message_size", "2048"),
					resource.TestCheckResourceAttr(resourceName, "message_retention_seconds", "86400"),
					resource.TestCheckResourceAttr(resourceName, "receive_wait_time_seconds", "10"),
					resource.TestCheckResourceAttr(resourceName, "visibility_timeout_seconds", "60"),
				),
			},
			{
				Config:   testAccQueueConfig_policyNewEquivalent(rName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccSQSQueue_recentlyDeleted(t *testing.T) {
	ctx := acctest.Context(t)
	var queueAttributes map[string]string
	resourceName := "aws_sqs_queue.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, sqs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &queueAttributes),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsqs.ResourceQueue(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccQueueConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &queueAttributes),
				),
			},
		},
	})
}

func TestAccSQSQueue_redrivePolicy(t *testing.T) {
	ctx := acctest.Context(t)
	var queueAttributes map[string]string
	resourceName := "aws_sqs_queue.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, sqs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_redrivePolicy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &queueAttributes),
					resource.TestCheckResourceAttr(resourceName, "delay_seconds", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "redrive_policy"),
					resource.TestCheckResourceAttr(resourceName, "visibility_timeout_seconds", "300"),
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

func TestAccSQSQueue_redriveAllowPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	var queueAttributes map[string]string
	resourceName := "aws_sqs_queue.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, sqs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_redriveAllowPolicy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &queueAttributes),
					resource.TestCheckResourceAttr(resourceName, "delay_seconds", "0"),
					//resource.TestCheckResourceAttrSet(resourceName, "redrive_allow_policy"),
					resource.TestCheckResourceAttr(resourceName, "visibility_timeout_seconds", "300"),
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

func TestAccSQSQueue_fifoQueue(t *testing.T) {
	ctx := acctest.Context(t)
	var queueAttributes map[string]string
	resourceName := "aws_sqs_queue.test"
	rName := fmt.Sprintf("%s.fifo", sdkacctest.RandomWithPrefix(acctest.ResourcePrefix))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, sqs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_fifo(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &queueAttributes),
					resource.TestCheckResourceAttr(resourceName, "deduplication_scope", "queue"),
					resource.TestCheckResourceAttr(resourceName, "fifo_queue", "true"),
					resource.TestCheckResourceAttr(resourceName, "fifo_throughput_limit", "perQueue"),
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

func TestAccSQSQueue_FIFOQueue_expectNameError(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, sqs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccQueueConfig_fifo(rName),
				ExpectError: regexp.MustCompile(`invalid queue name:`),
			},
		},
	})
}

func TestAccSQSQueue_FIFOQueue_contentBasedDeduplication(t *testing.T) {
	ctx := acctest.Context(t)
	var queueAttributes map[string]string
	resourceName := "aws_sqs_queue.test"
	rName := fmt.Sprintf("%s.fifo", sdkacctest.RandomWithPrefix(acctest.ResourcePrefix))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, sqs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_fifoContentBasedDeduplication(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &queueAttributes),
					resource.TestCheckResourceAttr(resourceName, "content_based_deduplication", "true"),
					resource.TestCheckResourceAttr(resourceName, "fifo_queue", "true"),
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

func TestAccSQSQueue_FIFOQueue_highThroughputMode(t *testing.T) {
	ctx := acctest.Context(t)
	var queueAttributes map[string]string
	resourceName := "aws_sqs_queue.test"
	rName := fmt.Sprintf("%s.fifo", sdkacctest.RandomWithPrefix(acctest.ResourcePrefix))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, sqs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_fifoHighThroughputMode(rName, "null", "null"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &queueAttributes),
					resource.TestCheckResourceAttr(resourceName, "deduplication_scope", "queue"),
					resource.TestCheckResourceAttr(resourceName, "fifo_queue", "true"),
					resource.TestCheckResourceAttr(resourceName, "fifo_throughput_limit", "perQueue"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccQueueConfig_fifoHighThroughputMode(rName, "messageGroup", "perMessageGroupId"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &queueAttributes),
					resource.TestCheckResourceAttr(resourceName, "deduplication_scope", "messageGroup"),
					resource.TestCheckResourceAttr(resourceName, "fifo_queue", "true"),
					resource.TestCheckResourceAttr(resourceName, "fifo_throughput_limit", "perMessageGroupId"),
				),
			},
		},
	})
}

func TestAccSQSQueue_StandardQueue_expectContentBasedDeduplicationError(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, sqs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccQueueConfig_standardExpectContentBasedDeduplicationError(rName),
				ExpectError: regexp.MustCompile(`content-based deduplication can only be set for FIFO queue`),
			},
		},
	})
}

func TestAccSQSQueue_encryption(t *testing.T) {
	ctx := acctest.Context(t)
	var queueAttributes map[string]string
	resourceName := "aws_sqs_queue.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, sqs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_encryption(rName, "null"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &queueAttributes),
					resource.TestCheckResourceAttr(resourceName, "kms_data_key_reuse_period_seconds", "300"),
					resource.TestCheckResourceAttr(resourceName, "kms_master_key_id", "alias/aws/sqs"),
					resource.TestCheckResourceAttr(resourceName, "sqs_managed_sse_enabled", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccQueueConfig_encryption(rName, "3600"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &queueAttributes),
					resource.TestCheckResourceAttr(resourceName, "kms_data_key_reuse_period_seconds", "3600"),
					resource.TestCheckResourceAttr(resourceName, "kms_master_key_id", "alias/aws/sqs"),
					resource.TestCheckResourceAttr(resourceName, "sqs_managed_sse_enabled", "false"),
				),
			},
			{
				Config: testAccQueueConfig_managedEncryption(rName, "true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &queueAttributes),
					resource.TestCheckResourceAttr(resourceName, "kms_data_key_reuse_period_seconds", strconv.Itoa(tfsqs.DefaultQueueKMSDataKeyReusePeriodSeconds)),
					resource.TestCheckResourceAttr(resourceName, "kms_master_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "sqs_managed_sse_enabled", "true"),
				),
			},
		},
	})
}

func TestAccSQSQueue_managedEncryption(t *testing.T) {
	ctx := acctest.Context(t)
	var queueAttributes map[string]string
	resourceName := "aws_sqs_queue.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, sqs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_managedEncryption(rName, "null"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &queueAttributes),
					resource.TestCheckResourceAttr(resourceName, "sqs_managed_sse_enabled", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccQueueConfig_managedEncryption(rName, "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &queueAttributes),
					resource.TestCheckResourceAttr(resourceName, "sqs_managed_sse_enabled", "false"),
				),
			},
			{
				Config: testAccQueueConfig_managedEncryption(rName, "true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &queueAttributes),
					resource.TestCheckResourceAttr(resourceName, "sqs_managed_sse_enabled", "true"),
				),
			},
		},
	})
}

func TestAccSQSQueue_zeroVisibilityTimeoutSeconds(t *testing.T) {
	ctx := acctest.Context(t)
	var queueAttributes map[string]string
	resourceName := "aws_sqs_queue.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, sqs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_zeroVisibilityTimeoutSeconds(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &queueAttributes),
					resource.TestCheckResourceAttr(resourceName, "visibility_timeout_seconds", "0"),
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

// https://github.com/hashicorp/terraform-provider-aws/issues/19786.
func TestAccSQSQueue_defaultKMSDataKeyReusePeriodSeconds(t *testing.T) {
	ctx := acctest.Context(t)
	var queueAttributes map[string]string
	resourceName := "aws_sqs_queue.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, sqs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_defaultKMSDataKeyReusePeriodSeconds(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &queueAttributes),
					resource.TestCheckResourceAttr(resourceName, "kms_data_key_reuse_period_seconds", strconv.Itoa(tfsqs.DefaultQueueKMSDataKeyReusePeriodSeconds)),
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

func testAccCheckQueuePolicyAttribute(queueAttributes *map[string]string, rName, policyTemplate string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		expectedPolicy := fmt.Sprintf(policyTemplate, acctest.Partition(), acctest.Region(), acctest.AccountID(), rName)

		var actualPolicyText string
		for key, value := range *queueAttributes {
			if key == sqs.QueueAttributeNamePolicy {
				actualPolicyText = value
				break
			}
		}

		equivalent, err := awspolicy.PoliciesAreEquivalent(actualPolicyText, expectedPolicy)
		if err != nil {
			return fmt.Errorf("Error testing policy equivalence: %s", err)
		}
		if !equivalent {
			return fmt.Errorf("Non-equivalent policy error:\n\nexpected: %s\n\n     got: %s\n", expectedPolicy, actualPolicyText)
		}

		return nil
	}
}

func testAccCheckQueueExists(ctx context.Context, resourceName string, v *map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SQS Queue URL is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SQSConn(ctx)

		output, err := tfsqs.FindQueueAttributesByURL(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = output

		return nil
	}
}

func testAccCheckQueueDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SQSConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sqs_queue" {
				continue
			}

			// SQS seems to be highly eventually consistent. Even if one connection reports that the queue is gone,
			// another connection may still report it as present.
			_, err := tfresource.RetryUntilNotFound(ctx, tfsqs.QueueDeletedTimeout, func() (any, error) {
				return tfsqs.FindQueueAttributesByURL(ctx, conn, rs.Primary.ID)
			})
			if errors.Is(err, tfresource.ErrFoundResource) {
				return fmt.Errorf("SQS Queue %s still exists", rs.Primary.ID)
			}
			if err != nil {
				return err
			}

			continue
		}

		return nil
	}
}

const testAccQueueConfig_nameGenerated = `
resource "aws_sqs_queue" "test" {}
`

const testAccQueueConfig_nameGeneratedFIFO = `
resource "aws_sqs_queue" "test" {
  fifo_queue = true
}
`

func testAccQueueConfig_name(rName string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "test" {
  name = %[1]q
}
`, rName)
}

func testAccQueueConfig_namePrefix(prefix string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "test" {
  name_prefix = %[1]q
}
`, prefix)
}

func testAccQueueConfig_namePrefixFIFO(prefix string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "test" {
  name_prefix = %[1]q
  fifo_queue  = true
}
`, prefix)
}

func testAccQueueConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccQueueConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccQueueConfig_updated(rName string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "test" {
  name                       = %[1]q
  delay_seconds              = 90
  max_message_size           = 2048
  message_retention_seconds  = 86400
  receive_wait_time_seconds  = 10
  visibility_timeout_seconds = 60
}
`, rName)
}

func testAccQueueConfig_policy(rName string) string {
	return fmt.Sprintf(`
locals {
  queue_name = %[1]q
}

resource "aws_sns_topic" "test" {
  name = %[1]q
}

data "aws_partition" "current" {}

data "aws_region" "current" {}

data "aws_caller_identity" "current" {}

resource "aws_sqs_queue" "test" {
  name                       = local.queue_name
  delay_seconds              = 90
  max_message_size           = 2048
  message_retention_seconds  = 86400
  receive_wait_time_seconds  = 10
  visibility_timeout_seconds = 60

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Id": "sqspolicy",
  "Statement": [
    {
      "Sid": "Stmt1451501026839",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "sqs:SendMessage",
      "Resource": "arn:${data.aws_partition.current.partition}:sqs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:${local.queue_name}",
      "Condition": {
        "ArnEquals": {
          "aws:SourceArn": "${aws_sns_topic.test.arn}"
        }
      }
    }
  ]
}
EOF
}

resource "aws_sns_topic_subscription" "test" {
  topic_arn = aws_sns_topic.test.arn
  protocol  = "sqs"
  endpoint  = aws_sqs_queue.test.arn
}
`, rName)
}

func testAccQueueConfig_policyEquivalent(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

data "aws_partition" "current" {}

data "aws_region" "current" {}

data "aws_caller_identity" "current" {}

resource "aws_sqs_queue" "test" {
  name                       = %[1]q
  delay_seconds              = 90
  max_message_size           = 2048
  message_retention_seconds  = 86400
  receive_wait_time_seconds  = 10
  visibility_timeout_seconds = 60

  policy = jsonencode({
    Version = "2012-10-17"
    Id      = "sqspolicy"
    Statement = [{
      Sid       = "SID1993561419"
      Effect    = "Allow"
      Principal = "*"
      Action = [
        "sqs:SendMessage",
        "sqs:DeleteMessage",
        "sqs:ListQueues",
      ]
      Resource = "arn:${data.aws_partition.current.partition}:sqs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:%[1]s"
      Condition = {
        ArnEquals = {
          "aws:SourceArn" = aws_sns_topic.test.arn
        }
      }
    }]
  })
}

resource "aws_sns_topic_subscription" "test" {
  topic_arn = aws_sns_topic.test.arn
  protocol  = "sqs"
  endpoint  = aws_sqs_queue.test.arn
}
`, rName)
}

func testAccQueueConfig_policyNewEquivalent(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

data "aws_partition" "current" {}

data "aws_region" "current" {}

data "aws_caller_identity" "current" {}

resource "aws_sqs_queue" "test" {
  name                       = %[1]q
  delay_seconds              = 90
  max_message_size           = 2048
  message_retention_seconds  = 86400
  receive_wait_time_seconds  = 10
  visibility_timeout_seconds = 60

  policy = jsonencode({
    Version = "2012-10-17"
    Id      = "sqspolicy"
    Statement = [{
      Sid       = "SID1993561419"
      Effect    = "Allow"
      Principal = ["*"]
      Action = [
        "sqs:ListQueues",
        "sqs:SendMessage",
        "sqs:DeleteMessage",
      ]
      Resource = "arn:${data.aws_partition.current.partition}:sqs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:%[1]s"
      Condition = {
        ArnEquals = {
          "aws:SourceArn" = aws_sns_topic.test.arn
        }
      }
    }]
  })
}

resource "aws_sns_topic_subscription" "test" {
  topic_arn = aws_sns_topic.test.arn
  protocol  = "sqs"
  endpoint  = aws_sqs_queue.test.arn
}
`, rName)
}

func testAccQueueConfig_redrivePolicy(rName string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "test" {
  name                       = "%[1]s-1"
  delay_seconds              = 0
  visibility_timeout_seconds = 300

  redrive_policy = <<EOF
{
  "maxReceiveCount": 3,
  "deadLetterTargetArn": "${aws_sqs_queue.dlq.arn}"
}
EOF
}

resource "aws_sqs_queue" "dlq" {
  name = "%[1]s-2"
}
`, rName)
}

func testAccQueueConfig_redriveAllowPolicy(rName string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "test" {
  name                       = "%[1]s-1"
  delay_seconds              = 0
  visibility_timeout_seconds = 300

  redrive_allow_policy = <<EOF
{
  "redrivePermission": "byQueue",
  "sourceQueueArns": ["${aws_sqs_queue.dlq.arn}"]
}
EOF
}

resource "aws_sqs_queue" "dlq" {
  name = "%[1]s-2"
}
`, rName)
}

func testAccQueueConfig_fifo(rName string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "test" {
  name       = %[1]q
  fifo_queue = true
}
`, rName)
}

func testAccQueueConfig_fifoContentBasedDeduplication(rName string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "test" {
  name                        = %[1]q
  fifo_queue                  = true
  content_based_deduplication = true
}
`, rName)
}

func testAccQueueConfig_fifoHighThroughputMode(rName, deduplicationScope, fifoThroughputLimit string) string {
	if deduplicationScope != "null" {
		deduplicationScope = strconv.Quote(deduplicationScope)
	}

	if fifoThroughputLimit != "null" {
		fifoThroughputLimit = strconv.Quote(fifoThroughputLimit)
	}

	return fmt.Sprintf(`
resource "aws_sqs_queue" "test" {
  name       = %[1]q
  fifo_queue = true

  deduplication_scope   = %[2]s
  fifo_throughput_limit = %[3]s
}
`, rName, deduplicationScope, fifoThroughputLimit)
}

func testAccQueueConfig_standardExpectContentBasedDeduplicationError(rName string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "test" {
  name                        = %[1]q
  content_based_deduplication = true
}
`, rName)
}

func testAccQueueConfig_encryption(rName, kmsDataKeyReusePeriodSeconds string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "test" {
  name                              = %[1]q
  kms_master_key_id                 = "alias/aws/sqs"
  kms_data_key_reuse_period_seconds = %[2]s
}
`, rName, kmsDataKeyReusePeriodSeconds)
}

func testAccQueueConfig_managedEncryption(rName, sqsManagedSseEnabled string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "test" {
  name                    = %[1]q
  sqs_managed_sse_enabled = %[2]s
}
`, rName, sqsManagedSseEnabled)
}

func testAccQueueConfig_zeroVisibilityTimeoutSeconds(rName string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "test" {
  name                       = %[1]q
  visibility_timeout_seconds = 0
}
`, rName)
}

func testAccQueueConfig_defaultKMSDataKeyReusePeriodSeconds(rName string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "test" {
  name                              = %[1]q
  kms_data_key_reuse_period_seconds = 300
}
`, rName)
}
