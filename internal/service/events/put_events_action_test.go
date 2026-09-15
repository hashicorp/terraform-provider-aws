// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package events_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEventsPutEventsAction_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		CheckDestroy: testAccCheckBusDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPutEventsActionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPutEventsDelivered(ctx, rName, 1),
				),
			},
		},
	})
}

func TestAccEventsPutEventsAction_multipleEntries(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		CheckDestroy: testAccCheckBusDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPutEventsActionConfig_multipleEntries(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPutEventsDelivered(ctx, rName, 2),
				),
			},
		},
	})
}

func TestAccEventsPutEventsAction_customBus(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		CheckDestroy: testAccCheckBusDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPutEventsActionConfig_customBus(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPutEventsDelivered(ctx, rName, 1),
				),
			},
		},
	})
}

// nosemgrep: ci.events-in-func-name -- Verification helper for PutEvents delivery
func testAccCheckPutEventsDelivered(ctx context.Context, rName string, expected int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		meta := acctest.Provider.Meta().(*conns.AWSClient)
		evConn := meta.EventsClient(ctx)
		sqsConn := meta.SQSClient(ctx)

		// Ensure bus exists (sanity)
		if _, err := evConn.DescribeEventBus(ctx, &eventbridge.DescribeEventBusInput{Name: &rName}); err != nil {
			return fmt.Errorf("event bus %s not found: %w", rName, err)
		}

		// Discover queue URL via name convention
		queueName := rName + "-events-test"
		getOut, err := sqsConn.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{QueueName: &queueName})
		if err != nil {
			return fmt.Errorf("getting queue url: %w", err)
		}

		deadline := time.Now().Add(2 * time.Minute)
		received := 0
		marker := rName
		for time.Now().Before(deadline) && received < expected {
			// Long poll
			msgOut, err := sqsConn.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
				QueueUrl:            getOut.QueueUrl,
				MaxNumberOfMessages: 10,
				WaitTimeSeconds:     10,
			})
			if err != nil {
				// transient network errors: retry
				continue
			}
			for _, m := range msgOut.Messages {
				if m.Body == nil {
					continue
				}
				// EventBridge SQS target wraps the event as JSON; look for marker inside detail
				if strings.Contains(*m.Body, marker) {
					// Optionally parse to verify structure
					var parsed map[string]any
					_ = json.Unmarshal([]byte(*m.Body), &parsed)
					received++
				}
			}
		}

		if received < expected {
			return fmt.Errorf("expected %d events delivered to SQS, received %d", expected, received)
		}
		return nil
	}
}

// nosemgrep: ci.events-in-func-name -- Function reflects PutEvents operation naming for consistency.
func testAccPutEventsActionConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_event_rule" "test" {
  name           = %[1]q
  event_bus_name = aws_cloudwatch_event_bus.test.name
  event_pattern = jsonencode({
    source = ["test.application"]
  })
}

resource "aws_sqs_queue" "events_target" {
  name = "%[1]s-events-test"
}

resource "aws_sqs_queue_policy" "events_target" {
  queue_url = aws_sqs_queue.events_target.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid       = "AllowEventBridgeSendMessage"
        Effect    = "Allow"
        Principal = { Service = "events.amazonaws.com" }
        Action    = "sqs:SendMessage"
        Resource  = aws_sqs_queue.events_target.arn
        Condition = {
          ArnEquals = { "aws:SourceArn" = aws_cloudwatch_event_rule.test.arn }
        }
      }
    ]
  })
}

resource "aws_cloudwatch_event_target" "test" {
  rule           = aws_cloudwatch_event_rule.test.name
  event_bus_name = aws_cloudwatch_event_bus.test.name
  target_id      = "sqs"
  arn            = aws_sqs_queue.events_target.arn
}

action "aws_events_put_events" "test" {
  config {
    entry {
      source         = "test.application"
      detail_type    = "Test Event"
      event_bus_name = aws_cloudwatch_event_bus.test.name
      detail = jsonencode({
        marker = %[1]q
        action = "test"
      })
    }
  }
}

resource "terraform_data" "trigger" {
  input = "trigger"
  lifecycle {
    action_trigger {
      events  = [after_create, before_update]
      actions = [action.aws_events_put_events.test]
    }
  }
  depends_on = [
    aws_cloudwatch_event_target.test,
    aws_sqs_queue_policy.events_target
  ]
}
`, rName)
}

// nosemgrep: ci.events-in-func-name -- Function reflects PutEvents operation naming for consistency.
func testAccPutEventsActionConfig_multipleEntries(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_event_rule" "test" {
  name           = %[1]q
  event_bus_name = aws_cloudwatch_event_bus.test.name
  event_pattern = jsonencode({
    source = ["test.application", "test.orders"]
  })
}

resource "aws_sqs_queue" "events_target" {
  name = "%[1]s-events-test"
}

resource "aws_sqs_queue_policy" "events_target" {
  queue_url = aws_sqs_queue.events_target.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid       = "AllowEventBridgeSendMessage"
        Effect    = "Allow"
        Principal = { Service = "events.amazonaws.com" }
        Action    = "sqs:SendMessage"
        Resource  = aws_sqs_queue.events_target.arn
        Condition = {
          ArnEquals = { "aws:SourceArn" = aws_cloudwatch_event_rule.test.arn }
        }
      }
    ]
  })
}

resource "aws_cloudwatch_event_target" "test" {
  rule           = aws_cloudwatch_event_rule.test.name
  event_bus_name = aws_cloudwatch_event_bus.test.name
  target_id      = "sqs"
  arn            = aws_sqs_queue.events_target.arn
}

action "aws_events_put_events" "test" {
  config {
    entry {
      source         = "test.application"
      detail_type    = "User Action"
      event_bus_name = aws_cloudwatch_event_bus.test.name
      detail = jsonencode({
        marker = %[1]q
        action = "login"
      })
    }

    entry {
      source         = "test.orders"
      detail_type    = "Order Created"
      event_bus_name = aws_cloudwatch_event_bus.test.name
      detail = jsonencode({
        marker = %[1]q
        amount = 99.99
      })
    }
  }
}

resource "terraform_data" "trigger" {
  input = "trigger"
  lifecycle {
    action_trigger {
      events  = [after_create, before_update]
      actions = [action.aws_events_put_events.test]
    }
  }
  depends_on = [
    aws_cloudwatch_event_target.test,
    aws_sqs_queue_policy.events_target
  ]
}
`, rName)
}

// nosemgrep: ci.events-in-func-name -- Function reflects PutEvents operation naming for consistency.
func testAccPutEventsActionConfig_customBus(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_cloudwatch_event_bus" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_event_rule" "test" {
  name           = %[1]q
  event_bus_name = aws_cloudwatch_event_bus.test.name
  event_pattern = jsonencode({
    source      = ["custom.source"]
    detail-type = ["Custom Event"]
  })
}

resource "aws_sqs_queue" "events_target" {
  name = "%[1]s-events-test"
}

resource "aws_sqs_queue_policy" "events_target" {
  queue_url = aws_sqs_queue.events_target.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid       = "AllowEventBridgeSendMessage"
        Effect    = "Allow"
        Principal = { Service = "events.amazonaws.com" }
        Action    = "sqs:SendMessage"
        Resource  = aws_sqs_queue.events_target.arn
        Condition = {
          ArnEquals = { "aws:SourceArn" = aws_cloudwatch_event_rule.test.arn }
        }
      }
    ]
  })
}

resource "aws_cloudwatch_event_target" "test" {
  rule           = aws_cloudwatch_event_rule.test.name
  event_bus_name = aws_cloudwatch_event_bus.test.name
  target_id      = "sqs"
  arn            = aws_sqs_queue.events_target.arn
}

action "aws_events_put_events" "test" {
  config {
    entry {
      source         = "custom.source"
      detail_type    = "Custom Event"
      event_bus_name = aws_cloudwatch_event_bus.test.name
      time           = "2023-01-01T12:00:00Z"
      resources      = ["arn:${data.aws_partition.current.partition}:s3:::example-bucket"]
      detail = jsonencode({
        custom_field = "custom_value"
        marker       = %[1]q
        timestamp    = "2023-01-01T12:00:00Z"
      })
    }
  }
}

resource "terraform_data" "trigger" {
  input = "trigger"
  lifecycle {
    action_trigger {
      events  = [after_create, before_update]
      actions = [action.aws_events_put_events.test]
    }
  }
  depends_on = [
    aws_cloudwatch_event_target.test,
    aws_sqs_queue_policy.events_target
  ]
}
`, rName)
}
