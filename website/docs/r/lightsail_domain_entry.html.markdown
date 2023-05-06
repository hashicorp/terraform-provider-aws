---
subcategory: "Lightsail"
layout: "aws"
page_title: "AWS: aws_lightsail_domain_entry"
description: |-
  Provides an Lightsail Domain Entry
---

# Resource: aws_lightsail_domain_entry

Creates a domain entry resource

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

The following arguments are supported:

* `domain_name` - (Required) The name of the Lightsail domain in which to create the entry
* `name` - (Required) Name of the entry record
* `type` - (Required) Type of record
* `target` - (Required) Target of the domain entry
* `is_alias` - (Optional) If the entry should be an alias Defaults to `false`

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - A combination of attributes to create a unique id: `name`\_`domain_name`\_`type`\_`target`

## Import

`aws_lightsail_domain_entry` can be imported by using the id attribute, e.g.,

```
$ terraform import aws_lightsail_domain_entry.example www_mydomain.com_A_127.0.0.1
```
