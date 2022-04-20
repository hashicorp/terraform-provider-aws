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

The following arguments are supported:

* `bucket` - (Required) The bucket name.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `policy` - IAM bucket policy.
