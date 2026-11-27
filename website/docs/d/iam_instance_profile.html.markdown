---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_instance_profile"
description: |-
  Get information on a Amazon IAM Instance Profile
---

# Data Source: aws_iam_instance_profile

This data source can be used to fetch information about a specific
IAM instance profile. By using this data source, you can reference IAM
instance profile properties without having to hard code ARNs as input.

## Example Usage

```terraform
data "aws_iam_instance_profile" "example" {
  name = "an_example_instance_profile_name"
}
```

## Argument Reference

This data source supports the following arguments:

* `account` - (Optional) AWS account ID where this data source will be read. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `name` - (Required) Friendly IAM instance profile name to match.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN.
* `create_date` - String representation of the date the instance profile was created.
* `path` - Path to the instance profile.
* `role_arn` - Role ARN associated with this instance profile.
* `role_id` - Role ID associated with this instance profile.
* `role_name` - Role name associated with this instance profile.
