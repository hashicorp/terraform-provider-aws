---
subcategory: "Glue"
layout: "aws"
page_title: "AWS: aws_glue_catalog_level_table_optimizer"
description: |-
  Manages catalog-level Iceberg table optimization settings in AWS Glue.
---

# Resource: aws_glue_catalog_level_table_optimizer

Manages catalog-level Iceberg table optimization settings in AWS Glue, including compaction, retention, and orphan file deletion.

## Example Usage
```terraform
resource "aws_iam_role" "glue_optimizer" {
  name = "glue-catalog-optimizer"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action    = "sts:AssumeRole"
      Effect    = "Allow"
      Principal = { Service = "glue.amazonaws.com" }
    }]
  })
}

resource "aws_glue_catalog_level_table_optimizer" "example" {
  catalog_id = data.aws_caller_identity.current.account_id

  iceberg_optimization {
    role_arn = aws_iam_role.glue_optimizer.arn

    compaction = {
      enabled               = "true"
      strategy              = "binpack"
      minInputFiles         = "100"
    }

    retention = {
      enabled                          = "true"
      snapshotRetentionPeriodInDays    = "7"
      numberOfSnapshotsToRetain        = "1"
    }

    orphan_file_deletion = {
      enabled                            = "true"
      orphanFileRetentionPeriodInDays    = "3"
    }
  }
}

data "aws_caller_identity" "current" {}
```

## Argument Reference

The following arguments are required:

* `catalog_id` - (Required, Forces new resource) The ID of the Glue catalog.

### iceberg_optimization

* `role_arn` - (Required) The ARN of the IAM role used to perform Iceberg table optimization operations.
* `compaction` - (Optional) A map of key-value pairs specifying compaction configuration parameters.
* `retention` - (Optional) A map of key-value pairs specifying retention configuration parameters.
* `orphan_file_deletion` - (Optional) A map of key-value pairs specifying orphan file deletion configuration parameters.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The catalog ID.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Glue Catalog Level Table Optimizers using the `catalog_id`. For example:
```terraform
import {
  to = aws_glue_catalog_level_table_optimizer.example
  id = "123456789012"
}
```

Using `terraform import`:
```console
% terraform import aws_glue_catalog_level_table_optimizer.example 123456789012
```
