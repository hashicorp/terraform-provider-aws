---
subcategory: "Service Catalog AppRegistry"
layout: "aws"
page_title: "AWS: aws_servicecatalogappregistry_application_attribute_group_associations"
description: |-
  Terraform data source for managing an AWS Service Catalog AppRegistry Application Attribute Group Associations.
---

# Data Source: aws_servicecatalogappregistry_application_attribute_group_associations

Terraform data source for managing an AWS Service Catalog AppRegistry Application Attribute Group Associations.

## Example Usage

### Basic Usage

```terraform
data "aws_servicecatalogappregistry_application_attribute_group_associations" "example" {
}
```

## Argument Reference

The following arguments are required:

* `example_arg` - (Required) Concise argument description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.

The following arguments are optional:

* `optional_arg` - (Optional) Concise argument description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Application Attribute Group Associations. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.
* `example_attribute` - Concise description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.