// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package workmail

import (
	"context"
	"fmt"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/workmail"
	awstypes "github.com/aws/aws-sdk-go-v2/service/workmail/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_workmail_default_domain", name="Default Domain")
func newDefaultDomainResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &defaultDomainResource{}

	return r, nil
}

const (
	ResNameDefaultDomain = "Default Domain"
)

type defaultDomainResource struct {
	framework.ResourceWithModel[defaultDomainResourceModel]
	framework.WithNoOpDelete
}

func (r *defaultDomainResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrDomainName: schema.StringAttribute{
				Required: true,
			},
			"organization_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *defaultDomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().WorkMailClient(ctx)

	var plan defaultDomainResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.putDefaultMailDomain(ctx, conn, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *defaultDomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().WorkMailClient(ctx)

	var state defaultDomainResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	domainName, err := findDefaultDomainByOrgID(ctx, conn, state.OrganizationId.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.DomainName.String())
		return
	}

	state.DomainName = flex.StringValueToFramework(ctx, domainName)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *defaultDomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().WorkMailClient(ctx)

	var plan defaultDomainResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.putDefaultMailDomain(ctx, conn, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *defaultDomainResource) putDefaultMailDomain(ctx context.Context, conn *workmail.Client, plan *defaultDomainResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	var input workmail.UpdateDefaultMailDomainInput
	smerr.AddEnrich(ctx, &diags, flex.Expand(ctx, plan, &input))
	if diags.HasError() {
		return diags
	}

	_, err := conn.UpdateDefaultMailDomain(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &diags, err, smerr.ID, plan.DomainName.String())
	}

	return diags
}

func (r *defaultDomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := intflex.ExpandResourceId(req.ID, domainIDParts, false)
	if err != nil {
		resp.Diagnostics.Append(fwdiag.NewParsingResourceIDErrorDiagnostic(err))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrDomainName), parts[1])...)
}

func findDefaultDomainByOrgID(ctx context.Context, conn *workmail.Client, orgID string) (string, error) {
	input := workmail.ListMailDomainsInput{
		OrganizationId: aws.String(orgID),
	}

	pages := workmail.NewListMailDomainsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			if errs.IsA[*awstypes.ResourceNotFoundException](err) {
				return "", smarterr.NewError(&retry.NotFoundError{
					LastError: err,
				})
			}
			return "", smarterr.NewError(err)
		}

		for _, d := range page.MailDomains {
			if d.DefaultDomain {
				return aws.ToString(d.DomainName), nil
			}
		}
	}

	return "", smarterr.NewError(&retry.NotFoundError{
		Message: fmt.Sprintf("no default domain found for WorkMail organization %s", orgID),
	})
}

type defaultDomainResourceModel struct {
	framework.WithRegionModel
	OrganizationId types.String `tfsdk:"organization_id"`
	DomainName     types.String `tfsdk:"domain_name"`
}
