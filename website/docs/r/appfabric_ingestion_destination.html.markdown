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
  app_bundle_identifier             = "aws_appfabric_app_bundle.arn"
  ingestion_identifier              = "aws_appfabric_ingestion.arn"
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
				bucket_name = "examplebucketname"
				prefix = "AuditLog"
			}
		}
    }
  }
}
```


## Argument Reference

The following arguments are required:

* `app_bundle_identifier` - (Required) The Amazon Resource Name (ARN) or Universal Unique Identifier (UUID) of the app bundle to use for the request.
* `ingestion_identifier` - (Required) The Amazon Resource Name (ARN) or Universal Unique Identifier (UUID) of the ingestion to use for the request.
* `destination_configuration` - (Required) Contains information about the destination of ingested data.
* `processing_configuration` - (Required) Contains information about how ingested data is processed.

Destination Configuration support the following:

* `audit_log` - (Optional) Contains information about an audit log destination configuration.

Audit Log Destination Configuration support the following:

* `destination` - (Required) Contains information about an audit log destination. Only one destination (Firehose Stream) or (S3 Bucket) can be specified. 

Destination support the following:

* `firehose_stream` - (Optional) Contains information about an Amazon Data Firehose delivery stream.
* `s3_bucket` - (Optional) Contains information about an Amazon S3 bucket.

Firehose Stream support the following:

* `streamName` - (Required) The name of the Amazon Data Firehose delivery stream.

S3 Bucket support the following:

* `bucketName` - (Required) The name of the Amazon S3 bucket.
* `prefix` - (Optional) The object key to use.

Processing Configuration support the following:

* `audit_log` - (optional) Contains information about an audit log processing configuration.

Audit Log Processing Configuration support the following:

* `format` - (Required) The format in which the audit logs need to be formatted. Valid values: json | parquet
* `schema` - (Required) The event schema in which the audit logs need to be formatted. Valid values: ocsf | raw

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Ingestion Destination. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.
* `app_bundle_arn` - The Amazon Resource Name (ARN) of the app bundle for the ingestion destination.
* `ingestion_arn` - The Amazon Resource Name (ARN) of the ingestion for the ingestion destination.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)