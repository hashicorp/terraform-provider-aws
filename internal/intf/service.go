package intf

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// ServiceData is data about a service.
type ServiceData interface {
	DataSourceTypes(context.Context) (map[string]tfsdk.DataSourceType, error)
}
