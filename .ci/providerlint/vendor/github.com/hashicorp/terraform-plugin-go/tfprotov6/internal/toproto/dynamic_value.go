// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto

import (
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/internal/tfplugin6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func DynamicValue(in *tfprotov6.DynamicValue) *tfplugin6.DynamicValue {
	if in == nil {
		return nil
	}

	resp := &tfplugin6.DynamicValue{
		Msgpack: in.MsgPack,
		Json:    in.JSON,
	}

	return resp
}

func CtyType(in tftypes.Type) []byte {
	if in == nil {
		return nil
	}

	// MarshalJSON is always error safe.
	// nolint:staticcheck // Intended first-party usage
	resp, _ := in.MarshalJSON()

	return resp
}
