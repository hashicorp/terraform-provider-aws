---
subcategory: "Security Hub"
layout: "aws"
page_title: "AWS: aws_securityhub_configuration_policy"
description: |-
  Provides a resource to manage Security Hub configuration policy
---

# Resource: aws_securityhub_configuration_policy

Manages Security Hub configuration policy

~> **NOTE:** This resource requires [`aws_securityhub_organization_configuration`](/docs/providers/aws/r/securityhub_organization_admin_account.html) to be configured of type `CENTRAL`. More information about Security Hub central configuration and configuration policies can be found in the [How Security Hub configuration policies work](https://docs.aws.amazon.com/securityhub/latest/userguide/configuration-policies-overview.html) documentation.

## Example Usage

### Default standards enabled

```terraform
resource "aws_securityhub_finding_aggregator" "example" {
  linking_mode = "ALL_REGIONS"
}

resource "aws_securityhub_organization_configuration" "example" {
  auto_enable           = false
  auto_enable_standards = "NONE"
  organization_configuration {
    configuration_type = "CENTRAL"
  }

  depends_on = [aws_securityhub_finding_aggregator.example]
}

resource "aws_securityhub_configuration_policy" "example" {
  name        = "Example"
  description = "This is an example configuration policy"

  configuration_policy {
    service_enabled = true
    enabled_standard_arns = [
      "arn:aws:securityhub:us-east-1::standards/aws-foundational-security-best-practices/v/1.0.0",
      "arn:aws:securityhub:::ruleset/cis-aws-foundations-benchmark/v/1.2.0",
    ]
    security_controls_configuration {
      disabled_control_identifiers = []
    }
  }

  depends_on = [aws_securityhub_organization_configuration.example]
}
```

### Disabled Policy

```terraform
resource "aws_securityhub_configuration_policy" "disabled" {
  name        = "Disabled"
  description = "This is an example of disabled configuration policy"

  configuration_policy {
    service_enabled = false
  }

  depends_on = [aws_securityhub_organization_configuration.example]
}
```

### Custom Control Configuration

```terraform
resource "aws_securityhub_configuration_policy" "disabled" {
  name        = "Custom Controls"
  description = "This is an example of configuration policy with custom control settings"

  configuration_policy {
    service_enabled = true
    enabled_standard_arns = [
      "arn:aws:securityhub:us-east-1::standards/aws-foundational-security-best-practices/v/1.0.0",
      "arn:aws:securityhub:::ruleset/cis-aws-foundations-benchmark/v/1.2.0",
    ]
    security_controls_configuration {
      enabled_control_identifiers = [
        "APIGateway.1",
        "IAM.7",
      ]
      security_control_custom_parameter {
        security_control_id = "APIGateway.1"
        parameter {
          name       = "loggingLevel"
          value_type = "CUSTOM"
          enum {
            value = "INFO"
          }
        }
      }
      security_control_custom_parameter {
        security_control_id = "IAM.7"
        parameter {
          name       = "RequireLowercaseCharacters"
          value_type = "CUSTOM"
          bool {
            value = false
          }
        }
        parameter {
          name       = "MaxPasswordAge"
          value_type = "CUSTOM"
          int {
            value = 60
          }
        }
      }
    }
  }

  depends_on = [aws_securityhub_organization_configuration.example]
}
```

## Argument Reference

This resource supports the following arguments:

* `configuration_policy` - (Required) Defines how Security Hub is configured. See [below](#configuration_policy).
* `description` - (Optional) The description of the configuration policy.
* `name` - (Required) The name of the configuration policy.

### configuration_policy

The `configuration_policy` block supports the following:

* `enabled_standard_arns` - (Optional) A list that defines which security standards are enabled in the configuration policy. It must be defined if `service_enabled` is set to true.
* `security_controls_configuration` - (Optional) Defines which security controls are enabled in the configuration policy and any customizations to parameters affecting them. See [below](#security_controls_configuration).
* `service_enabled` - (Required) Indicates whether Security Hub is enabled in the policy.

### security_controls_configuration

The `security_controls_configuration` block supports the following:

* `disabled_control_identifiers` - (Optional) A list of security controls that are disabled in the configuration policy Security Hub enables all other controls (including newly released controls) other than the listed controls. Conflicts with `enabled_control_identifiers`.
* `enabled_control_identifiers` - (Optional) A list of security controls that are enabled in the configuration policy. Security Hub disables all other controls (including newly released controls) other than the listed controls. Conflicts with `disabled_control_identifiers`.
* `security_control_custom_parameter` - (Optional) A list of control parameter customizations that are included in a configuration policy. Include multiple blocks to define multiple control custom parameters. See [below](#security_control_custom_parameter).

### security_control_custom_parameter

The `security_control_custom_parameter` block supports the following:

* `parameter` - (Required) An object that specifies parameter values for a control in a configuration policy. See [below](#parameter).
* `security_control_id` - (Required) The ID of the security control. For more information see the [Security Hub controls reference] documentation.

### parameter

The `parameter` block supports the following:

* `name`: (Required) The name of the control parameter. For more information see the [Security Hub controls reference] documentation.
* `value_type`: (Required) Identifies whether a control parameter uses a custom user-defined value or subscribes to the default Security Hub behavior. Valid values: `DEFAULT`, `CUSTOM`.
* `bool`: (Optional) The bool `value` for a Boolean-typed Security Hub Control Parameter.
* `double`: (Optional) The float `value` for a Double-typed Security Hub Control Parameter.
* `enum`: (Optional) The string `value` for a Enum-typed Security Hub Control Parameter.
* `enum_list`: (Optional) The string list `value` for a EnumList-typed Security Hub Control Parameter.
* `int`: (Optional) The int `value` for a Int-typed Security Hub Control Parameter.
* `int_list`: (Optional) The int list `value` for a IntList-typed Security Hub Control Parameter.
* `int_list`: (Optional) The int list `value` for a IntList-typed Security Hub Control Parameter.
* `string`: (Optional) The string `value` for a String-typed Security Hub Control Parameter.
* `string_list`: (Optional) The string list `value` for a StringList-typed Security Hub Control Parameter.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` -  The UUID of the configuration policy.
* `arn ` - The ARN of the configuration policy.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import an existing Security Hub configuration policy using the universally unique identifier (UUID) of the policy. For example:

```terraform
import {
  to = aws_securityhub_configuration_policy.example
  id = "00000000-1111-2222-3333-444444444444"
}
```

Using `terraform import`, import an existing Security Hub enabled account using the universally unique identifier (UUID) of the policy. For example:

```console
% terraform import aws_securityhub_configuration_policy.example "00000000-1111-2222-3333-444444444444"
```
