// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sesv2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @FrameworkResource("aws_sesv2_tenant_suppression_attributes", name="Tenant Suppression Attributes")
// @IdentityAttribute("tenant_name")
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdAttribute="tenant_name")
func newTenantSuppressionAttributesResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &tenantSuppressionAttributesResource{}

	return r, nil
}

type tenantSuppressionAttributesResource struct {
	framework.ResourceWithModel[tenantSuppressionAttributesResourceModel]
	framework.WithNoOpDelete
	framework.WithImportByIdentity
}

func (r *tenantSuppressionAttributesResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"suppressed_reasons": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringEnumType[awstypes.SuppressionListReason](),
				Optional:    true,
				Computed:    true,
				ElementType: fwtypes.StringEnumType[awstypes.SuppressionListReason](),
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"suppression_scope": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.SuppressionListScope](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tenant_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *tenantSuppressionAttributesResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data tenantSuppressionAttributesResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SESV2Client(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.TenantName)
	input := sesv2.PutTenantSuppressionAttributesInput{}
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.PutTenantSuppressionAttributes(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
		return
	}

	// Read back to populate any computed values (e.g. suppression_scope).
	out, err := findTenantSuppressionAttributesByName(ctx, conn, name)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *tenantSuppressionAttributesResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data tenantSuppressionAttributesResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SESV2Client(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.TenantName)
	out, err := findTenantSuppressionAttributesByName(ctx, conn, name)
	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *tenantSuppressionAttributesResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var data tenantSuppressionAttributesResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SESV2Client(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.TenantName)
	input := sesv2.PutTenantSuppressionAttributesInput{}
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.PutTenantSuppressionAttributes(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
		return
	}

	out, err := findTenantSuppressionAttributesByName(ctx, conn, name)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func findTenantSuppressionAttributesByName(ctx context.Context, conn *sesv2.Client, name string) (*awstypes.TenantSuppressionAttributes, error) {
	out, err := findTenantByName(ctx, conn, name)

	if err != nil {
		return nil, err
	}

	if out.SuppressionAttributes == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return out.SuppressionAttributes, nil
}

type tenantSuppressionAttributesResourceModel struct {
	framework.WithRegionModel
	SuppressedReasons fwtypes.SetOfStringEnum[awstypes.SuppressionListReason] `tfsdk:"suppressed_reasons"`
	SuppressionScope  fwtypes.StringEnum[awstypes.SuppressionListScope]       `tfsdk:"suppression_scope"`
	TenantName        types.String                                            `tfsdk:"tenant_name"`
}
