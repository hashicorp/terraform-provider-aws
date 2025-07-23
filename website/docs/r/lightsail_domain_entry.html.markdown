---
subcategory: "Lightsail"
layout: "aws"
page_title: "AWS: aws_lightsail_domain_entry"
description: |-
  Manages a Lightsail domain entry (DNS record).
---

# Resource: aws_lightsail_domain_entry

Manages a Lightsail domain entry (DNS record). Use this resource to define how DNS queries for your domain are handled.

~> **NOTE on `id`:** In an effort to simplify imports, this resource `id` field has been updated to the standard resource id separator, a comma (`,`). For backward compatibility, the previous separator (underscore `_`) can still be used to read and import existing resources. When state is refreshed, the `id` will be updated to use the new standard separator. The previous separator will be deprecated in a future major release.

## Example Usage

```terraform
resource "aws_lightsail_domain" "example" {
  domain_name = "example.com"
}

resource "aws_lightsail_domain_entry" "example" {
  domain_name = aws_lightsail_domain.example.domain_name
  name        = "www"
  type        = "A"
  target      = "127.0.0.1"
}
```

## Argument Reference

The following arguments are required:

* `domain_name` - (Required) Name of the Lightsail domain in which to create the entry.
* `name` - (Required) Name of the entry record.
* `target` - (Required) Target of the domain entry.
* `type` - (Required) Type of record. Valid values: `A`, `AAAA`, `CNAME`, `MX`, `NS`, `SOA`, `SRV`, `TXT`.

The following arguments are optional:

* `is_alias` - (Optional) Whether the entry should be an alias. Default: `false`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Combination of attributes to create a unique id: `name`,`domain_name`,`type`,`target`.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Lightsail Domain Entry using the id attribute. For example:

```terraform
import {
  to = aws_lightsail_domain_entry.example
  id = "www,example.com,A,127.0.0.1"
}
```

Using `terraform import`, import Lightsail Domain Entry using the id attribute. For example:

```console
% terraform import aws_lightsail_domain_entry.example www,example.com,A,127.0.0.1
```
