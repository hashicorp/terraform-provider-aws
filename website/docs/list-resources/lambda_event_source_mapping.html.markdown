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

* `region` - (Optional) Region to query. Defaults to provider region.
