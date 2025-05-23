// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package notifications

import (
	"context"
	"errors"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/notifications"
	awstypes "github.com/aws/aws-sdk-go-v2/service/notifications/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_notifications_notification_configuration", name="Notification Configuration")
// @Tags(identifierAttribute="arn")
func newResourceNotificationConfiguration(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceNotificationConfiguration{}

	return r, nil
}

const (
	ResNameNotificationConfiguration = "Notification Configuration"
)

type resourceNotificationConfiguration struct {
	framework.ResourceWithConfigure
}

func (r *resourceNotificationConfiguration) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"aggregation_duration": schema.StringAttribute{
				Optional:   true,
				Computed:   true,
				CustomType: fwtypes.StringEnumType[awstypes.AggregationDuration](),
				Default:    stringdefault.StaticString(string(awstypes.AggregationDurationNone)),
			},
			names.AttrDescription: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 256),
					stringvalidator.RegexMatches(regexache.MustCompile(`[^\x01-\x1F\x7F-\x9F]*`), ""),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 64),
					stringvalidator.RegexMatches(regexache.MustCompile(`[A-Za-z0-9_\-]+`), ""),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (r *resourceNotificationConfiguration) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().NotificationsClient(ctx)

	var plan resourceNotificationConfigurationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input notifications.CreateNotificationConfigurationInput
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input.Tags = getTagsIn(ctx)

	out, err := conn.CreateNotificationConfiguration(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Notifications, create.ErrActionCreating, ResNameNotificationConfiguration, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.Arn == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Notifications, create.ErrActionCreating, ResNameNotificationConfiguration, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceNotificationConfiguration) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().NotificationsClient(ctx)

	var state resourceNotificationConfigurationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findNotificationConfigurationByARN(ctx, conn, state.ARN.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Notifications, create.ErrActionReading, ResNameNotificationConfiguration, state.ARN.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceNotificationConfiguration) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().NotificationsClient(ctx)

	var plan, state resourceNotificationConfigurationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, plan, state)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input notifications.UpdateNotificationConfigurationInput
		resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.UpdateNotificationConfiguration(ctx, &input)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Notifications, create.ErrActionUpdating, ResNameNotificationConfiguration, plan.ARN.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Notifications, create.ErrActionUpdating, ResNameNotificationConfiguration, plan.ARN.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceNotificationConfiguration) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().NotificationsClient(ctx)

	var state resourceNotificationConfigurationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := notifications.DeleteNotificationConfigurationInput{
		Arn: state.ARN.ValueStringPointer(),
	}

	_, err := conn.DeleteNotificationConfiguration(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Notifications, create.ErrActionDeleting, ResNameNotificationConfiguration, state.ARN.String(), err),
			err.Error(),
		)
		return
	}

	_, err = waitNotificationConfigurationDeleted(ctx, conn, state.ARN.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Notifications, create.ErrActionWaitingForDeletion, ResNameNotificationConfiguration, state.ARN.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceNotificationConfiguration) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrARN), req, resp)
}

func waitNotificationConfigurationDeleted(ctx context.Context, conn *notifications.Client, id string) (*awstypes.NotificationConfigurationStructure, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.NotificationConfigurationStatusDeleting),
		Target:  []string{},
		Refresh: statusNotificationConfiguration(ctx, conn, id),
		Timeout: 10 * time.Minute,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.NotificationConfigurationStructure); ok {
		return out, err
	}

	return nil, err
}

func statusNotificationConfiguration(ctx context.Context, conn *notifications.Client, arn string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := findNotificationConfigurationByARN(ctx, conn, arn)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func findNotificationConfigurationByARN(ctx context.Context, conn *notifications.Client, id string) (*notifications.GetNotificationConfigurationOutput, error) {
	input := notifications.GetNotificationConfigurationInput{
		Arn: aws.String(id),
	}

	out, err := conn.GetNotificationConfiguration(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(&input)
	}

	return out, nil
}

type resourceNotificationConfigurationModel struct {
	ARN                 types.String                                     `tfsdk:"arn"`
	AggregationDuration fwtypes.StringEnum[awstypes.AggregationDuration] `tfsdk:"aggregation_duration"`
	Description         types.String                                     `tfsdk:"description"`
	Name                types.String                                     `tfsdk:"name"`
	Tags                tftags.Map                                       `tfsdk:"tags"`
	TagsAll             tftags.Map                                       `tfsdk:"tags_all"`
}

func sweepNotificationConfigurations(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := notifications.ListNotificationConfigurationsInput{}
	conn := client.NotificationsClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := notifications.NewListNotificationConfigurationsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.NotificationConfigurations {
			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newResourceNotificationConfiguration, client,
				sweepfw.NewAttribute(names.AttrARN, aws.ToString(v.Arn))),
			)
		}
	}

	return sweepResources, nil
}
