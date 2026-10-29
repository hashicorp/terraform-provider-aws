---
subcategory: "DMS (Database Migration)"
layout: "aws"
page_title: "AWS: aws_dms_migration_project"
description: |-
  Lists DMS (Database Migration) Migration Project resources.
---

# List Resource: aws_dms_migration_project

Lists DMS (Database Migration) Migration Project resources.

## Example Usage

### Basic Usage

```terraform
list "aws_dms_migration_project" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
