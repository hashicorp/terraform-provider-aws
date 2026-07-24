---
subcategory: "Neptune Analytics"
layout: "aws"
page_title: "AWS: aws_neptunegraph_private_graph_endpoint"
description: |-
  Lists Neptune Analytics Private Graph Endpoint resources.
---

# List Resource: aws_neptunegraph_private_graph_endpoint

Lists Neptune Analytics Private Graph Endpoint resources.

## Example Usage

```terraform
list "aws_neptunegraph_private_graph_endpoint" "example" {
  provider = aws

  config {
    graph_identifier = "g-1234567890"
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `graph_identifier` - (Required) The unique identifier of the Neptune Analytics graph.
* `region` - (Optional) Region to query. Defaults to provider region.
