---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_rds_custom_db_engine_version"
description: |-
  Provides an custom engine version (CEV) resource for Amazon RDS Custom.
---

# Resource: aws_rds_custom_db_engine_version

Provides an custom engine version (CEV) resource for Amazon RDS Custom. For additional information, see [Working with CEVs for RDS Custom for Oracle](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/custom-cev.html) and [Working with CEVs for RDS Custom for SQL Server](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/custom-cev-sqlserver.html) in the the [RDS User Guide](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Welcome.html).

## Example Usage

### RDS Custom for Oracle Usage

```terraform
resource "aws_kms_key" "example" {
  description = "KMS symmetric key for RDS Custom for Oracle"
}

resource "aws_rds_custom_db_engine_version" "example" {
  database_installation_files_s3_bucket_name = "DOC-EXAMPLE-BUCKET"
  database_installation_files_s3_prefix      = "1915_GI/"
  engine                                     = "custom-oracle-ee-cdb"
  engine_version                             = "19.cdb_cev1"
  kms_key_id                                 = aws_kms_key.example.arn
  manifest                                   = <<JSON
  {
	"databaseInstallationFileNames":["V982063-01.zip"]
  }
  JSON

  tags = {
    Name = "example"
    Key  = "value"
  }
}
```

### RDS Custom for Oracle External Manifest Usage

```terraform
resource "aws_kms_key" "example" {
  description = "KMS symmetric key for RDS Custom for Oracle"
}

resource "aws_rds_custom_db_engine_version" "example" {
  database_installation_files_s3_bucket_name = "DOC-EXAMPLE-BUCKET"
  database_installation_files_s3_prefix      = "1915_GI/"
  engine                                     = "custom-oracle-ee-cdb"
  engine_version                             = "19.cdb_cev1"
  kms_key_id                                 = aws_kms_key.example.arn
  filename                                   = "manifest_1915_GI.json"
  manifest_hash                              = filebase64sha256(manifest_1915_GI.json)

  tags = {
    Name = "example"
    Key  = "value"
  }
}
```

### RDS Custom for SQL Server Usage

```terraform
# CEV creation requires an AMI owned by the operator
resource "aws_rds_custom_db_engine_version" "test" {
  engine          = "custom-sqlserver-se"
  engine_version  = "15.00.4249.2.cev-1"
  source_image_id = "ami-0aa12345678a12ab1"
}
```

### RDS Custom for SQL Server Usage with AMI from another region

```terraform
resource "aws_ami_copy" "example" {
  name              = "sqlserver-se-2019-15.00.4249.2"
  description       = "A copy of ami-xxxxxxxx"
  source_ami_id     = "ami-xxxxxxxx"
  source_ami_region = "us-east-1"
}
# CEV creation requires an AMI owned by the operator
resource "aws_rds_custom_db_engine_version" "test" {
  engine          = "custom-sqlserver-se"
  engine_version  = "15.00.4249.2.cev-1"
  source_image_id = aws_ami_copy.example.id
}
```

## Argument Reference

This resource supports the following arguments:

* `database_installation_files_s3_bucket_name` - (Optional) Name of the Amazon S3 bucket that contains the database installation files.
* `database_installation_files_s3_prefix` - (Optional) Prefix for the Amazon S3 bucket that contains the database installation files.
* `description` - (Optional) Description of the CEV.
* `engine` - (Required) Name of the database engine. Valid values are `custom-oracle*`, `custom-sqlserver*`.
* `engine_version` - (Required) Version of the database engine.
* `filename` - (Optional) Name of the manifest file within the local filesystem. Conflicts with `manifest`.
* `kms_key_id` - (Optional) ARN of the AWS KMS key that is used to encrypt the database installation files. Required for RDS Custom for Oracle.
* `manifest` - (Optional) Manifest file, in JSON format, that contains the list of database installation files. Conflicts with `filename`.
* `manifest_hash` - (Optional) Triggers updates. Must be set to a base64-encoded SHA256 hash of the manifest source specified with `filename`. The usual way to set this is filebase64sha256("manifest.json") where "manifest.json" is the local filename of the manifest source.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `source_image_id` - (Optional) ID of the AMI to create the CEV from. Required for RDS Custom for SQL Server. For RDS Custom for Oracle, you can specify an AMI ID that was used in a different Oracle CEV.
* `status` - (Optional) Status of the CEV. Valid values are `available`, `inactive`, `inactive-except-restore`.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN for the custom engine version.
* `create_time` - Date and time that the CEV was created.
* `db_parameter_group_family` - Name of the DB parameter group family for the CEV.
* `image_id` - ID of the AMI that was created with the CEV.
* `major_engine_version` - Major version of the database engine.
* `manifest_computed` - Returned manifest file, in JSON format, service generated and often different from input `manifest`.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `240m`)
- `update` - (Default `10m`)
- `delete` - (Default `60m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import custom engine versions for Amazon RDS custom using the `engine` and `engine_version` separated by a colon (`:`). For example:

```terraform
import {
  to = aws_rds_custom_db_engine_version.example
  id = "custom-oracle-ee-cdb:19.cdb_cev1"
}
```

Using `terraform import`, import custom engine versions for Amazon RDS custom using the `engine` and `engine_version` separated by a colon (`:`). For example:

```console
% terraform import aws_rds_custom_db_engine_version.example custom-oracle-ee-cdb:19.cdb_cev1
```
