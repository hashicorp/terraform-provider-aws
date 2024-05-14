// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicequotas

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/servicequotas"
	awstypes "github.com/aws/aws-sdk-go-v2/service/servicequotas/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Template Association")
func newResourceTemplateAssociation(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceTemplateAssociation{}, nil
}

const (
	ResNameTemplateAssociation = "Template Association"
)

type resourceTemplateAssociation struct {
	framework.ResourceWithConfigure
}

func (r *resourceTemplateAssociation) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_servicequotas_template_association"
}

func (r *resourceTemplateAssociation) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			names.AttrSkipDestroy: schema.BoolAttribute{
				Optional: true,
			},
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *resourceTemplateAssociation) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().ServiceQuotasClient(ctx)

	var plan resourceTemplateAssociationData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = types.StringValue(r.Meta().AccountID)

	_, err := conn.AssociateServiceQuotaTemplate(ctx, &servicequotas.AssociateServiceQuotaTemplateInput{})
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ServiceQuotas, create.ErrActionCreating, ResNameTemplateAssociation, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	// Status is not returned from Associate API, so call Get to get computed value
	out, err := conn.GetAssociationForServiceQuotaTemplate(ctx, &servicequotas.GetAssociationForServiceQuotaTemplateInput{})
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ServiceQuotas, create.ErrActionCreating, ResNameTemplateAssociation, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	plan.Status = flex.StringValueToFramework(ctx, string(out.ServiceQuotaTemplateAssociationStatus))

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceTemplateAssociation) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().ServiceQuotasClient(ctx)

	var state resourceTemplateAssociationData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.GetAssociationForServiceQuotaTemplate(ctx, &servicequotas.GetAssociationForServiceQuotaTemplateInput{})
	if out == nil || out.ServiceQuotaTemplateAssociationStatus == awstypes.ServiceQuotaTemplateAssociationStatusDisassociated {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ServiceQuotas, create.ErrActionSetting, ResNameTemplateAssociation, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	state.Status = flex.StringValueToFramework(ctx, string(out.ServiceQuotaTemplateAssociationStatus))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update is a no-op
func (r *resourceTemplateAssociation) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *resourceTemplateAssociation) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().ServiceQuotasClient(ctx)

	var state resourceTemplateAssociationData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.SkipDestroy.ValueBool() {
		return
	}

	_, err := conn.DisassociateServiceQuotaTemplate(ctx, &servicequotas.DisassociateServiceQuotaTemplateInput{})
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ServiceQuotas, create.ErrActionDeleting, ResNameTemplateAssociation, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceTemplateAssociation) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

type resourceTemplateAssociationData struct {
	ID          types.String `tfsdk:"id"`
	SkipDestroy types.Bool   `tfsdk:"skip_destroy"`
	Status      types.String `tfsdk:"status"`
}
