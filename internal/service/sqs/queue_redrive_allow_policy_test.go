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

func TestAccSQSQueueRedriveAllowPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var queueAttributes map[types.QueueAttributeName]string
	resourceName := "aws_sqs_queue_redrive_allow_policy.test"
	queueResourceName := "aws_sqs_queue.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SQSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueRedriveAllowPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, queueResourceName, &queueAttributes),
					resource.TestCheckResourceAttrSet(resourceName, "redrive_allow_policy"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:   testAccQueueRedriveAllowPolicyConfig_basic(rName),
				PlanOnly: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "redrive_allow_policy", queueResourceName, "redrive_allow_policy"),
				),
			},
		},
	})
}

func TestAccSQSQueueRedriveAllowPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var queueAttributes map[types.QueueAttributeName]string
	resourceName := "aws_sqs_queue_redrive_allow_policy.test"
	queueResourceName := "aws_sqs_queue.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SQSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueRedriveAllowPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, queueResourceName, &queueAttributes),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsqs.ResourceQueueRedriveAllowPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSQSQueueRedriveAllowPolicy_Disappears_queue(t *testing.T) {
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
				Config: testAccQueueRedriveAllowPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, queueResourceName, &queueAttributes),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsqs.ResourceQueue(), queueResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSQSQueueRedriveAllowPolicy_update(t *testing.T) {
	ctx := acctest.Context(t)
	var queueAttributes map[types.QueueAttributeName]string
	resourceName := "aws_sqs_queue_redrive_allow_policy.test"
	queueResourceName := "aws_sqs_queue.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SQSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueRedriveAllowPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, queueResourceName, &queueAttributes),
					resource.TestCheckResourceAttrSet(resourceName, "redrive_allow_policy"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccQueueRedriveAllowPolicyConfig_updated(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "redrive_allow_policy"),
				),
			},
		},
	})
}

func TestAccSQSQueueRedriveAllowPolicy_byQueue(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SQSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueRedriveAllowPolicyConfig_byQueue(rName),
			},
		},
	})
}

func testAccQueueRedriveAllowPolicyConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "test" {
  name = %[1]q
}

resource "aws_sqs_queue" "test_src" {
  name = "%[1]s_src"
  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.test.arn
    maxReceiveCount     = 4
  })
}

resource "aws_sqs_queue_redrive_allow_policy" "test" {
  queue_url = aws_sqs_queue.test.id
  redrive_allow_policy = jsonencode({
    redrivePermission = "byQueue",
    sourceQueueArns   = [aws_sqs_queue.test_src.arn]
  })
}
`, rName)
}

func testAccQueueRedriveAllowPolicyConfig_updated(rName string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "test" {
  name = %[1]q
}

resource "aws_sqs_queue" "test_src" {
  name = "%[1]s_src"
  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.test.arn
    maxReceiveCount     = 4
  })
}

resource "aws_sqs_queue_redrive_allow_policy" "test" {
  queue_url = aws_sqs_queue.test.id
  redrive_allow_policy = jsonencode({
    redrivePermission = "allowAll"
  })
}
`, rName)
}

func testAccQueueRedriveAllowPolicyConfig_byQueue(rName string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "test" {
  name = %[1]q
}

resource "aws_sqs_queue" "test_src" {
  name = "%[1]s_src"
  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.test.arn
    maxReceiveCount     = 4
  })
}

resource "aws_sqs_queue_redrive_allow_policy" "test" {
  queue_url            = aws_sqs_queue.test.id
  redrive_allow_policy = "{\"redrivePermission\": \"byQueue\", \"sourceQueueArns\": [\"${aws_sqs_queue.test_src.arn}\"]}"
}`, rName)
}
