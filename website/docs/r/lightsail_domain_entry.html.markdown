---
subcategory: "Lightsail"
layout: "aws"
page_title: "AWS: aws_lightsail_domain_entry"
description: |-
  Provides an Lightsail Domain Entry
---

# Resource: aws_lightsail_domain_entry

Creates a domain entry resource

~> **NOTE on `id`:** In an effort to simplify imports, this resource `id` field has been updated to the standard resource id separator, a comma (`,`). For backward compatibility, the previous separator (underscore `_`) can still be used to read and import existing resources. When state is refreshed, the `id` will be updated to use the new standard separator. The previous separator will be deprecated in a future major release.

## Example Usage

```terraform
resource "aws_lightsail_domain" "test" {
  domain_name = "mydomain.com"
}

resource "aws_lightsail_domain_entry" "test" {
  domain_name = aws_lightsail_domain.domain_test.domain_name
  name        = "www"
  type        = "A"
  target      = "127.0.0.1"
}

```

## Argument Reference

This resource supports the following arguments:

* `domain_name` - (Required) The name of the Lightsail domain in which to create the entry
* `name` - (Required) Name of the entry record
* `type` - (Required) Type of record
* `target` - (Required) Target of the domain entry
* `is_alias` - (Optional) If the entry should be an alias Defaults to `false`

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - A combination of attributes to create a unique id: `name`,`domain_name`,`type`,`target`

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_lightsail_domain_entry` using the id attribute. For example:

```terraform
import {
  to = aws_lightsail_domain_entry.example
  id = "www,mydomain.com,A,127.0.0.1"
}
```

Using `terraform import`, import `aws_lightsail_domain_entry` using the id attribute. For example:

```console
% terraform import aws_lightsail_domain_entry.example www,mydomain.com,A,127.0.0.1
```
