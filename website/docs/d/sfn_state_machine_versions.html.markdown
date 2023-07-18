---
subcategory: "SFN (Step Functions)"
layout: "aws"
page_title: "AWS: aws_sfn_state_machine_versions"
description: |-
  Terraform data source for managing an AWS SFN (Step Functions) State Machine Versions.
---

# Data Source: aws_sfn_state_machine_versions

Terraform data source for managing an AWS SFN (Step Functions) State Machine Versions.

## Example Usage

### Basic Usage

```terraform
data "aws_sfn_state_machine_versions" "test" {
  statemachine_arn = aws_sfn_state_machine.test.arn
}

```

## Argument Reference

The following arguments are required:

* `statemachine_arn` - (Required) ARN of the State Machine.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `statemachine_versions` - ARN List identifying the statemachine versions.
