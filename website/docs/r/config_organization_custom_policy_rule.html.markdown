---
subcategory: "Config"
layout: "aws"
page_title: "AWS: aws_config_organization_custom_policy_rule"
description: |-
  Terraform resource for managing an AWS Config Organization Custom Policy.
---

# Resource: aws_config_organization_custom_policy_rule

Manages a Config Organization Custom Policy Rule. More information about these rules can be found in the [Enabling AWS Config Rules Across all Accounts in Your Organization](https://docs.aws.amazon.com/config/latest/developerguide/config-rule-multi-account-deployment.html) and [AWS Config Managed Rules](https://docs.aws.amazon.com/config/latest/developerguide/evaluate-config_use-managed-rules.html) documentation. For working with Organization Managed Rules (those invoking an AWS managed rule), see the [`aws_config_organization_managed__rule` resource](/docs/providers/aws/r/config_organization_managed_rule.html).

~> **NOTE:** This resource must be created in the Organization master account and rules will include the master account unless its ID is added to the `excluded_accounts` argument.

## Example Usage

### Basic Usage

```terraform
resource "aws_config_organization_custom_policy_rule" "example" {
  name = "example_rule_name"

  policy_runtime = "guard-2.x.x"
  policy_text    = <<-EOF
  let status = ['ACTIVE']

  rule tableisactive when
      resourceType == "AWS::DynamoDB::Table" {
      configuration.tableStatus == %status
  }

  rule checkcompliance when
      resourceType == "AWS::DynamoDB::Table"
      tableisactive {
          let pitr = supplementaryConfiguration.ContinuousBackupsDescription.pointInTimeRecoveryDescription.pointInTimeRecoveryStatus
          %pitr == "ENABLED"
      }
  EOF

  resource_types_scope = ["AWS::DynamoDB::Table"]
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) name of the rule
* `policy_text` - (Required) policy definition containing the logic for your organization AWS Config Custom Policy rule
* `policy_runtime` - (Required)  runtime system for your organization AWS Config Custom Policy rules
* `trigger_types` - (Required) List of notification types that trigger AWS Config to run an evaluation for the rule. Valid values: `ConfigurationItemChangeNotification`, `OversizedConfigurationItemChangeNotification`

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `description` - (Optional) Description of the rule
* `debug_log_delivery_accounts` - (Optional) List of AWS account identifiers to exclude from the rule
* `excluded_accounts` - (Optional) List of AWS account identifiers to exclude from the rule
* `input_parameters` - (Optional) A string in JSON format that is passed to the AWS Config Rule Lambda Function
* `maximum_execution_frequency` - (Optional) Maximum frequency with which AWS Config runs evaluations for a rule, if the rule is triggered at a periodic frequency. Defaults to `TwentyFour_Hours` for periodic frequency triggered rules. Valid values: `One_Hour`, `Three_Hours`, `Six_Hours`, `Twelve_Hours`, or `TwentyFour_Hours`.
* `resource_id_scope` - (Optional) Identifier of the AWS resource to evaluate
* `resource_types_scope` - (Optional) List of types of AWS resources to evaluate
* `tag_key_scope` - (Optional, Required if `tag_value_scope` is configured) Tag key of AWS resources to evaluate
* `tag_value_scope` - (Optional) Tag value of AWS resources to evaluate

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of the rule

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `20m`)
* `update` - (Default `20m`)
* `delete` - (Default `20m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a Config Organization Custom Policy Rule using the `name` argument. For example:

```terraform
import {
  to = aws_config_organization_custom_policy_rule.example
  id = "example_rule_name"
}
```

Using `terraform import`, import a Config Organization Custom Policy Rule using the `name` argument. For example:

```console
% terraform import aws_config_organization_custom_policy_rule.example example_rule_name
```
