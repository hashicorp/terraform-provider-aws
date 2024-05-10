---
subcategory: "Rekognition"
layout: "aws"
page_title: "AWS: aws_rekognition_collection"
description: |-
  Terraform resource for managing an AWS Rekognition Collection.
---

# Resource: aws_rekognition_collection

Terraform resource for managing an AWS Rekognition Collection.

## Example Usage

```terraform
resource "aws_rekognition_collection" "example" {
  collection_id = "my-collection"

  tags = {
    example = 1
  }
}
```

## Argument Reference

The following arguments are required:

* `collection_id` - (Required) The name of the collection

The following arguments are optional:

* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Collection.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `face_model_version` - The Face Model Version that the collection was initialized with

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `2m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Rekognition Collection using the `example_id_arg`. For example:

```terraform
import {
  to = aws_rekognition_collection.example
  id = "collection-id-12345678"
}
```

Using `terraform import`, import Rekognition Collection using the `example_id_arg`. For example:

```console
% terraform import aws_rekognition_collection.example collection-id-12345678
```
