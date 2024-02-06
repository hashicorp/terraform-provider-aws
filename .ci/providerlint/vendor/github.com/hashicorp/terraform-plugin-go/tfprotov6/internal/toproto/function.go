// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto

import (
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/internal/tfplugin6"
)

func CallFunction_Response(in *tfprotov6.CallFunctionResponse) *tfplugin6.CallFunction_Response {
	if in == nil {
		return nil
	}

	resp := &tfplugin6.CallFunction_Response{
		Diagnostics: Diagnostics(in.Diagnostics),
		Result:      DynamicValue(in.Result),
	}

	return resp
}

func Function(in *tfprotov6.Function) *tfplugin6.Function {
	if in == nil {
		return nil
	}

	resp := &tfplugin6.Function{
		Description:        in.Description,
		DescriptionKind:    StringKind(in.DescriptionKind),
		DeprecationMessage: in.DeprecationMessage,
		Parameters:         make([]*tfplugin6.Function_Parameter, 0, len(in.Parameters)),
		Return:             Function_Return(in.Return),
		Summary:            in.Summary,
		VariadicParameter:  Function_Parameter(in.VariadicParameter),
	}

	for _, parameter := range in.Parameters {
		resp.Parameters = append(resp.Parameters, Function_Parameter(parameter))
	}

	return resp
}

func Function_Parameter(in *tfprotov6.FunctionParameter) *tfplugin6.Function_Parameter {
	if in == nil {
		return nil
	}

	resp := &tfplugin6.Function_Parameter{
		AllowNullValue:     in.AllowNullValue,
		AllowUnknownValues: in.AllowUnknownValues,
		Description:        in.Description,
		DescriptionKind:    StringKind(in.DescriptionKind),
		Name:               in.Name,
		Type:               CtyType(in.Type),
	}

	return resp
}

func Function_Return(in *tfprotov6.FunctionReturn) *tfplugin6.Function_Return {
	if in == nil {
		return nil
	}

	resp := &tfplugin6.Function_Return{
		Type: CtyType(in.Type),
	}

	return resp
}

func GetFunctions_Response(in *tfprotov6.GetFunctionsResponse) *tfplugin6.GetFunctions_Response {
	if in == nil {
		return nil
	}

	resp := &tfplugin6.GetFunctions_Response{
		Diagnostics: Diagnostics(in.Diagnostics),
		Functions:   make(map[string]*tfplugin6.Function, len(in.Functions)),
	}

	for name, function := range in.Functions {
		resp.Functions[name] = Function(function)
	}

	return resp
}

func GetMetadata_FunctionMetadata(in *tfprotov6.FunctionMetadata) *tfplugin6.GetMetadata_FunctionMetadata {
	if in == nil {
		return nil
	}

	return &tfplugin6.GetMetadata_FunctionMetadata{
		Name: in.Name,
	}
}
