// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicequotas

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicequotas"
	awstypes "github.com/aws/aws-sdk-go-v2/service/servicequotas/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Template")
func newResourceTemplate(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceTemplate{}, nil
}

const (
	ResNameTemplate     = "Template"
	templateIDPartCount = 3
)

type resourceTemplate struct {
	framework.ResourceWithConfigure
}

func (r *resourceTemplate) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_servicequotas_template"
}

func (r *resourceTemplate) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"global_quota": schema.BoolAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"quota_code": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"quota_name": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrRegion: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"service_code": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrServiceName: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrUnit: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrValue: schema.Float64Attribute{
				Required: true,
			},
		},
	}
}

func (r *resourceTemplate) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().ServiceQuotasClient(ctx)

	var plan resourceTemplateData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	region := plan.Region.ValueString()
	quotaCode := plan.QuotaCode.ValueString()
	serviceCode := plan.ServiceCode.ValueString()

	parts := []string{region, quotaCode, serviceCode}
	id, err := flex.FlattenResourceId(parts, templateIDPartCount, false)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ServiceQuotas, create.ErrActionCreating, ResNameTemplate, id, err),
			err.Error(),
		)
	}
	plan.ID = fwflex.StringValueToFramework(ctx, id)

	in := &servicequotas.PutServiceQuotaIncreaseRequestIntoTemplateInput{
		AwsRegion:    aws.String(region),
		DesiredValue: aws.Float64(plan.Value.ValueFloat64()),
		QuotaCode:    aws.String(quotaCode),
		ServiceCode:  aws.String(serviceCode),
	}

	out, err := conn.PutServiceQuotaIncreaseRequestIntoTemplate(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ServiceQuotas, create.ErrActionCreating, ResNameTemplate, plan.ID.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.ServiceQuotaIncreaseRequestInTemplate == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ServiceQuotas, create.ErrActionCreating, ResNameTemplate, plan.ID.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	templateItem := out.ServiceQuotaIncreaseRequestInTemplate
	plan.GlobalQuota = types.BoolValue(templateItem.GlobalQuota)
	plan.QuotaName = fwflex.StringToFramework(ctx, templateItem.QuotaName)
	plan.ServiceName = fwflex.StringToFramework(ctx, templateItem.ServiceName)
	plan.Unit = fwflex.StringToFramework(ctx, templateItem.Unit)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceTemplate) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().ServiceQuotasClient(ctx)

	var state resourceTemplateData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := FindTemplateByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ServiceQuotas, create.ErrActionSetting, ResNameTemplate, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	state.GlobalQuota = types.BoolValue(out.GlobalQuota)
	state.QuotaCode = fwflex.StringToFramework(ctx, out.QuotaCode)
	state.QuotaName = fwflex.StringToFramework(ctx, out.QuotaName)
	state.Region = fwflex.StringToFramework(ctx, out.AwsRegion)
	state.ServiceCode = fwflex.StringToFramework(ctx, out.ServiceCode)
	state.ServiceName = fwflex.StringToFramework(ctx, out.ServiceName)
	state.Unit = fwflex.StringToFramework(ctx, out.Unit)
	state.Value = fwflex.Float64ToFramework(ctx, out.DesiredValue)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceTemplate) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().ServiceQuotasClient(ctx)

	var state, plan resourceTemplateData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Value.Equal(state.Value) {
		in := &servicequotas.PutServiceQuotaIncreaseRequestIntoTemplateInput{
			AwsRegion:    aws.String(plan.Region.ValueString()),
			DesiredValue: aws.Float64(plan.Value.ValueFloat64()),
			QuotaCode:    aws.String(plan.QuotaCode.ValueString()),
			ServiceCode:  aws.String(plan.ServiceCode.ValueString()),
		}

		out, err := conn.PutServiceQuotaIncreaseRequestIntoTemplate(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.ServiceQuotas, create.ErrActionUpdating, ResNameTemplate, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.ServiceQuotaIncreaseRequestInTemplate == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.ServiceQuotas, create.ErrActionUpdating, ResNameTemplate, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		templateItem := out.ServiceQuotaIncreaseRequestInTemplate
		plan.GlobalQuota = types.BoolValue(templateItem.GlobalQuota)
		plan.QuotaName = fwflex.StringToFramework(ctx, templateItem.QuotaName)
		plan.ServiceName = fwflex.StringToFramework(ctx, templateItem.ServiceName)
		plan.Unit = fwflex.StringToFramework(ctx, templateItem.Unit)

		resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	}
}

func (r *resourceTemplate) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().ServiceQuotasClient(ctx)

	var state resourceTemplateData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &servicequotas.DeleteServiceQuotaIncreaseRequestFromTemplateInput{
		AwsRegion:   aws.String(state.Region.ValueString()),
		QuotaCode:   aws.String(state.QuotaCode.ValueString()),
		ServiceCode: aws.String(state.ServiceCode.ValueString()),
	}

	_, err := conn.DeleteServiceQuotaIncreaseRequestFromTemplate(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.NoSuchResourceException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ServiceQuotas, create.ErrActionDeleting, ResNameTemplate, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceTemplate) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func FindTemplateByID(ctx context.Context, conn *servicequotas.Client, id string) (*awstypes.ServiceQuotaIncreaseRequestInTemplate, error) {
	parts, err := flex.ExpandResourceId(id, templateIDPartCount, false)
	if err != nil {
		return nil, err
	}
	region := parts[0]
	quotaCode := parts[1]
	serviceCode := parts[2]

	in := &servicequotas.GetServiceQuotaIncreaseRequestFromTemplateInput{
		AwsRegion:   aws.String(region),
		QuotaCode:   aws.String(quotaCode),
		ServiceCode: aws.String(serviceCode),
	}

	out, err := conn.GetServiceQuotaIncreaseRequestFromTemplate(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.NoSuchResourceException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.ServiceQuotaIncreaseRequestInTemplate == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.ServiceQuotaIncreaseRequestInTemplate, nil
}

type resourceTemplateData struct {
	GlobalQuota types.Bool    `tfsdk:"global_quota"`
	ID          types.String  `tfsdk:"id"`
	QuotaCode   types.String  `tfsdk:"quota_code"`
	QuotaName   types.String  `tfsdk:"quota_name"`
	Region      types.String  `tfsdk:"region"`
	ServiceCode types.String  `tfsdk:"service_code"`
	ServiceName types.String  `tfsdk:"service_name"`
	Unit        types.String  `tfsdk:"unit"`
	Value       types.Float64 `tfsdk:"value"`
}
