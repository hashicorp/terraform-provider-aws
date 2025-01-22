---
subcategory: "CloudFormation"
layout: "aws"
page_title: "AWS: aws_cloudformation_stack_set"
description: |-
  Manages a CloudFormation StackSet.
---

# Resource: aws_cloudformation_stack_set

Manages a CloudFormation StackSet. StackSets allow CloudFormation templates to be easily deployed across multiple accounts and regions via StackSet Instances ([`aws_cloudformation_stack_set_instance` resource](/docs/providers/aws/r/cloudformation_stack_set_instance.html)). Additional information about StackSets can be found in the [AWS CloudFormation User Guide](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/what-is-cfnstacksets.html).

~> **NOTE:** All template parameters, including those with a `Default`, must be configured or ignored with the `lifecycle` configuration block `ignore_changes` argument.

~> **NOTE:** All `NoEcho` template parameters must be ignored with the `lifecycle` configuration block `ignore_changes` argument.

~> **NOTE:** When using a delegated administrator account, ensure that your IAM User or Role has the `organizations:ListDelegatedAdministrators` permission. Otherwise, you may get an error like `ValidationError: Account used is not a delegated administrator`.

## Example Usage

```terraform
data "aws_iam_policy_document" "AWSCloudFormationStackSetAdministrationRole_assume_role_policy" {
  statement {
    actions = ["sts:AssumeRole"]
    effect  = "Allow"

    principals {
      identifiers = ["cloudformation.amazonaws.com"]
      type        = "Service"
    }
  }
}

resource "aws_iam_role" "AWSCloudFormationStackSetAdministrationRole" {
  assume_role_policy = data.aws_iam_policy_document.AWSCloudFormationStackSetAdministrationRole_assume_role_policy.json
  name               = "AWSCloudFormationStackSetAdministrationRole"
}

resource "aws_cloudformation_stack_set" "example" {
  administration_role_arn = aws_iam_role.AWSCloudFormationStackSetAdministrationRole.arn
  name                    = "example"

  parameters = {
    VPCCidr = "10.0.0.0/16"
  }

  template_body = jsonencode({
    Parameters = {
      VPCCidr = {
        Type        = "String"
        Default     = "10.0.0.0/16"
        Description = "Enter the CIDR block for the VPC. Default is 10.0.0.0/16."
      }
    }
    Resources = {
      myVpc = {
        Type = "AWS::EC2::VPC"
        Properties = {
          CidrBlock = {
            Ref = "VPCCidr"
          }
          Tags = [
            {
              Key   = "Name"
              Value = "Primary_CF_VPC"
            }
          ]
        }
      }
    }
  })
}

data "aws_iam_policy_document" "AWSCloudFormationStackSetAdministrationRole_ExecutionPolicy" {
  statement {
    actions   = ["sts:AssumeRole"]
    effect    = "Allow"
    resources = ["arn:aws:iam::*:role/${aws_cloudformation_stack_set.example.execution_role_name}"]
  }
}

resource "aws_iam_role_policy" "AWSCloudFormationStackSetAdministrationRole_ExecutionPolicy" {
  name   = "ExecutionPolicy"
  policy = data.aws_iam_policy_document.AWSCloudFormationStackSetAdministrationRole_ExecutionPolicy.json
  role   = aws_iam_role.AWSCloudFormationStackSetAdministrationRole.name
}
```

## Argument Reference

This resource supports the following arguments:

* `administration_role_arn` - (Optional) Amazon Resource Number (ARN) of the IAM Role in the administrator account. This must be defined when using the `SELF_MANAGED` permission model.
* `auto_deployment` - (Optional) Configuration block containing the auto-deployment model for your StackSet. This can only be defined when using the `SERVICE_MANAGED` permission model.
    * `enabled` - (Optional) Whether or not auto-deployment is enabled.
    * `retain_stacks_on_account_removal` - (Optional) Whether or not to retain stacks when the account is removed.
* `name` - (Required) Name of the StackSet. The name must be unique in the region where you create your StackSet. The name can contain only alphanumeric characters (case-sensitive) and hyphens. It must start with an alphabetic character and cannot be longer than 128 characters.
* `capabilities` - (Optional) A list of capabilities. Valid values: `CAPABILITY_IAM`, `CAPABILITY_NAMED_IAM`, `CAPABILITY_AUTO_EXPAND`.
* `operation_preferences` - (Optional) Preferences for how AWS CloudFormation performs a stack set update.
* `description` - (Optional) Description of the StackSet.
* `execution_role_name` - (Optional) Name of the IAM Role in all target accounts for StackSet operations. Defaults to `AWSCloudFormationStackSetExecutionRole` when using the `SELF_MANAGED` permission model. This should not be defined when using the `SERVICE_MANAGED` permission model.
* `managed_execution` - (Optional) Configuration block to allow StackSets to perform non-conflicting operations concurrently and queues conflicting operations.
    * `active` - (Optional) When set to true, StackSets performs non-conflicting operations concurrently and queues conflicting operations. After conflicting operations finish, StackSets starts queued operations in request order. Default is false.
* `parameters` - (Optional) Key-value map of input parameters for the StackSet template. All template parameters, including those with a `Default`, must be configured or ignored with `lifecycle` configuration block `ignore_changes` argument. All `NoEcho` template parameters must be ignored with the `lifecycle` configuration block `ignore_changes` argument.
* `permission_model` - (Optional) Describes how the IAM roles required for your StackSet are created. Valid values: `SELF_MANAGED` (default), `SERVICE_MANAGED`.
* `call_as` - (Optional) Specifies whether you are acting as an account administrator in the organization's management account or as a delegated administrator in a member account. Valid values: `SELF` (default), `DELEGATED_ADMIN`.
* `tags` - (Optional) Key-value map of tags to associate with this StackSet and the Stacks created from it. AWS CloudFormation also propagates these tags to supported resources that are created in the Stacks. A maximum number of 50 tags can be specified. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `template_body` - (Optional) String containing the CloudFormation template body. Maximum size: 51,200 bytes. Conflicts with `template_url`.
* `template_url` - (Optional) String containing the location of a file containing the CloudFormation template body. The URL must point to a template that is located in an Amazon S3 bucket. Maximum location file size: 460,800 bytes. Conflicts with `template_body`.

### `operation_preferences` Argument Reference

The `operation_preferences` configuration block supports the following arguments:

* `failure_tolerance_count` - (Optional) The number of accounts, per Region, for which this operation can fail before AWS CloudFormation stops the operation in that Region.
* `failure_tolerance_percentage` - (Optional) The percentage of accounts, per Region, for which this stack operation can fail before AWS CloudFormation stops the operation in that Region.
* `max_concurrent_count` - (Optional) The maximum number of accounts in which to perform this operation at one time.
* `max_concurrent_percentage` - (Optional) The maximum percentage of accounts in which to perform this operation at one time.
* `region_concurrency_type` - (Optional) The concurrency type of deploying StackSets operations in Regions, could be in parallel or one Region at a time.
* `region_order` - (Optional) The order of the Regions in where you want to perform the stack operation.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of the StackSet.
* `id` - Name of the StackSet.
* `stack_set_id` - Unique identifier of the StackSet.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `update` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudFormation StackSets using the `name`. For example:

```terraform
import {
  to = aws_cloudformation_stack_set.example
  id = "example"
}
```

Import CloudFormation StackSets when acting a delegated administrator in a member account using the `name` and `call_as` values separated by a comma (`,`). For example:

```terraform
import {
  to = aws_cloudformation_stack_set.example
  id = "example,DELEGATED_ADMIN"
}
```

Using `terraform import`, import CloudFormation StackSets using the `name`. For example:

```console
% terraform import aws_cloudformation_stack_set.example example
```

Using `terraform import`, import CloudFormation StackSets when acting a delegated administrator in a member account using the `name` and `call_as` values separated by a comma (`,`). For example:

```console
% terraform import aws_cloudformation_stack_set.example example,DELEGATED_ADMIN
```
