// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package verifiedpermissions

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/verifiedpermissions"
	awstypes "github.com/aws/aws-sdk-go-v2/service/verifiedpermissions/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
)

// @FrameworkListResource("aws_verifiedpermissions_policy_store_alias")
func newPolicyStoreAliasResourceAsListResource() list.ListResourceWithConfigure {
	return &policyStoreAliasListResource{}
}

var _ list.ListResource = &policyStoreAliasListResource{}

type policyStoreAliasListResource struct {
	policyStoreAliasResource
	framework.WithList
}

func (l *policyStoreAliasListResource) ListResourceConfigSchema(
	_ context.Context,
	_ list.ListResourceSchemaRequest,
	response *list.ListResourceSchemaResponse,
) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"policy_store_id": listschema.StringAttribute{
				Description: "Optional policy store ID used to filter policy store aliases.",
				Optional:    true,
				Validators:  policyStoreIDValidators(),
			},
		},
	}
}

func (l *policyStoreAliasListResource) List(
	ctx context.Context,
	request list.ListRequest,
	stream *list.ListResultsStream,
) {
	conn := l.Meta().VerifiedPermissionsClient(ctx)

	var query listPolicyStoreAliasModel

	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		diagnostics := request.Config.Get(ctx, &query)
		if diagnostics.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diagnostics)
			return
		}
	}

	input := verifiedpermissions.ListPolicyStoreAliasesInput{}

	if !query.PolicyStoreID.IsNull() &&
		!query.PolicyStoreID.IsUnknown() &&
		query.PolicyStoreID.ValueString() != "" {
		input.Filter = &awstypes.PolicyStoreAliasFilter{
			PolicyStoreId: query.PolicyStoreID.ValueStringPointer(),
		}
	}

	tflog.Info(ctx, "Listing Verified Permissions Policy Store Alias resources")

	stream.Results = func(yield func(list.ListResult) bool) {
		for item, err := range listPolicyStoreAliases(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			if item.State == awstypes.AliasStatePendingDeletion {
				continue
			}

			aliasName := aws.ToString(item.AliasName)

			result := request.NewListResult(ctx)
			var data policyStoreAliasResourceModel

			l.SetResult(
				ctx,
				l.Meta(),
				request.IncludeResource,
				&data,
				&result,
				func() {
					data.AliasName = fwflex.StringToFramework(
						ctx,
						item.AliasName,
					)

					if request.IncludeResource {
						result.Diagnostics.Append(
							flattenPolicyStoreAlias(ctx, &item, &data)...,
						)
						if result.Diagnostics.HasError() {
							return
						}
					}

					result.DisplayName = aliasName
				},
			)

			if result.Diagnostics.HasError() {
				yield(list.ListResult{
					Diagnostics: result.Diagnostics,
				})
				return
			}

			if !yield(result) {
				return
			}
		}
	}
}

func listPolicyStoreAliases(
	ctx context.Context,
	conn *verifiedpermissions.Client,
	input *verifiedpermissions.ListPolicyStoreAliasesInput,
) iter.Seq2[awstypes.PolicyStoreAliasItem, error] {
	return func(yield func(awstypes.PolicyStoreAliasItem, error) bool) {
		pages := verifiedpermissions.NewListPolicyStoreAliasesPaginator(
			conn,
			input,
		)

		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(
					awstypes.PolicyStoreAliasItem{},
					fmt.Errorf(
						"listing Verified Permissions Policy Store Alias resources: %w",
						err,
					),
				)
				return
			}

			for _, item := range page.PolicyStoreAliases {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}

type listPolicyStoreAliasModel struct {
	framework.WithRegionModel

	PolicyStoreID types.String `tfsdk:"policy_store_id"`
}
