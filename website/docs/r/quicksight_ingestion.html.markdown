---
subcategory: "QuickSight"
layout: "aws"
page_title: "AWS: aws_quicksight_ingestion"
description: |-
  Terraform resource for managing an AWS QuickSight Ingestion.
---

# Resource: aws_quicksight_ingestion

Terraform resource for managing an AWS QuickSight Ingestion.

## Example Usage

### Basic Usage

```terraform
resource "aws_quicksight_ingestion" "example" {
  data_set_id    = aws_quicksight_data_set.example.data_set_id
  ingestion_id   = "example-id"
  ingestion_type = "FULL_REFRESH"
}
```

## Argument Reference

The following arguments are required:

* `data_set_id` - (Required) ID of the dataset used in the ingestion.
* `ingestion_id` - (Required) ID for the ingestion.
* `ingestion_type` - (Required) Type of ingestion to be created. Valid values are `INCREMENTAL_REFRESH` and `FULL_REFRESH`.

The following arguments are optional:

* `aws_account_id` - (Optional) AWS account ID.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Ingestion.
* `id` - A comma-delimited string joining AWS account ID, data set ID, and ingestion ID.
* `ingestion_status` - Ingestion status.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import QuickSight Ingestion using the AWS account ID, data set ID, and ingestion ID separated by commas (`,`). For example:

```terraform
import {
  to = aws_quicksight_ingestion.example
  id = "123456789012,example-dataset-id,example-ingestion-id"
}
```

Using `terraform import`, import QuickSight Ingestion using the AWS account ID, data set ID, and ingestion ID separated by commas (`,`). For example:

```console
% terraform import aws_quicksight_ingestion.example 123456789012,example-dataset-id,example-ingestion-id
```
