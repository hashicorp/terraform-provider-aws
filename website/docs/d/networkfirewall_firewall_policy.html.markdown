---
subcategory: "Network Firewall"
layout: "aws"
page_title: "AWS: aws_networkfirewall_firewall_policy"
description: |-
  Retrieve information about a firewall policy.
---

# Data Source: aws_networkfirewall_firewall_policy

Retrieve information about a firewall policy.

## Example Usage

### Find firewall policy by name

```terraform
data "aws_networkfirewall_firewall_policy" "example" {
  name = var.firewall_policy_name
}
```

### Find firewall policy by ARN

```terraform
data "aws_networkfirewall_firewall_policy" "example" {
  arn = var.firewall_policy_arn
}
```

### Find firewall policy by name and ARN

```terraform
data "aws_networkfirewall_firewall_policy" "example" {
  arn  = var.firewall_policy_arn
  name = var.firewall_policy_name
}
```

AWS Network Firewall does not allow multiple firewall policies with the same name to be created in an account. It is possible, however, to have multiple firewall policies available in a single account with identical `name` values but distinct `arn` values, e.g. firewall policies shared via a [Resource Access Manager (RAM) share](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/ram_resource_share). In that case specifying `arn`, or `name` and `arn`, is recommended.

~> **Note:** If there are multiple firewall policies in an account with the same `name`, and `arn` is not specified, the default behavior will return the firewall policy with `name` that was created in the account.

## Argument Reference

This data source supports the following arguments:

* `arn` - (Optional) ARN of the firewall policy. One or more of `arn` or `name` must be specified.
* `name` - (Optional) Descriptive name of the firewall policy. One or more of `arn` or `name` must be specified.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `description` - Description of the firewall policy.
* `firewall_policy` - Configuration block describing the rule groups and policy actions in the firewall policy. See [`firewall_policy` Block](#firewall_policy-block) below.
* `tags` - Key-value tags for the firewall policy.
* `update_token` - Token used for optimistic locking.

### `firewall_policy` Block

The `firewall_policy` block exports the following attributes:

* `enable_tls_session_holding` - Whether TCP and TLS packets are prevented from reaching destination servers until TLS Inspection has evaluated Server Name Indication (SNI) rules.
* `policy_variables` - Variables used to override default Suricata settings. See [`policy_variables` Block](#policy_variables-block) below.
* `route53_resolved_ip_validation` - Resolved IP validation settings, which check that a connection's destination IP address was returned by the Route 53 Resolver in a VPC associated with a DNS Firewall rule group. See [`route53_resolved_ip_validation` Block](#route53_resolved_ip_validation-block) below.
* `stateful_default_actions` - Set of actions taken on a packet if it does not match any stateful rules in the policy.
* `stateful_engine_options` - Options on how the policy handles stateful rules. See [`stateful_engine_options` Block](#stateful_engine_options-block) below.
* `stateful_rule_group_reference` - References to the stateful rule groups used in the policy. See [`stateful_rule_group_reference` Block](#stateful_rule_group_reference-block) below.
* `stateless_custom_action` - Custom action definitions available for use in the firewall policy's `stateless_default_actions`. See [`stateless_custom_action` Block](#stateless_custom_action-block) below.
* `stateless_default_actions` - Set of actions taken on a packet if it does not match any of the stateless rules in the policy.
* `stateless_fragment_default_actions` - Set of actions taken on a fragmented packet if it does not match any of the stateless rules in the policy.
* `stateless_rule_group_reference` - References to the stateless rule groups used in the policy. See [`stateless_rule_group_reference` Block](#stateless_rule_group_reference-block) below.
* `tls_inspection_configuration_arn` - ARN of the TLS Inspection configuration attached to the firewall policy.

### `policy_variables` Block

The `policy_variables` block exports the following attribute:

* `rule_variables` - Variables used to override default Suricata settings. See [`rule_variables` Block](#rule_variables-block) below.

### `rule_variables` Block

The `rule_variables` block exports the following attributes:

* `ip_set` - Set of IP addresses. See [`ip_set` Block](#ip_set-block) below.
* `key` - Alphanumeric string that identifies the `ip_set`.

### `ip_set` Block

The `ip_set` block exports the following attribute:

* `definition` - Set of IPv4 or IPv6 addresses in CIDR notation used for the Suricata `HOME_NET` variable.

### `route53_resolved_ip_validation` Block

The `route53_resolved_ip_validation` block exports the following attributes:

* `action` - Action taken on connections that fail validation.
* `bypass_destination_validation_ip_ranges` - Set of destination IPv4 or IPv6 CIDR blocks that are exempt from the policy-level check.
* `dns_expiry_exception_policy` - Action taken when the recorded DNS resolution has expired from the resolved IP cache but traffic continues.
* `dns_firewall_rule_group_id` - ID of the Route 53 Resolver DNS Firewall rule group used for validation.
* `source_ip_ranges` - Set of client source IPv4 or IPv6 CIDR blocks whose connections are subject to the policy-level validation check.

### `stateful_engine_options` Block

The `stateful_engine_options` block exports the following attributes:

* `flow_timeouts` - Amount of time that can pass without any traffic sent through the firewall before the firewall determines that the connection is idle. See [`flow_timeouts` Block](#flow_timeouts-block) below.
* `rule_order` - How the order of stateful rule evaluation is managed for the policy.
* `stream_exception_policy` - How traffic that has broken midstream is treated.

### `flow_timeouts` Block

The `flow_timeouts` block exports the following attribute:

* `tcp_idle_timeout_seconds` - Number of seconds that can pass without any TCP traffic sent through the firewall before the firewall determines that the connection is idle.

### `stateful_rule_group_reference` Block

The `stateful_rule_group_reference` block exports the following attributes:

* `deep_threat_inspection` - Whether deep threat inspection is enabled for active threat defense managed rule groups.
* `override` - Override values applied to the rule group. See [`override` Block](#override-block) below.
* `priority` - Integer setting that indicates the order in which the stateful rule groups are applied in a single policy.
* `resource_arn` - ARN of the stateful rule group.

### `override` Block

The `override` block exports the following attribute:

* `action` - Action that changes the rule group from `DROP` to `ALERT`.

### `stateless_custom_action` Block

The `stateless_custom_action` block exports the following attributes:

* `action_definition` - Custom action associated with the `action_name`. See [`action_definition` Block](#action_definition-block) below.
* `action_name` - Name of the custom action.

### `action_definition` Block

The `action_definition` block exports the following attribute:

* `publish_metric_action` - Stateless inspection criteria that publishes the specified metrics to Amazon CloudWatch for the matching packet. See [`publish_metric_action` Block](#publish_metric_action-block) below.

### `publish_metric_action` Block

The `publish_metric_action` block exports the following attribute:

* `dimension` - Set of dimension settings used for Amazon CloudWatch custom metrics. See [`dimension` Block](#dimension-block) below.

### `dimension` Block

The `dimension` block exports the following attribute:

* `value` - String value used in the custom metric dimension.

### `stateless_rule_group_reference` Block

The `stateless_rule_group_reference` block exports the following attributes:

* `priority` - Integer setting that indicates the order in which the stateless rule groups are run in a single policy.
* `resource_arn` - ARN of the stateless rule group.
