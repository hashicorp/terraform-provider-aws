// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package devopsguru

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/devopsguru"
	awstypes "github.com/aws/aws-sdk-go-v2/service/devopsguru/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Service Integration")
func newResourceServiceIntegration(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceServiceIntegration{}, nil
}

const (
	ResNameServiceIntegration = "Service Integration"
)

type resourceServiceIntegration struct {
	framework.ResourceWithConfigure
}

func (r *resourceServiceIntegration) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_devopsguru_service_integration"
}

func (r *resourceServiceIntegration) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"kms_server_side_encryption": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[kmsServerSideEncryptionData](ctx),
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					listvalidator.IsRequired(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrKMSKeyID: schema.StringAttribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"opt_in_status": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.OptInStatus](),
							Optional:   true,
							Computed:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						names.AttrType: schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.ServerSideEncryptionType](),
							Optional:   true,
							Computed:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
					},
				},
			},
			"logs_anomaly_detection": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[logsAnomalyDetectionData](ctx),
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					listvalidator.IsRequired(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"opt_in_status": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.OptInStatus](),
							Optional:   true,
							Computed:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
					},
				},
			},
			"ops_center": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[opsCenterData](ctx),
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					listvalidator.IsRequired(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"opt_in_status": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.OptInStatus](),
							Optional:   true,
							Computed:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
					},
				},
			},
		},
	}
}

func (r *resourceServiceIntegration) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().DevOpsGuruClient(ctx)

	var plan resourceServiceIntegrationData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue(r.Meta().Region)

	integration := &awstypes.UpdateServiceIntegrationConfig{}
	resp.Diagnostics.Append(flex.Expand(ctx, plan, integration)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &devopsguru.UpdateServiceIntegrationInput{
		ServiceIntegration: integration,
	}

	_, err := conn.UpdateServiceIntegration(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DevOpsGuru, create.ErrActionCreating, ResNameServiceIntegration, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	// Update API returns an empty body. Use find to populate computed fields.
	out, err := findServiceIntegration(ctx, conn)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DevOpsGuru, create.ErrActionCreating, ResNameServiceIntegration, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceServiceIntegration) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().DevOpsGuruClient(ctx)

	var state resourceServiceIntegrationData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findServiceIntegration(ctx, conn)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DevOpsGuru, create.ErrActionReading, ResNameServiceIntegration, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceServiceIntegration) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().DevOpsGuruClient(ctx)

	var plan, state resourceServiceIntegrationData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.KMSServerSideEncryption.Equal(state.KMSServerSideEncryption) ||
		!plan.LogsAnomalyDetection.Equal(state.LogsAnomalyDetection) ||
		!plan.OpsCenter.Equal(state.OpsCenter) {
		integration := &awstypes.UpdateServiceIntegrationConfig{}
		resp.Diagnostics.Append(flex.Expand(ctx, plan, integration)...)
		if resp.Diagnostics.HasError() {
			return
		}

		in := &devopsguru.UpdateServiceIntegrationInput{
			ServiceIntegration: integration,
		}

		_, err := conn.UpdateServiceIntegration(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DevOpsGuru, create.ErrActionUpdating, ResNameServiceIntegration, plan.ID.String(), err),
				err.Error(),
			)
			return
		}

		// Update API returns an empty body. Use find to populate computed fields.
		out, err := findServiceIntegration(ctx, conn)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DevOpsGuru, create.ErrActionUpdating, ResNameServiceIntegration, plan.ID.String(), err),
				err.Error(),
			)
			return
		}

		resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceServiceIntegration) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Delete is a no-op to prevent unintentionally disabling account-wide settings.
	//
	// The registry documentation includes a description of this behavior, indicating
	// that if users want to disable any settings previously opt-ed into they should
	// do so by applying those changes to an existing configuration before destroying
	// this resource.
}

func (r *resourceServiceIntegration) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func (r *resourceServiceIntegration) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	r.kmsSSEPlanModifier(ctx, req, resp)
	r.destroyPlanModifier(ctx, req, resp)
}

// kmsSSEPlanModifier is a resource plan modifier to handle cases where KMS settings
// are changed from a customer managed key to an AWS owned key
func (r *resourceServiceIntegration) kmsSSEPlanModifier(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if !req.State.Raw.IsNull() && !req.Plan.Raw.IsNull() {
		var config, plan resourceServiceIntegrationData
		resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
		resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}

		if !config.KMSServerSideEncryption.IsNull() && !plan.KMSServerSideEncryption.IsNull() {
			var planKMS []kmsServerSideEncryptionData
			var configKMS []kmsServerSideEncryptionData
			resp.Diagnostics.Append(plan.KMSServerSideEncryption.ElementsAs(ctx, &planKMS, false)...)
			resp.Diagnostics.Append(config.KMSServerSideEncryption.ElementsAs(ctx, &configKMS, false)...)
			if resp.Diagnostics.HasError() {
				return
			}

			// To avoid a ValidationException, force a replacement when KMS SSE is changed
			// to an AWS owned key and the computed key ID is being copied in the plan.
			//
			// ValidationException: Cannot specify KMSKeyId for AWS_OWNED_KEY
			if planKMS[0].Type.ValueString() == string(awstypes.ServerSideEncryptionTypeAwsOwnedKmsKey) &&
				!planKMS[0].KMSKeyID.IsNull() && configKMS[0].KMSKeyID.IsNull() {
				resp.RequiresReplace = []path.Path{path.Root("kms_server_side_encryption")}
			}
		}
	}
}

// destroyPlanModifier provides context on how to disable configured settings
func (r *resourceServiceIntegration) destroyPlanModifier(_ context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// If the entire plan is null, the resource is planned for destruction.
	if req.Plan.Raw.IsNull() {
		resp.Diagnostics.AddWarning(
			"Resource Destruction Considerations",
			"To prevent unintentional deletion of account wide settings, applying this resource destruction "+
				"will only remove the resource from the Terraform state. To disable any configured settings, "+
				"explicitly set the opt-in value to `DISABLED` and apply again before destroying.",
		)
	}
}

func findServiceIntegration(ctx context.Context, conn *devopsguru.Client) (*awstypes.ServiceIntegrationConfig, error) {
	in := &devopsguru.DescribeServiceIntegrationInput{}
	out, err := conn.DescribeServiceIntegration(ctx, in)
	if err != nil {
		return nil, err
	}

	if out == nil || out.ServiceIntegration == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.ServiceIntegration, nil
}

type resourceServiceIntegrationData struct {
	ID                      types.String                                                 `tfsdk:"id"`
	KMSServerSideEncryption fwtypes.ListNestedObjectValueOf[kmsServerSideEncryptionData] `tfsdk:"kms_server_side_encryption"`
	LogsAnomalyDetection    fwtypes.ListNestedObjectValueOf[logsAnomalyDetectionData]    `tfsdk:"logs_anomaly_detection"`
	OpsCenter               fwtypes.ListNestedObjectValueOf[opsCenterData]               `tfsdk:"ops_center"`
}

type kmsServerSideEncryptionData struct {
	KMSKeyID    types.String                                          `tfsdk:"kms_key_id"`
	OptInStatus fwtypes.StringEnum[awstypes.OptInStatus]              `tfsdk:"opt_in_status"`
	Type        fwtypes.StringEnum[awstypes.ServerSideEncryptionType] `tfsdk:"type"`
}

type logsAnomalyDetectionData struct {
	OptInStatus fwtypes.StringEnum[awstypes.OptInStatus] `tfsdk:"opt_in_status"`
}

type opsCenterData struct {
	OptInStatus fwtypes.StringEnum[awstypes.OptInStatus] `tfsdk:"opt_in_status"`
}
