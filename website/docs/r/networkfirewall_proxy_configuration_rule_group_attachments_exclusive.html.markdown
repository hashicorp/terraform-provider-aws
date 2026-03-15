---
subcategory: "Network Firewall"
layout: "aws"
page_title: "AWS: aws_networkfirewall_proxy_configuration_rule_group_attachments_exclusive"
description: |-
  Manages an AWS Network Firewall Proxy Configuration Rule Group Attachments Exclusive resource.
---

# Resource: aws_networkfirewall_proxy_configuration_rule_group_attachments_exclusive

Manages an AWS Network Firewall Proxy Configuration Rule Group Attachments Exclusive resource. This resource attaches proxy rule groups to a proxy configuration.

~> **NOTE:** This resource requires an existing [`aws_networkfirewall_proxy_configuration`](networkfirewall_proxy_configuration.html) and [`aws_networkfirewall_proxy_rule_group`](networkfirewall_proxy_rule_group.html).

## Example Usage

### Basic Usage

```terraform
resource "aws_networkfirewall_proxy_configuration" "example" {
  name = "example"

  default_rule_phase_actions {
    post_response = "ALLOW"
    pre_dns       = "ALLOW"
    pre_request   = "ALLOW"
  }
}

resource "aws_networkfirewall_proxy_rule_group" "example" {
  name = "example"
}

resource "aws_networkfirewall_proxy_configuration_rule_group_attachments_exclusive" "example" {
  proxy_configuration_arn = aws_networkfirewall_proxy_configuration.example.arn

  rule_group {
    proxy_rule_group_name = aws_networkfirewall_proxy_rule_group.example.name
  }
}
```

### Multiple Rule Groups

```terraform
resource "aws_networkfirewall_proxy_configuration" "example" {
  name = "example"

  default_rule_phase_actions {
    post_response = "ALLOW"
    pre_dns       = "ALLOW"
    pre_request   = "ALLOW"
  }
}

resource "aws_networkfirewall_proxy_rule_group" "first" {
  name = "first"
}

resource "aws_networkfirewall_proxy_rule_group" "second" {
  name = "second"
}

resource "aws_networkfirewall_proxy_configuration_rule_group_attachments_exclusive" "example" {
  proxy_configuration_arn = aws_networkfirewall_proxy_configuration.example.arn

  rule_group {
    proxy_rule_group_name = aws_networkfirewall_proxy_rule_group.first.name
  }

  rule_group {
    proxy_rule_group_name = aws_networkfirewall_proxy_rule_group.second.name
  }
}
```

## Argument Reference

The following arguments are required:

* `proxy_configuration_arn` - (Required) ARN of the proxy configuration to attach rule groups to.

The following arguments are optional:

* `rule_group` - (Optional) One or more rule group blocks. See [Rule Group](#rule-group) below.

### Rule Group

Each `rule_group` block supports the following:

* `proxy_rule_group_name` - (Required) Name of the proxy rule group to attach.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - ARN of the Proxy Configuration.
* `update_token` - Token used for optimistic locking.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Network Firewall Proxy Configuration Rule Group Attachments Exclusive using the `proxy_configuration_arn`. For example:

```terraform
import {
  to = aws_networkfirewall_proxy_configuration_rule_group_attachments_exclusive.example
  id = "arn:aws:network-firewall:us-west-2:123456789012:proxy-configuration/example"
}
```

Using `terraform import`, import Network Firewall Proxy Configuration Rule Group Attachments Exclusive using the `proxy_configuration_arn`. For example:

```console
% terraform import aws_networkfirewall_proxy_configuration_rule_group_attachments_exclusive.example arn:aws:network-firewall:us-west-2:123456789012:proxy-configuration/example
```
