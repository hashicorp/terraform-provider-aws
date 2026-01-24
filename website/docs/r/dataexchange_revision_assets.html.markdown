---
subcategory: "Data Exchange"
layout: "aws"
page_title: "AWS: aws_dataexchange_revision_assets"
description: |-
  Terraform resource for managing AWS Data Exchange Revision Assets.
---

# Resource: aws_dataexchange_revision_assets

Terraform resource for managing AWS Data Exchange Revision Assets.

~> Note: This resource creates a new revision and adds associated assets. Destroying this resource will delete the revision and all associated assets.

## Example Usage

### Basic Usage

```terraform
resource "aws_dataexchange_revision_assets" "example" {
  data_set_id = "example-data-set-id"

  asset {
    create_s3_data_access_from_s3_bucket {
      asset_source {
        bucket = "example-bucket"
      }
    }
  }

  tags = {
    Environment = "Production"
  }
}
```

## Argument Reference

The following arguments are required:

* `data_set_id` - (Required) Unique identifier for the data set associated with the revision.
* `asset` - (Required) A block to define the asset associated with the revision. See [Asset](#asset) for more details.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `comment` - (Optional) A comment for the revision. Maximum length is 16,348 characters.
* `finalize` - (Optional) Finalized a revision. Defaults to `false`.
* `force_destoy` - (Optional) Force destroy the revision. Defaults to `false`.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### asset

* `create_s3_data_access_from_s3_bucket` - (Optional) A block to create S3 data access from an S3 bucket. See [Create S3 Data Access from S3 Bucket](#create_s3_data_access_from_s3_bucket) for more details.
* `import_assets_from_s3` - (Optional) A block to import assets from S3. See [Import Assets from S3](#import_assets_from_s3) for more details.
* `import_assets_from_signed_url` - (Optional) A block to import assets from a signed URL. See [Import Assets from Signed URL](#import_assets_from_signed_url) for more details.

#### create_s3_data_access_from_s3_bucket

* `asset_source` - (Required) A block specifying the source bucket for the asset. This block supports the following:
    * `bucket` - (Required) The name of the S3 bucket.
        * `keys` - (Required) List of object keys in the S3 bucket.
        * `key_prefixes` - (Optional) List of key prefixes in the S3 bucket.
        * `kms_key_to_grant` - (Optional) A block specifying the KMS key to grant access. This block supports the following:
            * `kms_key_arn` - (Required) The ARN of the KMS key.

### import_assets_from_s3

* `asset_source` - (Required) A block specifying the source bucket and key for the asset. This block supports the following:
    * `bucket` - (Required) The name of the S3 bucket.
    * `key` - (Required) The key of the object in the S3 bucket.

### import_assets_from_signed_url

* `filename` - (Required) The name of the file to import.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN of the Data Exchange Revision Assets.
* `id` - The unique identifier for the revision.
* `created_at` - The timestamp when the revision was created, in RFC3339 format.
* `updated_at` - The timestamp when the revision was last updated, in RFC3339 format.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

Configuration options:

* `create` - (Default 30m) Time to create the revision.
