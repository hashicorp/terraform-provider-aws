// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package servicequotas

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicequotas"
	awstypes "github.com/aws/aws-sdk-go-v2/service/servicequotas/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_servicequotas_template", name="Template")
// @Region(overrideEnabled=false)
func newTemplateResource(context.Context) (resource.ResourceWithConfigure, error) {
	return &templateResource{}, nil
}

const (
	templateResourceIDPartCount = 3
)

type templateResource struct {
	framework.ResourceWithModel[templateResourceModel]
	framework.WithImportByID
}

func (r *templateResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"aws_region": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
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
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				DeprecationMessage: "region is deprecated. Use aws_region instead.",
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

func (r *templateResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data templateResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ServiceQuotasClient(ctx)

	region, quotaCode, serviceCode := fwflex.StringValueFromFramework(ctx, data.AWSRegion), fwflex.StringValueFromFramework(ctx, data.QuotaCode), fwflex.StringValueFromFramework(ctx, data.ServiceCode)
	if region == "" {
		region = fwflex.StringValueFromFramework(ctx, data.Region)
	}

	id, err := flex.FlattenResourceId([]string{region, quotaCode, serviceCode}, templateResourceIDPartCount, false)
	if err != nil {
		response.Diagnostics.AddError("creating resource ID", err.Error())

		return
	}

	var input servicequotas.PutServiceQuotaIncreaseRequestIntoTemplateInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.AwsRegion = aws.String(region)

	output, err := conn.PutServiceQuotaIncreaseRequestIntoTemplate(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Service Quotas Template (%s)", id), err.Error())

		return
	}

	// Set values for unknowns.
	data.ID = fwflex.StringValueToFramework(ctx, id)
	response.Diagnostics.Append(fwflex.Flatten(ctx, output.ServiceQuotaIncreaseRequestInTemplate, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	data.Region = data.AWSRegion

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *templateResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data templateResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	parts, err := flex.ExpandResourceId(data.ID.ValueString(), templateResourceIDPartCount, false)
	if err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().ServiceQuotasClient(ctx)

	output, err := findTemplateByThreePartKey(ctx, conn, parts[0], parts[1], parts[2])

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Service Quotas Template (%s)", data.ID.ValueString()), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	data.Region = data.AWSRegion

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *templateResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new templateResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	parts, err := flex.ExpandResourceId(new.ID.ValueString(), templateResourceIDPartCount, false)
	if err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().ServiceQuotasClient(ctx)

	if !new.DesiredValue.Equal(old.DesiredValue) {
		input := servicequotas.PutServiceQuotaIncreaseRequestIntoTemplateInput{
			AwsRegion:    aws.String(parts[0]),
			DesiredValue: fwflex.Float64FromFramework(ctx, new.DesiredValue),
			QuotaCode:    aws.String(parts[1]),
			ServiceCode:  aws.String(parts[2]),
		}

		_, err := conn.PutServiceQuotaIncreaseRequestIntoTemplate(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Service Quotas Template (%s)", new.ID.ValueString()), err.Error())

			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, new)...)
}

func (r *templateResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data templateResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	parts, err := flex.ExpandResourceId(data.ID.ValueString(), templateResourceIDPartCount, false)
	if err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().ServiceQuotasClient(ctx)

	input := servicequotas.DeleteServiceQuotaIncreaseRequestFromTemplateInput{
		AwsRegion:   aws.String(parts[0]),
		QuotaCode:   aws.String(parts[1]),
		ServiceCode: aws.String(parts[2]),
	}

	_, err = conn.DeleteServiceQuotaIncreaseRequestFromTemplate(ctx, &input)

	if errs.IsA[*awstypes.NoSuchResourceException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Service Quotas Template (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

func (r *templateResource) ConfigValidators(context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("aws_region"),
			path.MatchRoot(names.AttrRegion),
		),
	}
}

func findTemplateByThreePartKey(ctx context.Context, conn *servicequotas.Client, region, quotaCode, serviceCode string) (*awstypes.ServiceQuotaIncreaseRequestInTemplate, error) {
	input := servicequotas.GetServiceQuotaIncreaseRequestFromTemplateInput{
		AwsRegion:   aws.String(region),
		QuotaCode:   aws.String(quotaCode),
		ServiceCode: aws.String(serviceCode),
	}

	return findTemplate(ctx, conn, &input)
}

func findTemplate(ctx context.Context, conn *servicequotas.Client, input *servicequotas.GetServiceQuotaIncreaseRequestFromTemplateInput) (*awstypes.ServiceQuotaIncreaseRequestInTemplate, error) {
	output, err := conn.GetServiceQuotaIncreaseRequestFromTemplate(ctx, input)

	if errs.IsA[*awstypes.NoSuchResourceException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ServiceQuotaIncreaseRequestInTemplate == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.ServiceQuotaIncreaseRequestInTemplate, nil
}

type templateResourceModel struct {
	AWSRegion    types.String  `tfsdk:"aws_region"`
	DesiredValue types.Float64 `tfsdk:"value"`
	GlobalQuota  types.Bool    `tfsdk:"global_quota"`
	ID           types.String  `tfsdk:"id"`
	QuotaCode    types.String  `tfsdk:"quota_code"`
	QuotaName    types.String  `tfsdk:"quota_name"`
	Region       types.String  `tfsdk:"region"`
	ServiceCode  types.String  `tfsdk:"service_code"`
	ServiceName  types.String  `tfsdk:"service_name"`
	Unit         types.String  `tfsdk:"unit"`
}
