// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package lakeformation

import (
	"context"
	"errors"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lakeformation"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lakeformation/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_lakeformation_identity_center_configuration", name="Identity Center Configuration")
// @IdentityAttribute("catalog_id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/lakeformation;lakeformation.DescribeLakeFormationIdentityCenterConfigurationOutput")
// @Testing(preCheckWithRegion="github.com/hashicorp/terraform-provider-aws/internal/acctest;acctest.PreCheckSSOAdminInstancesWithRegion")
// @Testing(hasNoPreExistingResource=true)
// @Testing(serialize=true)
// @Testing(importStateIdAttribute="catalog_id")
// @Testing(generator=false)
func newResourceIdentityCenterConfiguration(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceIdentityCenterConfiguration{}

	return r, nil
}

const (
	ResNameIdentityCenterConfiguration = "Identity Center Configuration"
)

type resourceIdentityCenterConfiguration struct {
	framework.ResourceWithModel[resourceIdentityCenterConfigurationModel]
	framework.WithImportByIdentity
}

func (r *resourceIdentityCenterConfiguration) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"application_arn": schema.StringAttribute{
				Computed: true,
			},
			names.AttrCatalogID: schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The ID of the Data Catalog.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"instance_arn": schema.StringAttribute{
				Required:    true,
				Description: "The ARN of the Identity Center instance.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"resource_share": schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			// TODO: "external_filtering"
			// TODO: "share_recipients"
		},
	}
}

func (r *resourceIdentityCenterConfiguration) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().LakeFormationClient(ctx)

	var plan resourceIdentityCenterConfigurationModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.CatalogID.IsNull() || plan.CatalogID.IsUnknown() {
		plan.CatalogID = flex.StringValueToFramework(ctx, r.Meta().AccountID(ctx))
	}

	var input lakeformation.CreateLakeFormationIdentityCenterConfigurationInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	createOut, err := conn.CreateLakeFormationIdentityCenterConfiguration(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.CatalogID.String())
		return
	}
	if createOut == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.CatalogID.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, createOut, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	readOut, err := findIdentityCenterConfigurationByID(ctx, conn, plan.CatalogID.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.CatalogID.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, readOut, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *resourceIdentityCenterConfiguration) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().LakeFormationClient(ctx)

	var state resourceIdentityCenterConfigurationModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findIdentityCenterConfigurationByID(ctx, conn, state.CatalogID.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.CatalogID.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *resourceIdentityCenterConfiguration) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *resourceIdentityCenterConfiguration) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().LakeFormationClient(ctx)

	var state resourceIdentityCenterConfigurationModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := lakeformation.DeleteLakeFormationIdentityCenterConfigurationInput{
		CatalogId: state.CatalogID.ValueStringPointer(),
	}

	_, err := conn.DeleteLakeFormationIdentityCenterConfiguration(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.EntityNotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.CatalogID.String())
		return
	}
}

func findIdentityCenterConfigurationByID(ctx context.Context, conn *lakeformation.Client, id string) (*lakeformation.DescribeLakeFormationIdentityCenterConfigurationOutput, error) {
	input := lakeformation.DescribeLakeFormationIdentityCenterConfigurationInput{
		CatalogId: aws.String(id),
	}

	out, err := conn.DescribeLakeFormationIdentityCenterConfiguration(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.EntityNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out, nil
}

type resourceIdentityCenterConfigurationModel struct {
	framework.WithRegionModel
	ApplicationARN types.String `tfsdk:"application_arn"`
	CatalogID      types.String `tfsdk:"catalog_id"`
	InstanceARN    types.String `tfsdk:"instance_arn"`
	ResourceShare  types.String `tfsdk:"resource_share"`
}
