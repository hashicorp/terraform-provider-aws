---
subcategory: "Transcribe"
layout: "aws"
page_title: "AWS: aws_transcribe_start_transcription_job"
description: |-
  Starts an Amazon Transcribe transcription job.
---

# Action: aws_transcribe_start_transcription_job

~> **Note:** `aws_transcribe_start_transcription_job` is in beta. Its interface and behavior may change as the feature evolves, and breaking changes are possible. It is offered as a technical preview without compatibility guarantees until Terraform 1.14 is generally available.

Starts an Amazon Transcribe transcription job to transcribe audio from a media file. The media file must be uploaded to an Amazon S3 bucket before starting the transcription job.

For information about Amazon Transcribe, see the [Amazon Transcribe Developer Guide](https://docs.aws.amazon.com/transcribe/latest/dg/). For specific information about starting transcription jobs, see the [StartTranscriptionJob](https://docs.aws.amazon.com/transcribe/latest/APIReference/API_StartTranscriptionJob.html) page in the Amazon Transcribe API Reference.

~> **Note:** This action starts the transcription job and waits for it to begin processing, but does not wait for the transcription to complete. The job will continue running asynchronously after the action completes.

## Example Usage

### Basic Usage

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "my-transcription-bucket"
}

resource "aws_s3_object" "audio" {
  bucket = aws_s3_bucket.example.bucket
  key    = "audio/meeting.mp3"
  source = "path/to/meeting.mp3"
}

action "aws_transcribe_start_transcription_job" "example" {
  config {
    transcription_job_name = "meeting-transcription-${timestamp()}"
    media_file_uri         = "s3://${aws_s3_bucket.example.bucket}/${aws_s3_object.audio.key}"
    language_code          = "en-US"
  }
}
```

### Automatic Language Detection

```terraform
action "aws_transcribe_start_transcription_job" "auto_detect" {
  config {
    transcription_job_name = "auto-detect-transcription"
    media_file_uri         = "s3://my-bucket/audio/multilingual-meeting.mp3"
    identify_language      = true
    timeout                = 600
  }
}
```

### Multiple Language Detection

```terraform
action "aws_transcribe_start_transcription_job" "multilingual" {
  config {
    transcription_job_name      = "multilingual-transcription"
    media_file_uri              = "s3://my-bucket/audio/conference-call.mp3"
    identify_multiple_languages = true
    media_format                = "mp3"
    media_sample_rate_hertz     = 44100
  }
}
```

### Custom Output Location

```terraform
action "aws_transcribe_start_transcription_job" "custom_output" {
  config {
    transcription_job_name = "custom-output-transcription"
    media_file_uri         = "s3://my-bucket/audio/interview.wav"
    language_code          = "en-US"
    output_bucket_name     = aws_s3_bucket.transcripts.bucket
    output_key             = "transcripts/interview-transcript.json"
  }
}
```

### CI/CD Pipeline Integration

```terraform
resource "terraform_data" "process_audio" {
  input = var.audio_files

  lifecycle {
    action_trigger {
      events  = [before_create, before_update]
      actions = [action.aws_transcribe_start_transcription_job.batch_process]
    }
  }

  depends_on = [aws_s3_object.uploaded_audio]
}

action "aws_transcribe_start_transcription_job" "batch_process" {
  config {
    transcription_job_name = "batch-${formatdate("YYYY-MM-DD-hhmm", timestamp())}"
    media_file_uri         = "s3://${aws_s3_bucket.audio.bucket}/${aws_s3_object.uploaded_audio.key}"
    language_code          = var.audio_language
    timeout                = 900
  }
}
```

## Argument Reference

This action supports the following arguments:

* `transcription_job_name` - (Required) Unique name for the transcription job within your AWS account. Must be 1-200 characters and contain only alphanumeric characters, hyphens, periods, and underscores.
* `media_file_uri` - (Required) S3 location of the media file to transcribe (e.g., `s3://bucket-name/file.mp3`). The file must be accessible to Amazon Transcribe.
* `language_code` - (Optional) Language code for the language used in the input media file. Required if `identify_language` and `identify_multiple_languages` are both false. Valid values can be found in the [Amazon Transcribe supported languages documentation](https://docs.aws.amazon.com/transcribe/latest/dg/supported-languages.html).
* `identify_language` - (Optional) Enable automatic language identification for single-language media files. Cannot be used with `identify_multiple_languages`. Default: `false`.
* `identify_multiple_languages` - (Optional) Enable automatic language identification for multi-language media files. Cannot be used with `identify_language`. Default: `false`.
* `media_format` - (Optional) Format of the input media file. If not specified, Amazon Transcribe will attempt to determine the format automatically. Valid values: `mp3`, `mp4`, `wav`, `flac`, `ogg`, `amr`, `webm`, `m4a`.
* `media_sample_rate_hertz` - (Optional) Sample rate of the input media file in Hertz. If not specified, Amazon Transcribe will attempt to determine the sample rate automatically. Valid range: 8000-48000.
* `output_bucket_name` - (Optional) Name of the S3 bucket where you want your transcription output stored. If not specified, output is stored in a service-managed bucket.
* `output_key` - (Optional) S3 object key for your transcription output. If not specified, a default key is generated.
* `region` - (Optional) Region where this action should be [run](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `timeout` - (Optional) Maximum time in seconds to wait for the transcription job to start. Must be between 60 and 3600 seconds. Default: `300`.
