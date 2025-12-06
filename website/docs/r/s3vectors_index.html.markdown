---
subcategory: "S3 Vectors"
layout: "aws"
page_title: "AWS: aws_s3vectors_index"
description: |-
  Terraform resource for managing an Amazon S3 Vectors Index.
---

# Resource: aws_s3vectors_index

Terraform resource for managing an Amazon S3 Vectors Index.

## Example Usage

### Basic Usage

```terraform
resource "aws_s3vectors_index" "example" {
  index_name         = "example-index"
  vector_bucket_name = aws_s3vectors_vector_bucket.example.vector_bucket_name

  data_type       = "float32"
  dimension       = 2
  distance_metric = "euclidean"
}
```

## Argument Reference

The following arguments are required:

* `data_type` - (Required, Forces new resource) Data type of the vectors to be inserted into the vector index. Valid values: `float32`.
* `dimension` - (Required, Forces new resource) Dimensions of the vectors to be inserted into the vector index.
* `distance_metric` - (Required, Forces new resource) Distance metric to be used for similarity search. Valid values: `cosine`, `euclidean`.
* `index_name` - (Required, Forces new resource) Name of the vector index.
* `vector_bucket_name` - (Required, Forces new resource) Name of the vector bucket for the vector index.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `encryption_configuration` - (Optional, Forces new resource) Block for encryption configuration for the vector index. See [`encyption_configuration` block](#encyption_configuration-block) below.
* `metadata_configuration` - (Optional, Forces new resource) Block for metadata configuration for the vector index. See [`metadata_configuration` block](#metadata_configuration-block) below.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `encyption_configuration` block

The `encryption_configuration` block supports the following attributes:

* `kms_key_id` - (Optional, Forces new resource) AWS Key Management Service (KMS) customer managed key ID to use for the encryption configuration. This parameter is allowed if and only if `sse_type` is set to `aws:kms`.
* `sse_type` - (Optional, Forces new resource) Type of encryption to use. Valid values: `AES256`, `aws:kms`. Defaults to `AES256`.

### `metadata_configuration` block

The `metadata_configuration` block supports the following attributes:

* `non_filterable_metadata_keys` - (Required, Forces new resource) List of non-filterable metadata keys.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `creation_time` - Date and time when the vector index was created.
* `index_arn` - ARN of the vector index.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import S3 Vectors Index using the `index_arn`. For example:

```terraform
import {
  to = aws_s3vectors_index.example
  id = "arn:aws:s3vectors:us-west-2:123456789012:bucket/example-bucket/index/example-index"
}
```

Using `terraform import`, import S3 Vectors Index using the `index_arn`. For example:

```console
% terraform import aws_s3vectors_index.example arn:aws:s3vectors:us-west-2:123456789012:bucket/example-bucket/index/example-index
```
