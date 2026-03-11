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
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_lambda_function")
func newFunctionResourceAsListResource() inttypes.ListResourceForSDK {
	l := listResourceFunction{}
	l.SetResourceSchema(resourceFunction())
	return &l
}

var _ list.ListResource = &listResourceFunction{}

type listResourceFunction struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *listResourceFunction) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().LambdaClient(ctx)

	var query listFunctionModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	tflog.Info(ctx, "Listing Lambda Functions")
	stream.Results = func(yield func(list.ListResult) bool) {
		var input lambda.ListFunctionsInput
		for item, err := range listFunctions(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			functionName := aws.ToString(item.FunctionName)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), functionName)

			result := request.NewListResult(ctx)
			rd := l.ResourceData()
			rd.SetId(functionName)
			rd.Set("function_name", functionName)

			tflog.Info(ctx, "Reading Lambda Function")
			getFunctionInput := lambda.GetFunctionInput{FunctionName: aws.String(functionName)}
			output, err := findFunction(ctx, conn, &getFunctionInput)
			if err != nil {
				tflog.Error(ctx, "Reading Lambda Function", map[string]any{
					names.AttrID: functionName,
					"err":        err.Error(),
				})
				continue
			}

			diags := resourceFunctionFlatten(ctx, l.Meta(), rd, output, getFunctionInput, false)
			if diags.HasError() {
				tflog.Error(ctx, "Reading Lambda Function", map[string]any{
					names.AttrID: functionName,
					"diags":      sdkdiag.DiagnosticsString(diags),
				})
				continue
			}

			result.DisplayName = functionName

			l.SetResult(ctx, l.Meta(), request.IncludeResource, &result, rd)
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

type listFunctionModel struct {
	framework.WithRegionModel
}

func listFunctions(ctx context.Context, conn *lambda.Client, input *lambda.ListFunctionsInput) iter.Seq2[awstypes.FunctionConfiguration, error] {
	return func(yield func(awstypes.FunctionConfiguration, error) bool) {
		pages := lambda.NewListFunctionsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.FunctionConfiguration{}, fmt.Errorf("listing Lambda Function resources: %w", err))
				return
			}

			for _, item := range page.Functions {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
