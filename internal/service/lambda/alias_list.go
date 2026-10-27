// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambda

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_lambda_alias")
func newAliasResourceAsListResource() inttypes.ListResourceForSDK {
	l := aliasListResource{}
	l.SetResourceSchema(resourceAlias())
	return &l
}

var _ list.ListResource = &aliasListResource{}

type aliasListResource struct {
	framework.ResourceWithConfigure
	framework.ListResourceWithSDKv2Resource
}

type aliasListResourceModel struct {
	framework.WithRegionModel
	FunctionName types.String `tfsdk:"function_name"`
}

func (l *aliasListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"function_name": listschema.StringAttribute{
				Required: true,
			},
		},
		Blocks: map[string]listschema.Block{},
	}
}

func (l *aliasListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	var query aliasListResourceModel
	if diags := request.Config.Get(ctx, &query); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	awsClient := l.Meta()
	conn := awsClient.LambdaClient(ctx)

	functionName := query.FunctionName.ValueString()

	tflog.Info(ctx, "Listing Lambda Aliases", map[string]any{
		"function_name": functionName,
	})

	stream.Results = func(yield func(list.ListResult) bool) {
		input := lambda.ListAliasesInput{
			FunctionName: aws.String(functionName),
		}

		for item, err := range listAliases(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}
			aliasARN := aws.ToString(item.AliasArn)
			aliasName := aws.ToString(item.Name)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), aliasARN)

			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(aliasARN)
			rd.Set("function_name", functionName)
			rd.Set(names.AttrName, aliasName)

			if request.IncludeResource {
				output, err := findAliasByTwoPartKey(ctx, conn, functionName, aliasName)
				if err != nil {
					tflog.Error(ctx, "Reading Lambda Alias", map[string]any{
						names.AttrARN: aliasARN,
						"error":       err.Error(),
					})
					continue
				}

				if diags := resourceAliasFlatten(ctx, awsClient, rd, output); diags.HasError() {
					tflog.Error(ctx, "Flattening Lambda Alias", map[string]any{
						names.AttrARN: aliasARN,
						"error":       sdkdiag.DiagnosticsString(diags),
					})
					continue
				}
			}

			result.DisplayName = aliasName

			l.SetResult(ctx, awsClient, request.IncludeResource, rd, &result)
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

func listAliases(ctx context.Context, conn *lambda.Client, input *lambda.ListAliasesInput) iter.Seq2[awstypes.AliasConfiguration, error] {
	return func(yield func(awstypes.AliasConfiguration, error) bool) {
		pages := lambda.NewListAliasesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.AliasConfiguration{}, fmt.Errorf("listing Lambda Aliases: %w", err))
				return
			}

			for _, item := range page.Aliases {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
