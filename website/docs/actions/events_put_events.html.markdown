---
subcategory: "EventBridge"
layout: "aws"
page_title: "AWS: aws_events_put_events"
description: |-
  Sends custom events to Amazon EventBridge so that they can be matched to rules.
---

# Action: aws_events_put_events

~> **Note:** `aws_events_put_events` is in beta. Its interface and behavior may change as the feature evolves, and breaking changes are possible. It is offered as a technical preview without compatibility guarantees until Terraform 1.14 is generally available.

Sends custom events to Amazon EventBridge so that they can be matched to rules. This action provides an imperative way to emit events from Terraform plans (e.g., deployment notifications) while still allowing Terraform to manage when the emission occurs through `action_trigger` lifecycle events.

## Example Usage

### Basic Event

```terraform
action "aws_events_put_events" "example" {
  config {
    entry {
      source      = "mycompany.myapp"
      detail_type = "User Action"
      detail = jsonencode({
        user_id = "12345"
        action  = "login"
      })
    }
  }
}
```

### Multiple Events

```terraform
action "aws_events_put_events" "batch" {
  config {
    entry {
      source      = "mycompany.orders"
      detail_type = "Order Created"
      detail = jsonencode({
        order_id = "order-123"
        amount   = 99.99
      })
    }

    entry {
      source      = "mycompany.orders"
      detail_type = "Order Updated"
      detail = jsonencode({
        order_id = "order-456"
        status   = "shipped"
      })
    }
  }
}
```

### Custom Event Bus

```terraform
resource "aws_cloudwatch_event_bus" "example" {
  name = "custom-bus"
}

action "aws_events_put_events" "custom_bus" {
  config {
    entry {
      source         = "mycompany.analytics"
      detail_type    = "Page View"
      event_bus_name = aws_cloudwatch_event_bus.example.name
      detail = jsonencode({
        page = "/home"
        user = "anonymous"
      })
    }
  }
}
```

### Event with Resources and Timestamp

```terraform
action "aws_events_put_events" "detailed" {
  config {
    entry {
      source      = "aws.ec2"
      detail_type = "EC2 Instance State-change Notification"
      time        = "2023-01-01T12:00:00Z" # RFC3339
      resources   = ["arn:aws:ec2:us-east-1:123456789012:instance/i-1234567890abcdef0"]
      detail = jsonencode({
        instance_id = "i-1234567890abcdef0"
        state       = "running"
      })
    }
  }
}
```

### Triggered by Terraform Data

```terraform
resource "terraform_data" "deploy" {
  input = var.deployment_id

  lifecycle {
    action_trigger {
      events  = [before_create, before_update]
      actions = [action.aws_events_put_events.deployment]
    }
  }
}

action "aws_events_put_events" "deployment" {
  config {
    entry {
      source      = "mycompany.deployments"
      detail_type = "Deployment Complete"
      detail = jsonencode({
        deployment_id = var.deployment_id
        environment   = var.environment
        timestamp     = timestamp()
      })
    }
  }
}
```

## Argument Reference

This action supports the following arguments:

* `entry` - (Required) One or more `entry` blocks defining events to send. Multiple blocks may be specified. See [below](#entry-block).
* `region` - (Optional) Region where this action should be [run](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `entry` Block

Each `entry` block supports:

* `source` - (Required) The source identifier for the event (e.g., `mycompany.myapp`).
* `detail_type` - (Optional) Free-form string used to decide what fields to expect in the event detail.
* `detail` - (Optional) JSON string (use `jsonencode()`) representing the event detail payload.
* `event_bus_name` - (Optional) Name or ARN of the event bus. Defaults to the account's default bus.
* `resources` - (Optional) List of ARNs the event primarily concerns.
* `time` - (Optional) RFC3339 timestamp for the event. If omitted, the receive time is used.
