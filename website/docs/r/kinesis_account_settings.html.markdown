---
subcategory: "Kinesis"
layout: "aws"
page_title: "AWS: aws_kinesis_account_settings"
description: |-
  Manages account-level settings for Amazon Kinesis Data Streams.
---

# Resource: aws_kinesis_account_settings

Manages account-level settings for Amazon Kinesis Data Streams.

~> Deletion of this resource will not modify any settings, only remove the resource from state.

## Example Usage

```terraform
resource "aws_kinesis_account_settings" "example" {
  minimum_throughput_billing_commitment {
    status = "ENABLED"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `minimum_throughput_billing_commitment` - (Optional) Minimum throughput billing commitment configuration. Detailed below.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `minimum_throughput_billing_commitment` Block

* `status` - (Required) Desired status of the minimum throughput billing commitment. Valid values: `ENABLED`, `DISABLED`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `minimum_throughput_billing_commitment[0].earliest_allowed_end_at` - Earliest timestamp when the commitment can be ended.
* `minimum_throughput_billing_commitment[0].ended_at` - Timestamp when the commitment was ended.
* `minimum_throughput_billing_commitment[0].started_at` - Timestamp when the commitment was started.
* `minimum_throughput_billing_commitment[0].status_actual` - Current status of the minimum throughput billing commitment. Values: `ENABLED`, `DISABLED`, `ENABLED_UNTIL_EARLIEST_ALLOWED_END`.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_kinesis_account_settings.example
  identity = {
    region = "us-west-2"
  }
}

resource "aws_kinesis_account_settings" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Kinesis account settings using the region name. For example:

```terraform
import {
  to = aws_kinesis_account_settings.example
  id = "us-west-2"
}
```

Using `terraform import`, import Kinesis account settings using the region name. For example:

```console
% terraform import aws_kinesis_account_settings.example us-west-2
```
