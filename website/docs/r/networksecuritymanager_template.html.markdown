---
subcategory: "Network Security Manager"
layout: "aws"
page_title: "AWS: aws_networksecuritymanager_template"
description: |-
  Manages an AWS Network Security Manager Template.
---

# Resource: aws_networksecuritymanager_template

Manages an AWS Network Security Manager Template. A template groups one or more rules so they can be reused across policies.

## Example Usage

### Basic Usage

```terraform
resource "aws_networksecuritymanager_rule" "example" {
  name          = "example"
  firewall_type = "WAF"
  type          = "CONFIGURATION"
  configuration = jsonencode({
    DefaultAction = {
      Allow = {}
    }
  })
}

resource "aws_networksecuritymanager_template" "example" {
  name          = "example"
  description   = "Baseline web ACL settings"
  firewall_type = "WAF"
  rule_arns     = [aws_networksecuritymanager_rule.example.arn]
}
```

## Argument Reference

The following arguments are required:

* `firewall_type` - (Required, Forces new resource) Firewall type of the template. Valid values: `WAF`.
* `name` - (Required, Forces new resource) Name of the template.
* `rule_arns` - (Required) Ordered list of ARNs of the rules to associate with the template. Specify 1 to 50 rules. Each rule can be associated only once.

The following arguments are optional:

* `description` - (Optional) Description of the template.
* `is_published` - (Optional) Whether to publish the template. When `true`, the template is saved in the published (`ACTIVE`) state. When `false`, it is saved as a draft (`DRAFT`); for a published template, the draft is saved on top of the published version. Defaults to `true`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#region).
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the template.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_networksecuritymanager_template.example
  identity = {
    arn = "arn:aws:network-security-manager:us-east-1:123456789012:template:0123456789abcdef"
  }
}
```

### Identity Schema

#### Required

* `arn` (String) ARN of the template.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Network Security Manager Template using the `arn`. For example:

```terraform
import {
  to = aws_networksecuritymanager_template.example
  id = "arn:aws:network-security-manager:us-east-1:123456789012:template:0123456789abcdef"
}
```

Using `terraform import`, import Network Security Manager Template using the `arn`. For example:

```console
% terraform import aws_networksecuritymanager_template.example arn:aws:network-security-manager:us-east-1:123456789012:template:0123456789abcdef
```
