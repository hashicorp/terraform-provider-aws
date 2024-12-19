---
subcategory: "Amazon Q Business"
layout: "aws"
page_title: "AWS: aws_qbusiness_index"
description: |-
  Provides a Q Business Index resource.
---

# Resource: aws_qbusiness_index

Provides a Q Business Index resource.

## Example Usage

```terraform
resource "aws_qbusiness_index" "example" {
  application_id = "aws_qbusiness_app.test.application_id"
  display_name   = "Index display name"
}
```

## Argument Reference

This resource supports the following arguments:

* `application_id` - (Required) Id of the Q Business application.
* `display_name` - (Required) Index display name.
* `capacity_configuration` - (Required) The number of additional storage units for the Amazon Q index.
* `description` - (Optional) The description for the Amazon Q index.
* `document_attribute_configuration` - (Optional) Configuration information for document metadata or fields.

`document_attribute_configuration` supports the following:

* `name` - (Required) The name of the document attribute.
* `search` - (Required) Information about whether the document attribute can be used by an end user to search for information on their web experience. Valid values are `ENABLED` and `DISABLED`
* `type` - (Required) The type of document attribute. Valid values are `STRING`, `STRING_LIST`, `NUMBER` and `DATE`

`capacity_configuration` supports the following:

* `units` - (Required) The number of storage units configured for an Amazon Q index.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `index_id` - Id of the Q Business index.
* `arn` - Amazon Resource Name (ARN) of the Q Business index.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
