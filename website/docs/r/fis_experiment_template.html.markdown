---
subcategory: "Fault Injection Simulator (FIS)"
layout: "aws"
page_title: "AWS: aws_fis_experiment_template"
description: |-
  Provides an FIS Experiment Template.
---

# Resource: aws_fis_experiment_template

Provides an FIS Experiment Template, which can be used to run an experiment.
An experiment template contains one or more actions to run on specified targets during an experiment.
It also contains the stop conditions that prevent the experiment from going out of bounds.
See [Amazon Fault Injection Simulator](https://docs.aws.amazon.com/fis/index.html)
for more information.

## Example Usage

```terraform
data "aws_partition" "current" {}
data "aws_region" "current" {}
data "aws_caller_identity" "current" {}

resource "aws_iam_policy" "example" {
  name        = "example"
  description = "example"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "ec2:TerminateInstances"
      ],
      "Effect": "Allow",
      "Resource": "arn:${data.aws_partition.current.partition}:ec2:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:instance/*"
    }
  ]
}
EOF
}

resource "aws_iam_role" "example" {
  name = "example"

  managed_policy_arns = [aws_iam_policy.example.arn]

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "fis.${data.aws_partition.current.dns_suffix}"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "example" {
  role       = aws_iam_role.example.name
  policy_arn = aws_iam_policy.example.arn
}

resource "aws_fis_experiment_template" "example" {
  description = "example"
  role_arn    = aws_iam_role.example.arn

  stop_condition {
    source = "none"
  }

  action {
    name      = "example-action"
    action_id = "aws:ec2:terminate-instances"

    target {
      key   = "Instances"
      value = "example-target"
    }
  }

  target {
    name           = "example-target"
    resource_type  = "aws:ec2:instance"
    selection_mode = "COUNT(1)"

    resource_tag {
      key   = "env"
      value = "example"
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `description` - (Required) Description for the experiment template.
* `role_arn` - (Required) The Amazon Resource Name (ARN) of an IAM role that grants the AWS FIS service permission to perform service actions on your behalf.
* `action` - (Required) Configuration block(s) defining an action to be performed during an experiment. Described below.
* `stop_condition` - (Required) Configuration block(s) defining when an ongoing experiment should be stopped. Described below.
* `target` - (Optional) Configuration block(s) defining a target of an action. Described below.
* `tags` - Key-value mapping of tags for the IAM role. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## action Configuration Block

Supported arguments for the `action` configuration block:

* `action_id` - (Required) The ID of the action. To find out what actions are supported see [AWS FIS actions reference](https://docs.aws.amazon.com/fis/latest/userguide/fis-actions-reference.html).
* `name` - (Required) Friendly name given to the action.
* `description` - (Optional) Description for the action.
* `parameter` - (Optional) Configuration block(s) defining the parameter(s) for the action, if applicable. Described below.
* `start_after` - (Optional) Set of action names that must complete before this action can be executed.
* `target` - (Optional) Configuration block defining action's target, if applicable. Described below.

### parameter

Attributes for action parameter block(s):

* `key` - (Required) Parameter name.
* `value` - (Required) Parameter value.

For a list of parameters supported by each action, see [AWS FIS actions reference](https://docs.aws.amazon.com/fis/latest/userguide/fis-actions-reference.html).

### target

Attributes for action target block(s):

* `key` - (Required) Target type. Valid values are:
    * `Clusters` - for ECS Clusters.
    * `DBInstances` - for RDS DB Instances.
    * `Instances` - for EC2 Instances.
    * `Nodegroups` - for EKS Node groups.
    * `Roles` - for IAM Roles.
* `value` - (Required) Target name, referencing a corresponding target block on the Experiment Template.

## stop_condition Configuration Block

Supported arguments for the `stop_condition` configuration block:

* `source` - (Required) Source of the condition. One of `none`, `aws:cloudwatch:alarm`.
* `value` - (Optional) The Amazon Resource Name (ARN) of the CloudWatch alarm. This is required if the source is a CloudWatch alarm.

## target Configuration Block

Supported arguments for the `target` configuration block:

* `name` - (Required) Friendly name given to the target.
* `resource_type` - (Required) The AWS resource type. The resource type must be supported for the specified action. To find out what resource types are supported, see [Targets for AWS FIS](https://docs.aws.amazon.com/fis/latest/userguide/targets.html#resource-types).
* `selection_mode` - (Required) Scopes the identified resources to a specific count of the resources at random, or a percentage of the resources. All identified resources are included in the target. Valid values are:
    * `ALL` - Select all identified resources.
    * `COUNT(n)` - Randomly select `n` of the identified resources. For example, `COUNT(1)` selects one of the resources.
    * `PERCENT(n)` - Randomly select `n` percent of the identified resources. For example, `PERCENT(25)` selects 25% of the resources.
* `filter` - (Optional) Configuration block(s) defining filter(s) for the target. Filters can be used to select resources based on specific attributes returned by the respective describe action of the resource type. For more information, see [Targets for AWS FIS](https://docs.aws.amazon.com/fis/latest/userguide/targets.html#target-filters). Described below.
* `resource_arns` - (Optional) Set of Amazon Resource Names (ARNs) of the resources to target with an action. Conflicts with `resource_tag`.
* `resource_tag` - (Optional) Configuration block(s) defining tag(s) the resources need to have to be considered a valid target for an action. Conflicts with `resource_arns`. Described below.

~> **NOTE:** The `target` configuration block requires to have either `resource_arns` or `resource_tag` defined to be valid.

### filter

Attributes for target filter block(s):

* `path` - (Required) Attribute path for the filter.
* `values` - (Required) Set of attribute values for the filter.

~> **NOTE:** Values specified in a `filter` are joined with an `OR` clause, while values across multiple `filter` blocks are joined with an `AND` clause. For more information, see [Targets for AWS FIS](https://docs.aws.amazon.com/fis/latest/userguide/targets.html#target-filters).

### resource_tag

Attributes for target resource_tag block(s):

* `key` - (Required) Tag key.
* `value` - (Required) Tag value.

## Import

FIS Experiment Templates can be imported using the `id`, e.g.

```
$ terraform import aws_fis_experiment_template.template EXT123AbCdEfGhIjK
```
