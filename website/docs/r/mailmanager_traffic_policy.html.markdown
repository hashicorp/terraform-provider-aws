---
subcategory: "SES Mail Manager"
layout: "aws"
page_title: "AWS: aws_mailmanager_traffic_policy"
description: |-
  Manages an SES Mail Manager Traffic Policy.
---

# Resource: aws_mailmanager_traffic_policy

Manages an SES Mail Manager Traffic Policy.

## Example Usage

### Basic Usage

```terraform
resource "aws_mailmanager_traffic_policy" "example" {
  default_action         = "ALLOW"
  max_message_size_bytes = 100000
  name                   = "example"

  policy_statement {
    action = "DENY"

    condition {
      ip_expression {
        operator = "CIDR_MATCHES"
        values   = ["192.0.2.0/24"]

        evaluate {
          attribute = "SENDER_IP"
        }
      }
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `default_action` - (Required) Default action applied to traffic that does not match any policy statement. Valid values are `ALLOW` and `DENY`.
* `name` - (Required) Name of the traffic policy.
* `policy_statement` - (Required) Traffic policy statements. See [`policy_statement` Block](#policy_statement-block) below.

The following arguments are optional:

* `max_message_size_bytes` - (Optional) Maximum message size, in bytes, allowed by the traffic policy.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `policy_statement` Block

* `action` - (Required) Action applied when all conditions match. Valid values are `ALLOW` and `DENY`.
* `condition` - (Required) Conditions evaluated by the statement. See [`condition` Block](#condition-block) below.

### `condition` Block

Exactly one of the following expression blocks must be configured:

* `boolean_expression` - (Optional) Boolean comparison. See [`boolean_expression` Block](#boolean_expression-block) below.
* `ip_expression` - (Optional) IPv4 address comparison. See [`ip_expression` Block](#ip_expression-block) below.
* `ipv6_expression` - (Optional) IPv6 address comparison. See [`ipv6_expression` Block](#ipv6_expression-block) below.
* `string_expression` - (Optional) String comparison. See [`string_expression` Block](#string_expression-block) below.
* `tls_expression` - (Optional) TLS policy comparison. See [`tls_expression` Block](#tls_expression-block) below.

### `boolean_expression` Block

* `evaluate` - (Required) Operand evaluated by the expression. See [`policy_statement.condition.boolean_expression.evaluate` Block](#policy_statementconditionboolean_expressionevaluate-block) below.
* `operator` - (Required) Boolean operator used for the comparison.

### `policy_statement.condition.boolean_expression.evaluate` Block

Exactly one of the following blocks must be configured:

* `analysis` - (Optional) Analysis result to evaluate. See [`policy_statement.condition.boolean_expression.evaluate.analysis` Block](#policy_statementconditionboolean_expressionevaluateanalysis-block) below.
* `is_in_address_list` - (Optional) Address list membership check. See [`is_in_address_list` Block](#is_in_address_list-block) below.

### `policy_statement.condition.boolean_expression.evaluate.analysis` Block

* `analyzer` - (Required) ARN of the Add On performing the analysis.
* `result_field` - (Required) Result field returned in the analysis.

### `is_in_address_list` Block

* `address_lists` - (Required) List containing exactly one address list ARN to check membership against.
* `attribute` - (Required) Email attribute to check against the address list.

### `ip_expression` Block

* `evaluate` - (Required) Operand evaluated by the expression. See [`policy_statement.condition.ip_expression.evaluate` Block](#policy_statementconditionip_expressionevaluate-block) below.
* `operator` - (Required) IP address operator used for the comparison.
* `values` - (Required) IPv4 CIDR ranges used for the comparison.

### `policy_statement.condition.ip_expression.evaluate` Block

* `attribute` - (Required) Message attribute to evaluate.

### `ipv6_expression` Block

* `evaluate` - (Required) Operand evaluated by the expression. See [`policy_statement.condition.ipv6_expression.evaluate` Block](#policy_statementconditionipv6_expressionevaluate-block) below.
* `operator` - (Required) IPv6 address operator used for the comparison.
* `values` - (Required) IPv6 CIDR ranges used for the comparison.

### `policy_statement.condition.ipv6_expression.evaluate` Block

* `attribute` - (Required) Message attribute to evaluate.

### `string_expression` Block

* `evaluate` - (Required) Operand evaluated by the expression. See [`policy_statement.condition.string_expression.evaluate` Block](#policy_statementconditionstring_expressionevaluate-block) below.
* `operator` - (Required) String operator used for the comparison.
* `values` - (Required) Strings used for the comparison.

### `policy_statement.condition.string_expression.evaluate` Block

Exactly one of the following must be configured:

* `analysis` - (Optional) Analysis result to evaluate. See [`policy_statement.condition.string_expression.evaluate.analysis` Block](#policy_statementconditionstring_expressionevaluateanalysis-block) below.
* `attribute` - (Optional) Email attribute to evaluate.

### `policy_statement.condition.string_expression.evaluate.analysis` Block

* `analyzer` - (Required) ARN of the Add On performing the analysis.
* `result_field` - (Required) Result field returned in the analysis.

### `tls_expression` Block

* `evaluate` - (Required) Operand evaluated by the expression. See [`policy_statement.condition.tls_expression.evaluate` Block](#policy_statementconditiontls_expressionevaluate-block) below.
* `operator` - (Required) TLS policy operator used for the comparison.
* `value` - (Required) TLS policy used for the comparison.

### `policy_statement.condition.tls_expression.evaluate` Block

* `attribute` - (Required) TLS attribute to evaluate.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the traffic policy.
* `created_timestamp` - Timestamp when the traffic policy was created.
* `id` - ID of the traffic policy.
* `last_updated_timestamp` - Timestamp when the traffic policy was last updated.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_mailmanager_traffic_policy.example
  identity = {
    id = "example-id"
  }
}
```

### Identity Schema

#### Required

* `id` (String) ID of the traffic policy.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import an SES Mail Manager Traffic Policy using its ID. For example:

```terraform
import {
  to = aws_mailmanager_traffic_policy.example
  id = "example-id"
}
```

Using `terraform import`, import an SES Mail Manager Traffic Policy using its ID. For example:

```console
% terraform import aws_mailmanager_traffic_policy.example example-id
```
