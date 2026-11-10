// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambda

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_lambda_resource_policy")
func newResourcePolicyResourceAsListResource() list.ListResourceWithConfigure {
	return &resourcePolicyListResource{}
}

var _ list.ListResource = &resourcePolicyListResource{}

type resourcePolicyListResource struct {
	resourcePolicyResource
	framework.WithList
}

func (l *resourcePolicyListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().LambdaClient(ctx)

	stream.Results = func(yield func(list.ListResult) bool) {
		var input lambda.ListFunctionsInput
		for item, err := range listFunctions(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			functionARN := aws.ToString(item.FunctionArn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), functionARN)

			out, err := findResourcePolicyByARN(ctx, conn, functionARN)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				yield(fwdiag.NewListResultErrorDiagnostic(err))
				return
			}

			result := request.NewListResult(ctx)

			var data resourcePolicyResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.ResourceARN = fwtypes.ARNValue(functionARN)
				result.Diagnostics.Append(l.flatten(ctx, out, &data)...)
				if result.Diagnostics.HasError() {
					return
				}

				result.DisplayName = aws.ToString(item.FunctionName)
			})

			if !yield(result) {
				return
			}
		}
	}
}
