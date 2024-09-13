---
subcategory: "S3 (Simple Storage)"
layout: "aws"
page_title: "AWS: aws_canonical_user_id"
description: |-
  Provides the canonical user ID for the AWS account associated with the provider
  connection to AWS.
---

# Data Source: aws_canonical_user_id

The Canonical User ID data source allows access to the [canonical user ID](http://docs.aws.amazon.com/general/latest/gr/acct-identifiers.html)
for the effective account in which Terraform is working.  

~> **NOTE:** To use this data source, you must have the `s3:ListAllMyBuckets` permission.

## Example Usage

```terraform
data "aws_canonical_user_id" "current" {}

output "canonical_user_id" {
  value = data.aws_canonical_user_id.current.id
}
```

## Argument Reference

There are no arguments available for this data source.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - Canonical user ID associated with the AWS account.

* `display_name` - Human-friendly name linked to the canonical user ID. The bucket owner's display name. **NOTE:** [This value](https://docs.aws.amazon.com/AmazonS3/latest/API/RESTServiceGET.html) is only included in the response in the US East (N. Virginia), US West (N. California), US West (Oregon), Asia Pacific (Singapore), Asia Pacific (Sydney), Asia Pacific (Tokyo), EU (Ireland), and South America (São Paulo) regions.
