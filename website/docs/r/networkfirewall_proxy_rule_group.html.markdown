---
subcategory: "Network Firewall"
layout: "aws"
page_title: "AWS: aws_networkfirewall_proxy_rule_group"
description: |-
  Manages an AWS Network Firewall Proxy Rule Group resource.
---

# Resource: aws_networkfirewall_proxy_rule_group

Manages an AWS Network Firewall Proxy Rule Group resource. A proxy rule group is a container for proxy rules that can be referenced by a proxy configuration.

~> **NOTE:** This resource creates an empty proxy rule group. Use the [`aws_networkfirewall_proxy_rules_exclusive`](networkfirewall_proxy_rules_exclusive.html) resource to add rules to the group.

## Example Usage

### Basic Usage

```terraform
resource "aws_networkfirewall_proxy_rule_group" "example" {
  name = "example"
}
```

### With Description and Tags

```terraform
resource "aws_networkfirewall_proxy_rule_group" "example" {
  name        = "example"
  description = "Example proxy rule group for HTTP traffic"

  tags = {
    Name        = "example"
    Environment = "production"
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Descriptive name of the proxy rule group.

The following arguments are optional:

* `description` - (Optional) Description of the proxy rule group.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Proxy Rule Group.
* `id` - ARN of the Proxy Rule Group.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `update_token` - Token used for optimistic locking.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Network Firewall Proxy Rule Group using the `arn`. For example:

```terraform
import {
  to = aws_networkfirewall_proxy_rule_group.example
  id = "arn:aws:network-firewall:us-west-2:123456789012:proxy-rule-group/example"
}
```

Using `terraform import`, import Network Firewall Proxy Rule Group using the `arn`. For example:

```console
% terraform import aws_networkfirewall_proxy_rule_group.example arn:aws:network-firewall:us-west-2:123456789012:proxy-rule-group/example
```
