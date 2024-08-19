---
subcategory: "QuickSight"
layout: "aws"
page_title: "AWS: aws_quicksight_user"
description: |-
  Use this data source to fetch information about a QuickSight User.
---

# Data Source: aws_quicksight_user

This data source can be used to fetch information about a specific
QuickSight user. By using this data source, you can reference QuickSight user
properties without having to hard code ARNs or unique IDs as input.

## Example Usage

### Basic Usage

```terraform
data "aws_quicksight_user" "example" {
  user_name = "example"
}
```

## Argument Reference

The following arguments are required:

* `user_name` - (Required) The name of the user that you want to match.

The following arguments are optional:

* `aws_account_id` - (Optional) AWS account ID.
* `namespace` - (Optional) QuickSight namespace. Defaults to `default`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `active` - The active status of user. When you create an Amazon QuickSight user that’s not an IAM user or an Active Directory user, that user is inactive until they sign in and provide a password.
* `arn` - The Amazon Resource Name (ARN) for the user.
* `email` - The user's email address.
* `identity_type` - The type of identity authentication used by the user.
* `principal_id` - The principal ID of the user.
* `user_role` - The Amazon QuickSight role for the user. The user role can be one of the following:.
    - `READER`: A user who has read-only access to dashboards.
    - `AUTHOR`: A user who can create data sources, datasets, analyzes, and dashboards.
    - `ADMIN`: A user who is an author, who can also manage Amazon QuickSight settings.
