---
subcategory: "QuickSight"
layout: "aws"
page_title: "AWS: aws_quicksight_iam_policy_assignment"
description: |-
  Terraform resource for managing an AWS QuickSight IAM Policy Assignment.
---

# Resource: aws_quicksight_iam_policy_assignment

Terraform resource for managing an AWS QuickSight IAM Policy Assignment.

## Example Usage

### Basic Usage

```terraform
resource "aws_quicksight_iam_policy_assignment" "example" {
  assignment_name   = "example"
  assignment_status = "ENABLED"
  policy_arn        = aws_iam_policy.example.arn
  identities {
    user = [aws_quicksight_user.example.user_name]
  }
}
```

## Argument Reference

The following arguments are required:

* `assignment_name` - (Required) Name of the assignment.
* `assignment_status` - (Required) Status of the assignment. Valid values are `ENABLED`, `DISABLED`, and `DRAFT`.

The following arguments are optional:

* `aws_account_id` - (Optional) AWS account ID.
* `identities` - (Optional) Amazon QuickSight users, groups, or both to assign the policy to. See [`identities`](#identities).
* `namespace` - (Optional) Namespace that contains the assignment. Defaults to `default`.
* `policy_arn` - (Optional) ARN of the IAM policy to apply to the Amazon QuickSight users and groups specified in this assignment.

### identities

* `groups` - (Optional) Array of Quicksight group names to assign the policy to.
* `user` - (Optional) Array of Quicksight user names to assign the policy to.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `assignment_id` - Assignment ID.
* `id` - A comma-delimited string joining AWS account ID, namespace, and assignment name.

## Import

QuickSight IAM Policy Assignment can be imported using the AWS account ID, namespace, and assignment name separated by commas (`,`) e.g.,

```
$ terraform import aws_quicksight_iam_policy_assignment.example 123456789012,default,example
```
