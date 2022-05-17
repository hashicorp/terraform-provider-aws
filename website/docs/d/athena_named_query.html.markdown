---
subcategory: "Athena"
layout: "aws"
page_title: "AWS: aws_athena_named_query"
description: |-
    Provides an Athena Named Query resource.
---

# Resource: aws_athena_named_query

Provides an Athena Named Query resource.

## Example Usage

```terraform
data "aws_athena_named_query" "example" {
  name = "athenaQueryName"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The plain language name for the query. Maximum length of 128.
* `workgroup` - (Optional) The workgroup to which the query belongs. Defaults to `primary`

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The unique ID of the query.

## Import

Athena Named Query can be imported using the query ID, e.g.,

```
$ terraform import aws_athena_named_query.example 0123456789
```
