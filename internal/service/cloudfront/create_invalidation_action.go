// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/actionwait"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @Action(aws_cloudfront_create_invalidation, name="Create Invalidation")
func newCreateInvalidationAction(_ context.Context) (action.ActionWithConfigure, error) {
	return &createInvalidationAction{}, nil
}

var (
	_ action.Action = (*createInvalidationAction)(nil)
)

type createInvalidationAction struct {
	framework.ActionWithModel[createInvalidationModel]
}

type createInvalidationModel struct {
	DistributionID  types.String         `tfsdk:"distribution_id"`
	Paths           fwtypes.ListOfString `tfsdk:"paths"`
	CallerReference types.String         `tfsdk:"caller_reference"`
	Timeout         types.Int64          `tfsdk:"timeout"`
}

func (a *createInvalidationAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Invalidates CloudFront distribution cache for specified paths. This action creates an invalidation request and waits for it to complete.",
		Attributes: map[string]schema.Attribute{
			"distribution_id": schema.StringAttribute{
				Description: "The ID of the CloudFront distribution to invalidate cache for",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexache.MustCompile(`^[A-Z0-9]+$`),
						"must be a valid CloudFront distribution ID (e.g., E1GHKQ2EXAMPLE)",
					),
				},
			},
			"paths": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				Description: "List of file paths or patterns to invalidate. Use /* to invalidate all files",
				Required:    true,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(3000), // CloudFront limit
				},
			},
			"caller_reference": schema.StringAttribute{
				Description: "Unique identifier for the invalidation request. If not provided, one will be generated automatically",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(128),
				},
			},
			names.AttrTimeout: schema.Int64Attribute{
				Description: "Timeout in seconds to wait for the invalidation to complete (default: 900)",
				Optional:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(60),
					int64validator.AtMost(3600),
				},
			},
		},
	}
}

func (a *createInvalidationAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var config createInvalidationModel

	// Parse configuration
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get AWS client
	conn := a.Meta().CloudFrontClient(ctx)

	distributionID := config.DistributionID.ValueString()

	// Convert paths list to string slice
	var paths []string
	resp.Diagnostics.Append(config.Paths.ElementsAs(ctx, &paths, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate paths
	for _, path := range paths {
		if path == "" {
			resp.Diagnostics.AddError(
				"Invalid Path",
				"Path cannot be empty",
			)
			return
		}
		if !regexache.MustCompile(`^(/.*|\*)$`).MatchString(path) {
			resp.Diagnostics.AddError(
				"Invalid Path Format",
				fmt.Sprintf("Path '%s' must start with '/' or be '*' for all files", path),
			)
			return
		}
	}

	// Set caller reference if not provided
	callerReference := config.CallerReference.ValueString()
	if callerReference == "" {
		callerReference = sdkid.UniqueId()
	}

	// Set default timeout if not provided
	timeout := 900 * time.Second
	if !config.Timeout.IsNull() {
		timeout = time.Duration(config.Timeout.ValueInt64()) * time.Second
	}

	tflog.Info(ctx, "Starting CloudFront cache invalidation action", map[string]any{
		"distribution_id":  distributionID,
		"paths":            paths,
		"caller_reference": callerReference,
		names.AttrTimeout:  timeout.String(),
	})

	// Send initial progress update
	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("Starting cache invalidation for CloudFront distribution %s...", distributionID),
	})

	// Check if distribution exists first
	_, err := findDistributionByID(ctx, conn, distributionID)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, "NoSuchDistribution") {
			resp.Diagnostics.AddError(
				"Distribution Not Found",
				fmt.Sprintf("CloudFront distribution %s was not found", distributionID),
			)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to Describe Distribution",
			fmt.Sprintf("Could not describe CloudFront distribution %s: %s", distributionID, err),
		)
		return
	}

	// Create invalidation request
	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("Creating invalidation request for %d path(s)...", len(paths)),
	})

	invalidationInput := &cloudfront.CreateInvalidationInput{
		DistributionId: aws.String(distributionID),
		InvalidationBatch: &awstypes.InvalidationBatch{
			CallerReference: aws.String(callerReference),
			Paths: &awstypes.Paths{
				Quantity: aws.Int32(int32(len(paths))),
				Items:    paths,
			},
		},
	}

	output, err := conn.CreateInvalidation(ctx, invalidationInput)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, "TooManyInvalidationsInProgress") {
			resp.Diagnostics.AddError(
				"Too Many Invalidations In Progress",
				fmt.Sprintf("CloudFront distribution %s has too many invalidations in progress. Please wait and try again.", distributionID),
			)
			return
		}
		if tfawserr.ErrCodeEquals(err, "InvalidArgument") {
			resp.Diagnostics.AddError(
				"Invalid Invalidation Request",
				fmt.Sprintf("Invalid invalidation request for distribution %s: %s", distributionID, err),
			)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to Create Invalidation",
			fmt.Sprintf("Could not create invalidation for CloudFront distribution %s: %s", distributionID, err),
		)
		return
	}

	invalidationID := aws.ToString(output.Invalidation.Id)

	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("Invalidation %s created, waiting for completion...", invalidationID),
	})

	// Wait for invalidation to complete with periodic progress updates using actionwait
	// Use fixed interval since CloudFront invalidations have predictable timing and
	// don't benefit from exponential backoff - status changes are infrequent and consistent
	_, err = actionwait.WaitForStatus(ctx, func(ctx context.Context) (actionwait.FetchResult[struct{}], error) {
		input := cloudfront.GetInvalidationInput{
			DistributionId: aws.String(distributionID),
			Id:             aws.String(invalidationID),
		}
		output, gerr := conn.GetInvalidation(ctx, &input)
		if gerr != nil {
			return actionwait.FetchResult[struct{}]{}, fmt.Errorf("getting invalidation status: %w", gerr)
		}
		status := aws.ToString(output.Invalidation.Status)
		return actionwait.FetchResult[struct{}]{Status: actionwait.Status(status)}, nil
	}, actionwait.Options[struct{}]{
		Timeout:          timeout,
		Interval:         actionwait.FixedInterval(actionwait.DefaultPollInterval),
		ProgressInterval: 60 * time.Second,
		SuccessStates:    []actionwait.Status{"Completed"},
		TransitionalStates: []actionwait.Status{
			"InProgress",
		},
		ProgressSink: func(fr actionwait.FetchResult[any], meta actionwait.ProgressMeta) {
			resp.SendProgress(action.InvokeProgressEvent{Message: fmt.Sprintf("Invalidation %s is currently '%s', continuing to wait for completion...", invalidationID, fr.Status)})
		},
	})
	if err != nil {
		var timeoutErr *actionwait.TimeoutError
		var unexpectedErr *actionwait.UnexpectedStateError
		if errors.As(err, &timeoutErr) {
			resp.Diagnostics.AddError(
				"Timeout Waiting for Invalidation to Complete",
				fmt.Sprintf("CloudFront invalidation %s did not complete within %s: %s", invalidationID, timeout, err),
			)
		} else if errors.As(err, &unexpectedErr) {
			resp.Diagnostics.AddError(
				"Invalid Invalidation State",
				fmt.Sprintf("CloudFront invalidation %s entered unexpected state: %s", invalidationID, err),
			)
		} else {
			resp.Diagnostics.AddError(
				"Failed While Waiting for Invalidation",
				fmt.Sprintf("Error waiting for CloudFront invalidation %s: %s", invalidationID, err),
			)
		}
		return
	}

	// Final success message
	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("CloudFront cache invalidation %s completed successfully for distribution %s", invalidationID, distributionID),
	})

	tflog.Info(ctx, "CloudFront invalidate cache action completed successfully", map[string]any{
		"distribution_id": distributionID,
		"invalidation_id": invalidationID,
		"paths":           paths,
	})
}
