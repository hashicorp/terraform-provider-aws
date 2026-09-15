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

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `statemachine_arn` - (Required) ARN of the State Machine.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `statemachine_versions` - ARN List identifying the statemachine versions.
