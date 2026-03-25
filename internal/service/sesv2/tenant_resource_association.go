// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sesv2

import (
	"context"
	"fmt"
	"strings"

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
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_sesv2_tenant_resource_association", name="Tenant Resource Association")
// @Testing(importStateIdAttribute="tenant_name")
func newTenantResourceAssociationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &tenantResourceAssociationResource{}
	return r, nil
}

type tenantResourceAssociationResource struct {
	framework.ResourceWithModel[tenantResourceAssociationResourceModel]
	framework.WithNoUpdate
}

func (r *tenantResourceAssociationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrResourceARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
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

func (r *tenantResourceAssociationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan tenantResourceAssociationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SESV2Client(ctx)

	tenantName, resourceARN := fwflex.StringValueFromFramework(ctx, plan.TenantName), fwflex.StringValueFromFramework(ctx, plan.ResourceArn)
	input := sesv2.CreateTenantResourceAssociationInput{
		ResourceArn: aws.String(resourceARN),
		TenantName:  aws.String(tenantName),
	}
	_, err := conn.CreateTenantResourceAssociation(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, tenantName)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *tenantResourceAssociationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state tenantResourceAssociationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SESV2Client(ctx)

	tenantName, resourceARN := fwflex.StringValueFromFramework(ctx, state.TenantName), fwflex.StringValueFromFramework(ctx, state.ResourceArn)
	_, err := findTenantResourceAssociationByTwoPartKey(ctx, conn, tenantName, resourceARN)

	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, tenantName)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *tenantResourceAssociationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state tenantResourceAssociationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SESV2Client(ctx)

	tenantName, resourceARN := fwflex.StringValueFromFramework(ctx, state.TenantName), fwflex.StringValueFromFramework(ctx, state.ResourceArn)
	input := sesv2.DeleteTenantResourceAssociationInput{
		ResourceArn: aws.String(resourceARN),
		TenantName:  aws.String(tenantName),
	}
	_, err := conn.DeleteTenantResourceAssociation(ctx, &input)
	if errs.IsA[*awstypes.NotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, tenantName)
		return
	}
}

func (r *tenantResourceAssociationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
	resp.State.SetAttribute(ctx, path.Root("tenant_name"), tenantName)
	resp.State.SetAttribute(ctx, path.Root(names.AttrResourceARN), resourceARN)
}

func findTenantResourceAssociationByTwoPartKey(ctx context.Context, conn *sesv2.Client, tenantName, resourceARN string) (*awstypes.TenantResource, error) {
	input := sesv2.ListTenantResourcesInput{
		TenantName: aws.String(tenantName),
	}

	output, err := findTenantResources(ctx, conn, &input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(tfslices.Filter(output, func(v awstypes.TenantResource) bool {
		return aws.ToString(v.ResourceArn) == resourceARN
	}))
}

func findTenantResources(ctx context.Context, conn *sesv2.Client, input *sesv2.ListTenantResourcesInput) ([]awstypes.TenantResource, error) {
	var output []awstypes.TenantResource

	pages := sesv2.NewListTenantResourcesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.NotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.TenantResources...)
	}

	return output, nil
}

type tenantResourceAssociationResourceModel struct {
	framework.WithRegionModel
	ResourceArn fwtypes.ARN  `tfsdk:"resource_arn"`
	TenantName  types.String `tfsdk:"tenant_name"`
}
