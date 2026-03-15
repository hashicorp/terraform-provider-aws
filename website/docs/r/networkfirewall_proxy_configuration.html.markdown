---
subcategory: "Network Firewall"
layout: "aws"
page_title: "AWS: aws_networkfirewall_proxy_configuration"
description: |-
  Manages an AWS Network Firewall Proxy Configuration.
---

# Resource: aws_networkfirewall_proxy_configuration

Manages an AWS Network Firewall Proxy Configuration.

## Example Usage

### Basic Usage

```terraform
resource "aws_networkfirewall_proxy_configuration" "example" {
  name = "example"

  default_rule_phase_actions {
    pre_dns       = "ALLOW"
    pre_request   = "ALLOW"
    post_response = "ALLOW"
  }
}
```

### With Description and Tags

```terraform
resource "aws_networkfirewall_proxy_configuration" "example" {
  name        = "example"
  description = "Example proxy configuration"

  default_rule_phase_actions {
    pre_dns       = "DROP"
    pre_request   = "ALLOW"
    post_response = "ALLOW"
  }

  tags = {
    Name        = "example"
    Environment = "production"
  }
}
```

## Argument Reference

The following arguments are required:

* `default_rule_phase_actions` - (Required) Default actions to take on proxy traffic. See [Default Rule Phase Actions](#default-rule-phase-actions) below.
* `name` - (Required) Descriptive name of the proxy configuration.

The following arguments are optional:

* `description` - (Optional) Description of the proxy configuration.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Default Rule Phase Actions

The `default_rule_phase_actions` block supports the following:

* `post_response` - (Required) Default action for the POST_RESPONSE phase. Valid values: `ALLOW`, `DROP`.
* `pre_dns` - (Required) Default action for the PRE_DNS phase. Valid values: `ALLOW`, `DROP`.
* `pre_request` - (Required) Default action for the PRE_REQUEST phase. Valid values: `ALLOW`, `DROP`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Proxy Configuration.
* `id` - ARN of the Proxy Configuration.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `update_token` - Token used for optimistic locking.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Network Firewall Proxy Configuration using the `arn`. For example:

```terraform
import {
  to = aws_networkfirewall_proxy_configuration.example
  id = "arn:aws:network-firewall:us-west-2:123456789012:proxy-configuration/example"
}
```

Using `terraform import`, import Network Firewall Proxy Configuration using the `arn`. For example:

```console
% terraform import aws_networkfirewall_proxy_configuration.example arn:aws:network-firewall:us-west-2:123456789012:proxy-configuration/example
```
