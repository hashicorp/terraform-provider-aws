---
subcategory: "Service Catalog AppRegistry"
layout: "aws"
page_title: "AWS: aws_servicecatalogappregistry_attribute_group"
description: |-
  Terraform data source for managing an AWS Service Catalog AppRegistry Attribute Group.
---

# Data Source: aws_servicecatalogappregistry_attribute_group

Terraform data source for managing an AWS Service Catalog AppRegistry Attribute Group.

## Example Usage

### Basic Usage

```terraform
data "aws_servicecatalogappregistry_attribute_group" "example" {
  name = "example_attribute_group"
}
```

## Argument Reference

The following arguments are required:


The following arguments are optional:

* `id`   - (Optional) ID of the Attribute Group to find.
* `name` - (Optional) Name of the Attribute Group to find.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `attributes` - A JSON string of nested key-value pairs that represents the attributes of the group.
* `description` - Description of the Attribute Group.
* `tags` - A map of tags assigned to the Attribute Group. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
