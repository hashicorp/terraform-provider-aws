---
subcategory: "Chime SDK Voice"
layout: "aws"
page_title: "AWS: aws_chimesdkvoice_tags"
description: |-
  Adds tags to an Amazon Chime SDK Voice resource
---

# Resource: aws_chimesdkvoice_tags

Adds tags to an Amazon Chime SDK Voice resource

## Example Usage

```terraform
resource "aws_chimesdkvoice_tags" "test" {
   resource_arn = aws_chime_voice_connector.test.voice_connector_arn
   tags = {
      "key1" : "value1"
   }
}
```

## Argument Reference

The following arguments are supported:

* `resource_arn` - (Required) The arn of the amazon Chime SDK Voice resource.
* `tags` - (Required) String map of tags to be associated with the resource.

## Import

Configuration Recorder can be imported using the name, e.g.,

```
$ terraform import aws_chimesdkvoice_tags.test example
```
