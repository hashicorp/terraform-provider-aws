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

## Attributes Reference

No additional attributes are exported.

## Timeouts

[Configuration options](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

EventBridge Schema Registry Policy can be imported using the `registry_name`, e.g.,

```
$ terraform import aws_schemas_registry_policy.example example
```
