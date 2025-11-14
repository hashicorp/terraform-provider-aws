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
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_notifications_notification_configuration", name="Notification Configuration")
// @Tags(identifierAttribute="arn")
func newNotificationConfigurationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &notificationConfigurationResource{}

	return r, nil
}

type notificationConfigurationResource struct {
	framework.ResourceWithModel[notificationConfigurationResourceModel]
}

func (r *notificationConfigurationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"aggregation_duration": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.AggregationDuration](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
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

func (r *notificationConfigurationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data notificationConfigurationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NotificationsClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.Name)
	var input notifications.CreateNotificationConfigurationInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.Tags = getTagsIn(ctx)

	outputCNC, err := conn.CreateNotificationConfiguration(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating User Notifications Notification Configuration (%s)", name), err.Error())

		return
	}

	arn := aws.ToString(outputCNC.Arn)
	outputGNC, err := findNotificationConfigurationByARN(ctx, conn, arn)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading User Notifications Notification Configuration (%s)", arn), err.Error())

		return
	}

	// Set values for unknowns.
	data.AggregationDuration = fwtypes.StringEnumValue(outputGNC.AggregationDuration)
	data.ARN = fwflex.StringValueToFramework(ctx, arn)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *notificationConfigurationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data notificationConfigurationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NotificationsClient(ctx)

	out, err := findNotificationConfigurationByARN(ctx, conn, data.ARN.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading User Notifications Notification Configuration (%s)", data.ARN.ValueString()), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, out, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *notificationConfigurationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old notificationConfigurationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NotificationsClient(ctx)

	if !new.AggregationDuration.Equal(old.AggregationDuration) || !new.Description.Equal(old.Description) || !new.Name.Equal(old.Name) {
		var input notifications.UpdateNotificationConfigurationInput
		response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateNotificationConfiguration(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating User Notifications Notification Configuration (%s)", new.ARN.ValueString()), err.Error())

			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *notificationConfigurationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data notificationConfigurationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NotificationsClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, data.ARN)
	input := notifications.DeleteNotificationConfigurationInput{
		Arn: aws.String(arn),
	}
	_, err := conn.DeleteNotificationConfiguration(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting User Notifications Notification Configuration (%s)", arn), err.Error())

		return
	}

	if _, err := waitNotificationConfigurationDeleted(ctx, conn, arn); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for User Notifications Notification Configuration (%s) delete", arn), err.Error())

		return
	}
}

func (r *notificationConfigurationResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrARN), request, response)
}

func findNotificationConfigurationByARN(ctx context.Context, conn *notifications.Client, id string) (*notifications.GetNotificationConfigurationOutput, error) {
	input := notifications.GetNotificationConfigurationInput{
		Arn: aws.String(id),
	}
	output, err := conn.GetNotificationConfiguration(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: &input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func statusNotificationConfiguration(ctx context.Context, conn *notifications.Client, arn string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findNotificationConfigurationByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitNotificationConfigurationDeleted(ctx context.Context, conn *notifications.Client, id string) (*notifications.GetNotificationConfigurationOutput, error) {
	const (
		timeout = 10 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.NotificationConfigurationStatusDeleting),
		Target:  []string{},
		Refresh: statusNotificationConfiguration(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*notifications.GetNotificationConfigurationOutput); ok {
		return output, err
	}

	return nil, err
}

type notificationConfigurationResourceModel struct {
	AggregationDuration fwtypes.StringEnum[awstypes.AggregationDuration] `tfsdk:"aggregation_duration"`
	ARN                 types.String                                     `tfsdk:"arn"`
	Description         types.String                                     `tfsdk:"description"`
	Name                types.String                                     `tfsdk:"name"`
	Tags                tftags.Map                                       `tfsdk:"tags"`
	TagsAll             tftags.Map                                       `tfsdk:"tags_all"`
}
