// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package function

{{- if .IncludeComments }}
// **PLEASE DELETE THIS AND ALL TIP COMMENTS BEFORE SUBMITTING A PR FOR REVIEW!**
//
// TIP: ==== INTRODUCTION ====
// Thank you for trying the skaff tool!
//
// You have opted to include these helpful comments. They all include "TIP:"
// to help you find and remove them when you're done with them.
//
// While some aspects of this file are customized to your input, the
// scaffold tool does *not* produce any function logic.
//
// In other words, as generated, this is a rough outline of the work you will
// need to do. If something doesn't make sense for your situation, get rid of
// it.{{- end }}

import (
{{- if .IncludeComments }}
	// TIP: ==== IMPORTS ====
	// This is a common set of imports but not customized to your code since
	// your code hasn't been written yet. Make sure you, your IDE, or
	// goimports -w <file> fixes these imports.
	//
	// The provider linter wants your imports to be in two groups: first,
	// standard library (i.e., "fmt" or "strings"), second, everything else.
{{- end }}
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/function"
)
{{- if .IncludeComments }}
// TIP: ==== FILE STRUCTURE ====
// All functions should follow this basic outline. Improve this functions's
// maintainability by sticking to it.
//
// 1. Package declaration
// 2. Imports
// 3. Function struct with New* initialization function
// 4. Metadata, Definition, and Run methods (in that order)
{{- end }}

var _ function.Function = {{ .FunctionLower }}Function{}
{{ if .IncludeComments }}
// TIP: ==== INITIALIZATION FUNCTION ====
// The New* function returns an instance of the provider function struct. Currently,
// functions DO NOT follow the self-registration process used by resources
// and data sources, so this registration function must be manually added
// to the providers `Functions` method in `internal/provider/fwprovider/provider.go`.
{{- end }}
func New{{ .Function }}Function() function.Function {
	return &{{ .FunctionLower }}Function{}
}

type {{ .FunctionLower }}Function struct{}

func (f {{ .FunctionLower }}Function) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "{{ .FunctionSnake }}"
}
{{ if .IncludeComments }}
// TIP: ==== DEFINITION METHOD ====
// This method contains function details such as description, arguments, and 
// return values. The types of argument and return values are explicitly
// defined in this method.
{{- end }}
func (f {{ .FunctionLower }}Function) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:             "{{ .FunctionSnake }} Function",
		MarkdownDescription: "{{ .Description }}",
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:                "arg",
				MarkdownDescription: "Example argument description",
			},
		},
		Return: function.StringReturn{},
	}
}
{{ if .IncludeComments }}
// TIP: ==== RUN METHOD ====
// This method contains the logic of the function, such as manipulating 
// arguments or returning static values.
{{- end }}
func (f {{ .FunctionLower }}Function) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var arg string

	resp.Error = function.ConcatFuncErrors(req.Arguments.Get(ctx, &arg))
	if resp.Error != nil {
		return
	}
{{ if .IncludeComments }}
	// TIP: ==== ERROR HANDLING ====
	// Depending on the function logic being applied, there may be multiple
	// points at which the function could error. 
	//
	// Whenever logic is executed that could return an error, `resp.Error` should 
	// be set to the return of the `function.ConcatFuncErrors` helper.
{{- end }}
	if arg != "foo" {
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewFuncError("argument isn't foo"))
		return
	}

	result := fmt.Sprintf("%s-bar", arg)
{{ if .IncludeComments }}
        // TIP: ==== SETTING THE RESULT ====
	// This should always be the last step of this method, and potential
	// errors from setting the value should always be handled.
{{- end }}
	resp.Error = function.ConcatFuncErrors(resp.Result.Set(ctx, result))
}
