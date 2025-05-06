---
subcategory: "QLDB (Quantum Ledger Database)"
layout: "aws"
page_title: "AWS: aws_qldb_ledger"
description: |-
  Get information on a Amazon Quantum Ledger Database (QLDB)
---

# Data Source: aws_qldb_ledger

Use this data source to fetch information about a Quantum Ledger Database.

## Example Usage

```terraform
data "aws_qldb_ledger" "example" {
  name = "an_example_ledger"
}
```

## Argument Reference

This data source supports the following arguments:

* `name` - (Required) Friendly name of the ledger to match.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

See the [QLDB Ledger Resource](/docs/providers/aws/r/qldb_ledger.html) for details on the
returned attributes - they are identical.
