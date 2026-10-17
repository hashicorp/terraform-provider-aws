---
subcategory: "Oracle Database@AWS"
layout: "aws"
page_title: "AWS: aws_odb_exascale_db_storage_vault"
description: |-
  Lists Oracle Database@AWS Exascale DB Storage Vault resources.
---

# List Resource: aws_odb_exascale_db_storage_vault

Lists Oracle Database@AWS Exascale DB Storage Vault resources.

## Example Usage

```terraform
list "aws_odb_exascale_db_storage_vault" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
