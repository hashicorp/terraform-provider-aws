---
layout: "aws"
page_title: "AWS: aws_qldb_ledger"
description: |-
  Get information on a Amazon Quantum Ledger Database (QLDB)
---

# Data Source: aws_qldb_ledger

Use this data source to get the ARN, deletion protection and permissions mode of a ledger in AWS Quantum Ledger Database 
(QLDB). By using this data source, you can reference QLDB Ledgers
without having to hard code the ARNs as input.

## Example Usage

```hcl
data "aws_qldb_ledger" "example" {
  name = "an_example_ledger"
}
```

## Argument Reference

* `name` - (Required) The friendly name of the ledger to match.

## Attributes Reference

* `arn` - Set to the ARN of the found topic, suitable for referencing in other resources that support QLDB ledgers.
* `permissions_mode` - Permissions mode attached to the QLDB Ledger instance. Currently outputs `ALLOW_ALL`
* `deletion_protection` - Deletion protection on the QLDB Ledger instance. Set to `true` by default. 
