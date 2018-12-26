---
layout: "aws"
page_title: "AWS: aws_s3_bucket_policy"
sidebar_current: "docs-aws-datasource-s3-bucket-policy"
description: |-
    Provides policy document of an S3 bucket
---

# Data Source: aws_s3_bucket_policy

The S3 policy data source allows access to the policy document of an object stored inside S3 bucket.


## Example Usage

```hcl
data "aws_s3_bucket_object" "policy" {
  bucket = "tf-test-bucket"
}
```

## Argument Reference

The following arguments are supported:

* `bucket` - (Required) The name of the bucket to read the policy from
## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `policy` - The text of the policy attached to the bucket.
