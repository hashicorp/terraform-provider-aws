---
layout: "aws"
page_title: "AWS: aws_organizations_unit"
sidebar_current: "docs-aws-resource-organizations-unit"
description: |-
  Provides a resource to create an organizational unit.
---

# aws_organizations_unit

Provides a resource to create an organizational unit.

## Example Usage:

```hcl
data "aws_organizations_unit" "root" {
  root = true
}

resource "aws_organizations_unit" "tenants" {
  parent_id = "${data.aws_organizations_unit.root.id}"
  name = "tenants"
}

resource "aws_organizations_account" "tenant" {
  name  = "my_new_account"
  email = "john@doe.org"
  parent_id = "${data.aws_organizations_unit.tenants.id}"
}
```

## Argument Reference

The following arguments are supported:

* `name` - The name for the organizational unit
* `parent_id` - ID of the parent organizational unit, which may be the root

## Attributes Reference

The following additional attributes are exported:

* `arn` - ARN of the organization
* `id` - Identifier of the organization
* `parent_id` - ID of the parent organizational unit

## Import

The AWS organization can be imported by using the `id`, e.g.

```
$ terraform import aws_organizations_unit.my_unit ou-1234567
```
