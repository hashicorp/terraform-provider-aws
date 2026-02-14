---
subcategory: "S3 Vectors"
layout: "aws"
page_title: "AWS: aws_s3vectors_vector_bucket"
description: |-
  Terraform resource for managing an Amazon S3 Vectors Vector Bucket.
---

# Resource: aws_s3vectors_vector_bucket

Terraform resource for managing an Amazon S3 Vectors Vector Bucket.

## Example Usage

### Basic Usage

```terraform
resource "aws_s3vectors_vector_bucket" "example" {
  vector_bucket_name = "example-bucket"
}
```

### Encryption

```terraform
resource "aws_s3vectors_vector_bucket" "example" {
  vector_bucket_name = "example-bucket"

  encryption_configuration {
    sse_type    = "aws:kms"
    kms_key_arn = aws_kms_key.example.arn
  }
}
```

## Argument Reference

The following arguments are required:

* `vector_bucket_name` - (Required, Forces new resource) Name of the vector bucket.

The following arguments are optional:

* `encryption_configuration` - (Optional, Forces new resource) Encryption configuration for the vector bucket. See [Encryption Configuration](#encryption-configuration) below for more details.
* `force_destroy` - (Optional, Default:`false`) Boolean that indicates all indexes and vectors should be deleted from the vector bucket *when the vector bucket is destroyed* so that the vector bucket can be destroyed without error. Once this parameter is set to `true`, there must be a successful `terraform apply` run before a destroy is required to update this value in the resource state. Without a successful `terraform apply` after this parameter is set, this flag will have no effect. If setting this field in the same operation that would require replacing the vector bucket or destroying the vector bucket, this flag will not work.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Encryption Configuration

The `encryption_configuration` block supports the following:

* `kms_key_arn` - (Optional, Forces new resource) AWS KMS CMK ARN to use for the default encryption of the vector bucket. Allowed if and only if `sse_type` is set to `aws:kms`.
* `sse_type` - (Optional, Forces new resource) Server-side encryption type to use for the default encryption of the vector bucket. Valid values: `AES256`, `aws:kms`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `creation_time` - Date and time when the vector bucket was created.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `vector_bucket_arn` - ARN of the vector bucket.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import S3 Vectors Vector Bucket using the `vector_bucket_arn`. For example:

```terraform
import {
  to = aws_s3vectors_vector_bucket.example
  id = "arn:aws:s3vectors:us-west-2:123456789012:bucket/example-bucket"
}
```

Using `terraform import`, import S3 Vectors Vector Bucket using the `vector_bucket_arn`. For example:

```console
% terraform import aws_s3vectors_vector_bucket.example arn:aws:s3vectors:us-west-2:123456789012:bucket/example-bucket
```
