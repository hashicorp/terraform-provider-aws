---
subcategory: "SDB (SimpleDB)"
layout: "aws"
page_title: "AWS: aws_simpledb_domain"
description: |-
  Provides a SimpleDB domain resource.
---

# Resource: aws_simpledb_domain

Provides a SimpleDB domain resource

## Example Usage

```terraform
resource "aws_simpledb_domain" "users" {
  name = "users"
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) The name of the SimpleDB domain

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The name of the SimpleDB domain

## Import

SimpleDB Domains can be imported using the `name`, e.g.,

```
$ terraform import aws_simpledb_domain.users users
```
