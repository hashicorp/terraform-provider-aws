---
subcategory: "CloudWatch Logs"
layout: "aws"
page_title: "AWS: aws_cloudwatch_log_delivery"
description: |-
  Terraform resource for managing an AWS CloudWatch Logs Delivery.
---

# Resource: aws_cloudwatch_log_delivery

Terraform resource for managing an AWS CloudWatch Logs Delivery. A delivery is a connection between an `aws_cloudwatch_log_delivery_source` and an `aws_cloudwatch_log_delivery_destination`.

## Example Usage

### Basic Usage

```terraform
resource "aws_cloudwatch_log_delivery" "example" {
  delivery_source_name     = aws_cloudwatch_log_delivery_source.example.name
  delivery_destination_arn = aws_cloudwatch_log_delivery_destination.example.arn

  field_delimiter = ","

  record_fields = ["event_timestamp", "event"]
}
```

## Argument Reference

This resource supports the following arguments:

* `delivery_destination_arn` - (Required) The ARN of the delivery destination to use for this delivery.
* `delivery_source_name` - (Required) The name of the delivery source to use for this delivery.
* `field_delimiter` - (Optional) The field delimiter to use between record fields when the final output format of a delivery is in `plain`, `w3c`, or `raw` format.
* `record_fields` - (Optional) The list of record fields to be delivered to the destination, in order.
* `s3_delivery_configuration` - (Optional) Parameters that are valid only when the delivery's delivery destination is an S3 bucket.
    * `enable_hive_compatible_path` - (Optional) This parameter causes the S3 objects that contain delivered logs to use a prefix structure that allows for integration with Apache Hive.
    * `suffix_path` - (Optional) This string allows re-configuring the S3 object prefix to contain either static or variable sections. The valid variables to use in the suffix path will vary by each log source.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) of the delivery.
* `id` - The unique ID that identifies this delivery in your account.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudWatch Logs Delivery using the `id`. For example:

```terraform
import {
  to = aws_cloudwatch_log_delivery.example
  id = "jsoGVi4Zq8VlYp9n"
}
```

Using `terraform import`, import CloudWatch Logs Delivery using the `id`. For example:

```console
% terraform import aws_cloudwatch_log_delivery.example jsoGVi4Zq8VlYp9n
```
