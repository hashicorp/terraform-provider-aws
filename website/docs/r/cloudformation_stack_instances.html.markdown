---
subcategory: "CloudFormation"
layout: "aws"
page_title: "AWS: aws_cloudformation_stack_instances"
description: |-
  Manages CloudFormation stack instances.
---

# Resource: aws_cloudformation_stack_instances

Manages CloudFormation stack instances for the specified accounts, within the specified regions. A stack instance refers to a stack in a specific account and region. Additional information about stacks can be found in the [AWS CloudFormation User Guide](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/stacks.html).

~> **NOTE:** This resource will manage all stack instances for the specified `stack_set_name`. If you create stack instances outside of Terraform or import existing infrastructure, ensure that your configuration includes all accounts and regions where stack instances exist for the stack set. Failing to include all accounts and regions will cause Terraform to continuously report differences between your configuration and the actual infrastructure.

~> **NOTE:** All target accounts must have an IAM Role created that matches the name of the execution role configured in the stack (the `execution_role_name` argument in the `aws_cloudformation_stack_set` resource) in a trust relationship with the administrative account or administration IAM Role. The execution role must have appropriate permissions to manage resources defined in the template along with those required for stacks to operate. See the [AWS CloudFormation User Guide](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/stacksets-prereqs.html) for more details.

~> **NOTE:** To retain the Stack during Terraform resource destroy, ensure `retain_stacks = true` has been successfully applied into the Terraform state first. This must be completed _before_ an apply that would destroy the resource.

## Example Usage

### Basic Usage

```terraform
resource "aws_cloudformation_stack_instances" "example" {
  accounts       = ["123456789012", "234567890123"]
  regions        = ["us-east-1", "us-west-2"]
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
resource "aws_cloudformation_stack_instances" "example" {
  deployment_targets {
    organizational_unit_ids = [aws_organizations_organization.example.roots[0].id]
  }

  regions        = ["us-west-2", "us-east-1"]
  stack_set_name = aws_cloudformation_stack_set.example.name
}
```

## Argument Reference

The following arguments are required:

* `stack_set_name` - (Required, Force new) Name of the stack set.

The following arguments are optional:

* `accounts` - (Optional) Accounts where you want to create stack instances in the specified `regions`. You can specify either `accounts` or `deployment_targets`, but not both.
* `deployment_targets` - (Optional) AWS Organizations accounts for which to create stack instances in the `regions`. stack sets doesn't deploy stack instances to the organization management account, even if the organization management account is in your organization or in an OU in your organization. Drift detection is not possible for most of this argument. See [deployment_targets](#deployment_targets) below.
* `parameter_overrides` - (Optional) Key-value map of input parameters to override from the stack set for these instances. This argument's drift detection is limited to the first account and region since each instance can have unique parameters.
* `regions` - (Optional) Regions where you want to create stack instances in the specified `accounts`.
* `retain_stacks` - (Optional) Whether to remove the stack instances from the stack set, but not delete the stacks. You can't reassociate a retained stack or add an existing, saved stack to a new stack set. To retain the stack, ensure `retain_stacks = true` has been successfully applied _before_ an apply that would destroy the resource. Defaults to `false`.
* `call_as` - (Optional) Whether you are acting as an account administrator in the organization's management account or as a delegated administrator in a member account. Valid values: `SELF` (default), `DELEGATED_ADMIN`.
* `operation_preferences` - (Optional) Preferences for how AWS CloudFormation performs a stack set operation. See [operation_preferences](#operation_preferences) below.

### `deployment_targets`

The `deployment_targets` configuration block supports the following arguments:

* `account_filter_type` - (Optional, Force new) Limit deployment targets to individual accounts or include additional accounts with provided OUs. Valid values: `INTERSECTION`, `DIFFERENCE`, `UNION`, `NONE`.
* `accounts` - (Optional) List of accounts to deploy stack set updates.
* `accounts_url` - (Optional) S3 URL of the file containing the list of accounts.
* `organizational_unit_ids` - (Optional) Organization root ID or organizational unit (OU) IDs to which stack sets deploy.

### `operation_preferences`

The `operation_preferences` configuration block supports the following arguments:

* `concurrency_mode` - (Optional) How the concurrency level behaves during the operation execution. Valid values are `STRICT_FAILURE_TOLERANCE` and `SOFT_FAILURE_TOLERANCE`.
* `failure_tolerance_count` - (Optional) Number of accounts, per region, for which this operation can fail before CloudFormation stops the operation in that region.
* `failure_tolerance_percentage` - (Optional) Percentage of accounts, per region, for which this stack operation can fail before CloudFormation stops the operation in that region.
* `max_concurrent_count` - (Optional) Maximum number of accounts in which to perform this operation at one time.
* `max_concurrent_percentage` - (Optional) Maximum percentage of accounts in which to perform this operation at one time.
* `region_concurrency_type` - (Optional) Concurrency type of deploying stack sets operations in regions, could be in parallel or one region at a time. Valid values are `SEQUENTIAL` and `PARALLEL`.
* `region_order` - (Optional) Order of the regions where you want to perform the stack operation.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `stack_instance_summaries` - List of stack instances created from an organizational unit deployment target. This may not always be set depending on whether CloudFormation returns summaries for your configuration. See [`stack_instance_summaries`](#stack_instance_summaries-attribute-reference).
* `stack_set_id` - Unique identifier of the stack set.

### `stack_instance_summaries`

* `account_id` - Account ID in which the instance is deployed.
* `detailed_status` - Detailed status of the stack instance. Values include `PENDING`, `RUNNING`, `SUCCEEDED`, `FAILED`, `CANCELLED`, `INOPERABLE`, `SKIPPED_SUSPENDED_ACCOUNT`, `FAILED_IMPORT`.
* `drift_status` - Status of the stack instance's actual configuration compared to the expected template and parameter configuration of the stack set to which it belongs. Values include `DRIFTED`, `IN_SYNC`, `UNKNOWN`, `NOT_CHECKED`.
* `organizational_unit_id` - Organization root ID or organizational unit (OU) IDs that you specified for `deployment_targets`.
* `region` - Region that the stack instance is associated with.
* `stack_id` - ID of the stack instance.
* `stack_set_id` - Name or unique ID of the stack set that the stack instance is associated with.
* `status` - Status of the stack instance, in terms of its synchronization with its associated stack set. Values include `CURRENT`, `OUTDATED`, `INOPERABLE`.
* `status_reason` - Explanation for the specific status code assigned to this stack instance.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudFormation stack instances using the stack set name and `call_as` separated by commas (`,`). If you are importing a stack instance targeting OUs, see the example below. For example:

```terraform
import {
  to = aws_cloudformation_stack_instances.example
  id = "example,SELF"
}
```

Import CloudFormation stack instances that target OUs, using the stack set name, `call_as`, and "OU" separated by commas (`,`). For example:

```terraform
import {
  to = aws_cloudformation_stack_instances.example
  id = "example,SELF,OU"
}
```

Using `terraform import`, import CloudFormation stack instances using the stack set name and `call_as` separated by commas (`,`). If you are importing a stack instance targeting OUs, see the example below. For example:

```console
% terraform import aws_cloudformation_stack_instances.example example,SELF
```

Using `terraform import`, Import CloudFormation stack instances that target OUs, using the stack set name, `call_as`, and "OU" separated by commas (`,`). For example:

```console
% terraform import aws_cloudformation_stack_instances.example example,SELF,OU
```
