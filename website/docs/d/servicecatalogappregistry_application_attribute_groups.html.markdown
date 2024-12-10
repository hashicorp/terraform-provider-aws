---
subcategory: "Service Catalog AppRegistry"
layout: "aws"
page_title: "AWS: aws_servicecatalogappregistry_application_attribute_groups"
description: |-
  Terraform data source for managing an AWS Service Catalog AppRegistry Application Attribute Groups.
---

# Data Source: aws_servicecatalogappregistry_application_attribute_groups

Terraform data source for managing an AWS Service Catalog AppRegistry Application Attribute Group Associations.

## Example Usage

### Basic Usage

```terraform
data "aws_servicecatalogappregistry_application_attribute_groups" "example" {
  id = "12456778723424sdffsdfsdq34,12234t3564dsfsdf34asff4ww3"
}
```

## Argument Reference

The following arguments are required:

The following arguments are optional:

~> Exactly one of `id`or `name` must be set.

* `id`   - (Optional) ID of the Attribute Group to find.
* `name` - (Optional) Name of the Attribute Group to find.

The following arguments are optional:

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `attribute_group_ids` - Set of Ids ot the Attribute Groups this Application is associated with.
