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
resource "aws_detective_graph" "example" {}
```

## Argument Reference

The following arguments are optional:

* `tags` -  (Optional) A map of tags to assign to the instance. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ARN of the Detective Graph.
* `created_time` - Date and time, in UTC and extended RFC 3339 format, when the Amazon Detective Graph was created.

## Import

`aws_detective_graph` can be imported using the arn, e.g.

```
$ terraform import aws_detective_graph.example arn:aws:detective:us-east-1:123456789101:graph:231684d34gh74g4bae1dbc7bd807d02d
```
