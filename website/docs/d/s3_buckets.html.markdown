---
subcategory: "S3 (Simple Storage)"
layout: "aws"
page_title: "AWS: aws_s3_buckets"
description: |-
  Get information about a set of S3 Buckets.
---

# Data Source: aws_s3_buckets

Use this data source to get the IDs of S3 Buckets.

## Example Usage

### All buckets in an account

```terraform
data "aws_s3_buckets" "buckets" {}
```

### Buckets filtered by name regex

Buckets whose role-name contains `project`

```terraform
data "aws_s3_buckets" "buckets" {
  name_regex = ".*project.*"
}
```

## Argument Reference

The following arguments are supported:

* `name_regex` - (Optional) A regex string to apply to the S3 buckets list returned by AWS. This allows more advanced filtering not supported from the AWS API. This filtering is done locally on what AWS returns, and could have a performance impact if the result is large.

## Attributes Reference

* `ids` - Set of IDs of the matched S3 Buckets.
