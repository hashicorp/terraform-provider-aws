---
subcategory: "Service Catalog AppRegistry"
layout: "aws"
page_title: "AWS: aws_servicecatalogappregistry_attribute_group_associations"
description: |-
  Terraform data source for managing AWS Service Catalog AppRegistry Attribute Group Associations.
---

# Data Source: aws_servicecatalogappregistry_attribute_group_associations

Terraform data source for managing AWS Service Catalog AppRegistry Attribute Group Associations.

## Example Usage

### Basic Usage

```terraform
data "aws_servicecatalogappregistry_attribute_group_associations" "example" {
  id = "12456778723424sdffsdfsdq34,12234t3564dsfsdf34asff4ww3"
}
```

## Argument Reference

The following arguments are optional:

~> Exactly one of `id`or `name` must be set.

* `id`   - (Optional) ID of the application to which attribute groups are associated.
* `name` - (Optional) Name of the application to which attribute groups are associated.

The following arguments are optional:

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `attribute_group_ids` - Set of attribute group IDs this application is associated with.
