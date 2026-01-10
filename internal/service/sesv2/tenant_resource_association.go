// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sesv2

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_sesv2_tenant_resource_association", name="Tenant Resource Association", tags=false)
// @Testing(importStateIdAttribute="tenant_name")
func newResourceTenantResourceAssociation(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceTenantResourceAssociation{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameTenantResourceAssociation = "Tenant Resource Association"
)

type resourceTenantResourceAssociation struct {
	framework.ResourceWithModel[resourceTenantResourceAssociationModel]
	framework.WithTimeouts
}

func (r *resourceTenantResourceAssociation) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: schema.StringAttribute{
				Computed: true,
			},
			names.AttrResourceARN: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
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

func (r *resourceTenantResourceAssociation) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().SESV2Client(ctx)

	var plan resourceTenantResourceAssociationModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input sesv2.CreateTenantResourceAssociationInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateTenantResourceAssociation(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.TenantName.String())
		return
	}
	if out == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.TenantName.String())
		return
	}

	plan.ID = types.StringValue(createID(plan.TenantName.ValueString(), plan.ResourceArn.ValueString()))

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *resourceTenantResourceAssociation) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().SESV2Client(ctx)

	var state resourceTenantResourceAssociationModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findTenantResourceAssociationByID(ctx, conn, state.ID.ValueString())

	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *resourceTenantResourceAssociation) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().SESV2Client(ctx)

	var state resourceTenantResourceAssociationModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := sesv2.DeleteTenantResourceAssociationInput{
		ResourceArn: state.ResourceArn.ValueStringPointer(),
		TenantName:  state.TenantName.ValueStringPointer(),
	}

	_, err := conn.DeleteTenantResourceAssociation(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.NotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}
}

func (r *resourceTenantResourceAssociation) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	customID := req.ID

	parts := strings.Split(customID, "|")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid ID Format",
			fmt.Sprintf("Expected ID in the format of tenant_name|resource_arn, got: %s", customID),
		)
	}
	tenantName := parts[0]
	resourceARN := parts[1]
	resp.State.SetAttribute(ctx, path.Root(names.AttrID), customID)
	resp.State.SetAttribute(ctx, path.Root("tenant_name"), tenantName)
	resp.State.SetAttribute(ctx, path.Root(names.AttrResourceARN), resourceARN)
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func findTenantResourceAssociationByID(
	ctx context.Context,
	conn *sesv2.Client,
	resourceID string,
) (*awstypes.TenantResource, error) {

	parts := strings.SplitN(resourceID, "|", 2)
	if len(parts) != 2 {
		return nil, smarterr.NewError(
			tfresource.NewEmptyResultError(),
		)
	}

	tenantName := parts[0]
	resourceARN := parts[1]

	input := &sesv2.ListTenantResourcesInput{
		TenantName: aws.String(tenantName),
	}

	p := sesv2.NewListTenantResourcesPaginator(conn, input)

	for p.HasMorePages() {
		out, err := p.NextPage(ctx)
		if err != nil {
			return nil, tfresource.ErrEmptyResult
		}

		for _, tenantResource := range out.TenantResources {
			if aws.ToString(tenantResource.ResourceArn) == resourceARN {
				return &tenantResource, nil
			}
		}
	}

	return nil, tfresource.ErrEmptyResult
}

type resourceTenantResourceAssociationModel struct {
	framework.WithRegionModel
	ResourceArn types.String `tfsdk:"resource_arn"`
	ID          types.String `tfsdk:"id"`
	TenantName  types.String `tfsdk:"tenant_name"`
}

func createID(tenantName, resourceARN string) string {
	return fmt.Sprintf("%s|%s", tenantName, resourceARN)
}
