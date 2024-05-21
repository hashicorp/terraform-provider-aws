---
subcategory: "Rekognition"
layout: "aws"
page_title: "AWS: aws_rekognition_stream_processor"
description: |-
  Terraform resource for managing an AWS Rekognition Stream Processor.
---

# Resource: aws_rekognition_stream_processor

Terraform resource for managing an AWS Rekognition Stream Processor.

~> **Note:** This resource must be configured specifically for your use case, and not all options are compatible with one another. See [Stream Processor API documentation](https://docs.aws.amazon.com/rekognition/latest/APIReference/API_CreateStreamProcessor.html#rekognition-CreateStreamProcessor-request-Input) for configuration information.

## Example Usage

### Basic Usage

```terraform
resource "aws_rekognition_stream_processor" "example" {
}
```

## Argument Reference

The following arguments are required:

* `input` - (Required) Input video stream. See [`input`](#input) definition.
* `name` - (Required) The name of the Stream Processor
* `role_arn` - (Required) The ARN of the IAM role that allows access to the stream processor.
* `output` - (Required) Kinesis data stream stream or Amazon S3 bucket location to which Amazon Rekognition Video puts the analysis results

The following arguments are optional:

* `kms_key_id` - (Optional) Optional parameter for label detection stream processors
* `data_sharing_preference` - (Optional) See [`data_sharing_preference`](#data_sharing_preference) definition.
* `notification_channel` - (Optional) The Amazon Simple Notification Service topic to which Amazon Rekognition publishes the completion status. See [`notification_channel`](#notification_channel) definition.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Nested Blocks

### `input`

* `kinesis_video_stream` - Kinesis input stream. See [`kinesis_video_stream`](#kinesis_video_stream) definition.

#### `kinesis_video_stream`

* `arn` - ARN of the Kinesis video stream stream that streams the source video

### `output`

* `kinesis_data_stream` - (Optional) The Amazon Kinesis Data Streams stream to which the Amazon Rekognition stream processor streams the analysis results. See [`kinesis_data_stream`](#kinesis_data_stream) definition.
* `s3_destination` - (Optiona) The Amazon S3 bucket location to which Amazon Rekognition publishes the detailed inference results of a video analysis operation. See [`s3_destination`](#s3_destination) definition.

#### `kinesis_data_stream`

* `arn` - ARN of the output Amazon Kinesis Data Streams stream.

#### `s3_destination`

* `bucket` - The name of the Amazon S3 bucket you want to associate with the streaming video project
* `key_prefixx` - The prefix value of the location within the bucket that you want the information to be published to

### `data_sharing_preference`

* `opt_in` - (Optional) Shows whether you are sharing data with Rekognition to improve model performance.

### `notification_channel`

* `sns_topic_arn` - The Amazon Resource Number (ARN) of the Amazon Amazon Simple Notification Service topic to which Amazon Rekognition posts the completion status.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Stream Processor. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Rekognition Stream Processor using the `example_id_arg`. For example:

```terraform
import {
  to = aws_rekognition_stream_processor.example
  id = "stream_processor-id-12345678"
}
```

Using `terraform import`, import Rekognition Stream Processor using the `example_id_arg`. For example:

```console
% terraform import aws_rekognition_stream_processor.example stream_processor-id-12345678
```
