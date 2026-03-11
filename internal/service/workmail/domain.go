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
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_workmail_domain", name="Domain")
func newDomainResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &domainResource{}

	return r, nil
}

const (
	ResNameDomain = "Domain"
	domainIDParts = 2
)

type domainResource struct {
	framework.ResourceWithModel[domainResourceModel]
	framework.WithNoUpdate
}

func (r *domainResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrDomainName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"dkim_verification_status": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.DnsRecordVerificationStatus](),
				Computed:   true,
			},
			"is_default": schema.BoolAttribute{
				Computed: true,
			},
			"is_test_domain": schema.BoolAttribute{
				Computed: true,
			},
			"organization_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ownership_verification_status": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.DnsRecordVerificationStatus](),
				Computed:   true,
			},
			"records": framework.ResourceComputedListOfObjectsAttribute[dnsRecordModel](ctx),
		},
	}
}

func (r *domainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().WorkMailClient(ctx)

	var plan domainResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input workmail.RegisterMailDomainInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.RegisterMailDomain(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.DomainName.String())
		return
	}

	// Read back to populate computed fields
	out, err := findDomainByOrgAndName(ctx, conn, plan.OrganizationId.ValueString(), plan.DomainName.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.DomainName.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *domainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().WorkMailClient(ctx)

	var state domainResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findDomainByOrgAndName(ctx, conn, state.OrganizationId.ValueString(), state.DomainName.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.DomainName.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *domainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().WorkMailClient(ctx)

	var state domainResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	var input workmail.DeregisterMailDomainInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, &state, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.DeregisterMailDomain(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.MailDomainNotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.DomainName.String())
		return
	}
}

func (r *domainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := intflex.ExpandResourceId(req.ID, domainIDParts, false)
	if err != nil {
		resp.Diagnostics.Append(fwdiag.NewParsingResourceIDErrorDiagnostic(err))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrDomainName), parts[1])...)
}

func findDomainByOrgAndName(ctx context.Context, conn *workmail.Client, orgID, domainName string) (*workmail.GetMailDomainOutput, error) {
	input := workmail.GetMailDomainInput{
		OrganizationId: aws.String(orgID),
		DomainName:     aws.String(domainName),
	}

	out, err := conn.GetMailDomain(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.MailDomainNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, smarterr.NewError(&retry.NotFoundError{
			Message: fmt.Sprintf("WorkMail Domain %s in organization %s not found", domainName, orgID),
		})
	}

	return out, nil
}

type domainResourceModel struct {
	framework.WithRegionModel
	DomainName                  types.String                                             `tfsdk:"domain_name"`
	DkimVerificationStatus      fwtypes.StringEnum[awstypes.DnsRecordVerificationStatus] `tfsdk:"dkim_verification_status"`
	IsDefault                   types.Bool                                               `tfsdk:"is_default"`
	IsTestDomain                types.Bool                                               `tfsdk:"is_test_domain"`
	OrganizationId              types.String                                             `tfsdk:"organization_id"`
	OwnershipVerificationStatus fwtypes.StringEnum[awstypes.DnsRecordVerificationStatus] `tfsdk:"ownership_verification_status"`
	Records                     fwtypes.ListNestedObjectValueOf[dnsRecordModel]          `tfsdk:"records"`
}

type dnsRecordModel struct {
	Hostname types.String `tfsdk:"hostname"`
	Type     types.String `tfsdk:"type"`
	Value    types.String `tfsdk:"value"`
}
