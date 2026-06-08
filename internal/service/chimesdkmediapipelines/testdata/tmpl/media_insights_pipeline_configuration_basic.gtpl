resource "aws_chimesdkmediapipelines_media_insights_pipeline_configuration" "test" {
{{- template "region" }}
  name                     = var.rName
  resource_access_role_arn = aws_iam_role.test.arn
  elements {
    type = "AmazonTranscribeCallAnalyticsProcessor"
    amazon_transcribe_call_analytics_processor_configuration {
      language_code = "en-US"
    }
  }

  elements {
    type = "KinesisDataStreamSink"
    kinesis_data_stream_sink_configuration {
      insights_target = aws_kinesis_stream.test.arn
    }
  }
{{- template "tags" . }}
}

# testAccMediaInsightsPipelineConfigurationConfigBase

data "aws_iam_policy_document" "assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["mediapipelines.chime.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "test" {
  name               = var.rName
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_kinesis_stream" "test" {
{{- template "region" }}
  name        = var.rName
  shard_count = 2
}
