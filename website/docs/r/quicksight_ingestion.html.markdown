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

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the Ingestion.
* `id` - A comma-delimited string joining AWS account ID, data set ID, and ingestion ID.
* `ingestion_status` - Ingestion status.

## Import

QuickSight Ingestion can be imported using the AWS account ID, data set ID, and ingestion ID separated by commas (`,`) e.g.,

```
$ terraform import aws_quicksight_ingestion.example 123456789012,example-dataset-id,example-ingestion-id
```
