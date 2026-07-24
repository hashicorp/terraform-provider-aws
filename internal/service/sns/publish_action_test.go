// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sns_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSNSPublishAction_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, "sns"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccPublishActionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPublishActionDelivered(ctx, t, "aws_sqs_queue.test"),
				),
			},
		},
	})
}

func TestAccSNSPublishAction_withAttributes(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, "sns"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccPublishActionConfig_withAttributes(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPublishActionDelivered(ctx, t, "aws_sqs_queue.test"),
				),
			},
		},
	})
}

func testAccCheckPublishActionDelivered(ctx context.Context, t *testing.T, queueResourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[queueResourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", queueResourceName)
		}

		conn := acctest.ProviderMeta(ctx, t).SQSClient(ctx)
		queueURL := rs.Primary.Attributes[names.AttrURL]

		// Poll for message with timeout - adjusted for 20min test
		timeout := time.After(18 * time.Minute)         // Leave 2min buffer for test cleanup
		ticker := time.NewTicker(5 * time.Second)       // Poll every 5 seconds
		debugTicker := time.NewTicker(30 * time.Second) // Debug output every 30 seconds
		defer ticker.Stop()
		defer debugTicker.Stop()

		pollCount := 0
		startTime := time.Now()

		for {
			select {
			case <-timeout:
				elapsed := time.Since(startTime)
				return fmt.Errorf("timeout after %v waiting for SNS message in SQS queue (polled %d times)", elapsed, pollCount)

			case <-ticker.C:
				pollCount++
				output, err := conn.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
					QueueUrl:            &queueURL,
					MaxNumberOfMessages: 1,
					WaitTimeSeconds:     3, // Short poll with wait
				})
				if err != nil {
					continue
				}

				if len(output.Messages) > 0 {
					// Clean up message
					_, err := conn.DeleteMessage(ctx, &sqs.DeleteMessageInput{
						QueueUrl:      &queueURL,
						ReceiptHandle: output.Messages[0].ReceiptHandle,
					})
					if err != nil {
						return fmt.Errorf("error deleting message from SQS: %w", err)
					}
					return nil
				}
			}
		}
	}
}

func testAccPublishActionConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_sqs_queue" "test" {
  name = %[1]q
}

resource "aws_sns_topic_subscription" "test" {
  topic_arn = aws_sns_topic.test.arn
  protocol  = "sqs"
  endpoint  = aws_sqs_queue.test.arn
}

resource "aws_sqs_queue_policy" "test" {
  queue_url = aws_sqs_queue.test.id
  policy = jsonencode({
    Version = "2008-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = "*"
      Action    = "sqs:SendMessage"
      Resource  = aws_sqs_queue.test.arn
      Condition = {
        ArnEquals = {
          "aws:SourceArn" = aws_sns_topic.test.arn
        }
      }
    }]
  })
}

action "aws_sns_publish" "test" {
  config {
    topic_arn = aws_sns_topic.test.arn
    message   = "Test message from Terraform"
  }
}

resource "terraform_data" "trigger" {
  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.aws_sns_publish.test]
    }
  }

  depends_on = [
    aws_sns_topic_subscription.test,
    aws_sqs_queue_policy.test
  ]
}
`, rName)
}

func testAccPublishActionConfig_withAttributes(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_sqs_queue" "test" {
  name = %[1]q
}

resource "aws_sns_topic_subscription" "test" {
  topic_arn = aws_sns_topic.test.arn
  protocol  = "sqs"
  endpoint  = aws_sqs_queue.test.arn
}

resource "aws_sqs_queue_policy" "test" {
  queue_url = aws_sqs_queue.test.id
  policy = jsonencode({
    Version = "2008-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = "*"
      Action    = "sqs:SendMessage"
      Resource  = aws_sqs_queue.test.arn
      Condition = {
        ArnEquals = {
          "aws:SourceArn" = aws_sns_topic.test.arn
        }
      }
    }]
  })
}

action "aws_sns_publish" "test" {
  config {
    topic_arn = aws_sns_topic.test.arn
    subject   = "Test Subject"
    message   = "Test message with attributes"

    message_attributes {
      map_block_key = "priority"
      data_type     = "String"
      string_value  = "high"
    }

    message_attributes {
      map_block_key = "source"
      data_type     = "String"
      string_value  = "terraform"
    }
  }
}

resource "terraform_data" "trigger" {
  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.aws_sns_publish.test]
    }
  }

  depends_on = [
    aws_sns_topic_subscription.test,
    aws_sqs_queue_policy.test
  ]
}
`, rName)
}
