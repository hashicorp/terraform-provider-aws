package toproto

import (
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/internal/tfplugin6"
)

func GetProviderSchema_ServerCapabilities(in *tfprotov6.ServerCapabilities) *tfplugin6.GetProviderSchema_ServerCapabilities {
	if in == nil {
		return nil
	}

	return &tfplugin6.GetProviderSchema_ServerCapabilities{
		PlanDestroy: in.PlanDestroy,
	}
}
