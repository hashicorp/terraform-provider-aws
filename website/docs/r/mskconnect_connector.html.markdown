---
subcategory: "Kafka Connect (MSK Connect)"
layout: "aws"
page_title: "AWS: aws_mskconnect_connector"
description: |-
  Provides an Amazon MSK Connect Connector resource.
---

# Resource: aws_mskconnect_connector

Provides an Amazon MSK Connect Connector resource.

## Example Usage

### Basic configuration

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "example"
}

resource "aws_s3_object" "example" {
  bucket = aws_s3_bucket.example.id
  key    = "debezium.zip"
  source = "debezium.zip"
}

resource "aws_mskconnect_custom_plugin" "example" {
  name         = "debezium-example"
  content_type = "ZIP"
  location {
    s3 {
      bucket_arn = aws_s3_bucket.example.arn
      file_key   = aws_s3_object.example.key
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) The name of the custom plugin..
* `content_type` - (Required) The type of the plugin file. Allowed values are `ZIP` and `JAR`.
* `location` - (Required) Information about the location of a custom plugin. See below.

The following arguments are optional:

* `description` - (Optional) A summary description of the custom plugin.

### location Argument Reference

* `s3` - (Required) Information of the plugin file stored in Amazon S3. See below.

#### location s3 Argument Reference

* `bucket_arn` - (Required) The Amazon Resource Name (ARN) of an S3 bucket.
* `file_key` - (Required) The file key for an object in an S3 bucket.
* `object_version` - (Optional) The version of an object in an S3 bucket.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - the Amazon Resource Name (ARN) of the connector.
* `version` - an ID of the latest successfully created version of the connector.
* `state` - the state of the connector.


## Timeouts

`aws_mskconnect_custom_plugin` provides the following
[Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) configuration options:

* `create` - (Default `20 minutes`) How long to wait for the MSK Connect Connector to be created.
* `update` - (Default `20 minutes`) How long to wait for the MSK Connect Connector to be created.
* `delete` - (Default `10 minutes`) How long to wait for the MSK Connect Connector to be created.

## Import

MSK Connect Connector can be imported using the connector's `arn`, e.g.,

```
$ terraform import aws_mskconnect_connector.example 'arn:aws:kafkaconnect:eu-central-1:123456789012:connector/example/264edee4-17a3-412e-bd76-6681cfc93805-3'
```
