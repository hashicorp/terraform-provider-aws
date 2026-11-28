---
subcategory: "OpenSearch Serverless"
layout: "aws"
page_title: "AWS: aws_opensearchserverless_access_policy"
description: |-
  Lists OpenSearch Serverless Access Policy resources.
---

# List Resource: aws_opensearchserverless_access_policy

Lists OpenSearch Serverless Access Policy resources.

## Example Usage

```terraform
list "aws_opensearchserverless_access_policy" "example" {
  provider = aws

  config {
    type = "data"
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region to query. Defaults to provider region.
* `type` - (Required) Type of access policy. Currently the only valid value is `data`.
