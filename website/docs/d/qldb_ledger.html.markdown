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
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the QLDB Ledger.
* `deletion_protection` - Deletion protection setting of the QLDB Ledger.
* `kms_key` - KMS key used for encryption of data at rest in the ledger.
* `permissions_mode` - Permissions mode of the QLDB Ledger.
* `tags` - Map of tags assigned to the resource.
