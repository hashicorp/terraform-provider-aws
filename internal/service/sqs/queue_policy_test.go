// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sqs_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfsqs "github.com/hashicorp/terraform-provider-aws/internal/service/sqs"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSQSQueuePolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var queueAttributes map[types.QueueAttributeName]string
	resourceName := "aws_sqs_queue_policy.test"
	queueResourceName := "aws_sqs_queue.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SQSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueuePolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, queueResourceName, &queueAttributes),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:   testAccQueuePolicyConfig_basic(rName),
				PlanOnly: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPolicy, queueResourceName, names.AttrPolicy),
				),
			},
		},
	})
}

func TestAccSQSQueuePolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var queueAttributes map[types.QueueAttributeName]string
	resourceName := "aws_sqs_queue_policy.test"
	queueResourceName := "aws_sqs_queue.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SQSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueuePolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, queueResourceName, &queueAttributes),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsqs.ResourceQueuePolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSQSQueuePolicy_Disappears_queue(t *testing.T) {
	ctx := acctest.Context(t)
	var queueAttributes map[types.QueueAttributeName]string
	queueResourceName := "aws_sqs_queue.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SQSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueuePolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, queueResourceName, &queueAttributes),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsqs.ResourceQueue(), queueResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSQSQueuePolicy_update(t *testing.T) {
	ctx := acctest.Context(t)
	var queueAttributes map[types.QueueAttributeName]string
	resourceName := "aws_sqs_queue_policy.test"
	queueResourceName := "aws_sqs_queue.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SQSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueuePolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, queueResourceName, &queueAttributes),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccQueuePolicyConfig_updated(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
				),
			},
		},
	})
}

func testAccQueuePolicyConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "test" {
  name = %[1]q
}

resource "aws_sqs_queue_policy" "test" {
  queue_url = aws_sqs_queue.test.id

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": "*",
    "Action": "sqs:*",
    "Resource": "${aws_sqs_queue.test.arn}",
    "Condition": {
      "ArnEquals": {
        "aws:SourceArn": "${aws_sqs_queue.test.arn}"
      }
    }
  }]
}
POLICY
}
`, rName)
}

func testAccQueuePolicyConfig_updated(rName string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "test" {
  name = %[1]q
}

resource "aws_sqs_queue_policy" "test" {
  queue_url = aws_sqs_queue.test.id

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": "*",
    "Action": [
      "sqs:SendMessage",
      "sqs:ReceiveMessage"
    ],
    "Resource": "${aws_sqs_queue.test.arn}"
  }]
}
POLICY
}
`, rName)
}
