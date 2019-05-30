---
layout: "aws"
page_title: "AWS: aws_cloudformation_stack_set_instance"
sidebar_current: "docs-aws-resource-cloudformation-stack-set-instance"
description: |-
  Manages a CloudFormation Stack Set Instance.
---

# Resource: aws_cloudformation_stack_set_instance

Manages a CloudFormation Stack Set Instance. Instances are managed in the account and region of the Stack Set after the target account permissions have been configured. Additional information about Stack Sets can be found in the [AWS CloudFormation User Guide](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/what-is-cfnstacksets.html).

~> **NOTE:** All target accounts must have an IAM Role created that matches the name of the execution role configured in the Stack Set (the `execution_role_name` argument in the `aws_cloudformation_stack_set` resource) in a trust relationship with the administrative account or administration IAM Role. The execution role must have appropriate permissions to manage resources defined in the template along with those required for Stack Sets to operate. See the [AWS CloudFormation User Guide](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/stacksets-prereqs.html) for more details.

~> **NOTE:** To retain the Stack during Terraform resource destroy, ensure `retain_stack = true` has been successfully applied into the Terraform state first. This must be completed _before_ an apply that would destroy the resource.

## Example Usage

```hcl
resource "aws_cloudformation_stack_set_instance" "example" {
  account_id     = "123456789012"
  region         = "us-east-1"
  stack_set_name = "${aws_cloudformation_stack_set.example.name}"
}
```

### Example IAM Setup in Target Account

```hcl
data "aws_iam_policy_document" "AWSCloudFormationStackSetExecutionRole_assume_role_policy" {
  statement {
    actions = ["sts:AssumeRole"]
    effect  = "Allow"

    principals {
      identifiers = ["${aws_iam_role.AWSCloudFormationStackSetAdministrationRole.arn}"]
      type        = "AWS"
    }
  }
}

resource "aws_iam_role" "AWSCloudFormationStackSetExecutionRole" {
  assume_role_policy = "${data.aws_iam_policy_document.AWSCloudFormationStackSetExecutionRole_assume_role_policy.json}"
  name               = "AWSCloudFormationStackSetExecutionRole"
}

# Documentation: https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/stacksets-prereqs.html
# Additional IAM permissions necessary depend on the resources defined in the Stack Set template
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
  policy = "${data.aws_iam_policy_document.AWSCloudFormationStackSetExecutionRole_MinimumExecutionPolicy.json}"
  role   = "${aws_iam_role.AWSCloudFormationStackSetExecutionRole.name}"
}
```

## Argument Reference

The following arguments are supported:

* `stack_set_name` - (Required) Name of the Stack Set.
* `account_id` - (Optional) Target AWS Account ID to create a Stack based on the Stack Set. Defaults to current account.
* `parameter_overrides` - (Optional) Key-value map of input parameters to override from the Stack Set for this Instance.
* `region` - (Optional) Target AWS Region to create a Stack based on the Stack Set. Defaults to current region.
* `retain_stack` - (Optional) During Terraform resource destroy, remove Instance from Stack Set while keeping the Stack and its associated resources. Must be enabled in Terraform state _before_ destroy operation to take effect. You cannot reassociate a retained Stack or add an existing, saved Stack to a new Stack Set. Defaults to `false`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Stack Set name, target AWS account ID, and target AWS region separated by commas (`,`)
* `stack_id` - Stack identifier

## Timeouts

`aws_cloudformation_stack_set_instance` provides the following [Timeouts](/docs/configuration/resources.html#timeouts) configuration options:

* `create` - (Default `30m`) How long to wait for a Stack to be created.
* `update` - (Default `30m`) How long to wait for a Stack to be updated.
* `delete` - (Default `30m`) How long to wait for a Stack to be deleted.

## Import

CloudFormation Stack Set Instances can be imported using the Stack Set name, target AWS account ID, and target AWS region separated by commas (`,`) e.g.

```
$ terraform import aws_cloudformation_stack_set_instance.example example,123456789012,us-east-1
```
