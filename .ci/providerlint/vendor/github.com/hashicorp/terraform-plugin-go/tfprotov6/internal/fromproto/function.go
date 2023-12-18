// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fromproto

import (
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/internal/tfplugin6"
)

func CallFunctionRequest(in *tfplugin6.CallFunction_Request) (*tfprotov6.CallFunctionRequest, error) {
	if in == nil {
		return nil, nil
	}

	resp := &tfprotov6.CallFunctionRequest{
		Arguments: make([]*tfprotov6.DynamicValue, 0, len(in.Arguments)),
		Name:      in.Name,
	}

	for _, argument := range in.Arguments {
		resp.Arguments = append(resp.Arguments, DynamicValue(argument))
	}

	return resp, nil
}

func GetFunctionsRequest(in *tfplugin6.GetFunctions_Request) (*tfprotov6.GetFunctionsRequest, error) {
	if in == nil {
		return nil, nil
	}

	resp := &tfprotov6.GetFunctionsRequest{}

	return resp, nil
}
