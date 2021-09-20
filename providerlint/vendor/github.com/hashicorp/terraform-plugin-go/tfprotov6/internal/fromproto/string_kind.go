package fromproto

import (
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/internal/tfplugin6"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func StringKind(in tfplugin6.StringKind) tfprotov6.StringKind {
	return tfprotov6.StringKind(in)
}
