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

* `asset_type` - (Required) The type of asset that is added to a data set. Valid values are: `S3_SNAPSHOT`, `REDSHIFT_DATA_SHARE`, and `API_GATEWAY_API`.
* `description` - (Required) A description for the data set.
* `name` - (Required) The name of the data set.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The Id of the data set.
* `arn` - The Amazon Resource Name of this data set.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

DataExchange DataSets can be imported by their arn:

```
$ terraform import aws_dataexchange_data_set.example arn:aws:dataexchange:us-west-2:123456789012:data-sets/4fa784c7-ccb4-4dbf-ba4f-02198320daa1
```
