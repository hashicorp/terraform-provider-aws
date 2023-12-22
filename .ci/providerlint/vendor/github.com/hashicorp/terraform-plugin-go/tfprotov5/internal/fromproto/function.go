// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fromproto

import (
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/internal/tfplugin5"
)

func CallFunctionRequest(in *tfplugin5.CallFunction_Request) (*tfprotov5.CallFunctionRequest, error) {
	if in == nil {
		return nil, nil
	}

	resp := &tfprotov5.CallFunctionRequest{
		Arguments: make([]*tfprotov5.DynamicValue, 0, len(in.Arguments)),
		Name:      in.Name,
	}

	for _, argument := range in.Arguments {
		resp.Arguments = append(resp.Arguments, DynamicValue(argument))
	}

	return resp, nil
}

func GetFunctionsRequest(in *tfplugin5.GetFunctions_Request) (*tfprotov5.GetFunctionsRequest, error) {
	if in == nil {
		return nil, nil
	}

	resp := &tfprotov5.GetFunctionsRequest{}

	return resp, nil
}
