// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package organizations

import (
	"context"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	awstypes "github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @FrameworkResource("aws_organizations_aws_service_access", name="Aws Service Access")
// @IdentityAttribute("service_principal")
// @Testing(hasNoPreExistingResource=true)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/organizations/types;awstypes;awstypes.EnabledServicePrincipal")
// @Testing(serialize=true)
// @Testing(preCheck="github.com/hashicorp/terraform-provider-aws/internal/acctest;acctest.PreCheckOrganizationManagementAccount")
// @Testing(generator=false)
func newAwsServiceAccessResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &awsServiceAccessResource{}

	return r, nil
}

const (
	ResNameAwsServiceAccess = "Aws Service Access"
)

type awsServiceAccessResource struct {
	framework.ResourceWithModel[awsServiceAccessResourceModel]
	framework.WithImportByIdentity
}

func (r *awsServiceAccessResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"service_principal": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"date_enabled": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
		},
	}
}

func (r *awsServiceAccessResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().OrganizationsClient(ctx)

	var plan awsServiceAccessResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input organizations.EnableAWSServiceAccessInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.EnableAWSServiceAccess(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ServicePrincipal.ValueString())
		return
	}

	enabledServicePrincipal, err := findAwsServiceAccessByServicePrincipal(ctx, conn, plan.ServicePrincipal.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ServicePrincipal.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, enabledServicePrincipal, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *awsServiceAccessResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().OrganizationsClient(ctx)

	var state awsServiceAccessResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findAwsServiceAccessByServicePrincipal(ctx, conn, state.ServicePrincipal.ValueString())

	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ServicePrincipal.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *awsServiceAccessResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().OrganizationsClient(ctx)

	var state awsServiceAccessResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := organizations.DisableAWSServiceAccessInput{
		ServicePrincipal: state.ServicePrincipal.ValueStringPointer(),
	}

	_, err := conn.DisableAWSServiceAccess(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ServicePrincipal.ValueString())
		return
	}
}

func findAwsServiceAccessByServicePrincipal(ctx context.Context, conn *organizations.Client, servicePrincipal string) (*awstypes.EnabledServicePrincipal, error) {
	var enabledServices []awstypes.EnabledServicePrincipal

	pages := organizations.NewListAWSServiceAccessForOrganizationPaginator(conn, nil)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, sp := range page.EnabledServicePrincipals {
			if aws.ToString(sp.ServicePrincipal) == servicePrincipal {
				enabledServices = append(enabledServices, sp)
			}
		}
	}

	return tfresource.AssertSingleValueResult(enabledServices)
}

type awsServiceAccessResourceModel struct {
	ServicePrincipal types.String      `tfsdk:"service_principal"`
	DateEnabled      timetypes.RFC3339 `tfsdk:"date_enabled"`
}
