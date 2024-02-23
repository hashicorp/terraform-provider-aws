// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto

import (
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/internal/tfplugin5"
)

func CallFunction_Response(in *tfprotov5.CallFunctionResponse) *tfplugin5.CallFunction_Response {
	if in == nil {
		return nil
	}

	resp := &tfplugin5.CallFunction_Response{
		Diagnostics: Diagnostics(in.Diagnostics),
		Result:      DynamicValue(in.Result),
	}

	return resp
}

func Function(in *tfprotov5.Function) *tfplugin5.Function {
	if in == nil {
		return nil
	}

	resp := &tfplugin5.Function{
		Description:        in.Description,
		DescriptionKind:    StringKind(in.DescriptionKind),
		DeprecationMessage: in.DeprecationMessage,
		Parameters:         make([]*tfplugin5.Function_Parameter, 0, len(in.Parameters)),
		Return:             Function_Return(in.Return),
		Summary:            in.Summary,
		VariadicParameter:  Function_Parameter(in.VariadicParameter),
	}

	for _, parameter := range in.Parameters {
		resp.Parameters = append(resp.Parameters, Function_Parameter(parameter))
	}

	return resp
}

func Function_Parameter(in *tfprotov5.FunctionParameter) *tfplugin5.Function_Parameter {
	if in == nil {
		return nil
	}

	resp := &tfplugin5.Function_Parameter{
		AllowNullValue:     in.AllowNullValue,
		AllowUnknownValues: in.AllowUnknownValues,
		Description:        in.Description,
		DescriptionKind:    StringKind(in.DescriptionKind),
		Name:               in.Name,
		Type:               CtyType(in.Type),
	}

	return resp
}

func Function_Return(in *tfprotov5.FunctionReturn) *tfplugin5.Function_Return {
	if in == nil {
		return nil
	}

	resp := &tfplugin5.Function_Return{
		Type: CtyType(in.Type),
	}

	return resp
}

func GetFunctions_Response(in *tfprotov5.GetFunctionsResponse) *tfplugin5.GetFunctions_Response {
	if in == nil {
		return nil
	}

	resp := &tfplugin5.GetFunctions_Response{
		Diagnostics: Diagnostics(in.Diagnostics),
		Functions:   make(map[string]*tfplugin5.Function, len(in.Functions)),
	}

	for name, function := range in.Functions {
		resp.Functions[name] = Function(function)
	}

	return resp
}

func GetMetadata_FunctionMetadata(in *tfprotov5.FunctionMetadata) *tfplugin5.GetMetadata_FunctionMetadata {
	if in == nil {
		return nil
	}

	resp := &tfplugin5.GetMetadata_FunctionMetadata{
		Name: in.Name,
	}

	return resp
}
