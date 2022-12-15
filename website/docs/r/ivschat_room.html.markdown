---
subcategory: "IVS (Interactive Video) Chat"
layout: "aws"
page_title: "AWS: aws_ivschat_room"
description: |-
  Terraform resource for managing an AWS IVS (Interactive Video) Chat Room.
---

# Resource: aws_ivschat_room

Terraform resource for managing an AWS IVS (Interactive Video) Chat Room.

## Example Usage

### Basic Usage

```terraform
resource "aws_ivschat_room" "example" {
  name = "tf-room"
}
```

## Usage with Logging Configuration to S3 Bucket

```terraform
resource "aws_s3_bucket" "example" {
  bucket_prefix = "tf-ivschat-logging-bucket-"
  force_destroy = true
}

resource "aws_ivschat_logging_configuration" "example" {
  name = "tf-ivschat-loggingconfiguration"

  lifecycle {
    create_before_destroy = true
  }

  destination_configuration {
    s3 {
      bucket_name = aws_s3_bucket.example.id
    }
  }
}

resource "aws_ivschat_room" "example" {
  name                              = "tf-ivschat-room"
  logging_configuration_identifiers = [aws_ivschat_logging_configuration.example.arn]
}
```

## Argument Reference

The following arguments are optional:

* `logging_configuration_identifiers` - (Optional) List of Logging Configuration
  ARNs to attach to the room.
* `maximum_message_length` - (Optional) Maximum number of characters in a single
  message. Messages are expected to be UTF-8 encoded and this limit applies
  specifically to rune/code-point count, not number of bytes.
* `maximum_message_rate_per_second` - (Optional) Maximum number of messages per
  second that can be sent to the room (by all clients).
* `message_review_handler` - (Optional) Configuration information for optional
  review of messages.
    * `fallback_result` - (Optional) The fallback behavior (whether the message
    is allowed or denied) if the handler does not return a valid response,
    encounters an error, or times out. Valid values: `ALLOW`, `DENY`.
    * `uri` - (Optional) ARN of the lambda message review handler function.
* `name` - (Optional) Room name.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the Room.
* `id` - Room ID
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

IVS (Interactive Video) Chat Room can be imported using the ARN, e.g.,

```
$ terraform import aws_ivschat_room.example arn:aws:ivschat:us-west-2:326937407773:room/GoXEXyB4VwHb
```
