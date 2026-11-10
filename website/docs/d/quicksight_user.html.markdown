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

* `user_name` - (Required) Name of the user that you want to match.

The following arguments are optional:

* `aws_account_id` - (Optional) AWS account ID. Defaults to automatically determined account ID of the Terraform AWS provider.
* `namespace` - (Optional) QuickSight namespace. Defaults to `default`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `active` - Active status of user. When you create an Amazon QuickSight user that’s not an IAM user or an Active Directory user, that user is inactive until they sign in and provide a password.
* `arn` - ARN for the user.
* `custom_permissions_name` - Custom permissions profile associated with this user.
* `email` - User's email address.
* `identity_type` - Type of identity authentication used by the user.
* `principal_id` - Principal ID of the user.
* `user_role` - Amazon QuickSight role for the user. Valid values are `READER` (read-only access to dashboards), `AUTHOR` (can create data sources, datasets, analyses, and dashboards), and `ADMIN` (an author who can also manage Amazon QuickSight settings).
