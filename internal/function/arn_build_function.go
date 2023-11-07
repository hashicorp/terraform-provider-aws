// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package function

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-framework/function"
)

var _ function.Function = arnBuildFunction{}

func NewARNBuildFunction() function.Function {
	return &arnBuildFunction{}
}

type arnBuildFunction struct{}

func (f arnBuildFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "arn_build"
}

func (f arnBuildFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Parameters: []function.Parameter{
			function.StringParameter{Name: "partition"},
			function.StringParameter{Name: "service"},
			function.StringParameter{Name: "region"},
			function.StringParameter{Name: "account_id"},
			function.StringParameter{Name: "resource"},
		},
		Return: function.StringReturn{},
	}
}

func (f arnBuildFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var partition, service, region, accountID, resource string

	resp.Diagnostics.Append(req.Arguments.Get(ctx, &partition, &service, &region, &accountID, &resource)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result := arn.ARN{
		Partition: partition,
		Service:   service,
		Region:    region,
		AccountID: accountID,
		Resource:  resource,
	}

	resp.Diagnostics.Append(resp.Result.Set(ctx, result.String())...)
}
