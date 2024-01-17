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

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `retriever_id` - Id of the Q Business retriever.
* `arn` - ARN of the Q Business retriever.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
