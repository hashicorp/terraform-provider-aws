---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_security_token_service_preferences"
description: |-
  Provides an IAM Security Token Service Preferences resource.
---

# Resource: aws_iam_security_token_service_preferences

Provides an IAM Security Token Service Preferences resource.

## Example Usage

```terraform
resource "aws_iam_security_token_service_preferences" "example" {
  global_endpoint_token_version = "v2Token"
}
```

## Argument Reference

This resource supports the following arguments:

* `global_endpoint_token_version` - (Required) The version of the STS global endpoint token. Valid values: `v1Token`, `v2Token`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The AWS Account ID.
