---
subcategory: "Security Hub"
layout: "aws"
page_title: "AWS: aws_securityhub_standards_control_association"
description: |-
  Terraform resource for managing an AWS Security Hub Standards Control Association.
---

# Resource: aws_securityhub_standards_control_association

Terraform resource for managing an AWS Security Hub Standards Control Association.

Disable/enable Security Hub security control in the standard.

The `aws_securityhub_standards_control_association`, similarly to `aws_securityhub_standards_control`,
behaves differently from normal resources, in that Terraform does not _create_ this resource, but instead "adopts" it
into management. When you _delete_ this resource configuration, Terraform "abandons" resource as is and just removes it from the state.

## Example Usage

### Basic usage

```terraform
resource "aws_securityhub_account" "example" {}

resource "aws_securityhub_standards_subscription" "cis_aws_foundations_benchmark" {
  standards_arn = "arn:aws:securityhub:::ruleset/cis-aws-foundations-benchmark/v/1.2.0"
  depends_on    = [aws_securityhub_account.example]
}

resource "aws_securityhub_standards_control_association" "cis_aws_foundations_benchmark_disable_iam_1" {
  standards_arn       = aws_securityhub_standards_subscription.cis_aws_foundations_benchmark.standards_arn
  security_control_id = "IAM.1"
  association_status  = "DISABLED"
  updated_reason      = "Not needed"
}
```

## Disabling security control in all standards

```terraform
resource "aws_securityhub_account" "example" {}

data "aws_securityhub_standards_control_associations" "iam_1" {
  security_control_id = "IAM.1"

  depends_on = [aws_securityhub_account.example]
}

resource "aws_securityhub_standards_control_association" "iam_1" {
  for_each = toset(data.aws_securityhub_standards_control_associations.iam_1.standards_control_associations[*].standards_arn)

  standards_arn       = each.key
  security_control_id = data.aws_securityhub_standards_control_associations.iam_1.security_control_id
  association_status  = "DISABLED"
  updated_reason      = "Not needed"
}
```

## Argument Reference

The following arguments are required:

* `association_status` - (Required) The desired enablement status of the control in the standard. Valid values: `ENABLED`, `DISABLED`.
* `security_control_id` - (Required) The unique identifier for the security control whose enablement status you want to update.
* `standards_arn` - (Required) The Amazon Resource Name (ARN) of the standard in which you want to update the control's enablement status.

The following arguments are optional:

* `updated_reason` - (Optional) The reason for updating the control's enablement status in the standard. Required when `association_status` is `DISABLED`.

## Attribute Reference

This resource exports no additional attributes.
