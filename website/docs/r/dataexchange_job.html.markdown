---
subcategory: "Data Exchange"
layout: "aws"
page_title: "AWS: aws_dataexchange_job"
description: |-
  Provides a DataExchange Job resource.
---

# Resource: aws_dataexchange_job

Provides a resource to manage AWS Data Exchange Jobs.
Jobs are asynchronous import or export operations used to create or copy assets.

## Example Usage

```terraform
resource "aws_dataexchange_job" "example" {
  type        = "IMPORT_ASSETS_FROM_S3"
  data_set_id = aws_dataexchange_data_set.example.id
  revision_id = aws_dataexchange_revision.example.id

  s3_asset_sources = [
    {
      bucket = "source-bucket"
      key    = "source/key/asset.csv"
    }
  ]

  start_on_creation = true
}
```

## Argument Reference

The following arguments are supported:

* `type` - (Required) The type of job to be created. Valid values: `IMPORT_ASSETS_FROM_S3`, `IMPORT_ASSET_FROM_SIGNED_URL`, `IMPORT_ASSETS_FROM_REDSHIFT_DATA_SHARES`, `IMPORT_ASSET_FROM_API_GATEWAY_API`, `EXPORT_ASSETS_TO_S3`, `EXPORT_ASSET_TO_SIGNED_URL`, `EXPORT_REVISIONS_TO_S3`, `IMPORT_ASSETS_FROM_LAKE_FORMATION_TAG_POLICY`.
* `data_set_id` - (Required) The ID of the data set associated with this job.
* `start_on_creation` - (Optional) If true, starts the job upon creation. Defaults to `false`.
* `revision_id` - (Optional) The ID of the revision associated with this job.

Depending on the job type, additional arguments are required:

### For EXPORT_ASSETS_TO_S3

* `s3_asset_destinations` - (Required) List of S3 destinations for the assets. Each destination includes:
    * `asset_id` - (Required) The ID of the asset to be exported.
    * `bucket` - (Required) The S3 bucket that is the destination for the asset.
    * `key` - (Required) The S3 key that is the destination for the asset.
* `s3_asset_destination_encryption_type` - (Optional) The type of encryption to use. Valid values: `aws:kms`.
* `s3_asset_destination_encryption_kms_key_arn` - (Optional) The ARN of the KMS key to use for encryption.

### For IMPORT_ASSETS_FROM_S3

* `s3_asset_sources` - (Required) List of S3 sources for the assets. Each source includes:
    * `bucket` - (Required) The S3 bucket that contains the asset.
    * `key` - (Required) The S3 key that identifies the asset.

[Additional sections for other job types...]

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the job.
* `arn` - The ARN of the job.
* `state` - The current state of the job.
* `created_at` - The date and time the job was created.
* `last_updated_time` - The date and time the job was last updated.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Data Exchange Job using the `id`. For example:

```terraform
import {
  to = aws_dataexchange_job.example
  id = "abc123def789"
}
```

Using `terraform import`, import Data Exchange Job using the `id`. For example:

```console
% terraform import aws_dataexchange_job.example abc123def789
```
