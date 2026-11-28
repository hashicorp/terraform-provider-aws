---
subcategory: "OpenSearch Serverless"
layout: "aws"
page_title: "AWS: aws_opensearchserverless_lifecycle_policy"
description: |-
  Lists OpenSearch Serverless Lifecycle Policy resources.
---

# List Resource: aws_opensearchserverless_lifecycle_policy

Lists OpenSearch Serverless Lifecycle Policy resources.

## Example Usage

```terraform
list "aws_opensearchserverless_lifecycle_policy" "example" {
  provider = aws

  config {
    type = "retention"
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region to query. Defaults to provider region.
* `type` - (Required) Type of lifecycle policy to list. Must be `retention`.
