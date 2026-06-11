---
subcategory: "Inspector"
layout: "aws"
page_title: "AWS: aws_inspector2_configuration"
description: |-
  Manages Amazon Inspector v2 EC2 scan mode and ECR re-scan duration configuration.
---

# Resource: aws_inspector2_configuration

Manages Amazon Inspector v2 account-level (or, when called by the delegated
administrator, organization-wide) configuration: the EC2 scan mode and the
ECR automated re-scan duration. Wraps the
[`UpdateConfiguration`](https://docs.aws.amazon.com/inspector/v2/APIReference/API_UpdateConfiguration.html)
and [`GetConfiguration`](https://docs.aws.amazon.com/inspector/v2/APIReference/API_GetConfiguration.html)
APIs.

The resource is regional. When the calling account is the Inspector delegated
administrator for the AWS Organization, the configuration applies to all
member accounts in that region. Otherwise it applies only to the calling
account.

~> **NOTE:** Inspector v2 has no `DeleteConfiguration` operation. On
`destroy`, this resource resets the configuration to AWS defaults
(`EC2_HYBRID`, ECR `LIFETIME` rescan with `LAST_IN_USE_AT` pull-date mode)
rather than removing it.

## Example Usage

### EC2 scan mode

```terraform
resource "aws_inspector2_configuration" "example" {
  ec2_configuration {
    scan_mode = "EC2_HYBRID"
  }
}
```

### ECR re-scan duration

```terraform
resource "aws_inspector2_configuration" "example" {
  ecr_configuration {
    rescan_duration           = "DAYS_14"
    pull_date_rescan_duration = "DAYS_14"
    pull_date_rescan_mode     = "LAST_IN_USE_AT"
  }
}
```

### Combined

```terraform
resource "aws_inspector2_configuration" "example" {
  ec2_configuration {
    scan_mode = "EC2_HYBRID"
  }

  ecr_configuration {
    rescan_duration = "DAYS_14"
  }
}
```

## Argument Reference

At least one of `ec2_configuration` or `ecr_configuration` must be set.

* `ec2_configuration` - (Optional) Configuration block for EC2 scanning. See [below](#ec2_configuration).
* `ecr_configuration` - (Optional) Configuration block for ECR re-scan settings. See [below](#ecr_configuration).

### ec2_configuration

* `scan_mode` - (Required) EC2 scan method. Valid values: `EC2_SSM_AGENT_BASED`, `EC2_HYBRID`.

### ecr_configuration

* `rescan_duration` - (Required) Image push-date re-scan duration. Valid values: `LIFETIME`, `DAYS_14`, `DAYS_30`, `DAYS_60`, `DAYS_90`, `DAYS_180`.
* `pull_date_rescan_duration` - (Optional) Image pull-date re-scan duration. Valid values: `DAYS_14`, `DAYS_30`, `DAYS_60`, `DAYS_90`, `DAYS_180`.
* `pull_date_rescan_mode` - (Optional) Pull-date re-scan mode. Valid values: `LAST_PULL_DATE`, `LAST_IN_USE_AT`.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Inspector v2 Configuration using the AWS account ID. For example:

```terraform
import {
  to = aws_inspector2_configuration.example
  id = "123456789012"
}
```

Using `terraform import`, import Inspector v2 Configuration using the AWS account ID. For example:

```console
% terraform import aws_inspector2_configuration.example 123456789012
```
