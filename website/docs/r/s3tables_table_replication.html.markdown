---
subcategory: "S3 Tables"
layout: "aws"
page_title: "AWS: aws_s3tables_table_replication"
description: |-
  Manages Amazon S3 Tables Table Replication configuration.
---

# Resource: aws_s3tables_table_replication

Manages Amazon S3 Tables Table Replication configuration.

## Example Usage

### Basic Usage

```terraform
resource "aws_s3tables_table_replication" "example" {
  table_arn = aws_s3tables_table.example.arn
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `table_arn` - (Required, Forces new resource) ARN referencing the Table that owns this replication configuration.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import S3 Tables Table Replication using the `table_arn`. For example:

```terraform
import {
  to = aws_s3tables_table_replication.example
  id = "arn:aws:s3tables:us-west-2:123456789012:table/example-table"
}
```

Using `terraform import`, import S3 Tables Table Replication using the `table_arn`. For example:

```console
% terraform import aws_s3tables_table_replication.example 'arn:aws:s3tables:us-west-2:123456789012:table/example-table'
```