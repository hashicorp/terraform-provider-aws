// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package transcribe

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/transcribe"
	awstypes "github.com/aws/aws-sdk-go-v2/service/transcribe/types"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @Action(aws_transcribe_start_transcription_job, name="Start Transcription Job")
func newStartTranscriptionJobAction(_ context.Context) (action.ActionWithConfigure, error) {
	return &startTranscriptionJobAction{}, nil
}

var (
	_ action.Action = (*startTranscriptionJobAction)(nil)
)

type startTranscriptionJobAction struct {
	framework.ActionWithModel[startTranscriptionJobActionModel]
}

type startTranscriptionJobActionModel struct {
	framework.WithRegionModel
	TranscriptionJobName      types.String                              `tfsdk:"transcription_job_name"`
	MediaFileUri              types.String                              `tfsdk:"media_file_uri"`
	LanguageCode              fwtypes.StringEnum[awstypes.LanguageCode] `tfsdk:"language_code"`
	IdentifyLanguage          types.Bool                                `tfsdk:"identify_language"`
	IdentifyMultipleLanguages types.Bool                                `tfsdk:"identify_multiple_languages"`
	MediaFormat               fwtypes.StringEnum[awstypes.MediaFormat]  `tfsdk:"media_format"`
	MediaSampleRateHertz      types.Int64                               `tfsdk:"media_sample_rate_hertz"`
	OutputBucketName          types.String                              `tfsdk:"output_bucket_name"`
	OutputKey                 types.String                              `tfsdk:"output_key"`
	Timeout                   types.Int64                               `tfsdk:"timeout"`
}

func (a *startTranscriptionJobAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Starts an Amazon Transcribe transcription job to transcribe audio from a media file. The media file must be uploaded to an Amazon S3 bucket before starting the transcription job.",
		Attributes: map[string]schema.Attribute{
			"transcription_job_name": schema.StringAttribute{
				Description: "A unique name for the transcription job within your AWS account.",
				Required:    true,
			},
			"media_file_uri": schema.StringAttribute{
				Description: "The Amazon S3 location of the media file to transcribe (e.g., s3://bucket-name/file.mp3).",
				Required:    true,
			},
			names.AttrLanguageCode: schema.StringAttribute{
				CustomType:  fwtypes.StringEnumType[awstypes.LanguageCode](),
				Description: "The language code for the language used in the input media file. Required if identify_language and identify_multiple_languages are both false.",
				Optional:    true,
			},
			"identify_language": schema.BoolAttribute{
				Description: "Enable automatic language identification for single-language media files. Cannot be used with identify_multiple_languages.",
				Optional:    true,
			},
			"identify_multiple_languages": schema.BoolAttribute{
				Description: "Enable automatic language identification for multi-language media files. Cannot be used with identify_language.",
				Optional:    true,
			},
			"media_format": schema.StringAttribute{
				CustomType:  fwtypes.StringEnumType[awstypes.MediaFormat](),
				Description: "The format of the input media file. If not specified, Amazon Transcribe will attempt to determine the format automatically.",
				Optional:    true,
			},
			"media_sample_rate_hertz": schema.Int64Attribute{
				Description: "The sample rate of the input media file in Hertz. If not specified, Amazon Transcribe will attempt to determine the sample rate automatically.",
				Optional:    true,
			},
			"output_bucket_name": schema.StringAttribute{
				Description: "The name of the Amazon S3 bucket where you want your transcription output stored. If not specified, output is stored in a service-managed bucket.",
				Optional:    true,
			},
			"output_key": schema.StringAttribute{
				Description: "The Amazon S3 object key for your transcription output. If not specified, a default key is generated.",
				Optional:    true,
			},
			names.AttrTimeout: schema.Int64Attribute{
				Description: "Maximum time in seconds to wait for the transcription job to start. Defaults to 300 seconds (5 minutes).",
				Optional:    true,
			},
		},
	}
}

func (a *startTranscriptionJobAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var config startTranscriptionJobActionModel

	// Parse configuration
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get AWS client
	conn := a.Meta().TranscribeClient(ctx)

	transcriptionJobName := config.TranscriptionJobName.ValueString()
	mediaFileUri := config.MediaFileUri.ValueString()

	// Set default timeout
	timeout := 5 * time.Minute
	if !config.Timeout.IsNull() {
		timeout = time.Duration(config.Timeout.ValueInt64()) * time.Second
	}

	tflog.Info(ctx, "Starting transcription job action", map[string]any{
		"transcription_job_name": transcriptionJobName,
		"media_file_uri":         mediaFileUri,
		"timeout_seconds":        int64(timeout.Seconds()),
	})

	// Send initial progress update
	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("Starting transcription job %s...", transcriptionJobName),
	})

	// Build the start transcription job input
	input := &transcribe.StartTranscriptionJobInput{
		TranscriptionJobName: aws.String(transcriptionJobName),
		Media: &awstypes.Media{
			MediaFileUri: aws.String(mediaFileUri),
		},
	}

	// Validate language configuration
	hasLanguageCode := !config.LanguageCode.IsNull() && !config.LanguageCode.IsUnknown()
	hasIdentifyLanguage := !config.IdentifyLanguage.IsNull() && config.IdentifyLanguage.ValueBool()
	hasIdentifyMultipleLanguages := !config.IdentifyMultipleLanguages.IsNull() && config.IdentifyMultipleLanguages.ValueBool()

	languageConfigCount := 0
	if hasLanguageCode {
		languageConfigCount++
	}
	if hasIdentifyLanguage {
		languageConfigCount++
	}
	if hasIdentifyMultipleLanguages {
		languageConfigCount++
	}

	if languageConfigCount == 0 {
		resp.Diagnostics.AddError(
			"Missing Language Configuration",
			"You must specify exactly one of: language_code, identify_language, or identify_multiple_languages",
		)
		return
	}

	if languageConfigCount > 1 {
		resp.Diagnostics.AddError(
			"Conflicting Language Configuration",
			"You can only specify one of: language_code, identify_language, or identify_multiple_languages",
		)
		return
	}

	// Set language configuration
	if hasLanguageCode {
		input.LanguageCode = config.LanguageCode.ValueEnum()
	}
	if hasIdentifyLanguage {
		input.IdentifyLanguage = aws.Bool(true)
	}
	if hasIdentifyMultipleLanguages {
		input.IdentifyMultipleLanguages = aws.Bool(true)
	}

	// Set optional parameters
	if !config.MediaFormat.IsNull() && !config.MediaFormat.IsUnknown() {
		input.MediaFormat = config.MediaFormat.ValueEnum()
	}

	if !config.MediaSampleRateHertz.IsNull() {
		input.MediaSampleRateHertz = aws.Int32(int32(config.MediaSampleRateHertz.ValueInt64()))
	}

	if !config.OutputBucketName.IsNull() {
		input.OutputBucketName = config.OutputBucketName.ValueStringPointer()
	}

	if !config.OutputKey.IsNull() {
		input.OutputKey = config.OutputKey.ValueStringPointer()
	}

	// Start the transcription job
	_, err := conn.StartTranscriptionJob(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Start Transcription Job",
			fmt.Sprintf("Could not start transcription job %s: %s", transcriptionJobName, err),
		)
		return
	}

	// Wait for job to be in progress or completed
	deadline := time.Now().Add(timeout)
	pollInterval := 5 * time.Second
	progressInterval := 30 * time.Second
	lastProgressUpdate := time.Now()

	for {
		// Check context cancellation
		select {
		case <-ctx.Done():
			resp.Diagnostics.AddError(
				"Context Cancelled",
				"Transcription job start operation was cancelled",
			)
			return
		default:
		}

		// Check timeout
		if time.Now().After(deadline) {
			resp.Diagnostics.AddError(
				"Timeout Waiting for Transcription Job",
				fmt.Sprintf("Transcription job %s did not start within %v", transcriptionJobName, timeout),
			)
			return
		}

		// Get job status
		getInput := &transcribe.GetTranscriptionJobInput{
			TranscriptionJobName: aws.String(transcriptionJobName),
		}

		getOutput, err := conn.GetTranscriptionJob(ctx, getInput)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to Get Transcription Job Status",
				fmt.Sprintf("Could not get status for transcription job %s: %s", transcriptionJobName, err),
			)
			return
		}

		if getOutput.TranscriptionJob == nil {
			resp.Diagnostics.AddError(
				"Transcription Job Not Found",
				fmt.Sprintf("Transcription job %s was not found", transcriptionJobName),
			)
			return
		}

		status := getOutput.TranscriptionJob.TranscriptionJobStatus

		// Send progress updates periodically
		if time.Since(lastProgressUpdate) >= progressInterval {
			resp.SendProgress(action.InvokeProgressEvent{
				Message: fmt.Sprintf("Transcription job %s is currently %s", transcriptionJobName, string(status)),
			})
			lastProgressUpdate = time.Now()
		}

		// Check if job has started successfully
		switch status {
		case awstypes.TranscriptionJobStatusInProgress, awstypes.TranscriptionJobStatusCompleted:
			// Job has started successfully
			resp.SendProgress(action.InvokeProgressEvent{
				Message: fmt.Sprintf("Transcription job %s started successfully and is %s", transcriptionJobName, string(status)),
			})

			tflog.Info(ctx, "Transcription job started successfully", map[string]any{
				"transcription_job_name": transcriptionJobName,
				"job_status":             string(status),
				names.AttrCreationTime:   getOutput.TranscriptionJob.CreationTime,
			})
			return

		case awstypes.TranscriptionJobStatusFailed:
			failureReason := ""
			if getOutput.TranscriptionJob.FailureReason != nil {
				failureReason = aws.ToString(getOutput.TranscriptionJob.FailureReason)
			}
			resp.Diagnostics.AddError(
				"Transcription Job Failed",
				fmt.Sprintf("Transcription job %s failed: %s", transcriptionJobName, failureReason),
			)
			return

		case awstypes.TranscriptionJobStatusQueued:
			// Job is still queued, continue waiting
			time.Sleep(pollInterval)
			continue

		default:
			resp.Diagnostics.AddError(
				"Unexpected Transcription Job Status",
				fmt.Sprintf("Transcription job %s has unexpected status: %s", transcriptionJobName, string(status)),
			)
			return
		}
	}
}
