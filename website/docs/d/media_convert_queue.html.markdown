---
subcategory: "Elemental MediaConvert"
layout: "aws"
page_title: "AWS: aws_media_convert_queue"
description: |-
  Retrieve information about a AWS Elemental MediaConvert Queue.
---

# Resource: aws_media_convert_queue

Retrieve information about a AWS Elemental MediaConvert Queue.

## Example Usage

```terraform
data "aws_media_convert_queue" "example" {
  id = "tf-example-queue"
}
```

## Argument Reference

This data source supports the following arguments:

* `id` - (Required) Unique identifier of the queue. The same as `name`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - The Arn of the queue.
* `name` - The same as `id`.
* `status` - The status of the queue.
* `tags` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
