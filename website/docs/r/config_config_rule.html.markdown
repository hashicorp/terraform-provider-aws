---
subcategory: "Config"
layout: "aws"
page_title: "AWS: aws_config_config_rule"
description: |-
  Provides an AWS Config Rule.
---

# Resource: aws_config_config_rule

Provides an AWS Config Rule.

~> **Note:** Config Rule requires an existing [Configuration Recorder](/docs/providers/aws/r/config_configuration_recorder.html) to be present. Use of `depends_on` is recommended (as shown below) to avoid race conditions.

## Example Usage

### AWS Managed Rules

AWS managed rules can be used by setting the source owner to `AWS` and the source identifier to the name of the managed rule. More information about AWS managed rules can be found in the [AWS Config Developer Guide](https://docs.aws.amazon.com/config/latest/developerguide/evaluate-config_use-managed-rules.html).

```terraform
resource "aws_config_config_rule" "r" {
  name = "example"

  source {
    owner             = "AWS"
    source_identifier = "S3_BUCKET_VERSIONING_ENABLED"
  }

  depends_on = [aws_config_configuration_recorder.foo]
}

resource "aws_config_configuration_recorder" "foo" {
  name     = "example"
  role_arn = aws_iam_role.r.arn
}

data "aws_iam_policy_document" "assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["config.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "r" {
  name               = "my-awsconfig-role"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

data "aws_iam_policy_document" "p" {
  statement {
    effect    = "Allow"
    actions   = ["config:Put*"]
    resources = ["*"]
  }
}

resource "aws_iam_role_policy" "p" {
  name   = "my-awsconfig-policy"
  role   = aws_iam_role.r.id
  policy = data.aws_iam_policy_document.p.json
}
```

### Custom Rules

Custom rules can be used by setting the source owner to `CUSTOM_LAMBDA` and the source identifier to the Amazon Resource Name (ARN) of the Lambda Function. The AWS Config service must have permissions to invoke the Lambda Function, e.g., via the [`aws_lambda_permission` resource](/docs/providers/aws/r/lambda_permission.html). More information about custom rules can be found in the [AWS Config Developer Guide](https://docs.aws.amazon.com/config/latest/developerguide/evaluate-config_develop-rules.html).

```terraform
resource "aws_config_configuration_recorder" "example" {
  # ... other configuration ...
}

resource "aws_lambda_function" "example" {
  # ... other configuration ...
}

resource "aws_lambda_permission" "example" {
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.example.arn
  principal     = "config.amazonaws.com"
  statement_id  = "AllowExecutionFromConfig"
}

resource "aws_config_config_rule" "example" {
  # ... other configuration ...

  source {
    owner             = "CUSTOM_LAMBDA"
    source_identifier = aws_lambda_function.example.arn
  }

  depends_on = [
    aws_config_configuration_recorder.example,
    aws_lambda_permission.example,
  ]
}
```

### Custom Policies

```terraform
resource "aws_config_config_rule" "example" {
  name = "example"

  source {
    owner = "CUSTOM_POLICY"

    source_detail {
      message_type = "ConfigurationItemChangeNotification"
    }

    custom_policy_details {
      policy_runtime = "guard-2.x.x"
      policy_text    = <<EOF
	  rule tableisactive when
		  resourceType == "AWS::DynamoDB::Table" {
		  configuration.tableStatus == ['ACTIVE']
	  }
	  
	  rule checkcompliance when
		  resourceType == "AWS::DynamoDB::Table"
		  tableisactive {
			  supplementaryConfiguration.ContinuousBackupsDescription.pointInTimeRecoveryDescription.pointInTimeRecoveryStatus == "ENABLED"
	  }
EOF					
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) The name of the rule
* `description` - (Optional) Description of the rule
* `evaluation_mode` - (Optional) The modes the Config rule can be evaluated in. See [Evaluation Mode](#evaluation-mode) for more details.
* `input_parameters` - (Optional) A string in JSON format that is passed to the AWS Config rule Lambda function.
* `maximum_execution_frequency` - (Optional) The maximum frequency with which AWS Config runs evaluations for a rule.
* `scope` - (Optional) Scope defines which resources can trigger an evaluation for the rule. See [Scope](#scope) Below.
* `source` - (Required) Source specifies the rule owner, the rule identifier, and the notifications that cause the function to evaluate your AWS resources. See [Source](#source) Below.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Evaluation Mode

* `mode` - (Optional) The mode of an evaluation.

### Scope

Defines which resources can trigger an evaluation for the rule.
If you do not specify a scope, evaluations are triggered when any resource in the recording group changes.

* `compliance_resource_id` - (Optional) The IDs of the only AWS resource that you want to trigger an evaluation for the rule. If you specify a resource ID, you must specify one resource type for `compliance_resource_types`.
* `compliance_resource_types` - (Optional) A list of resource types of only those AWS resources that you want to trigger an evaluation for the ruleE.g., `AWS::EC2::Instance`. You can only specify one type if you also specify a resource ID for `compliance_resource_id`. See [relevant part of AWS Docs](http://docs.aws.amazon.com/config/latest/APIReference/API_ResourceIdentifier.html#config-Type-ResourceIdentifier-resourceType) for available types.
* `tag_key` - (Optional, Required if `tag_value` is specified) The tag key that is applied to only those AWS resources that you want you want to trigger an evaluation for the rule.
* `tag_value` - (Optional) The tag value applied to only those AWS resources that you want to trigger an evaluation for the rule.

### Source

Provides the rule owner (AWS or customer), the rule identifier, and the notifications that cause the function to evaluate your AWS resources.

* `owner` - (Required) Indicates whether AWS or the customer owns and manages the AWS Config rule. Valid values are `AWS`, `CUSTOM_LAMBDA` or `CUSTOM_POLICY`. For more information about managed rules, see the [AWS Config Managed Rules documentation](https://docs.aws.amazon.com/config/latest/developerguide/evaluate-config_use-managed-rules.html). For more information about custom rules, see the [AWS Config Custom Rules documentation](https://docs.aws.amazon.com/config/latest/developerguide/evaluate-config_develop-rules.html). Custom Lambda Functions require permissions to allow the AWS Config service to invoke them, e.g., via the [`aws_lambda_permission` resource](/docs/providers/aws/r/lambda_permission.html).
* `source_identifier` - (Optional) For AWS Config managed rules, a predefined identifier, e.g `IAM_PASSWORD_POLICY`. For custom Lambda rules, the identifier is the ARN of the Lambda Function, such as `arn:aws:lambda:us-east-1:123456789012:function:custom_rule_name` or the [`arn` attribute of the `aws_lambda_function` resource](/docs/providers/aws/r/lambda_function.html#arn).
* `source_detail` - (Optional) Provides the source and type of the event that causes AWS Config to evaluate your AWS resources. Only valid if `owner` is `CUSTOM_LAMBDA` or `CUSTOM_POLICY`. See [Source Detail](#source-detail) Below.
* `custom_policy_details` - (Optional) Provides the runtime system, policy definition, and whether debug logging is enabled. Required when owner is set to `CUSTOM_POLICY`. See [Custom Policy Details](#custom-policy-details) Below.

#### Source Detail

* `event_source` - (Optional) The source of the event, such as an AWS service, that triggers AWS Config to evaluate your AWSresources. This defaults to `aws.config` and is the only valid value.
* `maximum_execution_frequency` - (Optional) The frequency that you want AWS Config to run evaluations for a rule that istriggered periodically. If specified, requires `message_type` to be `ScheduledNotification`.
* `message_type` - (Optional) The type of notification that triggers AWS Config to run an evaluation for a rule. You canspecify the following notification types:
    * `ConfigurationItemChangeNotification` - Triggers an evaluation when AWS Config delivers a configuration item as a result of a resource change.
    * `OversizedConfigurationItemChangeNotification` - Triggers an evaluation when AWS Config delivers an oversized configuration item. AWS Config may generate this notification type when a resource changes and the notification exceeds the maximum size allowed by Amazon SNS.
    * `ScheduledNotification` - Triggers a periodic evaluation at the frequency specified for `maximum_execution_frequency`.
    * `ConfigurationSnapshotDeliveryCompleted` - Triggers a periodic evaluation when AWS Config delivers a configuration snapshot.

#### Custom Policy Details

* `enable_debug_log_delivery` - (Optional) The boolean expression for enabling debug logging for your Config Custom Policy rule. The default value is `false`.
* `policy_runtime` - (Required) The runtime system for your Config Custom Policy rule. Guard is a policy-as-code language that allows you to write policies that are enforced by Config Custom Policy rules. For more information about Guard, see the [Guard GitHub Repository](https://github.com/aws-cloudformation/cloudformation-guard).
* `policy_text` - (Required) The policy definition containing the logic for your Config Custom Policy rule.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN of the config rule
* `rule_id` - The ID of the config rule
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Config Rule using the name. For example:

```terraform
import {
  to = aws_config_config_rule.foo
  id = "example"
}
```

Using `terraform import`, import Config Rule using the name. For example:

```console
% terraform import aws_config_config_rule.foo example
```
