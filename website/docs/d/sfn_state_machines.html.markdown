---
subcategory: "SFN (Step Functions)"
layout: "aws"
page_title: "AWS: aws_sfn_state_machines"
description: |-
  Terraform data source for managing AWS SFN (Step Functions) State Machines.
---

# Data Source: aws_sfn_state_machines

Terraform data source for managing AWS SFN (Step Functions) State Machines.

## Example Usage

### Basic Usage

```terraform
data "aws_sfn_state_machines" "example" {
}
```

```terraform
# Get all State Machines
data "aws_sfn_state_machines" "all" {
}

# Get more detailed information about each State Machine
data "aws_sfn_state_machine" "detailed" {
  for_each = data.aws_sfn_state_machines.all.names

  name = each.value
}
```

## Argument Reference

This data source does not accept any arguments.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arns` - List of ARNs of the State Machines.
* `names` - List of Names of the State Machines.
