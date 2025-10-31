---
subcategory: "S3 Vectors"
layout: "aws"
page_title: "AWS: aws_s3vectors_vector_bucket_policy"
description: |-
  Terraform resource for managing an Amazon S3 Vectors Vector Bucket policy.
---

# Resource: aws_s3vectors_vector_bucket_policy

Terraform resource for managing an Amazon S3 Vectors Vector Bucket policy.

## Example Usage

### Basic Usage

```terraform
resource "aws_s3vectors_vector_bucket_policy" "example" {
  vector_bucket_arn = aws_s3vectors_vector_bucket.example.arn

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Id": "writePolicy",
  "Statement": [{
    "Sid": "writeStatement",
    "Effect": "Allow",
    "Principal": {
      "AWS": "123456789012"
    },
    "Action": [
      "s3vectors:PutVectors"
    ],
    "Resource": "*"
  }]
}
EOF
}
```

## Argument Reference

The following arguments are required:

* `policy` - (Required) The policy document.
* `vector_bucket_arn` - (Required, Forces new resource) ARN of the vector bucket.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import S3 Vectors Vector Bucket policy using the `vector_bucket_arn`. For example:

```terraform
import {
  to = aws_s3vectors_vector_bucket_policy.example
  id = "arn:aws:s3vectors:us-west-2:123456789012:bucket/example-bucket"
}
```

Using `terraform import`, import S3 Vectors Vector Bucket policy using the `vector_bucket_arn`. For example:

```console
% terraform import aws_s3vectors_vector_bucket_policy.example arn:aws:s3vectors:us-west-2:123456789012:bucket/example-bucket
```
