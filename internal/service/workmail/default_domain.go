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
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
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
}

func (r *defaultDomainResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
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

	orgID := plan.OrganizationId.ValueString()
	domainName := plan.DomainName.ValueString()

	input := workmail.UpdateDefaultMailDomainInput{
		OrganizationId: aws.String(orgID),
		DomainName:     aws.String(domainName),
	}

	_, err := conn.UpdateDefaultMailDomain(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, orgID)
		return
	}

	plan.ID = types.StringValue(orgID)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *defaultDomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().WorkMailClient(ctx)

	var state defaultDomainResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := state.ID.ValueString()

	domainName, err := findDefaultDomainByOrgID(ctx, conn, orgID)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, orgID)
		return
	}

	state.OrganizationId = types.StringValue(orgID)
	state.DomainName = types.StringValue(domainName)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *defaultDomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().WorkMailClient(ctx)

	var plan defaultDomainResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := plan.OrganizationId.ValueString()
	domainName := plan.DomainName.ValueString()

	input := workmail.UpdateDefaultMailDomainInput{
		OrganizationId: aws.String(orgID),
		DomainName:     aws.String(domainName),
	}

	_, err := conn.UpdateDefaultMailDomain(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, orgID)
		return
	}

	plan.ID = types.StringValue(orgID)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *defaultDomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().WorkMailClient(ctx)

	var state defaultDomainResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := state.ID.ValueString()

	// Reset the default domain back to the auto-provisioned test domain
	testDomain, err := findTestDomainByOrgID(ctx, conn, orgID)
	if err != nil {
		// If we can't find the test domain, just remove from state without error.
		// The org may have been deleted already.
		return
	}

	// If the current default is already the test domain, nothing to do
	if testDomain == state.DomainName.ValueString() {
		return
	}

	input := workmail.UpdateDefaultMailDomainInput{
		OrganizationId: aws.String(orgID),
		DomainName:     aws.String(testDomain),
	}

	_, err = conn.UpdateDefaultMailDomain(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.MailDomainNotFoundException](err) {
			return
		}
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		if errs.IsA[*awstypes.OrganizationNotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, orgID)
		return
	}
}

func (r *defaultDomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrID), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), req.ID)...)
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

func findTestDomainByOrgID(ctx context.Context, conn *workmail.Client, orgID string) (string, error) {
	input := workmail.ListMailDomainsInput{
		OrganizationId: aws.String(orgID),
	}

	pages := workmail.NewListMailDomainsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return "", smarterr.NewError(err)
		}

		for _, d := range page.MailDomains {
			domainName := aws.ToString(d.DomainName)

			out, err := conn.GetMailDomain(ctx, &workmail.GetMailDomainInput{
				OrganizationId: aws.String(orgID),
				DomainName:     aws.String(domainName),
			})
			if err != nil {
				continue
			}

			if out.IsTestDomain {
				return domainName, nil
			}
		}
	}

	return "", fmt.Errorf("no test domain found for WorkMail organization %s", orgID)
}

type defaultDomainResourceModel struct {
	ID             types.String `tfsdk:"id"`
	OrganizationId types.String `tfsdk:"organization_id"`
	DomainName     types.String `tfsdk:"domain_name"`
}
