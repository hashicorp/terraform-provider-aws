---
subcategory: "Rekognition"
layout: "aws"
page_title: "AWS: aws_rekognition_collection"
description: |-
  Provides a Collection.
---

# Resource: aws_rekognition_collection

Provides a Collection.

## Example Usage

```terraform
resource "aws_rekognition_collection" "test" {
  collection_id = "test-collection"
}
```

## Argument Reference

The following arguments are supported:

* `collection_id` - (Required) The collection id.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `collection_arn` - The ARN assigned by AWS for this collection.
* `face_count` - Count of faces in this collection.
* `face_model_version` - Version of face model.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

Collection can be imported using the `collection_id`, e.g.,

```
$ terraform import aws_rekognition_collection.foo collection_id
```
