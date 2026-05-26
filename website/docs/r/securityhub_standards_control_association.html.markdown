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

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `updated_reason` - (Optional) The reason for updating the control's enablement status in the standard. Required when `association_status` is `DISABLED`.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_securityhub_standards_control_association.example
  identity = {
    security_control_id = "IAM.1"
    standards_arn       = "arn:aws:securityhub:us-east-1:123456789012:control/cis-aws-foundations-benchmark/v/1.2.0/1.10"
  }
}

resource "aws_securityhub_standards_control_association" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `security_control_id` (String) Security control ID.
* `standards_arn` (String) Standards ARN.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Security Hub standards control associations using `security_control_id` and `standards_arn` separated by a comma (`,`). For example:

```terraform
import {
  to = aws_securityhub_standards_control_association.example
  id = "IAM.1,arn:aws:securityhub:us-east-1:123456789012:control/cis-aws-foundations-benchmark/v/1.2.0/1.10"
}
```

Using `terraform import`, import Security Hub standards control associations using `security_control_id` and `standards_arn` separated by a comma (`,`). For example:

```console
% terraform import aws_securityhub_standards_control_association.example IAM.1,arn:aws:securityhub:us-east-1:123456789012:control/cis-aws-foundations-benchmark/v/1.2.0/1.10
```
