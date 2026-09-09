---
subcategory: "OpenSearch Serverless"
layout: "aws"
page_title: "AWS: aws_opensearchserverless_security_config"
description: |-
  Lists OpenSearch Serverless Security Config resources.
---

# List Resource: aws_opensearchserverless_security_config

Lists OpenSearch Serverless Security Config resources.

## Example Usage

```terraform
list "aws_opensearchserverless_security_config" "example" {
  provider = aws

  config {
    type = "saml"
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
* `type` - (Required) Type of security configuration to list. Valid values are `saml`, `iamidentitycenter`, and `iamfederation`.
