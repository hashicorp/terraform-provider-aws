---
subcategory: "Network Firewall"
layout: "aws"
page_title: "AWS: aws_networkfirewall_firewall_policy_v2"
description: |-
  Manages an AWS Network Firewall Next Generation (V2) Firewall Policy.
---

# Resource: aws_networkfirewall_firewall_policy_v2

Manages an AWS Network Firewall Next Generation (V2) Firewall Policy. For more information, see [Next Generation policies in AWS Network Firewall](https://docs.aws.amazon.com/network-firewall/latest/developerguide/firewall-policies-v2.html).

## Example Usage

### Basic Usage

```terraform
resource "aws_networkfirewall_firewall_policy_v2" "example" {
  name = "example"

  firewall_policy {
    firewall_default_action = "DROP"
    policy_mode             = "ENFORCE"

    firewall_rule_group_entry {
      priority = 100

      inline_rule_group {
        rules_source {
          firewall_rule {
            action       = "ACCEPT"
            action_scope = "FLOW"
            hook         = "TCP_FLOW_START"
            protocol     = "TCP"
            source_ips   = "$HOME_NET"
            source_ports = "any"
            dest_ips     = "any"
            dest_ports   = "443"

            rule_metadata {
              rule_description = "Allow HTTPS egress"
              rule_enabled     = true
            }

            rule_option {
              keyword  = "sid"
              settings = ["1000001"]
            }
          }
        }

        rule_variables {
          ip_set {
            key        = "HOME_NET"
            definition = ["10.0.0.0/16"]
          }
        }
      }
    }

    stateful_engine_options {
      stream_exception_policy = "DROP"

      flow_timeouts {
        tcp_idle_timeout_seconds = 350
      }
    }
  }

  tags = {
    Name = "example"
  }
}
```

### IPS Rule Groups

```terraform
data "aws_partition" "current" {}
data "aws_region" "current" {}

resource "aws_networkfirewall_firewall_policy_v2" "example" {
  name = "example"

  firewall_policy {
    ips_rule_group_entry {
      inline_rule_group {
        rules_source {
          rules_string = "alert tcp $HOME_NET any -> $EXTERNAL_NET 22 (msg:\"Outbound SSH\"; sid:2000001; rev:1;)"
        }
      }
    }

    ips_rule_group_entry {
      resource_arn = "arn:${data.aws_partition.current.partition}:network-firewall:${data.aws_region.current.region}:aws-managed:ips-rulegroup/ThreatSignaturesPhishing"

      override {
        action = "DROP_TO_ALERT"
      }
    }

    policy_variables {
      rule_variable {
        key        = "HOME_NET"
        definition = ["10.0.0.0/16", "10.1.0.0/16"]
      }
    }
  }
}
```

### Customer Managed KMS Key

```terraform
resource "aws_networkfirewall_firewall_policy_v2" "example" {
  name = "example"

  encryption_configuration {
    key_id = aws_kms_key.example.arn
    type   = "CUSTOMER_KMS"
  }

  firewall_policy {}
}
```

## Argument Reference

The following arguments are required:

* `firewall_policy` - (Required) Rule groups and other settings that define the policy. An empty block creates an empty policy. See [`firewall_policy` Block](#firewall_policy-block) below.
* `name` - (Required, Forces new resource) Descriptive name of the policy.

The following arguments are optional:

* `description` - (Optional) Description of the policy.
* `encryption_configuration` - (Optional) KMS encryption settings for the policy. If omitted, Network Firewall uses an AWS owned key. See [`encryption_configuration` Block](#encryption_configuration-block) below.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `encryption_configuration` Block

* `key_id` - (Optional) ARN of the AWS KMS customer managed key.
* `type` - (Required) Type of KMS key to use. Valid values: `AWS_OWNED_KMS_KEY`, `CUSTOMER_KMS`.

### `firewall_policy` Block

* `firewall_default_action` - (Optional) Action to take on traffic that doesn't match any rule. Valid values: `ACCEPT`, `ALERT`, `DROP`, `REJECT`.
* `firewall_rule_group_entry` - (Optional) Firewall rule group entries. See [`firewall_rule_group_entry` Block](#firewall_rule_group_entry-block) below.
* `ips_rule_group_entry` - (Optional) IPS rule group entries. See [`ips_rule_group_entry` Block](#ips_rule_group_entry-block) below.
* `policy_mode` - (Optional) Whether the policy enforces its rules or only monitors traffic. Valid values: `ENFORCE`, `MONITOR`.
* `policy_variables` - (Optional) Variables that override the IP set variables of the rule groups in the policy. See [`policy_variables` Block](#policy_variables-block) below.
* `route53_resolved_ip_validation` - (Optional) Validates destination IPs against Route 53 Resolver DNS responses. See [`route53_resolved_ip_validation` Block](#route53_resolved_ip_validation-block) below.
* `stateful_engine_options` - (Optional) Options for the stateful engine. See [`stateful_engine_options` Block](#stateful_engine_options-block) below.
* `tls_inspection_configuration_arn` - (Optional) ARN of the TLS inspection configuration to use with the policy.

### `firewall_rule_group_entry` Block

Specify exactly one of `inline_rule_group` or `resource_arn`.

* `inline_rule_group` - (Optional) Rule group defined inline in the policy. See [`firewall_rule_group_entry.inline_rule_group` Block](#firewall_rule_group_entryinline_rule_group-block) below.
* `priority` - (Optional) Priority of the entry. Network Firewall evaluates entries in order of increasing priority. Each entry must have a unique priority.
* `resource_arn` - (Optional) ARN of an existing firewall rule group to reference.

### `firewall_rule_group_entry.inline_rule_group` Block

* `reference_sets` - (Optional) IP set references. See [`reference_sets` Block](#reference_sets-block) below.
* `rule_variables` - (Optional) IP set and port set variables. See [`rule_variables` Block](#rule_variables-block) below.
* `rules_source` - (Required) Rules for the rule group. Specify exactly one of `firewall_rule` or `rules_string`. See [`firewall_rule_group_entry.inline_rule_group.rules_source` Block](#firewall_rule_group_entryinline_rule_grouprules_source-block) below.

### `firewall_rule_group_entry.inline_rule_group.rules_source` Block

* `firewall_rule` - (Optional) Structured firewall rules. See [`firewall_rule` Block](#firewall_rule-block) below.
* `rules_string` - (Optional) Firewall rules in Suricata-compatible firewall-mode format, for example `accept:flow tcp:flow_start any any -> any 80 (sid:1; rev:1;)`.

### `firewall_rule` Block

* `accept_prior_hooks` - (Optional) Whether the rule accepts traffic that was accepted at earlier hooks.
* `action` - (Required) Action to take on matching traffic. Valid values: `ACCEPT`, `DROP`, `REJECT`, `REJECT_DST`, `REJECT_BOTH`.
* `action_scope` - (Required) Scope that the action applies to. Valid values: `FLOW`, `TX`, `HOOK`, `PACKET`.
* `alert_enabled` - (Optional) Whether to generate an alert when the rule matches.
* `bypass_ips_enabled` - (Optional) Whether matching traffic bypasses IPS inspection.
* `dest_ips` - (Optional) Destination IP addresses, CIDR blocks, or variables, for example `$EXTERNAL_NET`.
* `dest_ports` - (Optional) Destination ports, port ranges, or variables.
* `hook` - (Required) Point in the traffic flow at which the rule is evaluated, for example `TCP_FLOW_START` or `TLS_CLIENT_HELLO`. See the [AWS API reference](https://docs.aws.amazon.com/network-firewall/latest/APIReference/API_FirewallRule.html) for valid values.
* `protocol` - (Required) Protocol to match, for example `TCP`, `UDP`, `TLS` or `HTTP`. See the [AWS API reference](https://docs.aws.amazon.com/network-firewall/latest/APIReference/API_FirewallRule.html) for valid values.
* `rule_metadata` - (Optional) Rule metadata. See [`rule_metadata` Block](#rule_metadata-block) below.
* `rule_option` - (Optional) Suricata rule options, such as `sid`. See [`rule_option` Block](#rule_option-block) below.
* `source_ips` - (Optional) Source IP addresses, CIDR blocks, or variables, for example `$HOME_NET`.
* `source_ports` - (Optional) Source ports, port ranges, or variables.

### `rule_metadata` Block

* `rule_description` - (Optional) Description of the rule.
* `rule_enabled` - (Optional) Whether the rule is enabled.

### `rule_option` Block

* `keyword` - (Required) Suricata rule option keyword, for example `sid`.
* `settings` - (Optional) Settings for the keyword.

### `ips_rule_group_entry` Block

Specify exactly one of `inline_rule_group` or `resource_arn`.

* `inline_rule_group` - (Optional) Rule group defined inline in the policy. See [`ips_rule_group_entry.inline_rule_group` Block](#ips_rule_group_entryinline_rule_group-block) below.
* `override` - (Optional) Overrides the actions of a referenced rule group. Only valid with `resource_arn`. See [`override` Block](#override-block) below.
* `resource_arn` - (Optional) ARN of an existing IPS rule group to reference, such as an AWS managed IPS rule group.

### `ips_rule_group_entry.inline_rule_group` Block

* `reference_sets` - (Optional) IP set references. See [`reference_sets` Block](#reference_sets-block) below.
* `rule_variables` - (Optional) IP set and port set variables. See [`rule_variables` Block](#rule_variables-block) below.
* `rules_source` - (Required) Rules for the rule group. See [`ips_rule_group_entry.inline_rule_group.rules_source` Block](#ips_rule_group_entryinline_rule_grouprules_source-block) below.

### `ips_rule_group_entry.inline_rule_group.rules_source` Block

* `rules_string` - (Required) IPS rules in Suricata-compatible format.

### `override` Block

* `action` - (Optional) Action override. Valid values: `DROP_TO_ALERT`.

### `reference_sets` Block

* `ip_set_reference` - (Optional) Set of IP set references. See [`ip_set_reference` Block](#ip_set_reference-block) below.

### `ip_set_reference` Block

* `key` - (Required) Name of the variable used in rules.
* `reference_arn` - (Required) ARN of the referenced resource, such as a VPC managed prefix list.

### `rule_variables` Block

* `ip_set` - (Optional) Set of IP set variables. See [`ip_set` Block](#ip_set-block) below.
* `port_set` - (Optional) Set of port set variables. See [`port_set` Block](#port_set-block) below.

### `ip_set` Block

* `definition` - (Required) List of IP addresses and CIDR blocks.
* `key` - (Required) Name of the variable.

### `port_set` Block

* `definition` - (Required) List of ports and port ranges.
* `key` - (Required) Name of the variable.

### `policy_variables` Block

* `rule_variable` - (Optional) Set of IP set variables. See [`rule_variable` Block](#rule_variable-block) below.

### `rule_variable` Block

* `definition` - (Required) List of IP addresses and CIDR blocks.
* `key` - (Required) Name of the variable, for example `HOME_NET`.

### `route53_resolved_ip_validation` Block

* `action` - (Required) Action to take on traffic whose destination IP wasn't resolved by Route 53 Resolver. Valid values: `ALERT`, `DROP`, `REJECT`.
* `bypass_destination_validation_ip_range` - (Optional) Destination ranges that skip validation. See [`bypass_destination_validation_ip_range` Block](#bypass_destination_validation_ip_range-block) below.
* `dns_expiry_exception_policy` - (Required) Action to take on traffic to an IP whose DNS record has expired. Valid values: `CONTINUE`, `DROP`, `REJECT`.
* `dns_firewall_rule_group_id` - (Required) ID of the Route 53 Resolver DNS Firewall rule group.
* `source_ip_range` - (Optional) Source ranges to validate. See [`source_ip_range` Block](#source_ip_range-block) below.

### `bypass_destination_validation_ip_range` Block

* `address_definition` - (Required) IP address or CIDR block.

### `source_ip_range` Block

* `address_definition` - (Required) IP address or CIDR block.

### `stateful_engine_options` Block

* `flow_timeouts` - (Optional) Flow timeout settings. See [`flow_timeouts` Block](#flow_timeouts-block) below.
* `stream_exception_policy` - (Optional) How to handle traffic when a network connection breaks midstream. Valid values: `DROP`, `REJECT`.

### `flow_timeouts` Block

* `tcp_idle_timeout_seconds` - (Optional) Seconds of TCP inactivity before a connection is considered idle. Valid values are between `60` and `6000`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the policy.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `update_token` - Token used for optimistic locking when updating the policy.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_networkfirewall_firewall_policy_v2.example
  identity = {
    "arn" = "arn:aws:network-firewall:us-west-2:123456789012:firewall-policy-v2/example"
  }
}

resource "aws_networkfirewall_firewall_policy_v2" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `arn` (String) ARN of the Network Firewall Firewall Policy V2.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Network Firewall Firewall Policy V2 using the `arn`. For example:

```terraform
import {
  to = aws_networkfirewall_firewall_policy_v2.example
  id = "arn:aws:network-firewall:us-west-2:123456789012:firewall-policy-v2/example"
}
```

Using `terraform import`, import Network Firewall Firewall Policy V2 using the `arn`. For example:

```console
% terraform import aws_networkfirewall_firewall_policy_v2.example arn:aws:network-firewall:us-west-2:123456789012:firewall-policy-v2/example
```
