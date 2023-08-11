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

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SimpleDB Domains using the `name`. For example:

```terraform
import {
  to = aws_simpledb_domain.users
  id = "users"
}
```

Using `terraform import`, import SimpleDB Domains using the `name`. For example:

```console
% terraform import aws_simpledb_domain.users users
```
