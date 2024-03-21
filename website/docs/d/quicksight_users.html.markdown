---
subcategory: "QuickSight"
layout: "aws"
page_title: "AWS: aws_quicksight_users"
description: |-
  Use this data source to fetch information about a QuickSight Users.
---

# Data Source: aws_quicksight_users

This data source can be used to fetch information about all QuickSight users.
The list of users can be narrow down by specific filter parameter or by combining multiple parameters.

## Example Usage

### Basic Usage

```terraform
data "aws_quicksight_users" "example" {}
```

### With `filter`

```terraform
data "aws_quicksight_users" "example" {
  filter {
    active          = true
    email_regex     = ".*@example.com"
    identity_type   = "IAM"
    user_name_regex = "IAMRoleName/.*"
    user_role       = "ADMIN"
  }
}
```

## Argument Reference

* `filter` - (Optional) Configuration block for filtering. Detailed below.

### filter Configuration block

* `active` - (Optional) The active status of user. When you create an Amazon QuickSight user that’s not an IAM user or an Active Directory user, that user is inactive until they sign in and provide a password. Valid values: `true` or `false`.
* `email_regex` - (Optional) Regex string to apply to the user email address returned by AWS.. This allows more advanced filtering not supported from the AWS API. This filtering is done locally on what AWS returns, and could have a performance impact if the result is large. Combine this with other options to narrow down the list AWS returns.
* `identity_type` - (Optional) The type of identity authentication used by the user. Valid values: `IAM`, `QUICKSIGHT`.
* `user_name_regex` - (Optional) Regex string to apply to the users name list returned by AWS. This allows more advanced filtering not supported from the AWS API. This filtering is done locally on what AWS returns, and could have a performance impact if the result is large. Combine this with other options to narrow down the list AWS returns.
* `user_role` - (Optional) The Amazon QuickSight role for the user. Valid values: `ADMIN`, `AUTHOR`, `READER`

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `aws_account_id` - AWS account ID.
* `namespace` - Name of the namespace ( by default is: `default` )
* `filter` - Block with the enabled filters to narrow down the list of users.
* `users` - List of users.

### Users

* `active` - The active status of user. When you create an Amazon QuickSight user that’s not an IAM user or an Active Directory user, that user is inactive until they sign in and provide a password.
* `arn` - The Amazon Resource Name (ARN) for the user.
* `email` - The user's email address.
* `identity_type` - The type of identity authentication used by the user.
* `principal_id` - The principal ID of the user.
* `user_name` - The user name. The value: `N\A` means that user identity type is `IAM` and the IAM role used for authenticate to the Quickservice has been deleted.  
* `user_role` - The Amazon QuickSight role for the user. The user role can be one of the following:.
    - `READER`: A user who has read-only access to dashboards.
    - `AUTHOR`: A user who can create data sources, datasets, analyses, and dashboards.
    - `ADMIN`: A user who is an author, who can also manage Amazon QuickSight settings.
