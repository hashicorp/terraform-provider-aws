---
subcategory: "Managed Streaming for Kafka"
layout: "aws"
page_title: "AWS: aws_msk_topic"
description: |-
  Lists Managed Streaming for Kafka Topic resources.
---

# List Resource: aws_msk_topic

Lists Managed Streaming for Kafka Topic resources.

## Example Usage

```terraform
list "aws_msk_topic" "example" {
  provider = aws

  config {
    cluster_arn = aws_msk_cluster.example.arn
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `cluster_arn` - (Required) ARN of the cluster to list Topics from.
* `region` - (Optional) Region to query. Defaults to provider region.
