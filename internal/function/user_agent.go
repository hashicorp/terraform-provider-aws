// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package function

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/function"
)

var _ function.Function = userAgentFunction{}

func NewUserAgentFunction() function.Function {
	return &userAgentFunction{}
}

type userAgentFunction struct{}

func (f userAgentFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "user_agent"
}

func (f userAgentFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:             "user_agent Function",
		MarkdownDescription: "Formats a User-Agent product for use with the user_agent argument in the provider or provider_meta block.",
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:                "product_name",
				MarkdownDescription: "Product name.",
			},
			function.StringParameter{
				Name:                "product_version",
				MarkdownDescription: "Product version.",
			},
			function.StringParameter{
				Name:                "comment",
				MarkdownDescription: "Comment describing any additional product details.",
			},
		},
		Return: function.StringReturn{},
	}
}

func (f userAgentFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var name, version, comment string

	resp.Error = function.ConcatFuncErrors(req.Arguments.Get(ctx, &name, &version, &comment))
	if resp.Error != nil {
		return
	}

	if name == "" {
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewFuncError("product_name must be set"))
		return
	}

	var sb strings.Builder

	sb.WriteString(name)
	if version != "" {
		sb.WriteString("/" + version)
	}
	if comment != "" {
		sb.WriteString(" (" + comment + ")")
	}

	resp.Error = function.ConcatFuncErrors(resp.Result.Set(ctx, sb.String()))
}
