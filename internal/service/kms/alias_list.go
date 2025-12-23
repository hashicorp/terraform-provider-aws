// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package kms

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_kms_alias")
func aliasResourceAsListResource() inttypes.ListResourceForSDK {
	l := aliasListResource{}
	l.SetResourceSchema(resourceAlias())
	return &l
}

type aliasListResource struct {
	framework.ResourceWithConfigure
	framework.ListResourceWithSDKv2Resource
}

type aliasListResourceModel struct {
	framework.WithRegionModel
}

func (l *aliasListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{},
		Blocks:     map[string]listschema.Block{},
	}
}

func (l *aliasListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	var query aliasListResourceModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	awsClient := l.Meta()
	conn := awsClient.KMSClient(ctx)

	tflog.Info(ctx, "Listing KMS aliases")
	stream.Results = func(yield func(list.ListResult) bool) {
		var input kms.ListAliasesInput
		pages := kms.NewListAliasesPaginator(conn, &input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			for _, alias := range page.Aliases {
				id := aws.ToString(alias.AliasName)
				ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)

				result := request.NewListResult(ctx)

				rd := l.ResourceData()
				rd.SetId(id)

				diags := resourceAliasRead(ctx, rd, awsClient)
				if diags.HasError() || rd.Id() == "" {
					// Resource can't be read or is logically deleted.
					// Log and continue.
					tflog.Error(ctx, "Reading KMS alias", map[string]any{
						names.AttrID: id,
						"diags":      sdkdiag.DiagnosticsString(diags),
					})
					continue
				}

				result.DisplayName = id

				l.SetResult(ctx, awsClient, request.IncludeResource, &result, rd)
				if result.Diagnostics.HasError() {
					yield(result)
					return
				}

				if !yield(result) {
					return
				}
			}
		}
	}
}
