---
subcategory: "CloudWatch Omni"
layout: "aws"
page_title: "AWS: aws_cloudwatchomni_access_grant"
description: |-
  Lists CloudWatch Omni Access Grant resources.
---

# List Resource: aws_cloudwatchomni_access_grant

Lists CloudWatch Omni Access Grant resources. Service-managed access grants are not listed.

## Example Usage

```terraform
list "aws_cloudwatchomni_access_grant" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region to query. Defaults to provider region.
