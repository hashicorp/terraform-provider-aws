---
subcategory: "DMS (Database Migration)"
layout: "aws"
page_title: "AWS: aws_dms_data_provider"
description: |-
  Lists DMS (Database Migration) Data Provider resources.
---

# List Resource: aws_dms_data_provider

Lists DMS (Database Migration) Data Provider resources.

## Example Usage

### Basic Usage

```terraform
list "aws_dms_data_provider" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
