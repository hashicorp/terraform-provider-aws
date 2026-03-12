// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambda

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
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

// @SDKListResource("aws_lambda_permission")
func permissionResourceAsListResource() inttypes.ListResourceForSDK {
	l := permissionListResource{}
	l.SetResourceSchema(resourcePermission())
	return &l
}

type permissionListResource struct {
	framework.ResourceWithConfigure
	framework.ListResourceWithSDKv2Resource
}

type permissionListResourceModel struct {
	framework.WithRegionModel
	FunctionName types.String `tfsdk:"function_name"`
	Qualifier    types.String `tfsdk:"qualifier"`
}

func (l *permissionListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"function_name": listschema.StringAttribute{
				Required: true,
			},
			"qualifier": listschema.StringAttribute{
				Optional: true,
			},
		},
		Blocks: map[string]listschema.Block{},
	}
}

func (l *permissionListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	var query permissionListResourceModel
	if diags := request.Config.Get(ctx, &query); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	awsClient := l.Meta()
	conn := awsClient.LambdaClient(ctx)

	functionName := query.FunctionName.ValueString()
	qualifier := query.Qualifier.ValueString()

	tflog.Info(ctx, "Listing Lambda permissions", map[string]any{
		"function_name": functionName,
	})

	stream.Results = func(yield func(list.ListResult) bool) {
		input := lambda.GetPolicyInput{
			FunctionName: aws.String(functionName),
		}
		if qualifier != "" {
			input.Qualifier = aws.String(qualifier)
		}

		output, err := findPolicy(ctx, conn, &input)
		if err != nil {
			result := fwdiag.NewListResultErrorDiagnostic(err)
			yield(result)
			return
		}

		var p policy
		if err := json.Unmarshal([]byte(aws.ToString(output.Policy)), &p); err != nil {
			result := fwdiag.NewListResultErrorDiagnostic(err)
			yield(result)
			return
		}

		for _, statement := range p.Statement {
			id := statement.Sid
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)

			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(id)
			rd.Set("function_name", functionName)
			rd.Set("statement_id", id)
			if qualifier != "" {
				rd.Set("qualifier", qualifier)
			}

			diags := resourcePermissionFlatten(ctx, rd, awsClient, &statement, functionName)
			if diags.HasError() || rd.Id() == "" {
				tflog.Error(ctx, "Reading Lambda permission", map[string]any{
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
