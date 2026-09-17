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

* `account_id` - (Optional) ID of the target account for the budget. Uses the current user's account ID by default if omitted.
* `action_threshold` - (Required) Trigger threshold of the action. See [`action_threshold` Block](#action_threshold-block).
* `action_type` - (Required) Type of action. This defines the type of tasks that can be carried out by this action. This field also determines the format for definition. Valid values are `APPLY_IAM_POLICY`, `APPLY_SCP_POLICY`, and `RUN_SSM_DOCUMENTS`.
* `approval_model` - (Required) Whether the action needs manual or automatic approval. Valid values are `AUTOMATIC` and `MANUAL`.
* `budget_name` - (Required) Name of a budget.
* `definition` - (Required) Type-specific parameters. See [`definition` Block](#definition-block).
* `execution_role_arn` - (Required) Role passed for action execution and reversion. Roles and actions must be in the same account.
* `notification_type` - (Required) Type of a notification. Valid values are `ACTUAL` or `FORECASTED`.
* `subscriber` - (Required) Set of subscribers. See [`subscriber` Block](#subscriber-block).
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `action_threshold` Block

* `action_threshold_type` - (Required) Type of threshold for a notification. Valid values are `PERCENTAGE` or `ABSOLUTE_VALUE`.
* `action_threshold_value` - (Required) Threshold of a notification.

### `subscriber` Block

* `address` - (Required) Address that AWS sends budget notifications to, either an SNS topic or an email.
* `subscription_type` - (Required) Type of notification that AWS sends to a subscriber. Valid values are `SNS` or `EMAIL`.

### `definition` Block

* `iam_action_definition` - (Optional) AWS Identity and Access Management (IAM) action definition details. See [`iam_action_definition` Block](#iam_action_definition-block).
* `scp_action_definition` - (Optional) Service control policies (SCPs) action definition details. See [`scp_action_definition` Block](#scp_action_definition-block).
* `ssm_action_definition` - (Optional) AWS Systems Manager (SSM) action definition details. See [`ssm_action_definition` Block](#ssm_action_definition-block).

#### `iam_action_definition` Block

* `groups` - (Optional) List of groups to be attached. There must be at least one group.
* `policy_arn` - (Required) ARN of the policy to be attached.
* `roles` - (Optional) List of roles to be attached. There must be at least one role.
* `users` - (Optional) List of users to be attached. There must be at least one user.

#### `scp_action_definition` Block

* `policy_id` - (Required) Policy ID attached.
* `target_ids` - (Optional) List of target IDs.

#### `ssm_action_definition` Block

* `action_sub_type` - (Required) Action subType. Valid values are `STOP_EC2_INSTANCES` or `STOP_RDS_INSTANCES`.
* `instance_ids` - (Required) EC2 and RDS instance IDs.
* `region` - (Required) Region to run the SSM document.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `action_id` - ID of the budget action.
* `arn` - ARN of the budget action.
* `id` - ID of resource.
* `status` - Status of the budget action.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)

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
