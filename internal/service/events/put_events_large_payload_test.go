// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package events_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEventsPutEventsAction_largePayload(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		CheckDestroy: testAccCheckBusDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPutEventsActionConfig_largePayload(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPutEventsDelivered(ctx, rName, 1),
				),
			},
		},
	})
}

// nosemgrep: ci.events-in-func-name -- Function reflects PutEvents operation naming for consistency.
func testAccPutEventsActionConfig_largePayload(rName string) string {
	// Create a large payload > 256KB. 300KB should suffice.
	// 300 * 1024 = 307200 bytes.
	largeString := strings.Repeat("A", 307200)

	return fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_event_rule" "test" {
  name           = %[1]q
  event_bus_name = aws_cloudwatch_event_bus.test.name
  event_pattern = jsonencode({
    source = ["test.large"]
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
      source         = "test.large"
      detail_type    = "Large Payload Event"
      event_bus_name = aws_cloudwatch_event_bus.test.name
      detail = jsonencode({
        marker = %[1]q
        data   = "%[2]s"
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
`, rName, largeString)
}
