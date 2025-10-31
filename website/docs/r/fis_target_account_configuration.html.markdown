---
subcategory: "FIS (Fault Injection Simulator)"
layout: "aws"
page_title: "AWS: aws_fis_target_account_configuration"
description: |-
  Manages an AWS Fault Injection Simulator (FIS) target account configuration.
---

# Resource: aws_fis_target_account_configuration

Manages an AWS Fault Injection Simulator (FIS) target account configuration used when an experiment template targets multiple accounts. The configuration registers a target AWS account and the IAM role FIS assumes in that account when running an experiment.

-> Only required when the parent experiment template sets `experiment_options.account_targeting` to `multi-account`.

See the AWS API: [CreateTargetAccountConfiguration](https://docs.aws.amazon.com/fis/latest/APIReference/API_CreateTargetAccountConfiguration.html).

## Example Usage

```terraform
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_iam_role" "template" {
  name = "example-template-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = ["fis.${data.aws_partition.current.dns_suffix}"]
      }
    }]
  })
}

resource "aws_iam_role" "target" {
  name = "example-target-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = ["fis.${data.aws_partition.current.dns_suffix}"]
      }
    }]
  })
}

resource "aws_fis_experiment_template" "example" {
  description = "example"
  role_arn    = aws_iam_role.template.arn

  stop_condition { source = "none" }

  experiment_options { account_targeting = "multi-account" }

  action {
    name      = "example-action"
    action_id = "aws:ec2:stop-instances"

    target {
      key   = "Instances"
      value = "example-target"
    }
  }

  target {
    name           = "example-target"
    resource_type  = "aws:ec2:instance"
    selection_mode = "ALL"

    resource_tag {
      key   = "env"
      value = "example"
    }
  }
}

resource "aws_fis_target_account_configuration" "example" {
  account_id             = data.aws_caller_identity.current.account_id
  description            = "Example target account"
  experiment_template_id = aws_fis_experiment_template.example.id
  role_arn               = aws_iam_role.target.arn
}
```

## Argument Reference

The following arguments are supported:

- `account_id` – (Required, ForceNew) AWS account ID to register as a target account. Cannot be changed after creation.
- `experiment_template_id` – (Required, ForceNew) ID of the parent FIS experiment template. Cannot be changed after creation.
- `role_arn` – (Required, ForceNew) ARN of the IAM role FIS assumes in the target account when running the experiment.
- `description` – (Optional) Description for the target account configuration.

## Attributes Reference

In addition to all arguments above, the following attribute is exported:

- `id` – Composite identifier in the form `<experiment_template_id>/<account_id>`.

## Import

FIS target account configurations can be imported using the composite identifier:

```bash
terraform import aws_fis_target_account_configuration.example et-1234567890abcdef/123456789012
```

## Timeouts

This resource provides the following timeouts:

- `create` – (Default `30m`)
- `update` – (Default `30m`)
- `delete` – (Default `30m`)

