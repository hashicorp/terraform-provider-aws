---
subcategory: "S3 (Simple Storage)"
layout: "aws"
page_title: "AWS: aws_s3_bucket_policy"
description: |-
    Provides IAM policy of an S3 bucket
---

# Data Source: aws_s3_bucket_policy

The bucket policy data source returns IAM policy of an S3 bucket.

## Example Usage

The following example retrieves IAM policy of a specified S3 bucket.

```terraform
data "aws_s3_bucket_policy" "example" {
  bucket = "example-bucket-name"
}

output "foo" {
  value = data.aws_s3_bucket_policy.example.policy
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `bucket` - (Required) Bucket name.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `policy` - IAM bucket policy.
