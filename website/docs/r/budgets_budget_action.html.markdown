---
subcategory: "Web Services Budgets"
layout: "aws"
page_title: "AWS: aws_budgets_budget_action"
description: |-
  Provides a budget action resource.
---

# Resource: aws_budgets_budget_action

Provides a budget action resource. Budget actions are cost savings controls that run either automatically on your behalf or by using a workflow approval process.

## Example Usage

```terraform
resource "aws_budgets_budget_action" "example" {
  budget_name        = aws_budgets_budget.example.name
  action_type        = "APPLY_IAM_POLICY"
  approval_model     = "AUTOMATIC"
  notification_type  = "ACTUAL"
  execution_role_arn = aws_iam_role.example.arn

  action_threshold {
    action_threshold_type  = "ABSOLUTE_VALUE"
    action_threshold_value = 100
  }

  definition {
    iam_action_definition {
      policy_arn = aws_iam_policy.example.arn
      roles      = [aws_iam_role.example.name]
    }
  }

  subscriber {
    address           = "example@example.example"
    subscription_type = "EMAIL"
  }

  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }
}

data "aws_iam_policy_document" "example" {
  statement {
    effect    = "Allow"
    actions   = ["ec2:Describe*"]
    resources = ["*"]
  }
}

resource "aws_iam_policy" "example" {
  name        = "example"
  description = "My example policy"
  policy      = data.aws_iam_policy_document.example.json
}

data "aws_partition" "current" {}

data "aws_iam_policy_document" "assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["budgets.${data.aws_partition.current.dns_suffix}"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "example" {
  name               = "example"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_budgets_budget" "example" {
  name              = "example"
  budget_type       = "USAGE"
  limit_amount      = "10.0"
  limit_unit        = "dollars"
  time_period_start = "2006-01-02_15:04"
  time_unit         = "MONTHLY"
}
```

## Argument Reference

This resource supports the following arguments:

* `account_id` - (Optional) The ID of the target account for budget. Will use current user's account_id by default if omitted.
* `budget_name` - (Required) The name of a budget.
* `action_threshold` - (Required) The trigger threshold of the action. See [Action Threshold](#action-threshold).
* `action_type` - (Required) The type of action. This defines the type of tasks that can be carried out by this action. This field also determines the format for definition. Valid values are `APPLY_IAM_POLICY`, `APPLY_SCP_POLICY`, and `RUN_SSM_DOCUMENTS`.
* `approval_model` - (Required) This specifies if the action needs manual or automatic approval. Valid values are `AUTOMATIC` and `MANUAL`.
* `definition` - (Required) Specifies all of the type-specific parameters. See [Definition](#definition).
* `execution_role_arn` - (Required) The role passed for action execution and reversion. Roles and actions must be in the same account.
* `notification_type` - (Required) The type of a notification. Valid values are `ACTUAL` or `FORECASTED`.
* `subscriber` - (Required) A list of subscribers. See [Subscriber](#subscriber).
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Action Threshold

* `action_threshold_type` - (Required) The type of threshold for a notification. Valid values are `PERCENTAGE` or `ABSOLUTE_VALUE`.
* `action_threshold_value` - (Required) The threshold of a notification.

### Subscriber

* `address` - (Required) The address that AWS sends budget notifications to, either an SNS topic or an email.
* `subscription_type` - (Required) The type of notification that AWS sends to a subscriber. Valid values are `SNS` or `EMAIL`.

### Definition

* `iam_action_definition` - (Optional) The AWS Identity and Access Management (IAM) action definition details. See [IAM Action Definition](#iam-action-definition).
* `ssm_action_definition` - (Optional) The AWS Systems Manager (SSM) action definition details. See [SSM Action Definition](#ssm-action-definition).
* `scp_action_definition` - (Optional) The service control policies (SCPs) action definition details. See [SCP Action Definition](#scp-action-definition).

#### IAM Action Definition

* `policy_arn` - (Required) The Amazon Resource Name (ARN) of the policy to be attached.
* `groups` - (Optional) A list of groups to be attached. There must be at least one group.
* `roles` - (Optional) A list of roles to be attached. There must be at least one role.
* `users` - (Optional) A list of users to be attached. There must be at least one user.

#### SCP Action Definition

* `policy_id` - (Required) The policy ID attached.
* `target_ids` - (Optional) A list of target IDs.

#### SSM Action Definition

* `action_sub_type` - (Required) The action subType. Valid values are `STOP_EC2_INSTANCES` or `STOP_RDS_INSTANCES`.
* `instance_ids` - (Required) The EC2 and RDS instance IDs.
* `region` - (Required) The Region to run the SSM document.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `action_id` - The id of the budget action.
* `id` - ID of resource.
* `arn` - The ARN of the budget action.
* `status` - The status of the budget action.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import budget actions using `AccountID:ActionID:BudgetName`. For example:

```terraform
import {
  to = aws_budgets_budget_action.myBudget
  id = "123456789012:some-id:myBudget"
}
```

Using `terraform import`, import budget actions using `AccountID:ActionID:BudgetName`. For example:

```console
% terraform import aws_budgets_budget_action.myBudget 123456789012:some-id:myBudget
```
