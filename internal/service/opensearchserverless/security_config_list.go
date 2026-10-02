// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package opensearchserverless

import (
	"context"
	"fmt"
	"iter"
	"strings"

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
// @FrameworkListResource("aws_opensearchserverless_security_config")
func newSecurityConfigResourceAsListResource() list.ListResourceWithConfigure {
	return &securityConfigListResource{}
}

var _ list.ListResource = &securityConfigListResource{}

type securityConfigListResource struct {
	securityConfigResource
	framework.WithList
}

func (l *securityConfigListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			names.AttrType: listschema.StringAttribute{
				CustomType:  fwtypes.StringEnumType[awstypes.SecurityConfigType](),
				Required:    true,
				Description: "Type of security configuration. Valid values: `saml`, `iamidentitycenter`, `iamfederation`.",
			},
		},
		Blocks: map[string]listschema.Block{},
	}
}

func (l *securityConfigListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.OpenSearchServerlessClient(ctx)

	var query listSecurityConfigModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		input := opensearchserverless.ListSecurityConfigsInput{
			Type: query.Type.ValueEnum(),
		}

		for item, err := range listSecurityConfigs(ctx, conn, &input) {
			if err != nil {
				yield(fwdiag.NewListResultErrorDiagnostic(err))
				return
			}

			id := aws.ToString(item.Id)
			// ID format: <type>/<account-id>/<name>
			parts := strings.SplitN(id, "/", 3)
			name := parts[len(parts)-1]
			ctx = tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrName), name)

			var securityConfig *awstypes.SecurityConfigDetail
			if request.IncludeResource {
				var err error
				securityConfig, err = findSecurityConfigByID(ctx, conn, id)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					yield(fwdiag.NewListResultErrorDiagnostic(err))
					return
				}
			}

			result := request.NewListResult(ctx)
			var data securityConfigResourceModel
			data.ID = fwflex.StringToFramework(ctx, item.Id)
			data.Name = fwflex.StringValueToFramework(ctx, name)
			data.Type = fwtypes.StringEnumValue(item.Type)

			l.SetResult(ctx, awsClient, request.IncludeResource, &data, &result, func() {
				if request.IncludeResource {
					result.Diagnostics.Append(l.flatten(ctx, securityConfig, &data)...)
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

type listSecurityConfigModel struct {
	framework.WithRegionModel
	Type fwtypes.StringEnum[awstypes.SecurityConfigType] `tfsdk:"type"`
}

func listSecurityConfigs(ctx context.Context, conn *opensearchserverless.Client, input *opensearchserverless.ListSecurityConfigsInput) iter.Seq2[awstypes.SecurityConfigSummary, error] {
	return func(yield func(awstypes.SecurityConfigSummary, error) bool) {
		pages := opensearchserverless.NewListSecurityConfigsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.SecurityConfigSummary{}, fmt.Errorf("listing OpenSearch Serverless Security Configs: %w", err))
				return
			}

			for _, item := range page.SecurityConfigSummaries {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
