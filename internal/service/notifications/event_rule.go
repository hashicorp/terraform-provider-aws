// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package notifications

import (
	"context"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/notifications"
	awstypes "github.com/aws/aws-sdk-go-v2/service/notifications/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/maps"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_notifications_event_rule", name="Event Rule")
func newEventRuleResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &eventRuleResource{}

	return r, nil
}

type eventRuleResource struct {
	framework.ResourceWithModel[eventRuleResourceModel]
}

func (r *eventRuleResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"event_pattern": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 4096),
				},
			},
			"event_type": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 128),
					stringvalidator.RegexMatches(regexache.MustCompile(`([a-zA-Z0-9 \-\(\)])+`), ""),
				},
			},
			"notification_configuration_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"regions": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				Required:    true,
				ElementType: types.StringType,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
			names.AttrSource: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 36),
					stringvalidator.RegexMatches(regexache.MustCompile(`aws.([a-z0-9\-])+`), ""),
				},
			},
		},
	}
}

func (r *eventRuleResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data eventRuleResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NotificationsClient(ctx)

	var input notifications.CreateEventRuleInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	output, err := conn.CreateEventRule(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError("creating User Notifications Event Rule", err.Error())

		return
	}

	// Set values for unknowns.
	arn := aws.ToString(output.Arn)
	data.ARN = fwflex.StringValueToFramework(ctx, arn)

	if _, err := waitEventRuleCreated(ctx, conn, arn); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for User Notifications Event Rule (%s) create", arn), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *eventRuleResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data eventRuleResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NotificationsClient(ctx)

	output, err := findEventRuleByARN(ctx, conn, data.ARN.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading User Notifications Event Rule (%s)", data.ARN.ValueString()), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *eventRuleResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old eventRuleResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NotificationsClient(ctx)

	if !new.EventPattern.Equal(old.EventPattern) || !new.Regions.Equal(old.Regions) {
		var input notifications.UpdateEventRuleInput
		response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateEventRule(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating User Notifications Event Rule (%s)", new.ARN.ValueString()), err.Error())

			return
		}

		if _, err := waitEventRuleUpdated(ctx, conn, new.ARN.ValueString()); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for User Notifications Event Rule (%s) update", new.ARN.ValueString()), err.Error())

			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *eventRuleResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data eventRuleResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NotificationsClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, data.ARN)
	input := notifications.DeleteEventRuleInput{
		Arn: aws.String(arn),
	}
	_, err := conn.DeleteEventRule(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting User Notifications Event Rule (%s)", arn), err.Error())

		return
	}

	if _, err := waitEventRuleDeleted(ctx, conn, arn); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for User Notifications Event Rule (%s) delete", arn), err.Error())

		return
	}
}

func (r *eventRuleResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrARN), request, response)
}

func findEventRuleByARN(ctx context.Context, conn *notifications.Client, arn string) (*notifications.GetEventRuleOutput, error) {
	input := notifications.GetEventRuleInput{
		Arn: aws.String(arn),
	}
	output, err := conn.GetEventRule(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Arn == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func statusEventRule(ctx context.Context, conn *notifications.Client, arn string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findEventRuleByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		allActive := true

		for _, v := range maps.Values(output.StatusSummaryByRegion) {
			switch status := v.Status; status {
			// If regions were added/deleted then rule status across regions can be a mix of "CREATING", "DELETING", "UPDATING"
			// Does not matter which is returned as any of these is valid for waitEventRuleUpdated implementation
			case awstypes.EventRuleStatusCreating,
				awstypes.EventRuleStatusUpdating,
				awstypes.EventRuleStatusDeleting:
				return output, string(status), nil
			case awstypes.EventRuleStatusInactive:
				allActive = false
			}
		}

		if allActive {
			return output, string(awstypes.EventRuleStatusActive), nil
		}

		return output, string(awstypes.EventRuleStatusInactive), nil
	}
}

func waitEventRuleCreated(ctx context.Context, conn *notifications.Client, arn string) (*notifications.GetEventRuleOutput, error) {
	const (
		timeout = 10 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.EventRuleStatusCreating),
		Target:                    enum.Slice(awstypes.EventRuleStatusActive, awstypes.EventRuleStatusInactive),
		Refresh:                   statusEventRule(ctx, conn, arn),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*notifications.GetEventRuleOutput); ok {
		return output, err
	}

	return nil, err
}

func waitEventRuleUpdated(ctx context.Context, conn *notifications.Client, id string) (*notifications.GetEventRuleOutput, error) {
	const (
		timeout = 10 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		// If regions were added/removed then rule status across regions can be a mix of "CREATING", "DELETING", "UPDATING"
		Pending:                   enum.Slice(awstypes.EventRuleStatusCreating, awstypes.EventRuleStatusUpdating, awstypes.EventRuleStatusDeleting),
		Target:                    enum.Slice(awstypes.EventRuleStatusActive, awstypes.EventRuleStatusInactive),
		Refresh:                   statusEventRule(ctx, conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*notifications.GetEventRuleOutput); ok {
		return output, err
	}

	return nil, err
}

func waitEventRuleDeleted(ctx context.Context, conn *notifications.Client, id string) (*notifications.GetEventRuleOutput, error) {
	const (
		timeout = 10 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.EventRuleStatusDeleting),
		Target:  []string{},
		Refresh: statusEventRule(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*notifications.GetEventRuleOutput); ok {
		return output, err
	}

	return nil, err
}

type eventRuleResourceModel struct {
	ARN                          types.String        `tfsdk:"arn"`
	EventPattern                 types.String        `tfsdk:"event_pattern"`
	EventType                    types.String        `tfsdk:"event_type"`
	NotificationConfigurationARN fwtypes.ARN         `tfsdk:"notification_configuration_arn"`
	Regions                      fwtypes.SetOfString `tfsdk:"regions"`
	Source                       types.String        `tfsdk:"source"`
}
