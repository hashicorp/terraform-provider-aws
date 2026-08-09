---
subcategory: "AppFabric"
layout: "aws"
page_title: "AWS: aws_appfabric_ingestion_destination"
description: |-
  Terraform resource for managing an AWS AppFabric Ingestion Destination.
---

# Resource: aws_appfabric_ingestion_destination

Terraform resource for managing an AWS AppFabric Ingestion Destination.

## Example Usage

### Basic Usage

```terraform
resource "aws_appfabric_ingestion_destination" "example" {
  app_bundle_arn = aws_appfabric_app_bundle.example.arn
  ingestion_arn  = aws_appfabric_ingestion.example.arn

  processing_configuration {
    audit_log {
      format = "json"
      schema = "raw"
    }
  }

  destination_configuration {
    audit_log {
      destination {
        s3_bucket {
          bucket_name = aws_s3_bucket.example.bucket
        }
      }
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `app_bundle_arn` - (Required) Amazon Resource Name (ARN) of the app bundle to use for the request.
* `destination_configuration` - (Required) Configuration for the destination of ingested data. See [`destination_configuration` Block](#destination_configuration-block) below.
* `ingestion_arn` - (Required) Amazon Resource Name (ARN) of the ingestion to use for the request.
* `processing_configuration` - (Required) Configuration for how ingested data is processed. See [`processing_configuration` Block](#processing_configuration-block) below.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `destination_configuration` Block

The `destination_configuration` block supports the following arguments:

* `audit_log` - (Required) Audit log destination configuration. See [`destination_configuration.audit_log` Block](#destination_configurationaudit_log-block) below.

### `destination_configuration.audit_log` Block

The `destination_configuration.audit_log` block supports the following arguments:

* `destination` - (Required) Destination for the audit log. Only one destination, either `firehose_stream` or `s3_bucket`, can be specified. See [`destination_configuration.audit_log.destination` Block](#destination_configurationaudit_logdestination-block) below.

### `destination_configuration.audit_log.destination` Block

The `destination_configuration.audit_log.destination` block supports the following arguments:

* `firehose_stream` - (Optional) Amazon Data Firehose delivery stream destination. See [`destination_configuration.audit_log.destination.firehose_stream` Block](#destination_configurationaudit_logdestinationfirehose_stream-block) below.
* `s3_bucket` - (Optional) Amazon S3 bucket destination. See [`destination_configuration.audit_log.destination.s3_bucket` Block](#destination_configurationaudit_logdestinations3_bucket-block) below.

### `destination_configuration.audit_log.destination.firehose_stream` Block

The `destination_configuration.audit_log.destination.firehose_stream` block supports the following arguments:

* `stream_name` - (Required) Name of the Amazon Data Firehose delivery stream.

### `destination_configuration.audit_log.destination.s3_bucket` Block

The `destination_configuration.audit_log.destination.s3_bucket` block supports the following arguments:

* `bucket_name` - (Required) Name of the Amazon S3 bucket.
* `prefix` - (Optional) Object key to use.

### `processing_configuration` Block

The `processing_configuration` block supports the following arguments:

* `audit_log` - (Required) Audit log processing configuration. See [`processing_configuration.audit_log` Block](#processing_configurationaudit_log-block) below.

### `processing_configuration.audit_log` Block

The `processing_configuration.audit_log` block supports the following arguments:

* `format` - (Required) Format in which the audit logs need to be formatted. Valid values: `json`, `parquet`.
* `schema` - (Required) Event schema in which the audit logs need to be formatted. Valid values: `ocsf`, `raw`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Ingestion Destination.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)
