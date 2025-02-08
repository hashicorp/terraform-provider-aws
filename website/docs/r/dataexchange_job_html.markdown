---
subcategory: "Data Exchange"
layout: "aws"
page_title: "AWS: aws_dataexchange_job"
description: |-
  Provides a DataExchange Job
---

# Resource: aws_dataexchange_job

Provides a resource to manage AWS Data Exchange Jobs.

## Example Usage

### Import Assets from S3

```terraform
resource "aws_dataexchange_data_set" "example" {
  asset_type  = "S3_SNAPSHOT"
  description = "example"
  name        = "example"
}

resource "aws_dataexchange_revision" "example" {
  data_set_id = aws_dataexchange_data_set.example.id
}

resource "aws_dataexchange_job" "example" {
  type = "IMPORT_ASSETS_FROM_S3"
  
  details {
    import_assets_from_s3 {
      data_set_id = aws_dataexchange_data_set.example.id
      revision_id = aws_dataexchange_revision.example.id
      
      asset_sources {
        bucket = "example-bucket"
        key    = "example-key"
      }
    }
  }
}
```

### Export Assets to S3

```terraform
resource "aws_dataexchange_job" "example" {
  type = "EXPORT_ASSETS_TO_S3"
  
  details {
    export_assets_to_s3 {
      data_set_id = aws_dataexchange_data_set.example.id
      revision_id = aws_dataexchange_revision.example.id
      
      asset_destinations {
        bucket   = "example-bucket"
        key      = "example-key"
        asset_id = "example-asset-id"
      }
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `type` - (Required) Type of the job. Valid values: `IMPORT_ASSETS_FROM_S3`, `EXPORT_ASSETS_TO_S3`, `IMPORT_ASSET_FROM_SIGNED_URL`, `EXPORT_ASSET_TO_SIGNED_URL`.
* `details` - (Required) Details for the job. See [Details](#details) below.

The following arguments are optional:

* `start_on_creation` - (Optional) Whether to start the job immediately upon creation.

### Details

The `details` block supports the following:

* `import_assets_from_s3` - (Optional) Configuration details for an import from S3 job. See [Import Assets from S3](#import-assets-from-s3) below.
* `export_assets_to_s3` - (Optional) Configuration details for an export to S3 job. See [Export Assets to S3](#export-assets-to-s3) below.
* `import_asset_from_signed_url` - (Optional) Configuration details for an import from signed URL job. See [Import Asset from Signed URL](#import-asset-from-signed-url) below.
* `export_asset_to_signed_url` - (Optional) Configuration details for an export to signed URL job. See [Export Asset to Signed URL](#export-asset-to-signed-url) below.

#### Import Assets from S3

The `import_assets_from_s3` block supports the following:

* `data_set_id` - (Required) The unique identifier for the data set.
* `revision_id` - (Required) The unique identifier for the revision.
* `asset_sources` - (Required) Source information for the assets to be imported. See [Asset Sources](#asset-sources) below.

#### Export Assets to S3

The `export_assets_to_s3` block supports the following:

* `data_set_id` - (Required) The unique identifier for the data set.
* `revision_id` - (Required) The unique identifier for the revision.
* `asset_destinations` - (Required) Destination information for the assets to be exported. See [Asset Destinations](#asset-destinations) below.
* `encryption` - (Optional) Encryption configuration for the exported assets. See [Encryption](#encryption) below.

#### Import Asset from Signed URL

The `import_asset_from_signed_url` block supports the following:

* `data_set_id` - (Required) The unique identifier for the data set.
* `revision_id` - (Required) The unique identifier for the revision.
* `asset_name` - (Required) The name of the asset.
* `md5_hash` - (Required) The MD5 hash of the asset.

#### Export Asset to Signed URL

The `export_asset_to_signed_url` block supports the following:

* `data_set_id` - (Required) The unique identifier for the data set.
* `revision_id` - (Required) The unique identifier for the revision.
* `asset_id` - (Required) The unique identifier for the asset.

### Asset Sources

The `asset_sources` block supports the following:

* `bucket` - (Required) The S3 bucket that contains the asset.
* `key` - (Required) The S3 key that identifies the asset.

### Asset Destinations

The `asset_destinations` block supports the following:

* `asset_id` - (Required) The unique identifier for the asset.
* `bucket` - (Required) The S3 bucket that is the destination for the asset.
* `key` - (Required) The S3 key that is the destination for the asset.

### Encryption

The `encryption` block supports the following:

* `type` - (Optional) Type of encryption. Valid values: `aws:kms`, `AES256`.
* `kms_key_arn` - (Optional) The ARN for the KMS key that will be used to encrypt the data.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the job.
* `id` - The unique identifier for the job.
* `state` - The current state of the job.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Data Exchange Job using the job ID. For example:

```terraform
import {
  to = aws_dataexchange_job.example
  id = "job-12345678"
}
```

Using `terraform import`, import Data Exchange Job using the job ID. For example:

```console
% terraform import aws_dataexchange_job.example job-12345678
```