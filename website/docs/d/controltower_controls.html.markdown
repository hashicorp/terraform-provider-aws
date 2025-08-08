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

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `target_identifier` - (Required) The ARN of the organizational unit.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `enabled_controls` - List of all the ARNs for the controls applied to the `target_identifier`.
