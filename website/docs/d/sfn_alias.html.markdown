---
subcategory: "SFN (Step Functions)"
layout: "aws"
page_title: "AWS: aws_sfn_alias"
description: |-
  Terraform data source for managing an AWS SFN (Step Functions) State Machine Alias.
---

# Data Source: aws_sfn_alias

Terraform data source for managing an AWS SFN (Step Functions) State Machine Alias.

## Example Usage

### Basic Usage

```terraform
data "aws_sfn_alias" "example" {
  name             = "my_sfn_alias"
  statemachine_arn = aws_sfn_state_machine.sfn_test.arn
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the State Machine alias.
* `statemachine_arn` - (Required) ARN of the State Machine.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN identifying the State Machine alias.
* `creation_date` - Date the state machine Alias was created.
* `description` - Description of state machine alias.
* `routing_configuration` - Routing Configuration of state machine alias
