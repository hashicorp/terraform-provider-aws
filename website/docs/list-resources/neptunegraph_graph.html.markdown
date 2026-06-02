---
subcategory: "Neptune Analytics"
layout: "aws"
page_title: "AWS: aws_neptunegraph_graph"
description: |-
  Lists Neptune Analytics Graph resources.
---

# List Resource: aws_neptunegraph_graph

Lists Neptune Analytics Graph resources.

## Example Usage

```terraform
list "aws_neptunegraph_graph" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
