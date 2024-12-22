---
subcategory: "Amazon Q Business"
layout: "aws"
page_title: "AWS: aws_qbusiness_dataaccessor"
description: |-
  Provides a Q Business Dataaccessor resource.
---

# Resource: aws_qbusiness_dataaccessor

Provides a Q Business Dataccessor resource.

## Example Usage

```terraform
resource "aws_qbusiness_dataaccessor" "example" {
  application_id = aws_qbusiness_app.example.id
  display_name   = "test-data-accessor"
  principal      = "arn:aws:iam::359246571101:role/zoom-ai-companion"

  action_configuration {
    action = "qbusiness:SearchRelevantContent"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `display_name` - (Required) A name for the Amazon Q Dataaccessor.
* `principal` - (Required) ARN of the IAM role for the ISV that will be accessing the data.
* `identity_center_instance_arn` - (Required) ARN of the IAM Identity Center instance you are either creating for — or connecting to — your Amazon Q Business Dataaccessor
* `action_configuration` - (Required) list of action configurations specifying the allowed actions and any associated filters. Maximum number of 10 items.

`action_configuration` supports the following:

* `action` - (Required) Amazon Q Business action that is allowed.
* `filter_configuration` - (Optional) Amazon Q Business action that is allowed.

`filter_configuration` supports the following:

* `document_attribute_filter` - (Required) Enables filtering of responses based on document attributes or metadata fields.

`document_attribute_filter` supports the following:

* `and_all_filters` - (Optional) Array of `document_attribute_filter`. Performs a logical AND operation on all supplied filters.
* `contains_all` - (Optional) Type of `document_attribute`. Returns true when a document contains all the specified document attributes or metadata fields. Supported for the following document attribute value types: `string_list_value`.
* `contains_any` - (Optional) Type of `document_attribute`. Returns true when a document contains any of the specified document attributes or metadata fields. Supported for the following document attribute value types: `string_list_value`.
* `equals_to` - (Optional) Type of `document_attribute`. Performs an equals operation on two document attributes or metadata fields. Supported for the following document attribute value types: `date_value`, `long_value`, `string_list_value` and `string_value`.
* `greater_than` - (Optional) Type of `document_attribute`. Performs a greater than operation on two document attributes or metadata fields. Supported for the following document attribute value types: `date_value` and `long_value`.
* `greater_than_or_equals` - (Optional) Type of `document_attribute`. Performs a greater or equals than operation on two document attributes or metadata fields. Supported for the following document attribute value types: `date_value` and `long_value`.
* `less_than` - (Optional) Type of `document_attribute`. Performs a less than operation on two document attributes or metadata fields. Supported for the following document attribute value types: `date_value` and `long_value`.
* `less_than_or_equals` - (Optional) Type of `document_attribute`. Performs a less than or equals operation on two document attributes or metadata fields.Supported for the following document attribute value type: `date_value` and `long_value`.
* `not_filter` - (Optional) Type of `document_attribute_filter`. Performs a logical NOT operation on all supplied filters.
* `or_all_filters` - (Optional) Array of `document_attribute_filter`. Performs a logical OR operation on all supplied filters.

`document_attribute` supports the following:

* `name` - (Required) Identifier for the attribute.
* `value` - (Required) Value for the attribute.

`value` supports the following:

* `date_value` - (Optional) A date expressed as an ISO 8601 string. Must be in UTC timezone
* `long_value` - (Optional) A long integer value.
* `string_list_value` - (Optional) A list of strings.
* `strings_value` - (Optional) A string.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `data_accessor_id` - Dataaccessor identifier.
* `arn` - ARN of the Amazon Q dataaccessor.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
