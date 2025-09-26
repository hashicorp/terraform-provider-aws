// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package transcribe

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/transcribe"
	awstypes "github.com/aws/aws-sdk-go-v2/service/transcribe/types"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/actionwait"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	transcriptionJobPollInterval     = 5 * time.Second
	transcriptionJobProgressInterval = 30 * time.Second
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

	// Validate language configuration - exactly one must be specified
	languageOptions := []bool{
		!config.LanguageCode.IsNull() && !config.LanguageCode.IsUnknown(),
		!config.IdentifyLanguage.IsNull() && config.IdentifyLanguage.ValueBool(),
		!config.IdentifyMultipleLanguages.IsNull() && config.IdentifyMultipleLanguages.ValueBool(),
	}

	activeCount := 0
	for _, active := range languageOptions {
		if active {
			activeCount++
		}
	}

	switch activeCount {
	case 0:
		resp.Diagnostics.AddError(
			"Missing Language Configuration",
			"You must specify exactly one of: language_code, identify_language, or identify_multiple_languages",
		)
		return
	case 1:
		// Valid - continue
	default:
		resp.Diagnostics.AddError(
			"Conflicting Language Configuration",
			"You can only specify one of: language_code, identify_language, or identify_multiple_languages",
		)
		return
	}

	// Set language configuration
	if languageOptions[0] {
		input.LanguageCode = config.LanguageCode.ValueEnum()
	}
	if languageOptions[1] {
		input.IdentifyLanguage = aws.Bool(true)
	}
	if languageOptions[2] {
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

	// Wait for job to move beyond QUEUED: treat IN_PROGRESS or COMPLETED as success, FAILED as failure, QUEUED transitional.
	fr, err := actionwait.WaitForStatus(ctx, func(ctx context.Context) (actionwait.FetchResult[*awstypes.TranscriptionJob], error) {
		input := transcribe.GetTranscriptionJobInput{TranscriptionJobName: aws.String(transcriptionJobName)}
		getOutput, gerr := conn.GetTranscriptionJob(ctx, &input)
		if gerr != nil {
			return actionwait.FetchResult[*awstypes.TranscriptionJob]{}, fmt.Errorf("get transcription job: %w", gerr)
		}
		if getOutput.TranscriptionJob == nil {
			return actionwait.FetchResult[*awstypes.TranscriptionJob]{}, fmt.Errorf("transcription job %s not found", transcriptionJobName)
		}
		status := getOutput.TranscriptionJob.TranscriptionJobStatus
		return actionwait.FetchResult[*awstypes.TranscriptionJob]{Status: actionwait.Status(status), Value: getOutput.TranscriptionJob}, nil
	}, actionwait.Options[*awstypes.TranscriptionJob]{
		Timeout:          timeout,
		Interval:         actionwait.FixedInterval(transcriptionJobPollInterval),
		ProgressInterval: transcriptionJobProgressInterval,
		SuccessStates: []actionwait.Status{
			actionwait.Status(awstypes.TranscriptionJobStatusInProgress),
			actionwait.Status(awstypes.TranscriptionJobStatusCompleted),
		},
		TransitionalStates: []actionwait.Status{
			actionwait.Status(awstypes.TranscriptionJobStatusQueued),
		},
		FailureStates: []actionwait.Status{
			actionwait.Status(awstypes.TranscriptionJobStatusFailed),
		},
		ProgressSink: func(fr actionwait.FetchResult[any], meta actionwait.ProgressMeta) {
			resp.SendProgress(action.InvokeProgressEvent{Message: fmt.Sprintf("Transcription job %s is currently %s", transcriptionJobName, fr.Status)})
		},
	})
	if err != nil {
		var timeoutErr *actionwait.TimeoutError
		var failureErr *actionwait.FailureStateError
		var unexpectedErr *actionwait.UnexpectedStateError

		if errors.As(err, &timeoutErr) {
			resp.Diagnostics.AddError(
				"Timeout Waiting for Transcription Job",
				fmt.Sprintf("Transcription job %s did not reach a running state within %v", transcriptionJobName, timeout),
			)
		} else if errors.As(err, &failureErr) {
			resp.Diagnostics.AddError(
				"Transcription Job Failed",
				fmt.Sprintf("Transcription job %s failed: %s", transcriptionJobName, failureErr.Status),
			)
		} else if errors.As(err, &unexpectedErr) {
			resp.Diagnostics.AddError(
				"Unexpected Transcription Job Status",
				fmt.Sprintf("Transcription job %s entered unexpected status: %s", transcriptionJobName, unexpectedErr.Status),
			)
		} else {
			resp.Diagnostics.AddError(
				"Error Waiting for Transcription Job",
				fmt.Sprintf("Error while waiting for transcription job %s: %s", transcriptionJobName, err),
			)
		}
		return
	}

	resp.SendProgress(action.InvokeProgressEvent{Message: fmt.Sprintf("Transcription job %s started successfully and is %s", transcriptionJobName, fr.Status)})
	logFields := map[string]any{
		"transcription_job_name": transcriptionJobName,
		"job_status":             fr.Status,
	}
	if fr.Value != nil {
		logFields[names.AttrCreationTime] = fr.Value.CreationTime
	}
	tflog.Info(ctx, "Transcription job started successfully", logFields)
}
