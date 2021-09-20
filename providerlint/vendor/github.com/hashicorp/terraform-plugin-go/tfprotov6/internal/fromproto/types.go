package fromproto

import (
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/internal/tfplugin6"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func DynamicValue(in *tfplugin6.DynamicValue) *tfprotov6.DynamicValue {
	return &tfprotov6.DynamicValue{
		MsgPack: in.Msgpack,
		JSON:    in.Json,
	}
}
