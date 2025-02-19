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
  application_id = "aws_qbusiness_application.test.application_id"
  display_name   = "Index display name"
  capacity_configuration {
    units = 1
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `application_id` - (Required) Id of the Q Business application.
* `display_name` - (Required) Index display name.
* `capacity_configuration` - (Required) Number of additional storage units for the Amazon Q index.

The following arguments are optional:

* `description` - (Optional) Description for the Amazon Q index.
* `document_attribute_configuration` - (Optional) Configuration information for document metadata or fields.

`document_attribute_configuration` supports the following:

* `name` - (Required) Name of the document attribute.
* `search` - (Required) Information about whether the document attribute can be used by an end user to search for information on their web experience. Valid values are `ENABLED` and `DISABLED`
* `type` - (Required) Type of document attribute. Valid values are `STRING`, `STRING_LIST`, `NUMBER` and `DATE`

`capacity_configuration` supports the following:

* `units` - (Required) Number of storage units configured for an Amazon Q index.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `index_id` - Id of the Q Business index.
* `arn` - Amazon Resource Name (ARN) of the Q Business index.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a Q Business Index using the `id`. For example:

```terraform
import {
  to = aws_qbusiness_index.example
  id = "a5006afd-3f45-42bc-abf7-c5374014b72d,7242e08d-94be-4c4a-aa22-6347944738cb"
}
```

Using `terraform import`, import a Q Business Index using the `id`. For example:

```console
% terraform import aws_qbusiness_index.example a5006afd-3f45-42bc-abf7-c5374014b72d,7242e08d-94be-4c4a-aa22-6347944738cb
```
