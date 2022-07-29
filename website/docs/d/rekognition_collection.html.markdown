---
subcategory: "Rekognition"
layout: "aws"
page_title: "AWS: aws_rekognition_collection"
description: |-
  Get information of Rekognition Collection
---

# Data Source: aws_rekognition_collection

Use this data source to get the ARN of a collection in Rekognition.

## Example Usage

```terraform
data "aws_rekognition_collection" "example" {
  collection_id = "an_example_collection"
}
```

## Argument Reference

* `collection_id` - (Required) Collection id.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `collection_arn` - The ARN assigned by AWS for this collection.
* `face_count` - Count of faces in this collection.
* `face_model_version` - Version of face model.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
