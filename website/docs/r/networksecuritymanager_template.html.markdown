---
subcategory: "Network Security Manager"
layout: "aws"
page_title: "AWS: aws_networksecuritymanager_template"
description: |-
  Terraform resource for managing an AWS Network Security Manager Template.
---

# Resource: aws_networksecuritymanager_template

Terraform resource for managing an AWS Network Security Manager Template. A template is an ordered list of published rules that together describe a firewall configuration; policies apply templates (and rules) to the resources of a scope. See [Templates](https://docs.aws.amazon.com/network-security-manager/latest/devguide/concepts.html) in the AWS Network Security Manager Developer Guide.

`rule_arns` is ordered: the position of an `INSPECTION` rule among the template's rules is its priority in the web ACL, so changing the order of the list is an update. Only published (`ACTIVE`) rules can be referenced.

~> A rule referenced by a published template cannot be deleted. When such a rule must be replaced (for example after a change of its `name` or `rule_type`), set `create_before_destroy = true` in the rule's `lifecycle` block so that the template is moved to the new rule before the old one is deleted; otherwise the apply fails with a `ConflictException`.

## Example Usage

### Basic Usage

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

resource "aws_networksecuritymanager_template" "example" {
  name          = "baseline"
  description   = "Baseline web ACL configuration"
  firewall_type = "WAF"
  rule_arns = [
    aws_networksecuritymanager_rule.default_action.arn,
    aws_networksecuritymanager_rule.common_rule_set.arn,
  ]
}
```

### Draft Template

```terraform
resource "aws_networksecuritymanager_template" "example" {
  name          = "baseline"
  firewall_type = "WAF"
  is_published  = false
  rule_arns     = [aws_networksecuritymanager_rule.default_action.arn]
}
```

### Replacing a Referenced Rule

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

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_networksecuritymanager_template" "example" {
  name          = "baseline"
  firewall_type = "WAF"
  rule_arns     = [aws_networksecuritymanager_rule.default_action.arn]
}
```

## Argument Reference

The following arguments are required:

* `firewall_type` - (Required, Forces new resource) Firewall type the template configures. Valid values: `WAF`.
* `name` - (Required, Forces new resource) Name of the template. Must start with an alphanumeric character and contain only alphanumeric characters, spaces, and `_.:/=+-@`. Template names are not unique; a template is identified by its `arn`.
* `rule_arns` - (Required) Ordered list of the ARNs of the published rules the template is made of, between 1 and 50 without duplicates. All rules must have the template's `firewall_type`.

The following arguments are optional:

* `description` - (Optional) Description of the template, up to 256 characters.
* `is_published` - (Optional) Whether the template is published. When `true` (the default) the template is `ACTIVE` and can be used by policies. When `false` the template is saved as a `DRAFT`; setting it to `false` on a published template saves a draft on top of the published version, which stays in use until the draft is published.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#region).
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the template.
* `has_published_version` - Whether a published version exists beneath a pending draft.
* `status` - Status of the template. Valid values: `DRAFT`, `ACTIVE`, `DISABLED`.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `template_id` - Service-generated ID of the template.
* `updated_at` - Time when the template was last updated.
* `version` - Version of the template. Incremented on every update.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_networksecuritymanager_template.example
  identity = {
    arn = "arn:aws:network-security-manager:us-east-1:123456789012:template:1uas2vkysvpubwb3kdwrm8b4c"
  }
}
```

### Identity Schema

#### Required

* `arn` - (String) ARN of the template.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Network Security Manager Template using the `arn`. For example:

```terraform
import {
  to = aws_networksecuritymanager_template.example
  id = "arn:aws:network-security-manager:us-east-1:123456789012:template:1uas2vkysvpubwb3kdwrm8b4c"
}
```

Using `terraform import`, import Network Security Manager Template using the `arn`. For example:

```console
% terraform import aws_networksecuritymanager_template.example arn:aws:network-security-manager:us-east-1:123456789012:template:1uas2vkysvpubwb3kdwrm8b4c
```
