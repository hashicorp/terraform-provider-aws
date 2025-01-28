---
subcategory: "CloudWatch Logs"
layout: "aws"
page_title: "AWS: aws_cloudwatch_log_delivery_destination"
description: |-
  Terraform resource for managing an AWS CloudWatch Logs Delivery Destination.
---

# Resource: aws_cloudwatch_log_delivery_destination

Terraform resource for managing an AWS CloudWatch Logs Delivery Destination.

## Example Usage

### Basic Usage

```terraform
resource "aws_cloudwatch_log_delivery_destination" "example" {
  name = "example"

  delivery_destination_configuration {
    destination_resource_arn = aws_cloudwatch_log_group.example.arn
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `delivery_destination_configuration` - (Required) The AWS resource that will receive the logs.
    * `destination_resource_arn` - (Required) The ARN of the AWS destination that this delivery destination represents.
* `name` - (Required) The name for this delivery destination.
* `output_format` - (Optional) The format of the logs that are sent to this delivery destination. Valid values: `json`, `plain`, `w3c`, `raw`, `parquet`.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) of the delivery destination.
* `delivery_destination_type` - Whether this delivery destination is CloudWatch Logs, Amazon S3, or Firehose.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudWatch Logs Delivery Destination using the `name`. For example:

```terraform
import {
  to = aws_cloudwatch_log_delivery_destination.example
  id = "example"
}
```

Using `terraform import`, import CloudWatch Logs Delivery Destination using the `name`. For example:

```console
% terraform import aws_cloudwatch_log_delivery_destination.example example
```
