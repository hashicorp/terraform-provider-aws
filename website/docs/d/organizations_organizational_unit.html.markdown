---
subcategory: "Organizations"
layout: "aws"
page_title: "AWS: aws_organizations_organizational_unit"
description: |-
  Terraform data source for getting an AWS Organizations Organizational Unit.
---

# Data Source: aws_organizations_organizational_unit

Terraform data source for getting an AWS Organizations Organizational Unit.

## Example Usage

### Basic Usage

```terraform
data "aws_organizations_organization" "org" {}

data "aws_organizations_organizational_unit" "ou" {
  parent_id = data.aws_organizations_organization.org.roots[0].id
  name      = "dev"
}
```

## Argument Reference

The following arguments are required:

* `parent_id` - (Required) Parent ID of the organizational unit.

* `name` - (Required) Name of the organizational unit

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the organizational unit

* `id` - ID of the organizational unit
