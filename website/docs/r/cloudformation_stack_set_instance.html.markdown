---
subcategory: "CloudFormation"
layout: "aws"
page_title: "AWS: aws_cloudformation_stack_set_instance"
description: |-
  Manages a CloudFormation StackSet Instance.
---

# Resource: aws_cloudformation_stack_set_instance

Manages a CloudFormation StackSet Instance. Instances are managed in the account and region of the StackSet after the target account permissions have been configured. Additional information about StackSets can be found in the [AWS CloudFormation User Guide](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/what-is-cfnstacksets.html).

~> **NOTE:** All target accounts must have an IAM Role created that matches the name of the execution role configured in the StackSet (the `execution_role_name` argument in the `aws_cloudformation_stack_set` resource) in a trust relationship with the administrative account or administration IAM Role. The execution role must have appropriate permissions to manage resources defined in the template along with those required for StackSets to operate. See the [AWS CloudFormation User Guide](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/stacksets-prereqs.html) for more details.

~> **NOTE:** To retain the Stack during Terraform resource destroy, ensure `retain_stack = true` has been successfully applied into the Terraform state first. This must be completed _before_ an apply that would destroy the resource.

## Example Usage

### Basic Usage

```terraform
resource "aws_cloudformation_stack_set_instance" "example" {
  account_id     = "123456789012"
  region         = "us-east-1"
  stack_set_name = aws_cloudformation_stack_set.example.name
}
```

### Example IAM Setup in Target Account

```terraform
data "aws_iam_policy_document" "AWSCloudFormationStackSetExecutionRole_assume_role_policy" {
  statement {
    actions = ["sts:AssumeRole"]
    effect  = "Allow"

    principals {
      identifiers = [aws_iam_role.AWSCloudFormationStackSetAdministrationRole.arn]
      type        = "AWS"
    }
  }
}

resource "aws_iam_role" "AWSCloudFormationStackSetExecutionRole" {
  assume_role_policy = data.aws_iam_policy_document.AWSCloudFormationStackSetExecutionRole_assume_role_policy.json
  name               = "AWSCloudFormationStackSetExecutionRole"
}

# Documentation: https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/stacksets-prereqs.html
# Additional IAM permissions necessary depend on the resources defined in the StackSet template
data "aws_iam_policy_document" "AWSCloudFormationStackSetExecutionRole_MinimumExecutionPolicy" {
  statement {
    actions = [
      "cloudformation:*",
      "s3:*",
      "sns:*",
    ]

    effect    = "Allow"
    resources = ["*"]
  }
}

resource "aws_iam_role_policy" "AWSCloudFormationStackSetExecutionRole_MinimumExecutionPolicy" {
  name   = "MinimumExecutionPolicy"
  policy = data.aws_iam_policy_document.AWSCloudFormationStackSetExecutionRole_MinimumExecutionPolicy.json
  role   = aws_iam_role.AWSCloudFormationStackSetExecutionRole.name
}
```

### Example Deployment across Organizations account

```terraform
resource "aws_cloudformation_stack_set_instance" "example" {
  deployment_targets {
    organizational_unit_ids = [aws_organizations_organization.example.roots[0].id]
  }

  region         = "us-east-1"
  stack_set_name = aws_cloudformation_stack_set.example.name
}
```

## Argument Reference

This resource supports the following arguments:

* `stack_set_name` - (Required) Name of the StackSet.
* `account_id` - (Optional) Target AWS Account ID to create a Stack based on the StackSet. Defaults to current account.
* `deployment_targets` - (Optional) AWS Organizations accounts to which StackSets deploys. StackSets doesn't deploy stack instances to the organization management account, even if the organization management account is in your organization or in an OU in your organization. Drift detection is not possible for this argument. See [deployment_targets](#deployment_targets-argument-reference) below.
* `parameter_overrides` - (Optional) Key-value map of input parameters to override from the StackSet for this Instance.
* `region` - (Optional) Target AWS Region to create a Stack based on the StackSet. Defaults to current region.
* `retain_stack` - (Optional) During Terraform resource destroy, remove Instance from StackSet while keeping the Stack and its associated resources. Must be enabled in Terraform state _before_ destroy operation to take effect. You cannot reassociate a retained Stack or add an existing, saved Stack to a new StackSet. Defaults to `false`.
* `call_as` - (Optional) Specifies whether you are acting as an account administrator in the organization's management account or as a delegated administrator in a member account. Valid values: `SELF` (default), `DELEGATED_ADMIN`.
* `operation_preferences` - (Optional) Preferences for how AWS CloudFormation performs a stack set operation.

### `deployment_targets` Argument Reference

The `deployment_targets` configuration block supports the following arguments:

* `organizational_unit_ids` - (Optional) Organization root ID or organizational unit (OU) IDs to which StackSets deploys.
* `account_filter_type` - (Optional) Limit deployment targets to individual accounts or include additional accounts with provided OUs. Valid values: `INTERSECTION`, `DIFFERENCE`, `UNION`, `NONE`.
* `accounts` - (Optional) List of accounts to deploy stack set updates.
* `accounts_url` - (Optional) S3 URL of the file containing the list of accounts.

### `operation_preferences` Argument Reference

The `operation_preferences` configuration block supports the following arguments:

* `failure_tolerance_count` - (Optional) Number of accounts, per Region, for which this operation can fail before AWS CloudFormation stops the operation in that Region.
* `failure_tolerance_percentage` - (Optional) Percentage of accounts, per Region, for which this stack operation can fail before AWS CloudFormation stops the operation in that Region.
* `max_concurrent_count` - (Optional) Maximum number of accounts in which to perform this operation at one time.
* `max_concurrent_percentage` - (Optional) Maximum percentage of accounts in which to perform this operation at one time.
* `concurrency_mode` - (Optional) Specifies how the concurrency level behaves during the operation execution. Valid values are `STRICT_FAILURE_TOLERANCE` and `SOFT_FAILURE_TOLERANCE`.
* `region_concurrency_type` - (Optional) Concurrency type of deploying StackSets operations in Regions, could be in parallel or one Region at a time. Valid values are `SEQUENTIAL` and `PARALLEL`.
* `region_order` - (Optional) Order of the Regions in where you want to perform the stack operation.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Unique identifier for the resource. If `deployment_targets` is set, this is a comma-delimited string combining stack set name, organizational unit IDs (`/`-delimited), and region (ie. `mystack,ou-123/ou-456,us-east-1`). Otherwise, this is a comma-delimited string combining stack set name, AWS account ID, and region (ie. `mystack,123456789012,us-east-1`).
* `organizational_unit_id` - Organization root ID or organizational unit (OU) ID in which the stack is deployed.
* `stack_id` - Stack identifier.
* `stack_instance_summaries` - List of stack instances created from an organizational unit deployment target. This will only be populated when `deployment_targets` is set. See [`stack_instance_summaries`](#stack_instance_summaries-attribute-reference).

### `stack_instance_summaries` Attribute Reference

* `account_id` - AWS account ID in which the stack is deployed.
* `organizational_unit_id` - Organizational unit ID in which the stack is deployed.
* `stack_id` - Stack identifier.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudFormation StackSet Instances that target an AWS Account ID using the StackSet name, target AWS account ID, and target AWS Region separated by commas (`,`). For example:

```terraform
import {
  to = aws_cloudformation_stack_set_instance.example
  id = "example,123456789012,us-east-1"
}
```

Import CloudFormation StackSet Instances that target AWS Organizational Units using the StackSet name, a slash (`/`) separated list of organizational unit IDs, and target AWS Region separated by commas (`,`). For example:

```terraform
import {
  to = aws_cloudformation_stack_set_instance.example
  id = "example,ou-sdas-123123123/ou-sdas-789789789,us-east-1"
}
```

Import CloudFormation StackSet Instances when acting a delegated administrator in a member account using the StackSet name, target AWS account ID or slash (`/`) separated list of organizational unit IDs, target AWS Region and `call_as` value separated by commas (`,`). For example:

```terraform
import {
  to = aws_cloudformation_stack_set_instance.example
  id = "example,ou-sdas-123123123/ou-sdas-789789789,us-east-1,DELEGATED_ADMIN"
}
```

Using `terraform import`, import CloudFormation StackSet Instances that target an AWS Account ID using the StackSet name, target AWS account ID, and target AWS Region separated by commas (`,`). For example:

```console
% terraform import aws_cloudformation_stack_set_instance.example example,123456789012,us-east-1
```

Using `terraform import`, import CloudFormation StackSet Instances that target AWS Organizational Units using the StackSet name, a slash (`/`) separated list of organizational unit IDs, and target AWS Region separated by commas (`,`). For example:

```console
% terraform import aws_cloudformation_stack_set_instance.example example,ou-sdas-123123123/ou-sdas-789789789,us-east-1
```

Using `terraform import`, import CloudFormation StackSet Instances when acting a delegated administrator in a member account using the StackSet name, target AWS account ID or slash (`/`) separated list of organizational unit IDs, target AWS Region and `call_as` value separated by commas (`,`). For example:

```console
% terraform import aws_cloudformation_stack_set_instance.example example,ou-sdas-123123123/ou-sdas-789789789,us-east-1,DELEGATED_ADMIN
```
