---
subcategory: "Connect"
layout: "aws"
page_title: "AWS: aws_connect_approved_origin"
description: |-
  Provides an Amazon Connect Approved Origin resource.
---

# Resource: aws_connect_approved_origin

Provides an Amazon Connect Approved Origin resource.
Approved origins allowlist external domains for embedding the Contact Control Panel (CCP).

## Example Usage
```terraform
resource "aws_connect_instance" "example" {
  identity_management_type = "CONNECT_MANAGED"
  inbound_calls_enabled    = true
  instance_alias           = "example"
  outbound_calls_enabled   = true
}

resource "aws_connect_approved_origin" "example" {
  instance_id = aws_connect_instance.example.id
  origin      = "https://example.com"
}
```

## Argument Reference

The following arguments are required:

* `instance_id` - (Required, Forces new resource) The identifier of the Amazon Connect instance.
* `origin` - (Required, Forces new resource) Domain name to be added as an approved origin, e.g., `https://example.com`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - A composite ID in the format `instance_id:origin`.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Connect Approved Origins using the `instance_id` and `origin` separated by `:`. For example:
```terraform
import {
  to = aws_connect_approved_origin.example
  id = "aaaaaaaa-bbbb-cccc-dddd-111111111111:https://example.com"
}
```

Using `terraform import`, import Connect Approved Origins using the `instance_id` and `origin` separated by `:`. For example:
```console
% terraform import aws_connect_approved_origin.example aaaaaaaa-bbbb-cccc-dddd-111111111111:https://example.com
```
