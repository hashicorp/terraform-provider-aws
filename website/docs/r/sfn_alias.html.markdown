---
subcategory: "SFN (Step Functions)"
layout: "aws"
page_title: "AWS: aws_sfn_alias"
description: |-
  Provides a Step Function State Machine Alias.
---

# Resource: aws_sfn_alias

Provides a Step Function State Machine Alias.

## Example Usage

### Basic Usage

```terraform
resource "aws_sfn_alias" "sfn_alias" {
  name = "my_sfn_alias"

  routing_configuration {
    state_machine_version_arn = aws_sfn_state_machine.sfn_test.state_machine_version_arn
    weight                    = 100
  }
}

resource "aws_sfn_alias" "my_sfn_alias" {
  name = "my_sfn_alias"

  routing_configuration {
    state_machine_version_arn = "arn:aws:states:us-east-1:12345:stateMachine:demo:3"
    weight                    = 50
  }

  routing_configuration {
    state_machine_version_arn = "arn:aws:states:us-east-1:12345:stateMachine:demo:2"
    weight                    = 50
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name for the alias you are creating.
* `description` - (Optional) Description of the alias.
* `routing_configuration` - (Required) The StateMachine alias' route configuration settings. Fields documented below

For **routing_configuration** the following attributes are supported:

* `state_machine_version_arn` - (Required) A version of the state machine.
* `weight` - (Required) Percentage of traffic routed to the state machine version.

The following arguments are optional:

* `optional_arg` - (Optional) Concise argument description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) identifying your state machine alias.
* `creation_date` - The date the state machine alias was created.

## Import

SFN (Step Functions) Alias can be imported using the `arn`, e.g.,

```
$ terraform import aws_sfn_alias.foo arn:aws:states:us-east-1:123456789098:stateMachine:myStateMachine:foo
```
