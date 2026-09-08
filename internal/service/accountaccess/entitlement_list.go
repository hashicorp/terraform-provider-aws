// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package accountaccess

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/accountaccess"
	awstypes "github.com/aws/aws-sdk-go-v2/service/accountaccess/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfiter "github.com/hashicorp/terraform-provider-aws/internal/iter"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_accountaccess_entitlement")
func newEntitlementResourceAsListResource() list.ListResourceWithConfigure {
	return &entitlementListResource{}
}

var _ list.ListResource = &entitlementListResource{}

type entitlementListResource struct {
	entitlementResource
	framework.WithList
}

func (l *entitlementListResource) ListResourceConfigSchema(ctx context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"application_arn": listschema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
		},
		Blocks: map[string]listschema.Block{
			names.AttrFilter: entitlementsFilterBlock(ctx),
		},
	}
}

func (l *entitlementListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().AccountAccessClient(ctx)

	var query listEntitlementModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	applicationARN := fwflex.StringValueFromFramework(ctx, query.ApplicationARN)
	var input accountaccess.ListEntitlementsInput
	if diags := fwflex.Expand(ctx, query, &input); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		for item, err := range listEntitlements(ctx, conn, &input) {
			if err != nil {
				yield(fwdiag.NewListResultErrorDiagnostic(err))
				return
			}

			entitlementID := aws.ToString(item.EntitlementId)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey("entitlement_id"), entitlementID)

			var out *accountaccess.GetEntitlementOutput
			if request.IncludeResource {
				out, err = findEntitlementByTwoPartKey(ctx, conn, applicationARN, entitlementID)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					yield(fwdiag.NewListResultErrorDiagnostic(err))
					return
				}
			}

			result := request.NewListResult(ctx)
			var data entitlementResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.ApplicationARN = fwtypes.ARNValue(applicationARN)
				data.EntitlementID = fwflex.StringValueToFramework(ctx, entitlementID)

				if request.IncludeResource {
					smerr.AddEnrich(ctx, &result.Diagnostics, l.flatten(ctx, out, &data))
					if result.Diagnostics.HasError() {
						return
					}
				}

				result.DisplayName = entitlementID
			})

			if !yield(result) {
				return
			}
		}
	}
}

func listEntitlements(ctx context.Context, conn *accountaccess.Client, input *accountaccess.ListEntitlementsInput, optFns ...func(*accountaccess.Options)) iter.Seq2[awstypes.EntitlementsListMember, error] {
	return tfiter.ConcatValuesWithError(listEntitlementPages(ctx, conn, input, optFns...))
}

func listEntitlementPages(ctx context.Context, conn *accountaccess.Client, input *accountaccess.ListEntitlementsInput, optFns ...func(*accountaccess.Options)) iter.Seq2[[]awstypes.EntitlementsListMember, error] {
	return func(yield func([]awstypes.EntitlementsListMember, error) bool) {
		pages := accountaccess.NewListEntitlementsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx, optFns...)
			if err != nil {
				yield(nil, fmt.Errorf("listing Account Access Entitlements: %w", err))
				return
			}

			if !yield(page.Entitlements, nil) {
				return
			}
		}
	}
}

type listEntitlementModel struct {
	framework.WithRegionModel
	ApplicationARN fwtypes.ARN                                             `tfsdk:"application_arn"`
	Filter         fwtypes.ListNestedObjectValueOf[entitlementFilterModel] `tfsdk:"filter"`
}

type entitlementFilterModel struct {
	PrincipalRole fwtypes.ListNestedObjectValueOf[principalRoleEntitlementFilterModel] `tfsdk:"principal_role"`
}

type principalRoleEntitlementFilterModel struct {
	Account   types.String                                          `tfsdk:"account_id"`
	Principal fwtypes.ListNestedObjectValueOf[principalFilterModel] `tfsdk:"principal"`
	RoleARN   fwtypes.ARN                                           `tfsdk:"role_arn"`
}

type principalFilterModel struct {
	IdentityCenter fwtypes.ListNestedObjectValueOf[identityCenterPrincipalFilterModel] `tfsdk:"identity_center"`
}

var (
	_ fwflex.Expander = principalFilterModel{}
)

func (m principalFilterModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.IdentityCenter.IsNull():
		model, d := m.IdentityCenter.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.PrincipalFilterMemberIdentityCenter
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, model, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}

	return nil, diags
}

type identityCenterPrincipalFilterModel struct {
	GroupID types.String `tfsdk:"group_id"`
	UserID  types.String `tfsdk:"user_id"`
}

var (
	_ fwflex.Expander = identityCenterPrincipalFilterModel{}
)

func (m identityCenterPrincipalFilterModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.GroupID.IsNull():
		r := awstypes.IdentityCenterPrincipalFilterMemberGroupId{
			Value: fwflex.StringValueFromFramework(ctx, m.GroupID),
		}
		return &r, diags

	case !m.UserID.IsNull():
		r := awstypes.IdentityCenterPrincipalFilterMemberUserId{
			Value: fwflex.StringValueFromFramework(ctx, m.UserID),
		}
		return &r, diags
	}

	return nil, diags
}
