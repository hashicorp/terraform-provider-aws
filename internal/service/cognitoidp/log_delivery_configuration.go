// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package cognitoidp

import (
	"context"
	"errors"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_cognito_log_delivery_configuration", name="Log Delivery Configuration")
// @IdentityAttribute("user_pool_id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types;awstypes;awstypes.LogDeliveryConfigurationType")
// @Testing(importStateIdFunc="testAccLogDeliveryConfigurationImportStateIdFunc")
// @Testing(importStateIdAttribute="user_pool_id")
// @Testing(hasNoPreExistingResource=true)
func newLogDeliveryConfigurationResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &logDeliveryConfigurationResource{}
	return r, nil
}

type logDeliveryConfigurationResource struct {
	framework.ResourceWithModel[resourceLogDeliveryConfigurationModel]
	framework.WithImportByIdentity
}

func (r *logDeliveryConfigurationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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
					listvalidator.IsRequired(),
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
									names.AttrStreamARN: schema.StringAttribute{
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

func (r *logDeliveryConfigurationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
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
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.UserPoolID.String())
		return
	}
	if out == nil || out.LogDeliveryConfiguration == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.UserPoolID.String())
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out.LogDeliveryConfiguration, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *logDeliveryConfigurationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().CognitoIDPClient(ctx)

	var state resourceLogDeliveryConfigurationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findLogDeliveryConfigurationByUserPoolID(ctx, conn, state.UserPoolID.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.UserPoolID.String())
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *logDeliveryConfigurationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.UserPoolID.String())
			return
		}
		if out == nil || out.LogDeliveryConfiguration == nil {
			smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.UserPoolID.String())
			return
		}

		resp.Diagnostics.Append(flex.Flatten(ctx, out.LogDeliveryConfiguration, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *logDeliveryConfigurationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().CognitoIDPClient(ctx)

	var state resourceLogDeliveryConfigurationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// set an empty configuration
	input := cognitoidentityprovider.SetLogDeliveryConfigurationInput{
		UserPoolId:        state.UserPoolID.ValueStringPointer(),
		LogConfigurations: []awstypes.LogConfigurationType{},
	}

	_, err := conn.SetLogDeliveryConfiguration(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.UserPoolID.String())
		return
	}
}

func findLogDeliveryConfigurationByUserPoolID(ctx context.Context, conn *cognitoidentityprovider.Client, userPoolID string) (*awstypes.LogDeliveryConfigurationType, error) {
	input := cognitoidentityprovider.GetLogDeliveryConfigurationInput{
		UserPoolId: aws.String(userPoolID),
	}

	out, err := conn.GetLogDeliveryConfiguration(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil || out.LogDeliveryConfiguration == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
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
