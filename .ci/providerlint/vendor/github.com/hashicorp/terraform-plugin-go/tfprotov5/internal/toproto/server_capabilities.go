package toproto

import (
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/internal/tfplugin5"
)

func GetProviderSchema_ServerCapabilities(in *tfprotov5.ServerCapabilities) *tfplugin5.GetProviderSchema_ServerCapabilities {
	if in == nil {
		return nil
	}

	return &tfplugin5.GetProviderSchema_ServerCapabilities{
		PlanDestroy: in.PlanDestroy,
	}
}
