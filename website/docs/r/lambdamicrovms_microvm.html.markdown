---
subcategory: "Lambda MicroVMs"
layout: "aws"
page_title: "AWS: aws_lambdamicrovms_microvm"
description: |-
  Manages an AWS Lambda MicroVMs MicroVM.
---

# Resource: aws_lambdamicrovms_microvm

Manages an AWS Lambda MicroVMs MicroVM. Use this resource to run a MicroVM from a [MicroVM image](lambdamicrovms_image.html.markdown), which provisions a dedicated HTTPS endpoint and starts your application from the image snapshot.

MicroVMs are immutable. Because the service has no update operation, changing any argument terminates the MicroVM and runs a new one.

## Example Usage

### Basic Usage

```terraform
resource "aws_lambdamicrovms_microvm" "example" {
  image_identifier = aws_lambdamicrovms_image.example.arn
}
```

### With Idle Policy and Logging

```terraform
resource "aws_lambdamicrovms_microvm" "example" {
  image_identifier            = aws_lambdamicrovms_image.example.arn
  execution_role_arn          = aws_iam_role.example.arn
  maximum_duration_in_seconds = 14400

  egress_network_connectors  = ["arn:aws:lambda:us-east-1:aws:network-connector:aws-network-connector:INTERNET_EGRESS"]
  ingress_network_connectors = ["arn:aws:lambda:us-east-1:aws:network-connector:aws-network-connector:ALL_INGRESS"]

  idle_policy {
    auto_resume_enabled        = true
    max_idle_duration_seconds  = 900
    suspended_duration_seconds = 1800
  }

  logging {
    cloud_watch {
      log_group = aws_cloudwatch_log_group.example.name
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `image_identifier` - (Required) ARN of the MicroVM image to run. Changing this value creates a new resource.

The following arguments are optional:

* `egress_network_connectors` - (Optional) List of egress network connector ARNs for outbound connectivity. Changing this value creates a new resource.
* `execution_role_arn` - (Optional) ARN of the IAM role assumed by the MicroVM during execution. Changing this value creates a new resource.
* `idle_policy` - (Optional) Configuration controlling automatic suspend and resume behavior. [See below](#idle_policy-block). Changing this value creates a new resource.
* `image_version` - (Optional) Version of the MicroVM image to run. Defaults to the latest active version. Changing this value creates a new resource.
* `ingress_network_connectors` - (Optional) List of ingress network connector ARNs for inbound connectivity. Changing this value creates a new resource.
* `logging` - (Optional) Logging configuration for the MicroVM. [See below](#logging-block). Changing this value creates a new resource.
* `maximum_duration_in_seconds` - (Optional) Maximum duration in seconds that the MicroVM can exist before the platform terminates it. Valid range: `1`–`28800` (8 hours). Changing this value creates a new resource.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `run_hook_payload` - (Optional) String payload (maximum 16 KB) delivered to the `/run` lifecycle hook when the MicroVM starts. Changing this value creates a new resource.

### `idle_policy` Block

The `idle_policy` block supports the following:

* `auto_resume_enabled` - (Required) Whether to automatically resume the MicroVM when it receives a request while suspended.
* `max_idle_duration_seconds` - (Required) Number of seconds without traffic after which the MicroVM is suspended.
* `suspended_duration_seconds` - (Required) Number of seconds a MicroVM remains suspended before it is automatically terminated.

### `logging` Block

The `logging` block supports exactly one of the following:

* `cloud_watch` - (Optional) Send logs to Amazon CloudWatch Logs. [See below](#cloud_watch-block).
* `disabled` - (Optional) Disable logging for the MicroVM. Specify an empty block: `disabled {}`.

### `cloud_watch` Block

The `cloud_watch` block supports the following:

* `log_group` - (Optional) Name of the CloudWatch Logs log group to send logs to.
* `log_stream` - (Optional) Name of the CloudWatch Logs log stream within the log group.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `endpoint` - HTTPS endpoint URL for communicating with the MicroVM.
* `image_arn` - ARN of the MicroVM image used to run this MicroVM.
* `microvm_id` - Unique identifier of the MicroVM.
* `started_at` - Timestamp when the MicroVM first started, in RFC 3339 format.
* `state` - Current lifecycle state of the MicroVM (e.g., `RUNNING`).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_lambdamicrovms_microvm.example
  identity = {
    "microvm_id" = "mvm-01234567-abcd-ef01-2345-6789abcdef01"
  }
}

resource "aws_lambdamicrovms_microvm" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `microvm_id` (String) Unique identifier of the MicroVM.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Lambda MicroVMs MicroVM using the `microvm_id`. For example:

```terraform
import {
  to = aws_lambdamicrovms_microvm.example
  id = "mvm-01234567-abcd-ef01-2345-6789abcdef01"
}
```

Using `terraform import`, import Lambda MicroVMs MicroVM using the `microvm_id`. For example:

```console
% terraform import aws_lambdamicrovms_microvm.example mvm-01234567-abcd-ef01-2345-6789abcdef01
```
