// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package opensearch

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/opensearch"
	awstypes "github.com/aws/aws-sdk-go-v2/service/opensearch/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
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
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrDomainName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"service": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.AWSServicePrincipal](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"authorized_principal": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[authorizedPrincipalData](ctx),
				Computed:   true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"service_options": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[serviceOptionsModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"supported_regions": schema.SetAttribute{
							CustomType:  fwtypes.SetOfStringType,
							ElementType: types.StringType,
							Optional:    true,
						},
					},
				},
			},
		},
	}
}

func (r *authorizeVPCEndpointAccessResource) ConfigValidators(context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("account"),
			path.MatchRoot("service"),
		),
	}
}

func (r *authorizeVPCEndpointAccessResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().OpenSearchClient(ctx)

	var plan authorizeVPCEndpointAccessResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &opensearch.AuthorizeVpcEndpointAccessInput{}
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

	resp.Diagnostics.Append(setPrincipalOnModel(ctx, out.AuthorizedPrincipal, &plan)...)
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

	out, err := findAuthorizeVPCEndpointAccessByPrincipal(ctx, conn, state.DomainName.ValueString(), state.Account.ValueString(), state.Service.ValueString())
	if retry.NotFound(err) {
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

	resp.Diagnostics.Append(setPrincipalOnModel(ctx, out, &state)...)
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
		DomainName: state.DomainName.ValueStringPointer(),
	}
	if !state.Account.IsNull() && state.Account.ValueString() != "" {
		in.Account = state.Account.ValueStringPointer()
	}
	if !state.Service.IsNull() && state.Service.ValueString() != "" {
		in.Service = awstypes.AWSServicePrincipal(state.Service.ValueString())
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

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrDomainName), parts[0])...)

	principal := parts[1]
	if inttypes.IsAWSAccountID(principal) {
		response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("account"), principal)...)
	} else {
		response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("service"), principal)...)
	}
}

func setPrincipalOnModel(ctx context.Context, ap *awstypes.AuthorizedPrincipal, model *authorizeVPCEndpointAccessResourceModel) (diags diag.Diagnostics) {
	principals := []authorizedPrincipalData{{
		Principal:     fwflex.StringToFramework(ctx, ap.Principal),
		PrincipalType: fwflex.StringValueToFramework(ctx, string(ap.PrincipalType)),
	}}
	list, d := fwtypes.NewListNestedObjectValueOfValueSlice[authorizedPrincipalData](ctx, principals)
	diags.Append(d...)
	if diags.HasError() {
		return
	}
	model.AuthorizedPrincipal = list

	principalStr := aws.ToString(ap.Principal)
	pt := string(ap.PrincipalType)
	switch {
	case pt == string(awstypes.PrincipalTypeAwsAccount) || pt == "AWS Account":
		model.Account = types.StringValue(principalStr)
		model.Service = fwtypes.StringEnumNull[awstypes.AWSServicePrincipal]()
	case pt == string(awstypes.PrincipalTypeAwsService) || pt == "AWS Service":
		model.Account = types.StringNull()
		model.Service = fwtypes.StringEnumValue(awstypes.AWSServicePrincipal(principalStr))
	default:
	}

	return
}

func findAuthorizeVPCEndpointAccessByPrincipal(ctx context.Context, conn *opensearch.Client, domainName, account, service string) (*awstypes.AuthorizedPrincipal, error) {
	input := opensearch.ListVpcEndpointAccessInput{
		DomainName: aws.String(domainName),
	}

	return findAuthorizeVPCEndpointAccess(ctx, conn, &input, func(ap *awstypes.AuthorizedPrincipal) bool {
		// The SDK defines PrincipalType as "AWS_ACCOUNT" / "AWS_SERVICE", but the
		// API returns "AWS Account" / "AWS Service" in practice.
		pt := string(ap.PrincipalType)
		principal := aws.ToString(ap.Principal)
		switch {
		case account != "":
			return (pt == string(awstypes.PrincipalTypeAwsAccount) || pt == "AWS Account") && principal == account
		case service != "":
			return (pt == string(awstypes.PrincipalTypeAwsService) || pt == "AWS Service") && principal == service
		}
		return false
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
	Service             fwtypes.StringEnum[awstypes.AWSServicePrincipal]         `tfsdk:"service"`
	ServiceOptions      fwtypes.ListNestedObjectValueOf[serviceOptionsModel]     `tfsdk:"service_options"`
	AuthorizedPrincipal fwtypes.ListNestedObjectValueOf[authorizedPrincipalData] `tfsdk:"authorized_principal"`
}

type serviceOptionsModel struct {
	SupportedRegions fwtypes.SetOfString `tfsdk:"supported_regions"`
}

type authorizedPrincipalData struct {
	Principal     types.String `tfsdk:"principal"`
	PrincipalType types.String `tfsdk:"principal_type"`
}
