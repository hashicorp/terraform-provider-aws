---
layout: "aws"
page_title: "AWS: aws_organizations_organizational_unit"
sidebar_current: "docs-aws-resource-organizations-organizational-unit"
description: |-
  Provides a resource to create an organizational unit.
---

# Resource: aws_organizations_organizational_unit

Provides a resource to create an organizational unit.

## Example Usage

```hcl
resource "aws_organizations_organizational_unit" "example" {
  name      = "example"
  parent_id = "${aws_organizations_organization.example.roots.0.id}"
}
```

## Argument Reference

The following arguments are supported:

* `name` - The name for the organizational unit
* `parent_id` - ID of the parent organizational unit, which may be the root

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the organizational unit
* `id` - Identifier of the organization unit

## Import

AWS Organizations Organizational Units can be imported by using the `id`, e.g.

```
$ terraform import aws_organizations_organizational_unit.example ou-1234567
```
