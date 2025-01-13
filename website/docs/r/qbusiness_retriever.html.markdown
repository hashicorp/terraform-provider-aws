---
subcategory: "Amazon Q Business"
layout: "aws"
page_title: "AWS: aws_qbusiness_retriever"
description: |-
  Provides a Q Business Retriever resource.
---

# Resource: aws_qbusiness_retriever

Provides a Q Business Retriever resource.

## Example Usage

```terraform
resource "aws_qbusiness_retriever" "example" {
  application_id = "aws_qbusiness_app.test.application_id"
  display_name   = "Retriever display name"

  native_index_configuration {
    index_id = aws_qbusiness_index.test.index_id

    string_boost_override {
      boost_key      = "string"
      boosting_level = "HIGH"
      attribute_value_boosting = {
        "key1" = "VERY_HIGH"
        "key2" = "VERY_HIGH"
      }
    }

    number_boost_override {
      boost_key      = "number1"
      boosting_level = "HIGH"
      boosting_type  = "PRIORITIZE_LARGER_VALUES"
    }

    number_boost_override {
      boost_key      = "number2"
      boosting_level = "LOW"
      boosting_type  = "PRIORITIZE_SMALLER_VALUES"
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `application_id` - (Required) Identifier of Amazon Q application.
* `display_name` - (Required) Name of retriever.
* `iam_service_role_arn` - (Optional) ARN of an IAM role used by Amazon Q to access the basic authentication credentials stored in a Secrets Manager secret.
* `kendra_index_configuration` - (Optional) Information on how the Amazon Kendra index used as a retriever for your Amazon Q application is configured. Conflicts with `native_index_configuration`
* `native_index_configuration` - (Optional) Information on how a Amazon Q index used as a retriever for your Amazon Q application is configured. Conflicts with `kendra_index_configuration`

`kendra_index_configuration` supports the following:

* `index_id` - (Required) Identifier of the Amazon Kendra index.

`native_index_configuration` supports the following:

* `index_id` - (Required) Identifier for the Amazon Q index.
* `date_boost_override` - (Optional) Provides information on boosting `DATE` type document attributes.
* `number_boost_override` - (Optional) Provides information on boosting `NUMBER` type document attributes.
* `string_boost_override` - (Optional) Provides information on boosting `STRING` type document attributes.
* `string_list_boost_override` - (Optional) Provides information on boosting `STRING_LIST` type document attributes.

`date_boost_override` supports the following:

* `boost_key` - (Required) Document attribute name
* `boosting_level` - (Required) Specifies how much a document attribute is boosted. Values are `NONE`, `LOW`, `MEDIUM`, `HIGH`, `VERY_HIGH`
* `boosting_duration` - (Required) Specifies the duration, in seconds, of a boost applies to a DATE type document attribute

`number_boost_override` supports the following:

* `boost_key` - (Required) Document attribute name
* `boosting_level` - (Required) Specifies how much a document attribute is boosted. Values are `NONE`, `LOW`, `MEDIUM`, `HIGH`, `VERY_HIGH`
* `boosting_type` - (Required) Specifies how much a document attribute is boosted. Values are `PRIORITIZE_LARGER_VALUES`, `PRIORITIZE_SMALLER_VALUES`

`string_boost_override` supports the following:

* `boost_key` - (Required) Document attribute name
* `boosting_level` - (Required) Specifies how much a document attribute is boosted. Values are `NONE`, `LOW`, `MEDIUM`, `HIGH`, `VERY_HIGH`
* `attribute_value_boosting` - (Required) Specifies specific values of a STRING type document attribute being boosted. String to string map

`string_list_boost_override` supports the following:

* `boost_key` - (Required) Document attribute name
* `boosting_level` - (Required) Specifies how much a document attribute is boosted. Values are `NONE`, `LOW`, `MEDIUM`, `HIGH`, `VERY_HIGH`

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `retriever_id` - Id of the Q Business retriever.
* `arn` - ARN of the Q Business retriever.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
