---
subcategory: "S3 Tables"
layout: "aws"
page_title: "AWS: aws_s3tables_table_bucket_metrics_configuration"
description: |-
  Terraform resource for managing an Amazon S3 Tables Table Bucket request metrics configuration.
---

# Resource: aws_s3tables_table_bucket_metrics_configuration

Terraform resource for managing an Amazon S3 Tables Table Bucket request metrics configuration.

## Example Usage

### Basic Usage

```terraform
resource "aws_s3tables_table_bucket_metrics_configuration" "example" {
  table_bucket_arn = aws_s3tables_table_bucket.example.arn
}

resource "aws_s3tables_table_bucket" "example" {
  name = "example-bucket"
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `table_bucket_arn` - (Required, Forces new resource) ARN referencing the Table Bucket that owns this request metrics configuration.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `metrics_configuration_id` - Unique identifier of the metrics configuration.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import S3 Tables Table Bucket Metrics Configuration using the `table_bucket_arn`. For example:

```terraform
import {
  to = aws_s3tables_table_bucket_metrics_configuration.example
  id = "arn:aws:s3tables:us-west-2:123456789012:bucket/example-bucket"
}
```

Using `terraform import`, import S3 Tables Table Bucket Metrics Configuration using the `table_bucket_arn`. For example:

```console
% terraform import aws_s3tables_table_bucket_metrics_configuration.example arn:aws:s3tables:us-west-2:123456789012:bucket/example-bucket
```
