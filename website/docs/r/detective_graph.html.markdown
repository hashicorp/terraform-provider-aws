---
subcategory: "Detective"
layout: "aws"
page_title: "AWS: aws_detective_graph"
description: |-
  Provides a resource to manage Amazon Detective on a Graph.
---

# Resource: aws_detective_graph

Provides a resource to manage an [AWS Detective Graph](https://docs.aws.amazon.com/detective/latest/APIReference/Welcome.html).

## Example Usage

```terraform
resource "aws_detective_graph" "test" {}
```

## Argument Reference

The following arguments are supported:

* `tags` -  (Optional) A map of key-value pairs that specifies the tags to associate with the detective graph. You can add up to 50 tags. For each tag, you provide the tag key and the tag value. Each tag key can contain up to 128 characters. Each tag value can contain up to 256 characters.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The unique identifier (ID) of the Detective Graph.
* `created_time` - The date and time, in UTC and extended RFC 3339 format, when the Amazon Detective Graph was created.

## Import

`aws_detective_graph` can be imported using the id, e.g.

```
$ terraform import aws_detective_graph.example abcd1
```
