---
subcategory: "OpenSearch Serverless"
layout: "aws"
page_title: "AWS: aws_opensearchserverless_vpc_endpoint"
description: |-
  Lists OpenSearch Serverless VPC Endpoint resources.
---

# List Resource: aws_opensearchserverless_vpc_endpoint

Lists OpenSearch Serverless VPC Endpoint resources.

## Example Usage

```terraform
list "aws_opensearchserverless_vpc_endpoint" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region to query. Defaults to provider region.
