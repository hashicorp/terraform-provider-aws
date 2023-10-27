---
subcategory: "EventBridge Schemas"
layout: "aws"
page_title: "AWS: aws_schemas_registry_policy"
description: |-
  Terraform resource for managing an AWS EventBridge Schemas Registry Policy.
---

# Resource: aws_schemas_registry_policy

Terraform resource for managing an AWS EventBridge Schemas Registry Policy.

## Example Usage

### Basic Usage

```terraform
data "aws_iam_policy_document" "example" {
  statement {
    sid    = "example"
    effect = "Allow"
    principals {
      type = "AWS"
      identifiers = [
        "109876543210"
      ]
    }
    actions = ["schemas:*"]
    resources = [
      "arn:aws:schemas:us-east-1:012345678901:registry/example",
      "arn:aws:schemas:us-east-1:012345678901:schema/example*"
    ]
  }
}

resource "aws_schemas_registry_policy" "example" {
  registry_name = "example"
  policy        = data.aws_iam_policy_document.example.json
}
```

## Argument Reference

The following arguments are required:

* `registry_name` - (Required) Name of EventBridge Schema Registry
* `policy` - (Required) Resource Policy for EventBridge Schema Registry

## Attribute Reference

This resource exports no additional attributes.

## Timeouts

[Configuration options](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import EventBridge Schema Registry Policy using the `registry_name`. For example:

```terraform
import {
  to = aws_schemas_registry_policy.example
  id = "example"
}
```

Using `terraform import`, import EventBridge Schema Registry Policy using the `registry_name`. For example:

```console
% terraform import aws_schemas_registry_policy.example example
```
