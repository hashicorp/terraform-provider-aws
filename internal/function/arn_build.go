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
		Summary:             "arn_build Function",
		MarkdownDescription: "Builds an ARN from its constituent parts",
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:                "partition",
				MarkdownDescription: "Partition in which the resource is located",
			},
			function.StringParameter{
				Name:                "service",
				MarkdownDescription: "Service namespace",
			},
			function.StringParameter{
				Name:                "region",
				MarkdownDescription: "Region code",
			},
			function.StringParameter{
				Name:                "account_id",
				MarkdownDescription: "AWS account identifier",
			},
			function.StringParameter{
				Name:                "resource",
				MarkdownDescription: "Resource section, typically composed of a resource type and identifier",
			},
		},
		Return: function.StringReturn{},
	}
}

func (f arnBuildFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var partition, service, region, accountID, resource string

	resp.Error = function.ConcatFuncErrors(req.Arguments.Get(ctx, &partition, &service, &region, &accountID, &resource))
	if resp.Error != nil {
		return
	}

	result := arn.ARN{
		Partition: partition,
		Service:   service,
		Region:    region,
		AccountID: accountID,
		Resource:  resource,
	}

	resp.Error = function.ConcatFuncErrors(resp.Result.Set(ctx, result.String()))
}
