---
subcategory: "CloudWatch Observability Admin"
layout: "aws"
page_title: "AWS: aws_observabilityadmin_telemetry_evaluation_for_organization"
description: |-
  Manages an AWS CloudWatch Observability Admin Telemetry Evaluation For Organization.
---

# Resource: aws_observabilityadmin_telemetry_evaluation_for_organization

Manages an AWS CloudWatch Observability Admin Telemetry Evaluation For Organization.

Telemetry evaluation for organization begins onboarding the organization and all member accounts to the telemetry config feature, enabling CloudWatch to evaluate and discover the state of telemetry configuration for AWS resources across the organization.

For more information, see the [AWS CloudWatch Observability Admin documentation](https://docs.aws.amazon.com/cloudwatch/latest/observabilityadmin/what-is-observabilityadmin.html).

~> **NOTE:** This resource can only be used from the management account or a delegated admin account of an AWS Organization. Only one telemetry evaluation for organization can exist per organization. Creating this resource enables the feature; destroying it disables the feature.

## Example Usage

### Basic Usage (Current Region Only)

```terraform
resource "aws_observabilityadmin_telemetry_evaluation_for_organization" "example" {}
```

### All Regions

```terraform
resource "aws_observabilityadmin_telemetry_evaluation_for_organization" "example" {
  all_regions = true
}
```

### Specific Regions

```terraform
resource "aws_observabilityadmin_telemetry_evaluation_for_organization" "example" {
  regions = ["us-east-1", "eu-west-1"]
}
```

## Argument Reference

The following arguments are optional:

* `all_regions` - (Optional) Whether to start telemetry evaluation in all AWS Regions where CloudWatch Observability Admin is available. The current region becomes the home region. Mutually exclusive with `regions`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `regions` - (Optional) List of AWS Regions to include in multi-region telemetry evaluation. The current region is always implicitly included. Mutually exclusive with `all_regions`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `failure_reason` - Reason for the failure status. Only populated if `status` is `FAILED_START` or `FAILED_STOP`.
* `home_region` - AWS Region designated as the home region for multi-region telemetry evaluation. Only present when multi-region evaluation is active.
* `status` - Onboarding status of the telemetry config feature for the organization. Valid values: `NOT_STARTED`, `STARTING`, `FAILED_START`, `RUNNING`, `STOPPING`, `FAILED_STOP`, `STOPPED`.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `5m`)
- `delete` - (Default `5m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_observabilityadmin_telemetry_evaluation_for_organization.example
  identity = {
    region = "us-west-2"
  }
}

resource "aws_observabilityadmin_telemetry_evaluation_for_organization" "example" {}
```

### Identity Schema

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudWatch Observability Admin Telemetry Evaluation For Organization using the region name. For example:

```terraform
import {
  to = aws_observabilityadmin_telemetry_evaluation_for_organization.example
  id = "us-west-2"
}
```

Using `terraform import`, import CloudWatch Observability Admin Telemetry Evaluation For Organization using the region name. For example:

```console
% terraform import aws_observabilityadmin_telemetry_evaluation_for_organization.example us-west-2
```
