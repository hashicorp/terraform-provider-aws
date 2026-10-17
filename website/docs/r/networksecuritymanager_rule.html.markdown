---
subcategory: "Network Security Manager"
layout: "aws"
page_title: "AWS: aws_networksecuritymanager_rule"
description: |-
  Terraform resource for managing an AWS Network Security Manager Rule.
---

# Resource: aws_networksecuritymanager_rule

Terraform resource for managing an AWS Network Security Manager Rule. A rule is the smallest unit of network security configuration: either a single AWS WAF web ACL setting (a `CONFIGURATION` rule) or an AWS WAF rule group to evaluate traffic with (an `INSPECTION` rule). Rules are grouped into templates and policies, which a deployment applies to the resources of a scope.

The rule's `configuration` is a JSON document whose structure depends on the rule type. For a `CONFIGURATION` rule, the document has exactly one root key naming the web ACL setting it contains: `DefaultAction`, `VisibilityConfig`, `CaptchaConfig`, `ChallengeConfig`, `CustomResponseBodies`, `LoggingConfiguration`, `DataProtectionConfig`, `AssociationConfig`, `OnSourceDDoSProtectionConfig` or `TokenDomains`. For an `INSPECTION` rule, the root key is `PreProcessFirewallManagerRuleGroups` or `PostProcessFirewallManagerRuleGroups` and holds a list with exactly one rule group; rule group priority is derived from position and must not be supplied. The nested structures follow the [AWS WAF API](https://docs.aws.amazon.com/waf/latest/APIReference/API_WebACL.html). See [Writing rule configurations](https://docs.aws.amazon.com/network-security-manager/latest/devguide/concepts.html) in the AWS Network Security Manager Developer Guide.

~> The API returns numeric values in the configuration document as strings. The resource treats a number and its string form as equal, so a configured `"ImmunityTime": 300` does not produce a diff.

## Example Usage

### Configuration Rule

```terraform
resource "aws_networksecuritymanager_rule" "example" {
  name          = "default-action-block"
  description   = "Block requests that match no rule"
  firewall_type = "WAF"
  rule_type     = "CONFIGURATION"
  configuration = jsonencode({
    DefaultAction = {
      Block = {}
    }
  })
}
```

### Inspection Rule

```terraform
resource "aws_networksecuritymanager_rule" "example" {
  name          = "common-rule-set"
  firewall_type = "WAF"
  rule_type     = "INSPECTION"
  configuration = jsonencode({
    PreProcessFirewallManagerRuleGroups = [{
      Name = "AWSManagedRulesCommonRuleSet"
      FirewallManagerStatement = {
        ManagedRuleGroupStatement = {
          VendorName = "AWS"
          Name       = "AWSManagedRulesCommonRuleSet"
        }
      }
      OverrideAction = {
        None = {}
      }
      VisibilityConfig = {
        SampledRequestsEnabled   = true
        CloudWatchMetricsEnabled = true
        MetricName               = "AWSManagedRulesCommonRuleSet"
      }
    }]
  })
}
```

### Draft Rule

```terraform
resource "aws_networksecuritymanager_rule" "example" {
  name          = "visibility-config"
  firewall_type = "WAF"
  rule_type     = "CONFIGURATION"
  is_published  = false
  configuration = jsonencode({
    VisibilityConfig = {
      SampledRequestsEnabled   = true
      CloudWatchMetricsEnabled = true
      MetricName               = "example"
    }
  })
}
```

## Argument Reference

The following arguments are required:

* `configuration` - (Required) Firewall configuration of the rule, as a JSON document. See the description above for its structure.
* `firewall_type` - (Required, Forces new resource) Firewall type the rule configures. Valid values: `WAF`.
* `name` - (Required, Forces new resource) Name of the rule. Must start with an alphanumeric character and contain only alphanumeric characters, spaces, and `_.:/=+-@`. Rule names are not unique; a rule is identified by its `arn`.
* `rule_type` - (Required, Forces new resource) Type of the rule. `CONFIGURATION` rules contain a web ACL setting, `INSPECTION` rules contain a rule group. Valid values: `CONFIGURATION`, `INSPECTION`.

The following arguments are optional:

* `description` - (Optional) Description of the rule, up to 256 characters. Must contain only alphanumeric characters, spaces, and `_.:/=+-@`. Removing the argument clears the description, which rewrites the rule and increments its `version`.
* `is_published` - (Optional) Whether the rule is published. When `true` (the default) the rule is `ACTIVE` and can be used by templates and policies. When `false` the rule is saved as a `DRAFT`; setting it to `false` on a published rule saves a draft on top of the published version, which stays in use until the draft is published.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#region).
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the rule.
* `has_published_version` - Whether a published version exists beneath a pending draft.
* `rule_id` - Service-generated ID of the rule.
* `status` - Status of the rule. Valid values: `DRAFT`, `ACTIVE`, `DISABLED`.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `updated_at` - Time when the rule was last updated.
* `version` - Version of the rule. Incremented on every update.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_networksecuritymanager_rule.example
  identity = {
    arn = "arn:aws:network-security-manager:us-east-1:123456789012:rule:1uas1v62rx5yjceq3o6a6wcpj"
  }
}
```

### Identity Schema

#### Required

* `arn` - (String) ARN of the rule.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Network Security Manager Rule using the `arn`. For example:

```terraform
import {
  to = aws_networksecuritymanager_rule.example
  id = "arn:aws:network-security-manager:us-east-1:123456789012:rule:1uas1v62rx5yjceq3o6a6wcpj"
}
```

Using `terraform import`, import Network Security Manager Rule using the `arn`. For example:

```console
% terraform import aws_networksecuritymanager_rule.example arn:aws:network-security-manager:us-east-1:123456789012:rule:1uas1v62rx5yjceq3o6a6wcpj
```
