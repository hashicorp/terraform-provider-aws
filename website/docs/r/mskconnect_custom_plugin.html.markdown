---
subcategory: "Managed Streaming for Kafka Connect"
layout: "aws"
page_title: "AWS: aws_mskconnect_custom_plugin"
description: |-
  Provides an Amazon MSK Connect custom plugin resource.
---

# Resource: aws_mskconnect_custom_plugin

Provides an Amazon MSK Connect Custom Plugin Resource.

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

* `arn` - the Amazon Resource Name (ARN) of the custom plugin.
* `latest_revision` - an ID of the latest successfully created revision of the custom plugin.
* `state` - the state of the custom plugin.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `delete` - (Default `10m`)

## Import

MSK Connect Custom Plugin can be imported using the plugin's `arn`, e.g.,

```
$ terraform import aws_mskconnect_custom_plugin.example 'arn:aws:kafkaconnect:eu-central-1:123456789012:custom-plugin/debezium-example/abcdefgh-1234-5678-9abc-defghijklmno-4'
```
