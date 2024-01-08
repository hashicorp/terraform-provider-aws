---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_access_keys"
description: |-
  Get information on IAM access keys associated with the specified IAM user.
---

# Data Source: aws_iam_access_keys

This data source can be used to fetch information about IAM access keys of a
specific IAM user.

## Example Usage

```terraform
data "aws_iam_access_keys" "example" {
  user = "an_example_user_name"
}
```

## Argument Reference

The following arguments are required:

* `user` - (Required) Name of the IAM user associated with the access keys.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `access_keys` - List of the IAM access keys associated with the specified user. See below.

The elements of the `access_keys` are exported with the following attributes:

* `access_key_id` - Access key ID.
* `create_date` - Date and time in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8) that the access key was created.
* `status` - Access key status. Possible values are `Active` and `Inactive`.
