---
subcategory: "Network Security Manager"
layout: "aws"
page_title: "AWS: aws_networksecuritymanager_rule"
description: |-
  Manages an AWS Network Security Manager Rule.
---

# Resource: aws_networksecuritymanager_rule

Manages an AWS Network Security Manager Rule. A rule defines a firewall configuration to enforce, such as an AWS WAF web ACL setting or an AWS WAF rule group. Rules can be grouped into templates and referenced from policies.

Every change to a rule publishes a new version of the rule. The resource never leaves an unpublished draft version behind.

## Example Usage

### Basic Usage

```terraform
resource "aws_networksecuritymanager_rule" "example" {
  name          = "example"
  description   = "Allow requests that match no other rule"
  firewall_type = "WAF"
  type          = "CONFIGURATION"
  configuration = jsonencode({
    DefaultAction = {
      Allow = {}
    }
  })
}
```

## Argument Reference

The following arguments are required:

* `configuration` - (Required) Firewall configuration of the rule, as a JSON document. The structure depends on `firewall_type` and `type`. For an AWS WAF `CONFIGURATION` rule, provide a single web ACL setting, such as `DefaultAction` or `VisibilityConfig`. For an AWS WAF `INSPECTION` rule, provide an AWS WAF rule group. For details, see [Writing rule configurations](https://docs.aws.amazon.com/network-security-manager/latest/devguide/what-is.html) in the AWS Network Security Manager Developer Guide.
* `firewall_type` - (Required, Forces new resource) Firewall type of the rule. Valid values: `WAF`.
* `name` - (Required, Forces new resource) Name of the rule.
* `type` - (Required) Type of the rule. `CONFIGURATION` rules contain firewall settings and `INSPECTION` rules contain rule groups. Valid values: `CONFIGURATION`, `INSPECTION`.

The following arguments are optional:

* `description` - (Optional) Description of the rule.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#region).
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the rule.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_networksecuritymanager_rule.example
  identity = {
    arn = "arn:aws:network-security-manager:us-east-1:123456789012:rule:0123456789abcdef"
  }
}
```

### Identity Schema

#### Required

* `arn` (String) ARN of the rule.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Network Security Manager Rule using the `arn`. For example:

```terraform
import {
  to = aws_networksecuritymanager_rule.example
  id = "arn:aws:network-security-manager:us-east-1:123456789012:rule:0123456789abcdef"
}
```

Using `terraform import`, import Network Security Manager Rule using the `arn`. For example:

```console
% terraform import aws_networksecuritymanager_rule.example arn:aws:network-security-manager:us-east-1:123456789012:rule:0123456789abcdef
```
