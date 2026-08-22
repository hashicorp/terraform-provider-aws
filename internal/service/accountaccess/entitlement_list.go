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
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @FrameworkListResource("aws_accountaccess_entitlement")
func newEntitlementResourceAsListResource() list.ListResourceWithConfigure {
	return &entitlementListResource{}
}

var _ list.ListResource = &entitlementListResource{}

type entitlementListResource struct {
	entitlementResource
	framework.WithList
}

func (l *entitlementListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			names.AttrAccountID: listschema.StringAttribute{
				Required:    true,
				Description: "AWS account ID to filter Entitlements by.",
			},
			"application_arn": listschema.StringAttribute{
				CustomType:  fwtypes.ARNType,
				Required:    true,
				Description: "ARN of the parent Account Access Application to list Entitlements from.",
			},
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

	applicationArn := query.ApplicationARN.ValueString()

	tflog.Info(ctx, "Listing Account Access Entitlements", map[string]any{
		logging.ResourceAttributeKey("application_arn"): applicationArn,
	})

	input := accountaccess.ListEntitlementsInput{
		ApplicationArn: aws.String(applicationArn),
		Filter: &awstypes.EntitlementFilter{
			PrincipalRole: &awstypes.PrincipalRoleEntitlementFilter{
				Account: query.AccountID.ValueStringPointer(),
			},
		},
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		for item, err := range listEntitlements(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			entitlementID := aws.ToString(item.EntitlementId)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey("entitlement_id"), entitlementID)

			var out *accountaccess.GetEntitlementOutput
			if request.IncludeResource {
				out, err = FindEntitlementByTwoPartKey(ctx, conn, applicationArn, entitlementID)
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
				data.ApplicationARN = fwtypes.ARNValue(applicationArn)
				data.EntitlementID = types.StringValue(entitlementID)
				data.ID = types.StringValue(buildEntitlementID(applicationArn, entitlementID))

				if request.IncludeResource {
					flattenEntitlement(ctx, out, &data)
				}

				result.DisplayName = entitlementID
			})

			if !yield(result) {
				return
			}
		}
	}
}

// listEntitlements returns an iterator over all EntitlementsListMember items
// using the SDK paginator.
func listEntitlements(ctx context.Context, conn *accountaccess.Client, input *accountaccess.ListEntitlementsInput) iter.Seq2[awstypes.EntitlementsListMember, error] {
	return func(yield func(awstypes.EntitlementsListMember, error) bool) {
		pages := accountaccess.NewListEntitlementsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.EntitlementsListMember{}, fmt.Errorf("listing Account Access Entitlement resources: %w", err))
				return
			}

			for _, item := range page.Entitlements {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}

// listEntitlementModel holds the query configuration for listing Entitlements.
type listEntitlementModel struct {
	framework.WithRegionModel
	AccountID      types.String `tfsdk:"account_id"`
	ApplicationARN fwtypes.ARN  `tfsdk:"application_arn"`
}
