---
subcategory: "Detective"
layout: "aws"
page_title: "AWS: aws_detective_graph"
description: |-
  Provides a resource to manage Amazon Detective on a Graph.
---

# Resource: aws_detective_graph

Provides a resource to manage an [AWS Detective Graph](https://docs.aws.amazon.com/detective/latest/APIReference/API_CreateGraph.html).

## Example Usage

```terraform
resource "aws_detective_graph" "test" {}
```

## Argument Reference

The following arguments are supported:

* `tags` -  (Optional) A map of key-value pairs that specifies the tags to associate with the detective graph. You can add up to 50 tags. For each tag, you provide the tag key and the tag value. Each tag key can contain up to 128 characters. Each tag value can contain up to 256 characters.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Unique identifier (ID) of the Detective Graph.
* `created_time` - Date and time, in UTC and extended RFC 3339 format, when the Amazon Detective Graph was created.

## Import

`aws_detective_graph` can be imported using the id, e.g.

```
$ terraform import aws_detective_graph.example arn:aws:detective:us-east-1:123456789101:graph:231684d34gh74g4bae1dbc7bd807d02d
```
