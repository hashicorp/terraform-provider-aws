---
subcategory: "Network Security Manager"
layout: "aws"
page_title: "AWS: aws_networksecuritymanager_policy"
description: |-
  Manages an AWS Network Security Manager Policy.
---

# Resource: aws_networksecuritymanager_policy

Manages an AWS Network Security Manager Policy. A policy combines templates and rules with enforcement settings for a firewall type, such as AWS WAF or AWS Shield Advanced.

~> **NOTE:** Terraform always publishes the policy on create and update, so a Terraform-managed policy never has a pending draft version.

## Example Usage

### AWS WAF Policy

```terraform
resource "aws_networksecuritymanager_policy" "example" {
  name          = "baseline-waf"
  firewall_type = "WAF"
  priority      = 1

  template_or_rule {
    template_arn = aws_networksecuritymanager_template.example.arn
  }

  template_or_rule {
    rule_arn = aws_networksecuritymanager_rule.example.arn
  }

  policy_configuration {
    remediation_enabled = true
    resources_clean_up  = false

    waf_config {
      conflict_resolution                  = "MERGE_WHERE_APPLICABLE"
      existing_customer_web_acl_resolution = "RETROFIT"
    }
  }
}
```

### AWS Shield Advanced Policy

```terraform
resource "aws_networksecuritymanager_policy" "example" {
  name          = "baseline-shield"
  firewall_type = "SHIELD_ADVANCED"
  priority      = 1

  policy_configuration {
    remediation_enabled = true
    resources_clean_up  = true
  }
}
```

## Argument Reference

The following arguments are required:

* `firewall_type` - (Required) Firewall type of the policy. Valid values: `WAF`, `SHIELD_ADVANCED`. Changing this value forces a new resource.
* `name` - (Required) Name of the policy. Changing this value forces a new resource.
* `policy_configuration` - (Required) Settings that control the policy's behavior. See [`policy_configuration`](#policy_configuration-block) below.
* `priority` - (Required) Priority of the policy. A lower number indicates a higher priority. Minimum value of `1`.

The following arguments are optional:

* `description` - (Optional) Description of the policy.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `template_or_rule` - (Optional) Ordered list of templates and rules to associate with the policy. For `WAF` policies, specify 1 to 100 entries, of which at most 2 can be templates. Must be omitted for `SHIELD_ADVANCED` policies. See [`template_or_rule`](#template_or_rule-block) below.

### `policy_configuration` Block

* `remediation_enabled` - (Required) Whether to automatically remediate noncompliant resources.
* `resources_clean_up` - (Required) Whether to automatically remove the resources the service created when they are no longer needed. Must be `true` for `SHIELD_ADVANCED` policies.
* `waf_config` - (Optional) AWS WAF-specific settings. Required for `WAF` policies. See [`waf_config`](#waf_config-block) below.

### `waf_config` Block

* `conflict_resolution` - (Required) Conflict-resolution strategy. Valid values: `MERGE_WHERE_APPLICABLE`.
* `existing_customer_web_acl_resolution` - (Required) How remediation is handled when a resource already has a customer-created web ACL. Valid values: `RETROFIT`, `OVERRIDE_ASSOCIATION`, `NO_REMEDIATION`.

### `template_or_rule` Block

Exactly one of the following must be specified:

* `rule_arn` - (Optional) ARN of the rule.
* `template_arn` - (Optional) ARN of the template.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the policy.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_networksecuritymanager_policy.example
  identity = {
    arn = "arn:aws:network-security-manager:us-east-1:123456789012:policy:abcd1234efgh"
  }
}

resource "aws_networksecuritymanager_policy" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `arn` (String) ARN of the policy.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Network Security Manager Policy using the `arn`. For example:

```terraform
import {
  to = aws_networksecuritymanager_policy.example
  id = "arn:aws:network-security-manager:us-east-1:123456789012:policy:abcd1234efgh"
}
```

Using `terraform import`, import Network Security Manager Policy using the `arn`. For example:

```console
% terraform import aws_networksecuritymanager_policy.example arn:aws:network-security-manager:us-east-1:123456789012:policy:abcd1234efgh
```
