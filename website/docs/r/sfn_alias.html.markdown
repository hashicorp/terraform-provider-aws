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

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) Name for the alias you are creating.
* `description` - (Optional) Description of the alias.
* `routing_configuration` - (Required) The StateMachine alias' route configuration settings. Fields documented below

`routing_configuration` supports the following arguments:

* `state_machine_version_arn` - (Required) The Amazon Resource Name (ARN) of the state machine version.
* `weight` - (Required) Percentage of traffic routed to the state machine version.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) identifying your state machine alias.
* `creation_date` - The date the state machine alias was created.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SFN (Step Functions) Alias using the `arn`. For example:

```terraform
import {
  to = aws_sfn_alias.foo
  id = "arn:aws:states:us-east-1:123456789098:stateMachine:myStateMachine:foo"
}
```

Using `terraform import`, import SFN (Step Functions) Alias using the `arn`. For example:

```console
% terraform import aws_sfn_alias.foo arn:aws:states:us-east-1:123456789098:stateMachine:myStateMachine:foo
```
