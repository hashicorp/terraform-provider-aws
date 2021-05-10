---
subcategory: "Service Catalog"
layout: "aws"
page_title: "AWS: aws_servicecatalog_tag_option"
description: |-
  Manages a Service Catalog Tag Option
---

# Resource: aws_servicecatalog_tag_option

Manages a Service Catalog Tag Option.

## Example Usage

### Basic Usage

```terraform
resource "aws_servicecatalog_tag_option" "example" {
  key   = "nyckel"
  value = "värde"
}
```

## Argument Reference

The following arguments are required:

* `key` - (Required) Tag option key.
* `value` - (Required) Tag option value.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `active` - Active state.
* `id` - Identifier.
* `owner` - AWS account ID of the owner account that created the tag option.

## Import

`aws_servicecatalog_tag_option` can be imported using the tag option ID, e.g.

```
$ terraform import aws_servicecatalog_tag_option.example tag-pjtvagohlyo3m
```
