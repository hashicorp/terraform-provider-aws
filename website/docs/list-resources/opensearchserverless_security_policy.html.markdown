---
subcategory: "OpenSearch Serverless"
layout: "aws"
page_title: "AWS: aws_opensearchserverless_security_policy"
description: |-
  Lists OpenSearch Serverless Security Policy resources.
---

# List Resource: aws_opensearchserverless_security_policy

Lists OpenSearch Serverless Security Policy resources.

## Example Usage

### List encryption policies

```terraform
list "aws_opensearchserverless_security_policy" "example" {
  provider = aws

  config {
    type = "encryption"
  }
}
```

### List network policies

```terraform
list "aws_opensearchserverless_security_policy" "example" {
  provider = aws

  config {
    type = "network"
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region to query. Defaults to provider region.
* `type` - (Required) Type of security policy. Valid values are `encryption` or `network`.
