---
subcategory: "DynamoDB"
layout: "aws"
page_title: "AWS: aws_dynamodb_table_export"
description: |-
  Terraform resource for managing an AWS DynamoDB Table Export.
---

# Resource: aws_dynamodb_table_export

Terraform resource for managing an AWS DynamoDB Table Export. Terraform will wait until the Table export reaches a status of `COMPLETED` or `FAILED`.

See the [AWS Documentation](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/S3DataExport.HowItWorks.html) for more information on how this process works.

~> **TIP:** Point-in-time Recovery must be enabled on the target DynamoDB Table.

~> **NOTE:** Once a AWS DynamoDB Table Export has been created it is immutable. The AWS API does not delete this resource. When you run destroy the provider will remove the resource from the Terraform state, no exported data will be deleted.

## Example Usage

### Basic Usage

```terraform

resource "aws_s3_bucket" "example" {
  bucket_prefix = "example"
  force_destroy = true
}

resource "aws_dynamodb_table" "example" {
  name         = "example-table-1"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "user_id"
  attribute {
    name = "user_id"
    type = "S"
  }

  point_in_time_recovery {
    enabled = true
  }
}

resource "aws_dynamodb_table_export" "example" {
  table_arn = aws_dynamodb_table.example.arn
  s3_bucket = aws_s3_bucket.example.id
}
```

### Example with export time

```terraform
resource "aws_dynamodb_table_export" "example" {
  export_time = "2023-04-02T11:30:13+01:00"
  s3_bucket   = aws_s3_bucket.example.id
  table_arn   = aws_dynamodb_table.example.arn
}
```

## Argument Reference

The following arguments are required:

* `s3_bucket` - (Required, Forces new resource) Name of the Amazon S3 bucket to export the snapshot to. See the [AWS Documentation](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/S3DataExport_Requesting.html#S3DataExport_Requesting_Permissions) for information on how configure this S3 bucket.
* `table_arn` - (Required, Forces new resource) ARN associated with the table to export.

The following arguments are optional:

* `export_format` - (Optional, Forces new resource) Format for the exported data. Valid values are `DYNAMODB_JSON` or `ION`. See the [AWS Documentation](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/S3DataExport.Output.html#S3DataExport.Output_Data) for more information on these export formats. Default is `DYNAMODB_JSON`.
* `export_time` - (Optional, Forces new resource) Time in RFC3339 format from which to export table data. The table export will be a snapshot of the table's state at this point in time. Omitting this value will result in a snapshot from the current time.
* `s3_bucket_owner` - (Optional, Forces new resource) ID of the AWS account that owns the bucket the export will be stored in.
* `s3_prefix` - (Optional, Forces new resource) Amazon S3 bucket prefix to use as the file name and path of the exported snapshot.
* `s3_sse_algorithm` - (Optional, Forces new resource) Type of encryption used on the bucket where export data will be stored. Valid values are: `AES256`, `KMS`.
* `s3_sse_kms_key_id` - (Optional, Forces new resource) ID of the AWS KMS managed key used to encrypt the S3 bucket where export data will be stored (if applicable).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Table Export.
* `billed_size_in_bytes` - Billable size of the table export.
* `end_time` - Time at which the export task completed.
* `export_status` - Status of the export - export can be in one of the following states `IN_PROGRESS`, `COMPLETED`, or `FAILED`.
* `item_count` - Number of items exported.
* `manifest_files_s3_key` - Name of the manifest file for the export task. See the [AWS Documentation](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/S3DataExport.Output.html#S3DataExport.Output_Manifest) for more information on this manifest file.
* `start_time` - Time at which the export task began.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `delete` - (Default `60m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DynamoDB table exports using the `arn`. For example:

```terraform
import {
  to = aws_dynamodb_table_export.example
  id = "arn:aws:dynamodb:us-west-2:12345678911:table/my-table-1/export/01580735656614-2c2f422e"
}
```

Using `terraform import`, import DynamoDB table exports using the `arn`. For example:

```console
% terraform import aws_dynamodb_table_export.example arn:aws:dynamodb:us-west-2:12345678911:table/my-table-1/export/01580735656614-2c2f422e
```
