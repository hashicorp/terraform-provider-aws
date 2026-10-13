---
subcategory: "EventBridge V2"
layout: "aws"
page_title: "AWS: aws_eventbridgev2_resource_policy"
description: |-
  Manages the resource policy of an Amazon EventBridge V2 event bus.
---

# Resource: aws_eventbridgev2_resource_policy

Manages the resource policy of an Amazon EventBridge V2 event bus. Use a resource policy to grant other AWS accounts access to the event bus.

There are two ways to give another account access to an event bus:

* **AWS Resource Access Manager (AWS RAM)** (recommended). Share the event bus with one of the EventBridge V2 managed permissions; see [Sharing with Another Account](eventbridgev2_event_bus.html#sharing-with-another-account). AWS RAM writes the event bus's `AWS_RAM` policy for you, gives you one place to see and revoke what you have shared, and works with AWS Organizations. The event bus also appears in the other account's [`aws_eventbridgev2_event_buses`](../d/eventbridgev2_event_buses.html) results.
* **A resource policy managed by this resource.** Use it when you need an explicit `Deny`, a condition key, or a grant that no managed permission expresses, either on its own or alongside an AWS RAM share. AWS RAM does not see this policy, so you own tracking it. An event bus shared only through a resource policy is not listed by `aws_eventbridgev2_event_buses` in the other account; give that account the event bus ARN instead.

EventBridge evaluates both policies when it authorizes a caller from another account, and an explicit `Deny` in either policy overrides an `Allow` in the other. The other account also needs the same actions in its own identity policies. See [Write a custom resource policy](https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-custom-bus-sharing.html#eb-custom-bus-access-policy).

-> **Note:** This resource manages the customer-managed (`default`) policy only. The `AWS_RAM` policy is managed exclusively by AWS RAM.

## Example Usage

### Grant Another Account Access to Subscribe

```terraform
resource "aws_eventbridgev2_event_bus" "example" {
  name = "example"
}

resource "aws_eventbridgev2_resource_policy" "example" {
  resource_arn = aws_eventbridgev2_event_bus.example.arn

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Sid    = "AllowPartnerAccountToSubscribe"
      Effect = "Allow"
      Principal = {
        AWS = "arn:aws:iam::444455556666:root"
      }
      Action = [
        "events:CreateSubscriber",
        "events:DescribeEventBus",
      ]
      Resource = aws_eventbridgev2_event_bus.example.arn
    }]
  })
}
```

The other account then creates a subscriber on the event bus by its ARN:

```terraform
resource "aws_eventbridgev2_subscriber" "example" {
  name          = "example"
  event_bus_arn = "arn:aws:events:us-east-1:111122223333:event-busv2/example/abcdefghij0123456789abcde"

  filter_configuration {
    filter {
      scope   = "METADATA"
      pattern = jsonencode({ source = ["example"] })
    }
  }

  invoke_configuration {
    role_arn   = aws_iam_role.example.arn
    target_arn = aws_sqs_queue.example.arn
  }
}
```

### Restrict an Event Bus Shared Through AWS RAM

```terraform
resource "aws_ram_resource_share" "example" {
  name                      = "example"
  allow_external_principals = false
  permission_arns           = ["arn:aws:ram::aws:permission/AWSRAMEventBridgeEventBusV2FullAccess"]
}

resource "aws_ram_resource_association" "example" {
  resource_arn       = aws_eventbridgev2_event_bus.example.arn
  resource_share_arn = aws_ram_resource_share.example.arn
}

resource "aws_ram_principal_association" "example" {
  principal          = "444455556666"
  resource_share_arn = aws_ram_resource_share.example.arn
}

# The managed permission allows publishing; deny it so the account can only subscribe and attach event sources.
resource "aws_eventbridgev2_resource_policy" "example" {
  resource_arn = aws_eventbridgev2_event_bus.example.arn

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Sid    = "DenyPartnerAccountToPublish"
      Effect = "Deny"
      Principal = {
        AWS = "arn:aws:iam::444455556666:root"
      }
      Action = [
        "events:PutEvents",
        "events:PutRawEvents",
      ]
      Resource = aws_eventbridgev2_event_bus.example.arn
    }]
  })
}
```

## Argument Reference

The following arguments are required:

* `policy` - (Required) Resource policy document, in JSON format.
* `resource_arn` - (Required) ARN of the event bus the policy is attached to. Changing this value creates a new resource.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `revision_id` - Revision ID of the policy. A new revision ID is generated each time the policy is updated.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_eventbridgev2_resource_policy.example
  identity = {
    "resource_arn" = "arn:aws:events:us-east-1:123456789012:event-busv2/example/abcdefghij0123456789abcde"
  }
}

resource "aws_eventbridgev2_resource_policy" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

- `resource_arn` (String) ARN of the event bus the policy is attached to.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import EventBridge V2 resource policies using the `resource_arn`. For example:

```terraform
import {
  to = aws_eventbridgev2_resource_policy.example
  id = "arn:aws:events:us-east-1:123456789012:event-busv2/example/abcdefghij0123456789abcde"
}
```

Using `terraform import`, import EventBridge V2 resource policies using the `resource_arn`. For example:

```console
% terraform import aws_eventbridgev2_resource_policy.example arn:aws:events:us-east-1:123456789012:event-busv2/example/abcdefghij0123456789abcde
```
