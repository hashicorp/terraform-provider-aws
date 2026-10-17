---
subcategory: "Network Security Manager"
layout: "aws"
page_title: "AWS: aws_networksecuritymanager_policy"
description: |-
  Terraform resource for managing an AWS Network Security Manager Policy.
---

# Resource: aws_networksecuritymanager_policy

Terraform resource for managing an AWS Network Security Manager Policy. A policy combines published templates and rules with enforcement settings for one firewall type; a deployment applies policies to the resources of a scope. See [Policies](https://docs.aws.amazon.com/network-security-manager/latest/devguide/concepts.html) in the AWS Network Security Manager Developer Guide.

An AWS WAF policy (`firewall_type = "WAF"`) needs a `waf_config` block and between 1 and 100 `associated_template_and_rule` blocks, of which at most 2 may reference templates. The blocks are ordered: the position of a template or rule in the list is its position in the web ACL the policy builds, so changing the order is an update. Only published (`ACTIVE`) templates and rules can be referenced.

An AWS Shield Advanced policy (`firewall_type = "SHIELD_ADVANCED"`) takes neither a `waf_config` block nor `associated_template_and_rule` blocks.

`priority` is unique across the policies of an account: two policies cannot share a priority, so swapping the priorities of two policies takes two applies through a third value.

~> A rule or template referenced by a published policy cannot be deleted. When such a rule or template must be replaced (for example after a change of its `name`), set `create_before_destroy = true` in its `lifecycle` block so that the policy is moved to the new resource before the old one is deleted; otherwise the apply fails with a `ConflictException`. A policy that exists only as a draft (`is_published = false` and never published) does not protect its references: they can be deleted, after which the draft cannot be published until its list is fixed.

## Example Usage

### WAF Policy

```terraform
resource "aws_networksecuritymanager_rule" "default_action" {
  name          = "default-action-block"
  firewall_type = "WAF"
  rule_type     = "CONFIGURATION"
  configuration = jsonencode({
    DefaultAction = {
      Block = {}
    }
  })
}

resource "aws_networksecuritymanager_rule" "common_rule_set" {
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

resource "aws_networksecuritymanager_template" "baseline" {
  name          = "baseline"
  firewall_type = "WAF"
  rule_arns     = [aws_networksecuritymanager_rule.default_action.arn]
}

resource "aws_networksecuritymanager_policy" "example" {
  name          = "web-acl-baseline"
  description   = "Baseline web ACL for every application load balancer"
  firewall_type = "WAF"
  priority      = 1

  policy_configuration {
    remediation_enabled = true
    resources_clean_up  = true

    waf_config {
      conflict_resolution                  = "MERGE_WHERE_APPLICABLE"
      existing_customer_web_acl_resolution = "RETROFIT"
    }
  }

  associated_template_and_rule {
    template_arn = aws_networksecuritymanager_template.baseline.arn
  }

  associated_template_and_rule {
    rule_arn = aws_networksecuritymanager_rule.common_rule_set.arn
  }
}
```

### Shield Advanced Policy

```terraform
resource "aws_networksecuritymanager_policy" "example" {
  name          = "shield-advanced-protection"
  firewall_type = "SHIELD_ADVANCED"
  priority      = 2

  policy_configuration {
    remediation_enabled = true
    resources_clean_up  = false
  }
}
```

### Draft Policy

```terraform
resource "aws_networksecuritymanager_policy" "example" {
  name          = "web-acl-baseline"
  firewall_type = "WAF"
  is_published  = false
  priority      = 1

  policy_configuration {
    remediation_enabled = false
    resources_clean_up  = false

    waf_config {
      conflict_resolution                  = "MERGE_WHERE_APPLICABLE"
      existing_customer_web_acl_resolution = "NO_REMEDIATION"
    }
  }

  associated_template_and_rule {
    rule_arn = aws_networksecuritymanager_rule.default_action.arn
  }
}
```

### Replacing a Referenced Template

```terraform
resource "aws_networksecuritymanager_template" "baseline" {
  name          = "baseline"
  firewall_type = "WAF"
  rule_arns     = [aws_networksecuritymanager_rule.default_action.arn]

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_networksecuritymanager_policy" "example" {
  name          = "web-acl-baseline"
  firewall_type = "WAF"
  priority      = 1

  policy_configuration {
    remediation_enabled = false
    resources_clean_up  = false

    waf_config {
      conflict_resolution                  = "MERGE_WHERE_APPLICABLE"
      existing_customer_web_acl_resolution = "NO_REMEDIATION"
    }
  }

  associated_template_and_rule {
    template_arn = aws_networksecuritymanager_template.baseline.arn
  }
}
```

## Argument Reference

The following arguments are required:

* `firewall_type` - (Required, Forces new resource) Firewall type the policy enforces. Valid values: `WAF`, `SHIELD_ADVANCED`.
* `name` - (Required, Forces new resource) Name of the policy. Must start with an alphanumeric character and contain only alphanumeric characters, spaces, and `_.:/=+-@`. Policy names are not unique; a policy is identified by its `arn`.
* `policy_configuration` - (Required) Settings that control the policy's behavior. See [`policy_configuration` Block](#policy_configuration-block) below.
* `priority` - (Required) Priority of the policy, at least `1`. A lower number is a higher priority. Unique across the policies of the account.

The following arguments are optional:

* `associated_template_and_rule` - (Optional) Ordered list of the published templates and rules the policy applies. Required for a `WAF` policy, between 1 and 100 blocks of which at most 2 reference templates, each referencing a distinct template or rule; not allowed for a `SHIELD_ADVANCED` policy. See [`associated_template_and_rule` Block](#associated_template_and_rule-block) below.
* `description` - (Optional) Description of the policy, up to 256 characters.
* `is_published` - (Optional) Whether the policy is published. When `true` (the default) the policy is `ACTIVE` and can be used by deployments. When `false` the policy is saved as a `DRAFT`; setting it to `false` on a published policy saves a draft on top of the published version, which stays in use until the draft is published.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#region).
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `associated_template_and_rule` Block

Exactly one of the following arguments must be set in each block:

* `rule_arn` - (Optional) ARN of a published rule.
* `template_arn` - (Optional) ARN of a published template.

### `policy_configuration` Block

The `policy_configuration` block supports the following arguments:

* `remediation_enabled` - (Required) Whether noncompliant resources are remediated automatically.
* `resources_clean_up` - (Required) Whether the resources the service created are removed automatically when they are no longer needed.
* `waf_config` - (Optional) AWS WAF settings. Required for a `WAF` policy, not allowed for a `SHIELD_ADVANCED` policy. See [`waf_config` Block](#waf_config-block) below.

### `waf_config` Block

The `waf_config` block supports the following arguments:

* `conflict_resolution` - (Required) How conflicting settings are resolved. Valid values: `MERGE_WHERE_APPLICABLE`.
* `existing_customer_web_acl_resolution` - (Required) How a resource that already has a customer-created web ACL is remediated. Valid values: `RETROFIT`, `OVERRIDE_ASSOCIATION`, `NO_REMEDIATION`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the policy.
* `has_published_version` - Whether a published version exists beneath a pending draft.
* `policy_id` - Service-generated ID of the policy.
* `status` - Status of the policy. Valid values: `DRAFT`, `ACTIVE`, `DISABLED`.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `updated_at` - Time when the policy was last updated.
* `version` - Version of the policy. Incremented on every update.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_networksecuritymanager_policy.example
  identity = {
    arn = "arn:aws:network-security-manager:us-east-1:123456789012:policy:1uas2vkysvpubwb3kdwrm8b4c"
  }
}
```

### Identity Schema

#### Required

* `arn` - (String) ARN of the policy.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Network Security Manager Policy using the `arn`. For example:

```terraform
import {
  to = aws_networksecuritymanager_policy.example
  id = "arn:aws:network-security-manager:us-east-1:123456789012:policy:1uas2vkysvpubwb3kdwrm8b4c"
}
```

Using `terraform import`, import Network Security Manager Policy using the `arn`. For example:

```console
% terraform import aws_networksecuritymanager_policy.example arn:aws:network-security-manager:us-east-1:123456789012:policy:1uas2vkysvpubwb3kdwrm8b4c
```
