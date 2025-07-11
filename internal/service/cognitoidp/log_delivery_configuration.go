// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_cognitoidp_log_delivery_configuration", name="Log Delivery Configuration")
func newLogDeliveryConfigurationResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &LogDeliveryConfigurationResource{}
	return r, nil
}

const (
	ResNameLogDeliveryConfiguration = "Log Delivery Configuration"
)

/*
	func newManagedUserPoolClientResource(context.Context) (resource.ResourceWithConfigure, error) {
		return &managedUserPoolClientResource{}, nil
	}

	type managedUserPoolClientResource struct {
		framework.ResourceWithModel[managedUserPoolClientResourceModel]
		framework.WithNoOpDelete
	}
*/
type LogDeliveryConfigurationResource struct {
	framework.ResourceWithModel[resourceLogDeliveryConfigurationModel]
}

func (r *LogDeliveryConfigurationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrUserPoolID: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"log_configurations": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[logConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"event_source": schema.StringAttribute{
							Required:   true,
							CustomType: fwtypes.StringEnumType[awstypes.EventSourceName](),
						},
						"log_level": schema.StringAttribute{
							Required:   true,
							CustomType: fwtypes.StringEnumType[awstypes.LogLevel](),
						},
					},
					Blocks: map[string]schema.Block{
						"cloud_watch_logs_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[cloudWatchLogsConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"log_group_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Optional:   true,
									},
								},
							},
						},
						"firehose_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[firehoseConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"stream_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Optional:   true,
									},
								},
							},
						},
						"s3_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[s3ConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"bucket_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Optional:   true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *LogDeliveryConfigurationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().CognitoIDPClient(ctx)

	var plan resourceLogDeliveryConfigurationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input cognitoidentityprovider.SetLogDeliveryConfigurationInput
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.SetLogDeliveryConfiguration(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CognitoIDP, create.ErrActionCreating, ResNameLogDeliveryConfiguration, plan.UserPoolID.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.LogDeliveryConfiguration == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CognitoIDP, create.ErrActionCreating, ResNameLogDeliveryConfiguration, plan.UserPoolID.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out.LogDeliveryConfiguration, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *LogDeliveryConfigurationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().CognitoIDPClient(ctx)

	var state resourceLogDeliveryConfigurationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findLogDeliveryConfigurationByUserPoolID(ctx, conn, state.UserPoolID.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CognitoIDP, create.ErrActionReading, ResNameLogDeliveryConfiguration, state.UserPoolID.String(), err),
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

func (r *LogDeliveryConfigurationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().CognitoIDPClient(ctx)

	var plan, state resourceLogDeliveryConfigurationModel
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
		var input cognitoidentityprovider.SetLogDeliveryConfigurationInput
		resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.SetLogDeliveryConfiguration(ctx, &input)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.CognitoIDP, create.ErrActionUpdating, ResNameLogDeliveryConfiguration, plan.UserPoolID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.LogDeliveryConfiguration == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.CognitoIDP, create.ErrActionUpdating, ResNameLogDeliveryConfiguration, plan.UserPoolID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		resp.Diagnostics.Append(flex.Flatten(ctx, out.LogDeliveryConfiguration, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *LogDeliveryConfigurationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().CognitoIDPClient(ctx)

	var state resourceLogDeliveryConfigurationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// To "delete" a log delivery configuration, we set an empty configuration
	input := cognitoidentityprovider.SetLogDeliveryConfigurationInput{
		UserPoolId:        state.UserPoolID.ValueStringPointer(),
		LogConfigurations: []awstypes.LogConfigurationType{},
	}

	_, err := conn.SetLogDeliveryConfiguration(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CognitoIDP, create.ErrActionDeleting, ResNameLogDeliveryConfiguration, state.UserPoolID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *LogDeliveryConfigurationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrUserPoolID), req, resp)
}

func findLogDeliveryConfigurationByUserPoolID(ctx context.Context, conn *cognitoidentityprovider.Client, userPoolID string) (*awstypes.LogDeliveryConfigurationType, error) {
	input := cognitoidentityprovider.GetLogDeliveryConfigurationInput{
		UserPoolId: aws.String(userPoolID),
	}

	out, err := conn.GetLogDeliveryConfiguration(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			}
		}

		return nil, err
	}

	if out == nil || out.LogDeliveryConfiguration == nil {
		return nil, tfresource.NewEmptyResultError(&input)
	}

	return out.LogDeliveryConfiguration, nil
}

type resourceLogDeliveryConfigurationModel struct {
	framework.WithRegionModel
	UserPoolID        types.String                                           `tfsdk:"user_pool_id"`
	LogConfigurations fwtypes.ListNestedObjectValueOf[logConfigurationModel] `tfsdk:"log_configurations"`
}

type logConfigurationModel struct {
	EventSource                 fwtypes.StringEnum[awstypes.EventSourceName]                      `tfsdk:"event_source"`
	LogLevel                    fwtypes.StringEnum[awstypes.LogLevel]                             `tfsdk:"log_level"`
	CloudWatchLogsConfiguration fwtypes.ListNestedObjectValueOf[cloudWatchLogsConfigurationModel] `tfsdk:"cloud_watch_logs_configuration"`
	FirehoseConfiguration       fwtypes.ListNestedObjectValueOf[firehoseConfigurationModel]       `tfsdk:"firehose_configuration"`
	S3Configuration             fwtypes.ListNestedObjectValueOf[s3ConfigurationModel]             `tfsdk:"s3_configuration"`
}

type cloudWatchLogsConfigurationModel struct {
	LogGroupArn fwtypes.ARN `tfsdk:"log_group_arn"`
}

type firehoseConfigurationModel struct {
	StreamArn fwtypes.ARN `tfsdk:"stream_arn"`
}

type s3ConfigurationModel struct {
	BucketArn fwtypes.ARN `tfsdk:"bucket_arn"`
}

func sweepLogDeliveryConfigurations(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	// Log delivery configurations are tied to user pools, so we don't need a separate sweeper
	// They will be cleaned up when user pools are swept
	return nil, nil
}
