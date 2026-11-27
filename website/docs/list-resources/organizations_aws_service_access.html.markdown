---
subcategory: "Organizations"
layout: "aws"
page_title: "AWS: aws_organizations_aws_service_access"
description: |-
  Lists Organizations AWS Service Access resources.
---

# List Resource: aws_organizations_aws_service_access

Lists Organizations AWS Service Access resources.

## Example Usage

```terraform
list "aws_organizations_aws_service_access" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
