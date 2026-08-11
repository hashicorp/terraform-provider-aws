// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package sesv2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_sesv2_tenant", name="Tenant")
// @Tags(identifierAttribute="tenant_arn")
// @Testing(importStateIdAttribute="tenant_name")
func newTenantResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &tenantResource{}
	return r, nil
}

type tenantResource struct {
	framework.ResourceWithModel[tenantResourceModel]
}

func (r *tenantResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"sending_status": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"tenant_arn":      framework.ARNAttributeComputedOnly(),
			"tenant_id":       framework.IDAttribute(),
			"tenant_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *tenantResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan tenantResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SESV2Client(ctx)

	name := fwflex.StringValueFromFramework(ctx, plan.TenantName)
	input := sesv2.CreateTenantInput{
		Tags:       getTagsIn(ctx),
		TenantName: aws.String(name),
	}

	out, err := conn.CreateTenant(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *tenantResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state tenantResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SESV2Client(ctx)

	name := fwflex.StringValueFromFramework(ctx, state.TenantName)
	out, err := findTenantByName(ctx, conn, name)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	setTagsOut(ctx, out.Tags)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *tenantResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state tenantResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SESV2Client(ctx)

	name := fwflex.StringValueFromFramework(ctx, state.TenantName)
	input := sesv2.DeleteTenantInput{
		TenantName: aws.String(name),
	}

	_, err := conn.DeleteTenant(ctx, &input)
	if errs.IsA[*awstypes.NotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
		return
	}
}

func (r *tenantResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("tenant_name"), req, resp)
}

func findTenantByName(ctx context.Context, conn *sesv2.Client, name string) (*awstypes.Tenant, error) {
	input := sesv2.GetTenantInput{
		TenantName: aws.String(name),
	}

	return findTenant(ctx, conn, &input)
}

func findTenant(ctx context.Context, conn *sesv2.Client, input *sesv2.GetTenantInput) (*awstypes.Tenant, error) {
	out, err := conn.GetTenant(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.Tenant == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return out.Tenant, nil
}

type tenantResourceModel struct {
	framework.WithRegionModel
	SendingStatus types.String `tfsdk:"sending_status"`
	Tags          tftags.Map   `tfsdk:"tags"`
	TagsAll       tftags.Map   `tfsdk:"tags_all"`
	TenantARN     types.String `tfsdk:"tenant_arn"`
	TenantID      types.String `tfsdk:"tenant_id"`
	TenantName    types.String `tfsdk:"tenant_name"`
}
