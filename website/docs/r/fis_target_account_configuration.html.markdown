---
subcategory: "FIS (Fault Injection Simulator)"
layout: "aws"
page_title: "AWS: aws_fis_target_account_configuration"
description: |-
  Manages an AWS FIS (Fault Injection Simulator) Target Account Configuration.
---

# Resource: aws_fis_target_account_configuration

Manages an AWS FIS (Fault Injection Simulator) Target Account Configuration.

## Example Usage

### Basic Usage

```terraform
resource "aws_fis_target_account_configuration" "example" {
  experiment_template_id = aws_fis_experiment_template.example.id
  account_id             = data.aws_caller_identity.current.account_id
  role_arn               = aws_iam_role.fis_role.arn
  description            = "Example"
}
```

## Argument Reference

The following arguments are required:

* `account_id` - (Required) Account ID of the target account.
* `experiment_template_id` - (Required) Experiment Template ID.

The following arguments are optional:

* `description` - (Optional) Description of the target account.
* `role_arn` - (Optional) ARN of the IAM Role for the target account.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports no additional attributes.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import FIS (Fault Injection Simulator) Target Account Configuration using the `account_id,experiment_template_id`. For example:

```terraform
import {
  to = aws_fis_target_account_configuration.example
  id = "abcd123456789,123456789012"
}
```

Using `terraform import`, import FIS (Fault Injection Simulator) Target Account Configuration using the `account_id,experiment_template_id`. For example:

```console
% terraform import aws_fis_target_account_configuration.example abcd123456789,123456789012
```
