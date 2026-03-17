---
subcategory: "Managed Streaming for Kafka"
layout: "aws"
page_title: "AWS: aws_msk_cluster"
description: |-
  Lists Managed Streaming for Kafka Cluster resources.
---

# List Resource: aws_msk_cluster

Lists Managed Streaming for Kafka Cluster resources.

## Example Usage

```terraform
list "aws_msk_cluster" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
