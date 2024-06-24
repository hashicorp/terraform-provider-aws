---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_instance_profiles"
description: |-
  Get information on a Amazon IAM Instance Profiles from IAM role
---

# Data Source: aws_iam_instance_profiles

This data source can be used to fetch information about all
IAM instance profiles under a role. By using this data source, you can reference IAM
instance profile properties without having to hard code ARNs as input.

## Example Usage

```terraform
data "aws_iam_instance_profiles" "example" {
  role_name = "an_example_iam_role_name"
}
```

## Argument Reference

* `role_name` - (Required) IAM role name.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arns` - Set of ARNs of instance profiles.
* `names` - Set of IAM instance profile names.
* `paths` - Set of IAM instance profile paths.
