---
subcategory: "CloudWatch Observability Admin"
layout: "aws"
page_title: "AWS: aws_observabilityadmin_telemetry_evaluation"
description: |-
  Manages an AWS CloudWatch Observability Admin Telemetry Evaluation.
---

# Resource: aws_observabilityadmin_telemetry_evaluation

Manages an AWS CloudWatch Observability Admin Telemetry Evaluation.

Telemetry evaluation begins onboarding the caller AWS account to the telemetry config feature, enabling CloudWatch to evaluate and discover the state of telemetry configuration for your AWS resources.

For more information, see the [AWS CloudWatch Observability Admin documentation](https://docs.aws.amazon.com/cloudwatch/latest/observabilityadmin/what-is-observabilityadmin.html).

~> **NOTE:** Only one telemetry evaluation can exist per account per region. Creating this resource enables the feature; destroying it disables the feature.

## Example Usage

### Basic Usage

```terraform
resource "aws_observabilityadmin_telemetry_evaluation" "example" {}
```

## Argument Reference

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `failure_reason` - Reason for the failure status. Only populated if `status` is `FAILED_START` or `FAILED_STOP`.
* `status` - Onboarding status of the telemetry config feature. Valid values: `NOT_STARTED`, `STARTING`, `FAILED_START`, `RUNNING`, `STOPPING`, `FAILED_STOP`, `STOPPED`.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `5m`)
- `delete` - (Default `5m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_observabilityadmin_telemetry_evaluation.example
  identity = {
    region = "us-west-2"
  }
}

resource "aws_observabilityadmin_telemetry_evaluation" "example" {}
```

### Identity Schema

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudWatch Observability Admin Telemetry Evaluation using the region name. For example:

```terraform
import {
  to = aws_observabilityadmin_telemetry_evaluation.example
  id = "us-west-2"
}
```

Using `terraform import`, import CloudWatch Observability Admin Telemetry Evaluation using the region name. For example:

```console
% terraform import aws_observabilityadmin_telemetry_evaluation.example us-west-2
```
