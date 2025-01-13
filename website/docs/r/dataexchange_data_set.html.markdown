---
subcategory: "Data Exchange"
layout: "aws"
page_title: "AWS: aws_dataexchange_data_set"
description: |-
  Provides a DataExchange DataSet
---

# Resource: aws_dataexchange_data_set

Provides a resource to manage AWS Data Exchange DataSets.

## Example Usage

```terraform
resource "aws_dataexchange_data_set" "example" {
  asset_type  = "S3_SNAPSHOT"
  description = "example"
  name        = "example"
}
```

## Argument Reference

* `asset_type` - (Required) The type of asset that is added to a data set. Valid values include `API_GATEWAY_API`, `LAKE_FORMATION_DATA_PERMISSION`, `REDSHIFT_DATA_SHARE`, `S3_DATA_ACCESS`, `S3_SNAPSHOT`.
* `description` - (Required) A description for the data set.
* `name` - (Required) The name of the data set.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The Id of the data set.
* `arn` - The Amazon Resource Name of this data set.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DataExchange DataSets using their `id`. For example:

```terraform
import {
  to = aws_dataexchange_data_set.example
  id = "4fa784c7-ccb4-4dbf-ba4f-02198320daa1"
}
```

Using `terraform import`, import DataExchange DataSets using their `id`. For example:

```console
% terraform import aws_dataexchange_data_set.example 4fa784c7-ccb4-4dbf-ba4f-02198320daa1
```
