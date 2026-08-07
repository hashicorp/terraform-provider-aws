---
subcategory: "FMS (Firewall Manager)"
layout: "aws"
page_title: "AWS: aws_fms_policy"
description: |-
  Provides a resource to create an AWS Firewall Manager policy
---

# Resource: aws_fms_policy

Provides a resource to create an AWS Firewall Manager policy. You need to be using AWS organizations and have enabled the Firewall Manager administrator account.

~> **NOTE:** Due to limitations with testing, we provide it as best effort. If you find it useful, and have the ability to help test or notice issues, consider reaching out to us on [GitHub](https://github.com/hashicorp/terraform-provider-aws).

## Example Usage

```terraform
resource "aws_fms_policy" "example" {
  name                  = "FMS-Policy-Example"
  exclude_resource_tags = false
  remediation_enabled   = false
  resource_type         = "AWS::ElasticLoadBalancingV2::LoadBalancer"

  security_service_policy_data {
    type = "WAF"

    managed_service_data = jsonencode({
      type = "WAF",
      ruleGroups = [{
        id = aws_wafregional_rule_group.example.id
        overrideAction = {
          type = "COUNT"
        }
      }]
      defaultAction = {
        type = "BLOCK"
      }
      overrideCustomerWebACLAssociation = false
    })
  }

  tags = {
    Name = "example-fms-policy"
  }
}

resource "aws_wafregional_rule_group" "example" {
  metric_name = "WAFRuleGroupExample"
  name        = "WAF-Rule-Group-Example"
}
```

## Argument Reference

This resource supports the following arguments:

* `delete_all_policy_resources` - (Optional) If true, the request will also perform a clean-up process. Defaults to `true`. More information can be found here [AWS Firewall Manager delete policy](https://docs.aws.amazon.com/fms/2018-01-01/APIReference/API_DeletePolicy.html)
* `delete_unused_fm_managed_resources` - (Optional) If true, Firewall Manager will automatically remove protections from resources that leave the policy scope. Defaults to `false`. More information can be found here [AWS Firewall Manager policy contents](https://docs.aws.amazon.com/fms/2018-01-01/APIReference/API_Policy.html)
* `description` - (Optional) Description of the AWS Network Firewall firewall policy.
* `exclude_map` - (Optional) Map of lists of accounts and OUs to exclude from the policy. See the [`exclude_map`](#exclude_map-block) block.
* `exclude_resource_tags` - (Required, Forces new resource) Whether resources with the tags specified in `resource_tags` are excluded from protection. If `true`, tagged resources are not protected by this policy. If `false` and `resource_tags` are populated, resources that contain those tags are protected by this policy.
* `include_map` - (Optional) Map of lists of accounts and OUs to include in the policy. See the [`include_map`](#include_map-block) block.
* `name` - (Required, Forces new resource) Friendly name of the AWS Firewall Manager Policy.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `remediation_enabled` - (Optional) Whether the policy is automatically applied to resources that already exist in the account.
* `resource_set_ids` - (Optional) Set of resource set IDs associated with the policy.
* `resource_tag_logical_operator` - (Optional) Controls how multiple resource tags are combined: with AND, so that a resource must have all tags to be included or excluded, or OR, so that a resource must have at least one tag. The valid values are `AND` and `OR`.
* `resource_tags` - (Optional) Map of resource tags that, if present, filter protections on resources based on `exclude_resource_tags`.
* `resource_type` - (Optional) Resource type to protect. Conflicts with `resource_type_list`. See the [FMS API Reference](https://docs.aws.amazon.com/fms/2018-01-01/APIReference/API_Policy.html#fms-Type-Policy-ResourceType) for more information about supported values.
* `resource_type_list` - (Optional) List of resource types to protect. Conflicts with `resource_type`. See the [FMS API Reference](https://docs.aws.amazon.com/fms/2018-01-01/APIReference/API_Policy.html#fms-Type-Policy-ResourceType) for more information about supported values. Lists with only one element are not supported, instead use `resource_type`.
* `security_service_policy_data` - (Required) Objects to include in Security Service Policy Data. See the [`security_service_policy_data`](#security_service_policy_data-block) block.
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `exclude_map` Block

You can specify inclusions or exclusions, but not both. If you specify an `include_map`, AWS Firewall Manager applies the policy to all accounts specified by the `include_map`, and does not evaluate any `exclude_map` specifications. If you do not specify an `include_map`, then Firewall Manager applies the policy to all accounts except for those specified by the `exclude_map`.

* `account` - (Optional) List of AWS Organization member accounts to exclude from this AWS FMS Policy.
* `orgunit` - (Optional) List of IDs of the AWS Organizational Units to exclude from this AWS FMS Policy. Specifying an OU is equivalent to specifying all accounts in the OU and in any of its child OUs, including any child OUs and accounts that are added at a later time.

### `include_map` Block

You can specify inclusions or exclusions, but not both. If you specify an `include_map`, AWS Firewall Manager applies the policy to all accounts specified by the `include_map`, and does not evaluate any `exclude_map` specifications. If you do not specify an `include_map`, then Firewall Manager applies the policy to all accounts except for those specified by the `exclude_map`.

* `account` - (Optional) List of AWS Organization member accounts to include for this AWS FMS Policy.
* `orgunit` - (Optional) List of IDs of the AWS Organizational Units to include for this AWS FMS Policy. Specifying an OU is equivalent to specifying all accounts in the OU and in any of its child OUs, including any child OUs and accounts that are added at a later time.

### `security_service_policy_data` Block

* `managed_service_data` - (Optional) Details about the service that are specific to the service type, in JSON format. For service type `SHIELD_ADVANCED`, this is an empty string. Examples depending on `type` can be found in the [AWS Firewall Manager SecurityServicePolicyData API Reference](https://docs.aws.amazon.com/fms/2018-01-01/APIReference/API_SecurityServicePolicyData.html).
* `policy_option` - (Optional) Contains the Network Firewall firewall policy options to configure a centralized deployment model. See the [`policy_option`](#policy_option-block) block.
* `type` - (Required, Forces new resource) Service that the policy uses to protect the resources. For the current list of supported types, refer to the [AWS Firewall Manager SecurityServicePolicyData API Type Reference](https://docs.aws.amazon.com/fms/2018-01-01/APIReference/API_SecurityServicePolicyData.html#fms-Type-SecurityServicePolicyData-Type).

### `policy_option` Block

* `network_acl_common_policy` - (Optional) Network ACL rules applied across accounts in the AWS Organization. See the [`network_acl_common_policy`](#network_acl_common_policy-block) block.
* `network_firewall_policy` - (Optional) Network Firewall policy options that configure a centralized deployment model. See the [`network_firewall_policy`](#network_firewall_policy-block) block.
* `third_party_firewall_policy` - (Optional) Third-party firewall policy options. See the [`third_party_firewall_policy`](#third_party_firewall_policy-block) block.

### `network_acl_common_policy` Block

* `network_acl_entry_set` - (Optional) Network ACL entries for the Network ACL policy. See the [`network_acl_entry_set`](#network_acl_entry_set-block) block.

### `network_acl_entry_set` Block

* `first_entry` - (Optional) Rules to run first in the Firewall Manager managed network ACLs. Firewall Manager creates entries with ID value between 1 and 5000. See the [`first_entry`](#first_entry-block) block.
* `force_remediate_for_first_entries` - (Required) Whether Firewall Manager applies this setting to first-entry policy violations that involve conflicts between the custom entries and the policy entries. If `false`, Firewall Manager marks the network ACL as noncompliant and does not try to remediate.
* `force_remediate_for_last_entries` - (Required) Whether Firewall Manager applies this setting to last-entry policy violations that involve conflicts between the custom entries and the policy entries. If `false`, Firewall Manager marks the network ACL as noncompliant and does not try to remediate.
* `last_entry` - (Optional) Rules to run last in the Firewall Manager managed network ACLs. Firewall Manager creates entries with ID value between 32000 and 32766. See the [`last_entry`](#last_entry-block) block.

### `first_entry` Block

* `cidr_block` - (Optional) IPv4 network range to allow or deny, in CIDR notation.
* `egress` - (Required) Whether Firewall Manager creates an egress rule. If `false`, Firewall Manager creates an ingress rule.
* `icmp_type_code` - (Optional) ICMP protocol configuration specifying the ICMP type and code. See the [`icmp_type_code`](#icmp_type_code-block) block.
* `ipv6_cidr_block` - (Optional) IPv6 network range to allow or deny, in CIDR notation.
* `port_range` - (Optional) Port range configuration for the rule. See the [`port_range`](#port_range-block) block.
* `protocol` - (Required) Protocol number. A value of `-1` means all protocols.
* `rule_action` - (Required) Whether to allow or deny the traffic that matches the rule. Valid values: `allow`, `deny`.

### `last_entry` Block

* `cidr_block` - (Optional) IPv4 network range to allow or deny, in CIDR notation.
* `egress` - (Required) Whether Firewall Manager creates an egress rule. If `false`, Firewall Manager creates an ingress rule.
* `icmp_type_code` - (Optional) ICMP protocol configuration specifying the ICMP type and code. See the [`icmp_type_code`](#icmp_type_code-block) block.
* `ipv6_cidr_block` - (Optional) IPv6 network range to allow or deny, in CIDR notation.
* `port_range` - (Optional) Port range configuration for the rule. See the [`port_range`](#port_range-block) block.
* `protocol` - (Required) Protocol number. A value of `-1` means all protocols.
* `rule_action` - (Required) Whether to allow or deny the traffic that matches the rule. Valid values: `allow`, `deny`.

### `icmp_type_code` Block

* `code` - (Optional) ICMP code.
* `type` - (Optional) ICMP type.

### `port_range` Block

* `from` - (Optional) Beginning port number of the range.
* `to` - (Optional) Ending port number of the range.

### `network_firewall_policy` Block

* `firewall_deployment_model` - (Optional) Deployment model for the firewall policy. To use a distributed model, remove the `policy_option` section. Valid values are `CENTRALIZED` and `DISTRIBUTED`.

### `third_party_firewall_policy` Block

* `firewall_deployment_model` - (Optional) Deployment model for the third-party firewall policy. Valid values are `CENTRALIZED` and `DISTRIBUTED`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the policy.
* `id` - ID of the AWS Firewall Manager policy.
* `policy_update_token` - Unique identifier for each update to the policy.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Firewall Manager policies using the policy ID. For example:

```terraform
import {
  to = aws_fms_policy.example
  id = "5be49585-a7e3-4c49-dde1-a179fe4a619a"
}
```

Using `terraform import`, import Firewall Manager policies using the policy ID. For example:

```console
% terraform import aws_fms_policy.example 5be49585-a7e3-4c49-dde1-a179fe4a619a
```
