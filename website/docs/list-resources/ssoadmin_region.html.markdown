---
subcategory: "SSO Admin"
layout: "aws"
page_title: "AWS: aws_ssoadmin_region"
description: |-
  Lists SSO Admin Region resources.
---

# List Resource: aws_ssoadmin_region

Lists SSO Admin Region resources.

## Example Usage

```terraform
list "aws_ssoadmin_region" "example" {
  provider = aws

  config {
    instance_arn = tolist(data.aws_ssoadmin_instances.example.arns)[0]
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `instance_arn` - (Required) ARN of the IAM Identity Center instance to list Regions from.
* `region` - (Optional) Region to query. Defaults to provider region.
