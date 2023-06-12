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

### Basic Usage

```terraform
resource "aws_rekognition_collection" "example" {
  collection_id = example
  tags = {
    key1 = value1
  }
}
```

## Argument Reference

* `collection_id` - (Required) ID for the collection that you are creating.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the Collection. 
* `id` - Collection ID
* `face_count` - The number of faces that are indexed into the collection. 
* `face_model_version` - The version of the face model that's used by the collection for face detection.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block). 

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `delete` - (Default `90m`)

## Import

Rekognition Collection can be imported using the `collection_id`, e.g.,

```
$ terraform import aws_rekognition_collection.example rft-8012925589
```
