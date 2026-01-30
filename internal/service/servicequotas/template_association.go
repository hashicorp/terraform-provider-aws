// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package servicequotas

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/servicequotas"
	awstypes "github.com/aws/aws-sdk-go-v2/service/servicequotas/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_servicequotas_template_association", name="Template Association")
func newTemplateAssociationResource(context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceTemplateAssociation{}, nil
}

type resourceTemplateAssociation struct {
	framework.ResourceWithModel[templateAssociationResourceModel]
	framework.WithImportByID
}

func (r *resourceTemplateAssociation) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			names.AttrSkipDestroy: schema.BoolAttribute{
				Optional: true,
			},
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *resourceTemplateAssociation) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data templateAssociationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ServiceQuotasClient(ctx)

	var input servicequotas.AssociateServiceQuotaTemplateInput
	_, err := conn.AssociateServiceQuotaTemplate(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError("associating Service Quotas Template", err.Error())

		return
	}

	// Set values for unknowns.
	data.ID = fwflex.StringValueToFramework(ctx, r.Meta().AccountID(ctx))
	data.Status = fwflex.StringValueToFramework(ctx, awstypes.ServiceQuotaTemplateAssociationStatusAssociated)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *resourceTemplateAssociation) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data templateAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ServiceQuotasClient(ctx)

	output, err := findTemplateAssociation(ctx, conn)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError("reading Service Quotas Template Association", err.Error())

		return
	}

	// Set attributes for import.
	data.Status = fwflex.StringValueToFramework(ctx, output.ServiceQuotaTemplateAssociationStatus)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceTemplateAssociation) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data templateAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if data.SkipDestroy.ValueBool() {
		return
	}

	conn := r.Meta().ServiceQuotasClient(ctx)

	var input servicequotas.DisassociateServiceQuotaTemplateInput
	_, err := conn.DisassociateServiceQuotaTemplate(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError("disassociating Service Quotas Template", err.Error())

		return
	}
}

func findTemplateAssociation(ctx context.Context, conn *servicequotas.Client) (*servicequotas.GetAssociationForServiceQuotaTemplateOutput, error) {
	var input servicequotas.GetAssociationForServiceQuotaTemplateInput
	output, err := conn.GetAssociationForServiceQuotaTemplate(ctx, &input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	if status := output.ServiceQuotaTemplateAssociationStatus; status == awstypes.ServiceQuotaTemplateAssociationStatusDisassociated {
		return nil, &retry.NotFoundError{
			Message: string(status),
		}
	}

	return output, nil
}

type templateAssociationResourceModel struct {
	framework.WithRegionModel
	ID          types.String `tfsdk:"id"`
	SkipDestroy types.Bool   `tfsdk:"skip_destroy"`
	Status      types.String `tfsdk:"status"`
}
