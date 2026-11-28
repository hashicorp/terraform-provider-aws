---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_event_source_mapping"
description: |-
  Lists Lambda Event Source Mapping resources.
---

# List Resource: aws_lambda_event_source_mapping

Lists Lambda Event Source Mapping resources.

## Example Usage

```terraform
list "aws_lambda_event_source_mapping" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region to query. Defaults to provider region.
