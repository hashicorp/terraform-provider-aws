---
subcategory: "Lake Formation"
layout: "aws"
page_title: "AWS: aws_lakeformation_lf_tag"
description: |-
    Creates a tag with the specified name and values.
---

# Resource: aws_lakeformation_lf_tag

Creates an LF-Tag with the specified name and values. Each key must have at least one value. The maximum number of values permitted is 1000.

## Example Usage

```terraform
resource "aws_lakeformation_lf_tag" "example" {
  key    = "module"
  values = ["Orders", "Sales", "Customers"]
}
```

## Argument Reference

This resource supports the following arguments:

* `catalog_id` - (Optional) ID of the Data Catalog to create the tag in. If omitted, this defaults to the AWS Account ID.
* `key` - (Required) Key-name for the tag.
* `values` - (Required) List of possible values an attribute can take.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Catalog ID and key-name of the tag

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Lake Formation LF-Tags using the `catalog_id:key`. If you have not set a Catalog ID specify the AWS Account ID that the database is in. For example:

```terraform
import {
  to = aws_lakeformation_lf_tag.example
  id = "123456789012:some_key"
}
```

Using `terraform import`, import Lake Formation LF-Tags using the `catalog_id:key`. If you have not set a Catalog ID specify the AWS Account ID that the database is in. For example:

```console
% terraform import aws_lakeformation_lf_tag.example 123456789012:some_key
```
