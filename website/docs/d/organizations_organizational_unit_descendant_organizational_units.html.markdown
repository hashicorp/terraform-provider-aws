---
subcategory: "Organizations"
layout: "aws"
page_title: "AWS: aws_organizations_organizational_unit_descendant_organizational_units"
description: |-
  Get all direct child organizational units under a parent organizational unit. This provides all children.
---

# Data Source: aws_organizations_organizational_unit_descendant_organizational_units

Get all direct child organizational units under a parent organizational unit. This provides all children.

## Example Usage

```terraform
data "aws_organizations_organization" "org" {}

data "aws_organizations_organizational_unit_descendant_organizational_units" "ous" {
  parent_id = data.aws_organizations_organization.org.roots[0].id
}
```

## Argument Reference

* `parent_id` - (Required) Parent ID of the organizational unit.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `children` - List of child organizational units, which have the following attributes:
    * `arn` - ARN of the organizational unit
    * `name` - Name of the organizational unit
    * `id` - ID of the organizational unit
* `id` - Parent identifier of the organizational units.
