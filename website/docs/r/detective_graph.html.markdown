---
subcategory: "Detective"
layout: "aws"
page_title: "AWS: aws_detective_graph"
description: |-
  Provides a resource to manage an Amazon Detective graph.
---

# Resource: aws_detective_graph

Provides a resource to manage an [AWS Detective Graph](https://docs.aws.amazon.com/detective/latest/APIReference/API_CreateGraph.html). As an AWS account may own only one Detective graph per region, provisioning multiple Detective graphs requires a separate provider configuration for each graph.

## Example Usage

```terraform
resource "aws_detective_graph" "example" {
  tags = {
    Name = "example-detective-graph"
  }
}
```

## Argument Reference

The following arguments are optional:

* `tags` -  (Optional) A map of tags to assign to the instance. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ARN of the Detective Graph.
* `graph_arn` - ARN of the Detective Graph.
* `created_time` - Date and time, in UTC and extended RFC 3339 format, when the Amazon Detective Graph was created.

## Import

`aws_detective_graph` can be imported using the ARN, e.g.

```
$ terraform import aws_detective_graph.example arn:aws:detective:us-east-1:123456789101:graph:231684d34gh74g4bae1dbc7bd807d02d
```
