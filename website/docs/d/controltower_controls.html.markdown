---
subcategory: "Control Tower"
layout: "aws"
page_title: "AWS: aws_controltower_controls"
description: |-
  List of Control Tower controls applied to an OU.
---

# Data Source: aws_controltower_controls

List of Control Tower controls applied to an OU.

## Example Usage

```terraform
data "aws_organizations_organization" "this" {}

data "aws_organizations_organizational_units" "this" {
  parent_id = data.aws_organizations_organization.this.roots[0].id
}

data "aws_controltower_controls" "this" {

  target_identifier = [
    for x in data.aws_organizations_organizational_units.this.children :
    x.arn if x.name == "Security"
  ][0]

}

```

## Argument Reference

The following arguments are required:

* `target_identifier` - (Required) The ARN of the organizational unit.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `enabled_controls` - List of all the ARNs for the controls applied to the `target_identifier`.
