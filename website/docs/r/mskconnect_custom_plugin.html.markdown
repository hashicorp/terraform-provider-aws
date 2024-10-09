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

* `name` - (Required, Forces new resource) The name of the custom plugin..
* `content_type` - (Required, Forces new resource) The type of the plugin file. Allowed values are `ZIP` and `JAR`.
* `location` - (Required, Forces new resource) Information about the location of a custom plugin. See [`location` Block](#location-block) for details.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

The following arguments are optional:

* `description` - (Optional, Forces new resource) A summary description of the custom plugin.

### `location` Block

The `location` configuration block supports the following arguments:

* `s3` - (Required, Forces new resource) Information of the plugin file stored in Amazon S3. See [`s3` Block](#s3-block) for details..

### `s3` Block

The `s3` configuration Block supports the following arguments:

* `bucket_arn` - (Required, Forces new resource) The Amazon Resource Name (ARN) of an S3 bucket.
* `file_key` - (Required, Forces new resource) The file key for an object in an S3 bucket.
* `object_version` - (Optional, Forces new resource) The version of an object in an S3 bucket.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - the Amazon Resource Name (ARN) of the custom plugin.
* `latest_revision` - an ID of the latest successfully created revision of the custom plugin.
* `state` - the state of the custom plugin.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `delete` - (Default `10m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import MSK Connect Custom Plugin using the plugin's `arn`. For example:

```terraform
import {
  to = aws_mskconnect_custom_plugin.example
  id = "arn:aws:kafkaconnect:eu-central-1:123456789012:custom-plugin/debezium-example/abcdefgh-1234-5678-9abc-defghijklmno-4"
}
```

Using `terraform import`, import MSK Connect Custom Plugin using the plugin's `arn`. For example:

```console
% terraform import aws_mskconnect_custom_plugin.example 'arn:aws:kafkaconnect:eu-central-1:123456789012:custom-plugin/debezium-example/abcdefgh-1234-5678-9abc-defghijklmno-4'
```
