---
subcategory: "Resilience Hub V2"
layout: "aws"
page_title: "AWS: aws_resiliencehubv2_assertion"
description: |-
  Lists Resilience Hub V2 Assertion resources for a service.
---

# List Resource: aws_resiliencehubv2_assertion

Lists Resilience Hub V2 Assertion resources for a given service.

## Example Usage

### Basic

```terraform
list "aws_resiliencehubv2_assertion" "example" {
  provider = aws

  config {
    service_arn = aws_resiliencehubv2_service.example.arn
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region to list resources in. Defaults to the provider's configured region.
* `service_arn` - (Required) ARN of the service whose assertions are listed.
