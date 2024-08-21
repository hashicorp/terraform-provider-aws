---
subcategory: "Rekognition"
layout: "aws"
page_title: "AWS: aws_rekognition_stream_processor"
description: |-
  Terraform resource for managing an AWS Rekognition Stream Processor.
---

# Resource: aws_rekognition_stream_processor

Terraform resource for managing an AWS Rekognition Stream Processor.

~> This resource must be configured specifically for your use case, and not all options are compatible with one another. See [Stream Processor API documentation](https://docs.aws.amazon.com/rekognition/latest/APIReference/API_CreateStreamProcessor.html#rekognition-CreateStreamProcessor-request-Input) for configuration information.

~> Stream Processors configured for Face Recognition cannot have _any_ properties updated after the fact, and it will result in an AWS API error.

## Example Usage

### Label Detection

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "example-bucket"
}

resource "aws_sns_topic" "example" {
  name = "example-topic"
}

resource "aws_kinesis_video_stream" "example" {
  name                    = "example-kinesis-input"
  data_retention_in_hours = 1
  device_name             = "kinesis-video-device-name"
  media_type              = "video/h264"
}

resource "aws_iam_role" "example" {
  name = "example-role"

  inline_policy {
    name = "Rekognition-Access"
    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Action   = ["s3:PutObject"]
          Effect   = "Allow"
          Resource = ["${aws_s3_bucket.example.arn}/*"]
        },
        {
          Action   = ["sns:Publish"]
          Effect   = "Allow"
          Resource = ["${aws_sns_topic.example.arn}"]
        },
        {
          Action = [
            "kinesis:Get*",
            "kinesis:DescribeStreamSummary"
          ]
          Effect   = "Allow"
          Resource = ["${aws_kinesis_video_stream.example.arn}"]
        },
      ]
    })
  }

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "rekognition.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_rekognition_stream_processor" "example" {
  role_arn = aws_iam_role.example.arn
  name     = "example-processor"

  data_sharing_preference {
    opt_in = false
  }

  output {
    s3_destination {
      bucket = aws_s3_bucket.example.bucket
    }
  }

  settings {
    connected_home {
      labels = ["PERSON", "PET"]
    }
  }

  input {
    kinesis_video_stream {
      arn = aws_kinesis_video_stream.example.arn
    }
  }

  notification_channel {
    sns_topic_arn = aws_sns_topic.example.arn
  }
}
```

### Face Detection Usage

```terraform
resource "aws_kinesis_video_stream" "example" {
  name                    = "example-kinesis-input"
  data_retention_in_hours = 1
  device_name             = "kinesis-video-device-name"
  media_type              = "video/h264"
}

resource "aws_kinesis_stream" "example" {
  name        = "terraform-kinesis-example"
  shard_count = 1
}

resource "aws_iam_role" "example" {
  name = "example-role"

  inline_policy {
    name = "Rekognition-Access"
    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Action = [
            "kinesis:Get*",
            "kinesis:DescribeStreamSummary"
          ]
          Effect   = "Allow"
          Resource = ["${aws_kinesis_video_stream.example.arn}"]
        },
        {
          Action = [
            "kinesis:PutRecord"
          ]
          Effect   = "Allow"
          Resource = ["${aws_kinesis_stream.example.arn}"]
        },
      ]
    })
  }

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "rekognition.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_rekognition_collection" "example" {
  collection_id = "example-collection"
}

resource "aws_rekognition_stream_processor" "example" {
  role_arn = aws_iam_role.example.arn
  name     = "example-processor"

  data_sharing_preference {
    opt_in = false
  }

  regions_of_interest {
    polygon {
      x = 0.5
      y = 0.5
    }
    polygon {
      x = 0.5
      y = 0.5
    }
    polygon {
      x = 0.5
      y = 0.5
    }
  }

  input {
    kinesis_video_stream {
      arn = aws_kinesis_video_stream.example.arn
    }
  }

  output {
    kinesis_data_stream {
      arn = aws_kinesis_stream.example.arn
    }
  }

  settings {
    face_search {
      collection_id = aws_rekognition_collection.example.id
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `input` - (Required) Input video stream. See [`input`](#input).
* `name` - (Required) The name of the Stream Processor.
* `output` - (Required) Kinesis data stream stream or Amazon S3 bucket location to which Amazon Rekognition Video puts the analysis results. See [`output`](#output).
* `role_arn` - (Required) The Amazon Resource Number (ARN) of the IAM role that allows access to the stream processor. The IAM role provides Rekognition read permissions for a Kinesis stream. It also provides write permissions to an Amazon S3 bucket and Amazon Simple Notification Service topic for a label detection stream processor. This is required for both face search and label detection stream processors.
* `settings` - (Required) Input parameters used in a streaming video analyzed by a stream processor. See [`settings`](#settings).

The following arguments are optional:

* `data_sharing_preference` - (Optional) See [`data_sharing_preference`](#data_sharing_preference).
* `kms_key_id` - (Optional) Optional parameter for label detection stream processors.
* `notification_channel` - (Optional) The Amazon Simple Notification Service topic to which Amazon Rekognition publishes the completion status. See [`notification_channel`](#notification_channel).
* `regions_of_interest` - (Optional) Specifies locations in the frames where Amazon Rekognition checks for objects or people. See [`regions_of_interest`](#regions_of_interest).
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `input`

* `kinesis_video_stream` - (Optional) Kinesis input stream. See [`kinesis_video_stream`](#kinesis_video_stream).

### `kinesis_video_stream`

* `arn` - (Optional) ARN of the Kinesis video stream stream that streams the source video.

### `output`

* `kinesis_data_stream` - (Optional) The Amazon Kinesis Data Streams stream to which the Amazon Rekognition stream processor streams the analysis results. See [`kinesis_data_stream`](#kinesis_data_stream).
* `s3_destination` - (Optional) The Amazon S3 bucket location to which Amazon Rekognition publishes the detailed inference results of a video analysis operation. See [`s3_destination`](#s3_destination).

### `kinesis_data_stream`

* `arn` - (Optional) ARN of the output Amazon Kinesis Data Streams stream.

### `s3_destination`

* `bucket` - (Optional) Name of the Amazon S3 bucket you want to associate with the streaming video project.
* `key_prefixx` - (Optional) Prefix value of the location within the bucket that you want the information to be published to.

### `data_sharing_preference`

* `opt_in` - (Optional) Whether you are sharing data with Rekognition to improve model performance.

### `regions_of_interest`

* `bounding_box` - (Optional) Box representing a region of interest on screen. Only 1 per region is allowed. See [`bounding_box`](#bounding_box).
* `polygon` - (Optional) Shape made up of up to 10 Point objects to define a region of interest. See [`polygon`](#polygon).

### `bounding_box`

A region can only have a single `bounding_box`

* `height` - (Required) Height of the bounding box as a ratio of the overall image height.
* `wight` - (Required) Width of the bounding box as a ratio of the overall image width.
* `left` - (Required) Left coordinate of the bounding box as a ratio of overall image width.
* `top` - (Required) Top coordinate of the bounding box as a ratio of overall image height.

### `polygon`

If using `polygon`, a minimum of 3 per region is required, with a maximum of 10.

* `x` - (Required) The value of the X coordinate for a point on a Polygon.
* `y` - (Required) The value of the Y coordinate for a point on a Polygon.

### `notification_channel`

* `sns_topic_arn` - (Required) The Amazon Resource Number (ARN) of the Amazon Amazon Simple Notification Service topic to which Amazon Rekognition posts the completion status.

### `settings`

* `connected_home` - (Optional) Label detection settings to use on a streaming video. See [`connected_home`](#connected_home).
* `face_search` - (Optional) Input face recognition parameters for an Amazon Rekognition stream processor. See [`face_search`](#face_search).

### `connected_home`

* `labels` - (Required) Specifies what you want to detect in the video, such as people, packages, or pets. The current valid labels you can include in this list are: `PERSON`, `PET`, `PACKAGE`, and `ALL`.
* `min_confidence` - (Optional) Minimum confidence required to label an object in the video.

### `face_search`

* `collection_id` - (Optional) ID of a collection that contains faces that you want to search for.
* `face_match_threshold` - (Optional) Minimum face match confidence score that must be met to return a result for a recognized face.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `stream_processor_arn` - ARN of the Stream Processor.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Rekognition Stream Processor using the `name`. For example:

```terraform
import {
  to = aws_rekognition_stream_processor.example
  id = "my-stream"
}
```

Using `terraform import`, import Rekognition Stream Processor using the `name`. For example:

```console
% terraform import aws_rekognition_stream_processor.example my-stream 
```
