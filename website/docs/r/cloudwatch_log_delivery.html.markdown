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

### CloudFront Standard Logging (v2)

CloudFront delivers access logs through CloudWatch Logs, so a distribution's standard logging (v2) configuration is expressed as a delivery source, a delivery destination, and a delivery. The `record_fields` list selects the access log fields, including `viewer-request-log-data` and `viewer-response-log-data`, which carry the custom data that a viewer request or viewer response [CloudFront Function](cloudfront_function.html) logs with `cf.logCustomData()`.

```terraform
resource "aws_cloudwatch_log_delivery_source" "example" {
  name         = "cloudfront-access-logs"
  log_type     = "ACCESS_LOGS"
  resource_arn = aws_cloudfront_distribution.example.arn
}

resource "aws_cloudwatch_log_delivery_destination" "example" {
  name          = "cloudfront-access-logs"
  output_format = "json"

  delivery_destination_configuration {
    destination_resource_arn = aws_cloudwatch_log_group.example.arn
  }
}

resource "aws_cloudwatch_log_delivery" "example" {
  delivery_source_name     = aws_cloudwatch_log_delivery_source.example.name
  delivery_destination_arn = aws_cloudwatch_log_delivery_destination.example.arn

  record_fields = [
    "date",
    "time",
    "c-ip",
    "sc-status",
    "viewer-request-log-data",
    "viewer-response-log-data",
  ]
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `delivery_destination_arn` - (Required) The ARN of the delivery destination to use for this delivery.
* `delivery_source_name` - (Required) The name of the delivery source to use for this delivery.
* `field_delimiter` - (Optional) The field delimiter to use between record fields when the final output format of a delivery is in `plain`, `w3c`, or `raw` format.
* `record_fields` - (Optional) The list of record fields to be delivered to the destination, in order. The valid field names vary by the `log_type` of the delivery source. For a CloudFront `ACCESS_LOGS` source, see [Configure standard logging (v2)](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/standard-logging.html#standard-logging-real-time-log-selection) for the supported values.
* `s3_delivery_configuration` - (Optional) Parameters that are valid only when the delivery's delivery destination is an S3 bucket.
    * `enable_hive_compatible_path` - (Optional) This parameter causes the S3 objects that contain delivered logs to use a prefix structure that allows for integration with Apache Hive.
    * `suffix_path` - (Optional) This string allows re-configuring the S3 object prefix to contain either static or variable sections. The valid variables to use in the suffix path will vary by each log source. **Note:** AWS automatically prepends account and service-specific prefixes (e.g., `AWSLogs/{account-id}/CloudFront/` for CloudFront sources) to the configured value. Specify only your custom suffix path without these AWS-managed prefixes.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) of the delivery.
* `id` - The unique ID that identifies this delivery in your account.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_cloudwatch_log_delivery.example
  identity = {
    id = "jsoGVi4Zq8VlYp9n"
  }
}

resource "aws_cloudwatch_log_delivery" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `id` (String) ID of the delivery.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Deliveries using `id`. For example:

```terraform
import {
  to = aws_cloudwatch_log_delivery.example
  id = "jsoGVi4Zq8VlYp9n"
}
```

Using `terraform import`, import Deliveries using `id`. For example:

```console
% terraform import aws_cloudwatch_log_delivery.example jsoGVi4Zq8VlYp9n
```
