---
subcategory: "Rekognition"
layout: "aws"
page_title: "AWS: aws_rekognition_stream_processor"
description: |-
  Terraform resource for managing an AWS Rekognition Stream Processor.
---

# Resource: aws_rekognition_stream_processor

Provides a Rekognition Stream Processor resource. Amazon Rekognition Video is a consumer of live video from Amazon Kinesis Video Streams that can detect and recognize faces or labels in a streaming video.

## Example Usage

### Label Detection

```terraform
resource "aws_rekognition_stream_processor" "example" {
  name = example
  input {
    kinesis_video_stream {
      arn = aws_kinesis_video_stream.test.arn
    }
  }	
  output {
    s3_destination {
      bucket = aws_s3_bucket.test.id
    }
  }
  notification_channel {
    sns_topic_arn = aws_sns_topic.test.arn
  }
  role_arn = aws_iam_role.test.arn
  settings {
    connected_home {
      labels = ["ALL"]
    }
  }
}

resource "aws_iam_role" "example" {
  name = example
  path = "/service-role/"
  assume_role_policy = jsonencode({
    "Version" = "2012-10-17",
    "Statement" = [
      {
        "Effect" = "Allow",
        "Principal" = {
          "Service" = [
            "rekognition.amazonaws.com",
          ]
        },
        "Action" = "sts:AssumeRole"
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "example" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonRekognitionServiceRole"
}

resource "aws_iam_role_policy" "example" {
	name   = example
	role   = aws_iam_role.test.id
	policy = jsonencode({
    "Version" = "2012-10-17",
    "Statement" = [
      {
        "Effect" = "Allow",
        "Action" = [
          "s3:PutObject"
        ],
        "Resource" = [
          "${aws_s3_bucket.test.arn}/*"
        ]
      }
    ]
  })
}

resource "aws_kinesis_video_stream" "example" {
  name = example 
}

resource "aws_s3_bucket" "example" {
  bucket = example
}	

resource "aws_sns_topic" "example" {
  name = format("%%s-%%s", "AmazonRekognition", "example")
}
```

### Face Search

```terraform
resource "aws_rekognition_stream_processor" "example" {
  name = example
  input {
    kinesis_video_stream {
      arn = aws_kinesis_video_stream.test.arn
    }
  }	
  output {
    kinesis_data_stream {
      arn = aws_kinesis_stream.test.arn
    }
  }
  role_arn = aws_iam_role.test.arn
  settings {
    face_search {
      collection_id = aws_rekognition_collection.test.collection_id
    }
  }
}

resource "aws_iam_role" "example" {
  name = example
  path = "/service-role/"
  assume_role_policy = jsonencode({
    "Version" = "2012-10-17",
    "Statement" = [
      {
        "Effect" = "Allow",
        "Principal" = {
          "Service" = [
            "rekognition.amazonaws.com",
          ]
        },
        "Action" = "sts:AssumeRole"
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "example" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonRekognitionServiceRole"
}

resource "aws_kinesis_video_stream" "example" {
  name = example
}

resource "aws_kinesis_stream" "example" {
  name             = example
  shard_count      = 1
  stream_mode_details {
    stream_mode    = "PROVISIONED"
  }
}

resource "aws_rekognition_collection" "example" {
  collection_id = example
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) A name to identify the stream processor. This is unique to the AWS account and region the Stream Processor is created in.

* `input` - (Required) A configuration that define the Kinesis video stream that provides the source streaming video. This is required for both face search and label detection stream processors.
	* `kinesis_video_stream` - (Optional) The Kinesis video stream input stream for the source streaming video
		* `arn` - (Optional) ARN of the Kinesis video stream that streams the source video.

* `role_arn` - (Required) ARN of the IAM role to be assumed by Rekognition. The IAM role must have read permissions for a Kinesis stream and write permissions to an Amazon S3 bucket and Amazon Simple Notification Service topic for a label detection. This is required for both face search and label detection stream processors.

* `output` - (Required) Allows the ability to specify the Kinesis data stream stream or S3 bucket location to which Amazon Rekognition Video puts the analysis results.
	* `kinesis_data_stream` - (Optional) The Amazon Kinesis Data Streams stream to which the Amazon Rekognition stream processor streams the analysis results.
		* `arn` - (Optional) ARN of the output Amazon Kinesis Data Streams stream.
	* `s3_destination` - (Optional)
		* `bucket` - (Optional) The name of the Amazon S3 bucket to associate with the stream processor.
		* `key_prefix` - (Optional) The prefix value of the location within the bucket.

* `settings` - (Required) A configurations used in a streaming video analyzed by a stream processor. 
	* `connected_home` - (Optional) The label detection settings to use on a streaming video. More details are given below.
	* `face_search` - (Optional) The face search settings to use on a streaming video. More details are given below. 

The `connected_home` object supports the following:

* `labels` - (Required) Type of detection in the video. Valid values are `PERSON`, `PET`, and `ALL`.
* `min_confidence` - (Optional) The minimum confidence required to label an object in the video. 

The `face_search` object supports the following:

* `collection_id` - (Optional) The ID of a collection that contains faces to detect
* `face_match_threashold` - (Optional) Minimum face match confidence score that must be met to return a result for a recognized face. The default value is `80`, the minimum is `0`, and the maximum is `100`.


The following arguments are optional:

* `data_sharing_preference` - (Optional) Configuration option to opt-in whether you are sharing data with Rekognition to improve model performance. 
	* `opt_in` - (Required) If this option is set to true, you choose to share data with Rekognition. 

* `kms_key_id` - (Optional) Specifies the identifier for a KMS key that the stream will use to encrypt results and data published to Amazon S3. This is an optional parameter for label detection stream processors and should not be used to create a face search stream processor. 

* `notification_channel` - (Optional) Configuration option to specify the Amazon SNS topic to which Amazon Rekognition publishes the object detection results and completion status of a video analysis operation.
	* `sns_topic_arn` - (Required) ARN of the output Amazon Simple Notification Service topic.. 

* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

* `regions_of_interest` - (Optional) Configuration option to specify locations in the frames where Amazon Rekognition checks for objects or people. This is an optional parameter for label detection stream processors and should not be used to create a face search stream processor. The maximum number of regions is `10`.
	* `bounding_box` - (Optional) The box representing a region of interest on screen.
	* `polygon` - (Optional) The shape made up of up to 10 points.

The `bounding_box` object supports the following:

* `height` - (Optional) Height of the bounding box as a ratio of the overall image height.
* `left` - (Optional) Left coordinate of the bounding box as a ratio of overall image width.
* `top` - (Optional) Top coordinate of the bounding box as a ratio of overall image height.
* `width` - (Optional) Width of the bounding box as a ratio of the overall image width.

The `polygon` object supports the following:

* `point` - (Required) The X and Y coordinates of a point on an image or video frame. The X and Y values are ratios of the overall image size or video resolution.
	* `x` - (Optional) The value of the X coordinate for a point on a Polygon
	* `y` - (Optional) The value of the X coordinate for a point on a Polygon

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the Stream Processor. 
* `id` - Stream Processor ID
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block). 

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

Rekognition Stream Processor can be imported using the `name`, e.g.,

```
$ terraform import aws_rekognition_stream_processor.example rft-8012925589
```
