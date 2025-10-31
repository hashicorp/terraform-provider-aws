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

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `creation_time` - Date and time when the vector index was created.
* `index_arn` - ARN of the vector index.

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
