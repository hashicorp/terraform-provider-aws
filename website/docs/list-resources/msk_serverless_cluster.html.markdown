---
subcategory: "Managed Streaming for Kafka"
layout: "aws"
page_title: "AWS: aws_msk_serverless_cluster"
description: |-
  Lists Managed Streaming for Kafka Serverless Cluster resources.
---

# List Resource: aws_msk_serverless_cluster

Lists Managed Streaming for Kafka Serverless Cluster resources.

## Example Usage

```terraform
list "aws_msk_serverless_cluster" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region to query. Defaults to provider region.
