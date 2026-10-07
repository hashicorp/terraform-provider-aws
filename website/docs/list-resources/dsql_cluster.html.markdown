---
subcategory: "DSQL"
layout: "aws"
page_title: "AWS: aws_dsql_cluster"
description: |-
  Lists Aurora DSQL Cluster resources.
---

# List Resource: aws_dsql_cluster

Lists Aurora DSQL Cluster resources.

## Example Usage

```terraform
list "aws_dsql_cluster" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
