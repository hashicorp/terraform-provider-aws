---
subcategory: "Chime SDK Media Pipelines"
layout: "aws"
page_title: "AWS: aws_chimesdkmediapipelines_media_insights_pipeline_configuration"
description: |-
  Terraform resource for managing an AWS Chime SDK Media Pipelines Media Insights Pipeline Configuration.
---

# Resource: aws_chimesdkmediapipelines_media_insights_pipeline_configuration

Terraform resource for managing an AWS Chime SDK Media Pipelines Media Insights Pipeline Configuration.
Consult the [Call analytics developer guide](https://docs.aws.amazon.com/chime-sdk/latest/dg/call-analytics.html) for more detailed information about usage.

## Example Usage

### Basic Usage

```terraform
resource "aws_chimesdkmediapipelines_media_insights_pipeline_configuration" "my_configuration" {
  name                     = "MyBasicConfiguration"
  resource_access_role_arn = aws_iam_role.call_analytics_role.arn
  elements {
    type = "AmazonTranscribeCallAnalyticsProcessor"
    amazon_transcribe_call_analytics_processor_configuration {
      language_code = "en-US"
    }
  }

  elements {
    type = "KinesisDataStreamSink"
    kinesis_data_stream_sink_configuration {
      insights_target = aws_kinesis_stream.example.arn
    }
  }

  tags = {
    Key1 = "Value1"
    Key2 = "Value2"
  }
}

resource "aws_kinesis_stream" "example" {
  name        = "example"
  shard_count = 2
}

data "aws_iam_policy_document" "media_pipelines_assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["mediapipelines.chime.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "call_analytics_role" {
  name               = "CallAnalyticsRole"
  assume_role_policy = data.aws_iam_policy_document.media_pipelines_assume_role.json
}
```

- The required policies on `call_analytics_role` will vary based on the selected processors. See [Call analytics resource access role](https://docs.aws.amazon.com/chime-sdk/latest/dg/ca-resource-access-role.html) for directions on choosing appropriate policies.

### Transcribe Call Analytics processor usage

```terraform
resource "aws_chimesdkmediapipelines_media_insights_pipeline_configuration" "my_configuration" {
  name                     = "MyCallAnalyticsConfiguration"
  resource_access_role_arn = aws_iam_role.example.arn
  elements {
    type = "AmazonTranscribeCallAnalyticsProcessor"
    amazon_transcribe_call_analytics_processor_configuration {
      call_analytics_stream_categories = [
        "category_1",
        "category_2"
      ]
      content_redaction_type               = "PII"
      enable_partial_results_stabilization = true
      filter_partial_results               = true
      language_code                        = "en-US"
      language_model_name                  = "MyLanguageModel"
      partial_results_stability            = "high"
      pii_entity_types                     = "ADDRESS,BANK_ACCOUNT_NUMBER"
      post_call_analytics_settings {
        content_redaction_output     = "redacted"
        data_access_role_arn         = aws_iam_role.post_call_role.arn
        output_encryption_kms_key_id = "MyKmsKeyId"
        output_location              = "s3://MyBucket"
      }
      vocabulary_filter_method = "mask"
      vocabulary_filter_name   = "MyVocabularyFilter"
      vocabulary_name          = "MyVocabulary"
    }
  }

  elements {
    type = "KinesisDataStreamSink"
    kinesis_data_stream_sink_configuration {
      insights_target = aws_kinesis_stream.example.arn
    }
  }
}

data "aws_iam_policy_document" "transcribe_assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["transcribe.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "post_call_role" {
  name               = "PostCallAccessRole"
  assume_role_policy = data.aws_iam_policy_document.transcribe_assume_role.json
}
```

### Real time alerts usage

```terraform
resource "aws_chimesdkmediapipelines_media_insights_pipeline_configuration" "my_configuration" {
  name                     = "MyRealTimeAlertConfiguration"
  resource_access_role_arn = aws_iam_role.call_analytics_role.arn
  elements {
    type = "AmazonTranscribeCallAnalyticsProcessor"
    amazon_transcribe_call_analytics_processor_configuration {
      language_code = "en-US"
    }
  }

  elements {
    type = "KinesisDataStreamSink"
    kinesis_data_stream_sink_configuration {
      insights_target = aws_kinesis_stream.example.arn
    }
  }

  real_time_alert_configuration {
    disabled = false

    rules {
      type = "IssueDetection"
      issue_detection_configuration {
        rule_name = "MyIssueDetectionRule"
      }
    }

    rules {
      type = "KeywordMatch"
      keyword_match_configuration {
        keywords  = ["keyword1", "keyword2"]
        negate    = false
        rule_name = "MyKeywordMatchRule"
      }
    }

    rules {
      type = "Sentiment"
      sentiment_configuration {
        rule_name      = "MySentimentRule"
        sentiment_type = "NEGATIVE"
        time_period    = 60
      }
    }
  }
}
```

### Transcribe processor usage

```terraform
resource "aws_chimesdkmediapipelines_media_insights_pipeline_configuration" "my_configuration" {
  name                     = "MyTranscribeConfiguration"
  resource_access_role_arn = aws_iam_role.example.arn
  elements {
    type = "AmazonTranscribeProcessor"
    amazon_transcribe_processor_configuration {
      content_identification_type          = "PII"
      enable_partial_results_stabilization = true
      filter_partial_results               = true
      language_code                        = "en-US"
      language_model_name                  = "MyLanguageModel"
      partial_results_stability            = "high"
      pii_entity_types                     = "ADDRESS,BANK_ACCOUNT_NUMBER"
      show_speaker_label                   = true
      vocabulary_filter_method             = "mask"
      vocabulary_filter_name               = "MyVocabularyFilter"
      vocabulary_name                      = "MyVocabulary"
    }
  }

  elements {
    type = "KinesisDataStreamSink"
    kinesis_data_stream_sink_configuration {
      insights_target = aws_kinesis_stream.example.arn
    }
  }
}
```

### Voice analytics processor usage

```terraform
resource "aws_chimesdkmediapipelines_media_insights_pipeline_configuration" "my_configuration" {
  name                     = "MyVoiceAnalyticsConfiguration"
  resource_access_role_arn = aws_iam_role.example.arn
  elements {
    type = "VoiceAnalyticsProcessor"
    voice_analytics_processor_configuration {
      speaker_search_status      = "Enabled"
      voice_tone_analysis_status = "Enabled"
    }
  }

  elements {
    type = "LambdaFunctionSink"
    lambda_function_sink_configuration {
      insights_target = "arn:aws:lambda:us-west-2:1111111111:function:MyFunction"
    }
  }

  elements {
    type = "SnsTopicSink"
    sns_topic_sink_configuration {
      insights_target = "arn:aws:sns:us-west-2:1111111111:topic/MyTopic"
    }
  }

  elements {
    type = "SqsQueueSink"
    sqs_queue_sink_configuration {
      insights_target = "arn:aws:sqs:us-west-2:1111111111:queue/MyQueue"
    }
  }

  elements {
    type = "KinesisDataStreamSink"
    kinesis_data_stream_sink_configuration {
      insights_target = aws_kinesis_stream.test.arn
    }
  }
}
```

### S3 Recording sink usage

```terraform
resource "aws_chimesdkmediapipelines_media_insights_pipeline_configuration" "my_configuration" {
  name                     = "MyS3RecordingConfiguration"
  resource_access_role_arn = aws_iam_role.example.arn
  elements {
    type = "S3RecordingSink"
    s3_recording_sink_configuration {
      destination = "arn:aws:s3:::MyBucket"
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) Configuration name.
* `resource_access_role_arn` - (Required) ARN of IAM Role used by service to invoke processors and sinks specified by configuration elements.
* `elements` - (Required) Collection of processors and sinks to transform media and deliver data.
* `real_time_alert_configuration` - (Optional) Configuration for real-time alert rules to send EventBridge notifications when certain conditions are met.
* `tags` - (Optional) Key-value map of tags for the resource.

### Elements

* `type` - (Required) Element type.
* `amazon_transcribe_call_analytics_processor_configuration` - (Optional) Configuration for Amazon Transcribe Call Analytics processor.
    * `call_analytics_stream_categories` - (Optional) Filter for category events to be delivered to insights target.
    * `content_identification_type` - (Optional) Labels all personally identifiable information (PII) identified in Utterance events.
    * `content_redaction_type` - (Optional) Redacts all personally identifiable information (PII) identified in Utterance events.
    * `enable_partial_results_stabilization` - (Optional) Enables partial result stabilization in Utterance events.
    * `filter_partial_results` - (Optional) Filters partial Utterance events from delivery to the insights target.
    * `language_code` - (Required) Language code for the transcription model.
    * `language_model_name` - (Optional) Name of custom language model for transcription.
    * `partial_results_stability` - (Optional) Level of stability to use when partial results stabilization is enabled.
    * `pii_entity_types` - (Optional) Types of personally identifiable information (PII) to redact from an Utterance event.
    * `post_call_analytics_settings` - (Optional) Settings for post call analytics.
        * `content_redaction_output` - (Optional) Should output be redacted.
        * `data_access_role_arn` - (Required) ARN of the role used by AWS Transcribe to upload your post call analysis.
        * `output_encryption_kms_key_id` - (Optional) ID of the KMS key used to encrypt the output.
        * `output_location` - (Required) The Amazon S3 location where you want your Call Analytics post-call transcription output stored.
    * `vocabulary_filter_method` - (Optional) Method for applying a vocabulary filter to Utterance events.
    * `vocabulary_filter_name` - (Optional) Name of the custom vocabulary filter to use when processing Utterance events.
    * `vocabulary_name` - (Optional) Name of the custom vocabulary to use when processing Utterance events.
* `amazon_transcribe_processor_configuration` - (Optional) Configuration for Amazon Transcribe processor.
    * `content_identification_type` - (Optional) Labels all personally identifiable information (PII) identified in Transcript events.
    * `content_redaction_type` - (Optional) Redacts all personally identifiable information (PII) identified in Transcript events.
    * `enable_partial_results_stabilization` - (Optional) Enables partial result stabilization in Transcript events.
    * `filter_partial_results` - (Optional) Filters partial Utterance events from delivery to the insights target.
    * `language_code` - (Required) Language code for the transcription model.
    * `language_model_name` - (Optional) Name of custom language model for transcription.
    * `partial_results_stability` - (Optional) Level of stability to use when partial results stabilization is enabled.
    * `pii_entity_types` - (Optional) Types of personally identifiable information (PII) to redact from a Transcript event.
    * `show_speaker_label` - (Optional) Enables speaker partitioning (diarization) in your Transcript events.
    * `vocabulary_filter_method` - (Optional) Method for applying a vocabulary filter to Transcript events.
    * `vocabulary_filter_name` - (Optional) Name of the custom vocabulary filter to use when processing Transcript events.
    * `vocabulary_name` - (Optional) Name of the custom vocabulary to use when processing Transcript events.
* `kinesis_data_stream_sink_configuration` - (Optional) Configuration for Kinesis Data Stream sink.
    * `insights_target` - (Required) Kinesis Data Stream to deliver results.
* `lambda_function_sink_configuration` - (Optional) Configuration for Lambda Function sink.
    * `insights_target` - (Required) Lambda Function to deliver results.
* `sns_topic_sink_configuration` - (Optional) Configuration for SNS Topic sink.
    * `insights_target` - (Required) SNS topic to deliver results.
* `sqs_queue_sink_configuration` - (Optional) Configuration for SQS Queue sink.
    * `insights_target` - (Required) SQS queue to deliver results.
* `s3_recording_sink_configuration` - (Optional) Configuration for S3 recording sink.
    * `destination` - (Required) S3 URI to deliver recordings.
* `voice_analytics_processor_configuration` - (Optional) Configuration for Voice analytics processor.
    * `speaker_search_status` - (Required) Enable speaker search.
    * `voice_tone_analysis_status` - (Required) Enable voice tone analysis.

### Real time alert configuration

* `rules` - (Required) Collection of real time alert rules
    * `type` - (Required) Rule type.
    * `issue_detection_configuration` - (Optional) Configuration for an issue detection rule.
        * `rule_name` - (Required) Rule name.
    * `keyword_match_configuration` - (Optional) Configuration for a keyword match rule.
        * `rule_name` - (Required) Rule name.
        * `keywords` - (Required) Collection of keywords to match.
        * `negate` - (Optional) Negate the rule.
    * `sentiment_configuration` - (Optional) Configuration for a sentiment rule.
        * `rule_name` - (Required) Rule name.
        * `sentiment_type` - (Required) Sentiment type to match.
        * `time_period` - (Optional) Analysis interval.
* `disabled` - (Optional) Disables real time alert rules.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Media Insights Pipeline Configuration.
* `id` - Unique ID of the Media Insights Pipeline Configuration.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `3m`)
* `update` - (Default `3m`)
* `delete` - (Default `30s`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Chime SDK Media Pipelines Media Insights Pipeline Configuration using the `id`. For example:

```terraform
import {
  to = aws_chimesdkmediapipelines_media_insights_pipeline_configuration.example
  id = "abcdef123456"
}
```

Using `terraform import`, import Chime SDK Media Pipelines Media Insights Pipeline Configuration using the `id`. For example:

```console
% terraform import aws_chimesdkmediapipelines_media_insights_pipeline_configuration.example abcdef123456
```
