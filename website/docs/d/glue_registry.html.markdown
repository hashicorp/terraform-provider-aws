---
subcategory: "Glue"
layout: "aws"
page_title: "AWS: aws_glue_registry"
description: |-
  Terraform data source for managing an AWS Glue Registry.
---

# Data Source: aws_glue_registry

Terraform data source for managing an AWS Glue Registry.

## Example Usage

### Basic Usage

```terraform
data "aws_glue_registry" "example" {
  name = "example"
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the Glue Registry.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of Glue Registry.
* `description` - A description of the registry.
