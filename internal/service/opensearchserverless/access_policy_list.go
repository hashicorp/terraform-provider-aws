// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package opensearchserverless

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless"
	awstypes "github.com/aws/aws-sdk-go-v2/service/opensearchserverless/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @FrameworkListResource("aws_opensearchserverless_access_policy")
func newAccessPolicyResourceAsListResource() list.ListResourceWithConfigure {
	return &accessPolicyListResource{}
}

var _ list.ListResource = &accessPolicyListResource{}

type accessPolicyListResource struct {
	accessPolicyResource
	framework.WithList
}

func (l *accessPolicyListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			names.AttrType: listschema.StringAttribute{
				CustomType:  fwtypes.StringEnumType[awstypes.AccessPolicyType](),
				Required:    true,
				Description: "Type of access policy. Currently the only valid value is `data`.",
			},
		},
		Blocks: map[string]listschema.Block{},
	}
}

func (l *accessPolicyListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.OpenSearchServerlessClient(ctx)

	var query listAccessPolicyModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		input := opensearchserverless.ListAccessPoliciesInput{
			Type: query.Type.ValueEnum(),
		}

		for item, err := range listAccessPolicies(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			name := aws.ToString(item.Name)
			ctx = tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrName), name)

			var accessPolicy *awstypes.AccessPolicyDetail
			if request.IncludeResource {
				var err error
				accessPolicy, err = findAccessPolicyByNameAndType(ctx, conn, name, string(item.Type))
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					yield(fwdiag.NewListResultErrorDiagnostic(err))
					return
				}
			}

			result := request.NewListResult(ctx)
			var data accessPolicyResourceModel
			data.ID = fwflex.StringToFramework(ctx, item.Name)
			data.Name = fwflex.StringToFramework(ctx, item.Name)
			data.Type = fwtypes.StringEnumValue(item.Type)

			l.SetResult(ctx, awsClient, request.IncludeResource, &data, &result, func() {
				if request.IncludeResource {
					result.Diagnostics.Append(l.flatten(ctx, accessPolicy, &data)...)
					if result.Diagnostics.HasError() {
						return
					}
				}

				result.DisplayName = name
			})

			if result.Diagnostics.HasError() {
				result = list.ListResult{Diagnostics: result.Diagnostics}
				yield(result)
				return
			}

			if !yield(result) {
				return
			}
		}
	}
}

type listAccessPolicyModel struct {
	framework.WithRegionModel
	Type fwtypes.StringEnum[awstypes.AccessPolicyType] `tfsdk:"type"`
}

func listAccessPolicies(ctx context.Context, conn *opensearchserverless.Client, input *opensearchserverless.ListAccessPoliciesInput) iter.Seq2[awstypes.AccessPolicySummary, error] {
	return func(yield func(awstypes.AccessPolicySummary, error) bool) {
		pages := opensearchserverless.NewListAccessPoliciesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.AccessPolicySummary{}, fmt.Errorf("listing OpenSearch Serverless Access Policies: %w", err))
				return
			}

			for _, item := range page.AccessPolicySummaries {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
