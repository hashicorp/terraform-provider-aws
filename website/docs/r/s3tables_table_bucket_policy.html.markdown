---
subcategory: "S3 Tables"
layout: "aws"
page_title: "AWS: aws_s3tables_table_bucket_policy"
description: |-
  Terraform resource for managing an Amazon S3 Tables Table Bucket Policy.
---

# Resource: aws_s3tables_table_bucket_policy

Terraform resource for managing an Amazon S3 Tables Table Bucket Policy.

## Example Usage

### Basic Usage

```terraform
resource "aws_s3tables_table_bucket_policy" "example" {
  resource_policy  = data.aws_iam_policy_document.example.json
  table_bucket_arn = aws_s3tables_table_bucket.example.arn
}

data "aws_iam_policy_document" "example" {
  statement {
    # ...
  }
}

resource "aws_s3tables_table_bucket" "test" {
  name = "example-bucket"
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `resource_policy` - (Required) Amazon Web Services resource-based policy document in JSON format.
* `table_bucket_arn` - (Required, Forces new resource) ARN referencing the Table Bucket that owns this policy.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import S3 Tables Table Bucket Policy using the `table_bucket_arn`. For example:

```terraform
import {
  to = aws_s3tables_table_bucket_policy.example
  id = "arn:aws:s3tables:us-west-2:123456789012:bucket/example-bucket;example-namespace"
}
```

Using `terraform import`, import S3 Tables Table Bucket Policy using the `table_bucket_arn`. For example:

```console
% terraform import aws_s3tables_table_bucket_policy.example 'arn:aws:s3tables:us-west-2:123456789012:bucket/example-bucket;example-namespace'
```
