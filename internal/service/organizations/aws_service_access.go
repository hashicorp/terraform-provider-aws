// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package organizations

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	awstypes "github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @FrameworkResource("aws_organizations_aws_service_access", name="AWS Service Access")
// @IdentityAttribute("service_principal")
// @Testing(hasNoPreExistingResource=true)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/organizations/types;awstypes;awstypes.EnabledServicePrincipal")
// @Testing(serialize=true)
// @Testing(preCheck="github.com/hashicorp/terraform-provider-aws/internal/acctest;acctest.PreCheckOrganizationManagementAccount")
// @Testing(generator=false)
// @Testing(importStateIdAttribute="service_principal")
func newAWSServiceAccessResource(_ context.Context) (resource.ResourceWithConfigure, error) { // nosemgrep:ci.aws-in-func-name
	r := &awsServiceAccessResource{}

	return r, nil
}

type awsServiceAccessResource struct {
	framework.ResourceWithModel[awsServiceAccessResourceModel]
	framework.WithImportByIdentity
}

func (r *awsServiceAccessResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"date_enabled": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"service_principal": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					fwvalidators.ServicePrincipal(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
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

	servicePrincipal := fwflex.StringValueFromFramework(ctx, plan.ServicePrincipal)
	var input organizations.EnableAWSServiceAccessInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.EnableAWSServiceAccess(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, servicePrincipal)
		return
	}

	enabledServicePrincipal, err := findAWSServiceAccessByServicePrincipal(ctx, conn, servicePrincipal)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ServicePrincipal.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, enabledServicePrincipal, &plan))
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

	servicePrincipal := fwflex.StringValueFromFramework(ctx, state.ServicePrincipal)
	out, err := findAWSServiceAccessByServicePrincipal(ctx, conn, servicePrincipal)

	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, servicePrincipal)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &state))
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

	servicePrincipal := fwflex.StringValueFromFramework(ctx, state.ServicePrincipal)
	input := organizations.DisableAWSServiceAccessInput{
		ServicePrincipal: aws.String(servicePrincipal),
	}
	_, err := conn.DisableAWSServiceAccess(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, servicePrincipal)
		return
	}
}

func (r *awsServiceAccessResource) flatten(ctx context.Context, enabledServicePrincipal *awstypes.EnabledServicePrincipal, data *awsServiceAccessResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	diags.Append(fwflex.Flatten(ctx, enabledServicePrincipal, data)...)
	return diags
}

func findAWSServiceAccessByServicePrincipal(ctx context.Context, conn *organizations.Client, servicePrincipal string) (*awstypes.EnabledServicePrincipal, error) { // nosemgrep:ci.aws-in-func-name
	var input organizations.ListAWSServiceAccessForOrganizationInput
	output, err := findEnabledServicePrincipals(ctx, conn, &input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(tfslices.Filter(output, func(v awstypes.EnabledServicePrincipal) bool {
		return aws.ToString(v.ServicePrincipal) == servicePrincipal
	}))
}

type awsServiceAccessResourceModel struct {
	DateEnabled      timetypes.RFC3339 `tfsdk:"date_enabled"`
	ServicePrincipal types.String      `tfsdk:"service_principal"`
}
