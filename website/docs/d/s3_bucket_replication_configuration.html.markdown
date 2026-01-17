---
subcategory: "S3 (Simple Storage)"
layout: "aws"
page_title: "AWS: aws_s3_bucket_replication_configuration"
description: |-
  Terraform data source for managing an AWS S3 (Simple Storage) Bucket Replication Configuration.
---

# Data Source: aws_s3_bucket_replication_configuration

Terraform data source for managing an AWS S3 (Simple Storage) Bucket Replication Configuration.

## Example Usage

### Basic Usage

```terraform
data "aws_s3_bucket_replication_configuration" "example" {
  bucket = "example-bucket"
}
```

## Argument Reference

This data source supports the following arguments:

* `bucket` - (Required) The name of the bucket to get the replication configuration for.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `role` – The ARN of the IAM role that Amazon S3 assumes when replicating objects.
* `rule` – An unordered list of configuration blocks that define the rules managing replication.

See the [`aws_s3_replication_configuration`](../r/s3_bucket_replication_configuration.html.markdown#rule) resource documentation for details about the `rule` block (excluding the `and` block within the `filter` block).

### `and`

* `prefix` – The object key name prefix that identifies the subset of objects to which the rule applies.
* `tag` – An unordered list of tags that identify a subset of objects to which the rule applies.
  
    * `key` – The key of the tag.
    * `value` – The value of the tag.
