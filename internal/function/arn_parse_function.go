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
		Parameters: []function.Parameter{
			function.StringParameter{},
		},
		Return: function.ObjectReturn{
			AttributeTypes: arnParseResultAttrTypes,
		},
	}
}

func (f arnParseFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var arg string

	resp.Diagnostics.Append(req.Arguments.Get(ctx, &arg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	parts, err := arn.Parse(arg)
	if err != nil {
		resp.Diagnostics.AddError("arn parsing failed", err.Error())
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
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.Result.Set(ctx, result)...)
}
