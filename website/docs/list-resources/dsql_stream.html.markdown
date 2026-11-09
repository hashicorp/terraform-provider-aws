---
subcategory: "DSQL"
layout: "aws"
page_title: "AWS: aws_dsql_stream"
description: |-
  Lists DSQL Stream resources.
---

# List Resource: aws_dsql_stream

Lists DSQL Stream resources.

## Example Usage

```terraform
list "aws_dsql_stream" "example" {
  provider = aws

  config {
    cluster_identifier = aws_dsql_cluster.example.identifier
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `cluster_identifier` - (Required) Identifier of the cluster to list streams for.
* `region` - (Optional) Region to query. Defaults to provider region.
