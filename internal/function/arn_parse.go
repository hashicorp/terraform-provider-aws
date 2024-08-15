// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package function

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var arnParseResultAttrTypes = map[string]attr.Type{
	"partition":  types.StringType,
	"service":    types.StringType,
	"region":     types.StringType,
	"account_id": types.StringType,
	"resource":   types.StringType,
}

var _ function.Function = arnParseFunction{}

func NewARNParseFunction() function.Function {
	return &arnParseFunction{}
}

type arnParseFunction struct{}

func (f arnParseFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "arn_parse"
}

func (f arnParseFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:             "arn_parse Function",
		MarkdownDescription: "Parses an ARN into its constituent parts",
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:                "arn",
				MarkdownDescription: "ARN (Amazon Resource Name) to parse",
			},
		},
		Return: function.ObjectReturn{
			AttributeTypes: arnParseResultAttrTypes,
		},
	}
}

func (f arnParseFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var arg string

	resp.Error = function.ConcatFuncErrors(req.Arguments.Get(ctx, &arg))
	if resp.Error != nil {
		return
	}

	parts, err := arn.Parse(arg)
	if err != nil {
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewFuncError(err.Error()))
		return
	}

	value := map[string]attr.Value{
		"partition":  types.StringValue(parts.Partition),
		"service":    types.StringValue(parts.Service),
		"region":     types.StringValue(parts.Region),
		"account_id": types.StringValue(parts.AccountID),
		"resource":   types.StringValue(parts.Resource),
	}

	result, d := types.ObjectValue(arnParseResultAttrTypes, value)
	if d.HasError() {
		resp.Error = function.ConcatFuncErrors(resp.Error, function.FuncErrorFromDiags(ctx, d))
		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Result.Set(ctx, result))
}
