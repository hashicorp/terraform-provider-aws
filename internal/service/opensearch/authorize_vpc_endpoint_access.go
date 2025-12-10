// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package opensearch

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/opensearch"
	awstypes "github.com/aws/aws-sdk-go-v2/service/opensearch/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_opensearch_authorize_vpc_endpoint_access", name="Authorize VPC Endpoint Access")
func newAuthorizeVPCEndpointAccessResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &authorizeVPCEndpointAccessResource{}

	return r, nil
}

const (
	ResNameAuthorizeVPCEndpointAccess = "Authorize Vpc Endpoint Access"
)

type authorizeVPCEndpointAccessResource struct {
	framework.ResourceWithModel[authorizeVPCEndpointAccessResourceModel]
	framework.WithNoUpdate
}

func (r *authorizeVPCEndpointAccessResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"account": schema.StringAttribute{
				Required: true, PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrDomainName: schema.StringAttribute{
				Required: true,
			},
			"authorized_principal": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[authorizedPrincipalData](ctx),
				Computed:   true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *authorizeVPCEndpointAccessResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().OpenSearchClient(ctx)

	var plan authorizeVPCEndpointAccessResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &opensearch.AuthorizeVpcEndpointAccessInput{
		Account:    plan.Account.ValueStringPointer(),
		DomainName: plan.DomainName.ValueStringPointer(),
	}

	resp.Diagnostics.Append(fwflex.Expand(ctx, plan, in)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.AuthorizeVpcEndpointAccess(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearch, create.ErrActionCreating, ResNameAuthorizeVPCEndpointAccess, plan.DomainName.String(), err),
			err.Error(),
		)
		return
	}

	if out == nil || out.AuthorizedPrincipal == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearch, create.ErrActionCreating, ResNameAuthorizeVPCEndpointAccess, plan.DomainName.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, out, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *authorizeVPCEndpointAccessResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().OpenSearchClient(ctx)

	var state authorizeVPCEndpointAccessResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findAuthorizeVPCEndpointAccessByTwoPartKey(ctx, conn, state.DomainName.ValueString(), state.Account.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearch, create.ErrActionSetting, ResNameAuthorizeVPCEndpointAccess, state.DomainName.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *authorizeVPCEndpointAccessResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().OpenSearchClient(ctx)

	var state authorizeVPCEndpointAccessResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &opensearch.RevokeVpcEndpointAccessInput{
		Account:    state.Account.ValueStringPointer(),
		DomainName: state.DomainName.ValueStringPointer(),
	}

	_, err := conn.RevokeVpcEndpointAccess(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearch, create.ErrActionDeleting, ResNameAuthorizeVPCEndpointAccess, state.DomainName.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *authorizeVPCEndpointAccessResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	const (
		authorizeVPCEndpointAccessImportIDParts = 2
	)
	parts, err := intflex.ExpandResourceId(request.ID, authorizeVPCEndpointAccessImportIDParts, true)

	if err != nil {
		response.Diagnostics.Append(fwdiag.NewParsingResourceIDErrorDiagnostic(err))

		return
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("account"), parts[1])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrDomainName), parts[0])...)
}

func findAuthorizeVPCEndpointAccessByTwoPartKey(ctx context.Context, conn *opensearch.Client, domainName, account string) (*awstypes.AuthorizedPrincipal, error) {
	input := opensearch.ListVpcEndpointAccessInput{
		DomainName: aws.String(domainName),
	}

	return findAuthorizeVPCEndpointAccess(ctx, conn, &input, func(ap *awstypes.AuthorizedPrincipal) bool {
		// AWS API documentation, and the SDK for Go following it, seems to be wrong for the possible values for PrincipalType.
		// It states it can be "AWS_ACCOUNT" or "AWS_SERVICE", but in practice for accounts the value is "AWS Account".
		// Hence, not using the constant awstypes.PrincipalTypeAwsAccount from the SDK.
		return ap.PrincipalType == "AWS Account" && aws.ToString(ap.Principal) == account
	})
}

func findAuthorizeVPCEndpointAccess(ctx context.Context, conn *opensearch.Client, input *opensearch.ListVpcEndpointAccessInput, filter tfslices.Predicate[*awstypes.AuthorizedPrincipal]) (*awstypes.AuthorizedPrincipal, error) {
	output, err := findAuthorizeVPCEndpointAccesses(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findAuthorizeVPCEndpointAccesses(ctx context.Context, conn *opensearch.Client, input *opensearch.ListVpcEndpointAccessInput, filter tfslices.Predicate[*awstypes.AuthorizedPrincipal]) ([]awstypes.AuthorizedPrincipal, error) {
	var output []awstypes.AuthorizedPrincipal

	err := listVPCEndpointAccessPages(ctx, conn, input, func(page *opensearch.ListVpcEndpointAccessOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.AuthorizedPrincipalList {
			if filter(&v) {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}

type authorizeVPCEndpointAccessResourceModel struct {
	framework.WithRegionModel
	Account             types.String                                             `tfsdk:"account"`
	DomainName          types.String                                             `tfsdk:"domain_name"`
	AuthorizedPrincipal fwtypes.ListNestedObjectValueOf[authorizedPrincipalData] `tfsdk:"authorized_principal"`
}

type authorizedPrincipalData struct {
	Principal     types.String `tfsdk:"principal"`
	PrincipalType types.String `tfsdk:"principal_type"`
}
