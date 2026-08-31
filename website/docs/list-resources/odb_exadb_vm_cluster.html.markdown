---
subcategory: "Oracle Database@AWS"
layout: "aws"
page_title: "AWS: aws_odb_exadb_vm_cluster"
description: |-
  Lists Oracle Database@AWS ExaDB VM Cluster resources.
---

# List Resource: aws_odb_exadb_vm_cluster

Lists Oracle Database@AWS ExaDB VM Cluster resources.

## Example Usage

```terraform
list "aws_odb_exadb_vm_cluster" "example" {
  provider = aws

  config {
    exascale_db_storage_vault_id = "xsvault_0123456789"
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `exascale_db_storage_vault_id` - (Optional) Limits results to ExaDB VM Clusters associated with the specified Exascale DB Storage Vault ID. Length must be between `6` and `2048` characters.
* `region` - (Optional) Region to query. Defaults to provider region.
