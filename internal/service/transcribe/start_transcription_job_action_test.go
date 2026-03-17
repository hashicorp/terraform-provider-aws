// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package transcribe_test

import (
	"context"
	"fmt"
	"slices"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/transcribe"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccTranscribeStartTranscriptionJobAction_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	bucketName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TranscribeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStartTranscriptionJobActionConfig_basic(rName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTranscriptionJobExists(ctx, t, rName),
					testAccCheckTranscriptionJobStatus(ctx, t, rName, "IN_PROGRESS", "COMPLETED"),
				),
			},
		},
	})
}

func TestAccTranscribeStartTranscriptionJobAction_identifyLanguage(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	bucketName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TranscribeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStartTranscriptionJobActionConfig_identifyLanguage(rName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTranscriptionJobExists(ctx, t, rName),
					testAccCheckTranscriptionJobStatus(ctx, t, rName, "IN_PROGRESS", "COMPLETED"),
					testAccCheckTranscriptionJobIdentifyLanguage(ctx, t, rName, true),
				),
			},
		},
	})
}

func TestAccTranscribeStartTranscriptionJobAction_withOutputLocation(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	bucketName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TranscribeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStartTranscriptionJobActionConfig_withOutputLocation(rName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTranscriptionJobExists(ctx, t, rName),
					testAccCheckTranscriptionJobStatus(ctx, t, rName, "IN_PROGRESS", "COMPLETED"),
				),
			},
		},
	})
}

func testAccCheckTranscriptionJobExists(ctx context.Context, t *testing.T, jobName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).TranscribeClient(ctx)

		input := &transcribe.GetTranscriptionJobInput{
			TranscriptionJobName: &jobName,
		}

		_, err := conn.GetTranscriptionJob(ctx, input)
		if err != nil {
			return fmt.Errorf("transcription job %s not found: %w", jobName, err)
		}

		return nil
	}
}

func testAccCheckTranscriptionJobStatus(ctx context.Context, t *testing.T, jobName string, expectedStatuses ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).TranscribeClient(ctx)

		input := &transcribe.GetTranscriptionJobInput{
			TranscriptionJobName: &jobName,
		}

		output, err := conn.GetTranscriptionJob(ctx, input)
		if err != nil {
			return fmt.Errorf("error getting transcription job %s: %w", jobName, err)
		}

		if output.TranscriptionJob == nil {
			return fmt.Errorf("transcription job %s not found", jobName)
		}

		actualStatus := string(output.TranscriptionJob.TranscriptionJobStatus)
		if slices.Contains(expectedStatuses, actualStatus) {
			return nil
		}

		return fmt.Errorf("expected transcription job %s status to be one of %v, got %s", jobName, expectedStatuses, actualStatus)
	}
}

func testAccCheckTranscriptionJobIdentifyLanguage(ctx context.Context, t *testing.T, jobName string, expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).TranscribeClient(ctx)

		input := &transcribe.GetTranscriptionJobInput{
			TranscriptionJobName: &jobName,
		}

		output, err := conn.GetTranscriptionJob(ctx, input)
		if err != nil {
			return fmt.Errorf("error getting transcription job %s: %w", jobName, err)
		}

		if output.TranscriptionJob == nil {
			return fmt.Errorf("transcription job %s not found", jobName)
		}

		actual := output.TranscriptionJob.IdentifyLanguage != nil && *output.TranscriptionJob.IdentifyLanguage
		if actual != expected {
			return fmt.Errorf("expected transcription job %s identify_language to be %t, got %t", jobName, expected, actual)
		}

		return nil
	}
}

func testAccStartTranscriptionJobActionConfig_basic(rName, bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[2]q
  force_destroy = true
}

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.bucket
  key    = "test-audio.wav"
  source = "test-fixtures/test-audio.wav"
}

action "aws_transcribe_start_transcription_job" "test" {
  config {
    transcription_job_name = %[1]q
    media_file_uri         = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.test.key}"
    language_code          = "en-US"
    timeout                = 600
  }
}

resource "terraform_data" "test" {
  triggers_replace = [
    aws_s3_object.test.etag
  ]

  input = "completed"

  lifecycle {
    action_trigger {
      events  = [before_create, before_update]
      actions = [action.aws_transcribe_start_transcription_job.test]
    }
  }
}
`, rName, bucketName)
}

func testAccStartTranscriptionJobActionConfig_identifyLanguage(rName, bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[2]q
  force_destroy = true
}

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.bucket
  key    = "test-audio.wav"
  source = "test-fixtures/test-audio.wav"
}

action "aws_transcribe_start_transcription_job" "test" {
  config {
    transcription_job_name = %[1]q
    media_file_uri         = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.test.key}"
    identify_language      = true
    timeout                = 600
  }
}

resource "terraform_data" "test" {
  triggers_replace = [
    aws_s3_object.test.etag
  ]

  input = "completed"

  lifecycle {
    action_trigger {
      events  = [before_create, before_update]
      actions = [action.aws_transcribe_start_transcription_job.test]
    }
  }
}
`, rName, bucketName)
}

func testAccStartTranscriptionJobActionConfig_withOutputLocation(rName, bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[2]q
  force_destroy = true
}

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.bucket
  key    = "test-audio.wav"
  source = "test-fixtures/test-audio.wav"
}

action "aws_transcribe_start_transcription_job" "test" {
  config {
    transcription_job_name = %[1]q
    media_file_uri         = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.test.key}"
    language_code          = "en-US"
    output_bucket_name     = aws_s3_bucket.test.bucket
    output_key             = "transcripts/%[1]s.json"
    timeout                = 600
  }
}

resource "terraform_data" "test" {
  triggers_replace = [
    aws_s3_object.test.etag
  ]

  input = "completed"

  lifecycle {
    action_trigger {
      events  = [before_create, before_update]
      actions = [action.aws_transcribe_start_transcription_job.test]
    }
  }
}
`, rName, bucketName)
}
