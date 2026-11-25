---
subcategory: "DSQL"
layout: "aws"
page_title: "AWS: aws_dsql_cluster_policy"
description: |-
  Lists Aurora DSQL Cluster Policy resources.
---

# List Resource: aws_dsql_cluster_policy

Lists Aurora DSQL Cluster Policy resources.

## Example Usage

```terraform
list "aws_dsql_cluster_policy" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
