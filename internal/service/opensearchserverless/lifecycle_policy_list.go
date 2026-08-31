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
// @FrameworkListResource("aws_opensearchserverless_lifecycle_policy")
func newLifecyclePolicyResourceAsListResource() list.ListResourceWithConfigure {
	return &lifecyclePolicyListResource{}
}

var _ list.ListResource = &lifecyclePolicyListResource{}

type lifecyclePolicyListResource struct {
	lifecyclePolicyResource
	framework.WithList
}

func (l *lifecyclePolicyListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			names.AttrType: listschema.StringAttribute{
				CustomType:  fwtypes.StringEnumType[awstypes.LifecyclePolicyType](),
				Required:    true,
				Description: "Type of lifecycle policy. Must be `retention`.",
			},
		},
		Blocks: map[string]listschema.Block{},
	}
}

func (l *lifecyclePolicyListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.OpenSearchServerlessClient(ctx)

	var query listLifecyclePolicyModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	policyType := query.Type.ValueEnum()

	stream.Results = func(yield func(list.ListResult) bool) {
		input := opensearchserverless.ListLifecyclePoliciesInput{
			Type: policyType,
		}

		for item, err := range listLifecyclePolicies(ctx, conn, &input) {
			if err != nil {
				yield(fwdiag.NewListResultErrorDiagnostic(err))
				return
			}

			name := aws.ToString(item.Name)
			ctx = tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrName), name)

			var policy *awstypes.LifecyclePolicyDetail
			if request.IncludeResource {
				var err error
				policy, err = findLifecyclePolicyByNameAndType(ctx, conn, name, string(item.Type))
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					yield(fwdiag.NewListResultErrorDiagnostic(err))
					return
				}
			}

			result := request.NewListResult(ctx)
			var data lifecyclePolicyResourceModel
			data.ID = fwflex.StringToFramework(ctx, item.Name)
			data.Name = fwflex.StringToFramework(ctx, item.Name)
			data.Type = fwtypes.StringEnumValue(item.Type)

			l.SetResult(ctx, awsClient, request.IncludeResource, &data, &result, func() {
				if request.IncludeResource {
					result.Diagnostics.Append(l.flatten(ctx, policy, &data)...)
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

type listLifecyclePolicyModel struct {
	framework.WithRegionModel
	Type fwtypes.StringEnum[awstypes.LifecyclePolicyType] `tfsdk:"type"`
}

func listLifecyclePolicies(ctx context.Context, conn *opensearchserverless.Client, input *opensearchserverless.ListLifecyclePoliciesInput) iter.Seq2[awstypes.LifecyclePolicySummary, error] {
	return func(yield func(awstypes.LifecyclePolicySummary, error) bool) {
		pages := opensearchserverless.NewListLifecyclePoliciesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.LifecyclePolicySummary{}, fmt.Errorf("listing OpenSearch Serverless Lifecycle Policies: %w", err))
				return
			}

			for _, item := range page.LifecyclePolicySummaries {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
