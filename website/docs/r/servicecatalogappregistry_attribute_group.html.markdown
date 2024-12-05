---
subcategory: "Service Catalog AppRegistry"
layout: "aws"
page_title: "AWS: aws_servicecatalogappregistry_attribute_group"
description: |-
  Terraform resource for managing an AWS Service Catalog AppRegistry Attribute Group.
---
# Resource: aws_servicecatalogappregistry_attribute_group

Terraform resource for managing an AWS Service Catalog AppRegistry Attribute Group.

## Example Usage

### Basic Usage

```terraform
resource "aws_servicecatalogappregistry_attribute_group" "example" {
  name        = "example"
  description = "example description"

  attributes = jsonencode({
    app   = "exampleapp"
    group = "examplegroup"
  })
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the Attribute Group.
* `attributes` - (Required) A JSON string of nested key-value pairs that represents the attributes of the group.

The following arguments are optional:

* `description` - (Optional) Description of the Attribute Group.
* `tags` - (Optional) A map of tags assigned to the Attribute Group. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Attribute Group.
* `id` - ID of the Attribute Group.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Service Catalog AppRegistry Attribute Group using the `id`. For example:

```terraform
import {
  to = aws_servicecatalogappregistry_attribute_group.example
  id = "1234567890abcfedhijk09876s"
}
```

Using `terraform import`, import Service Catalog AppRegistry Attribute Group using the `id`. For example:

```console
% terraform import aws_servicecatalogappregistry_attribute_group.example 1234567890abcfedhijk09876s
```
