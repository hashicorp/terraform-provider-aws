package intf

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/provider"
)

// ServiceData is data about a service.
type ServiceData interface {
	Configure(context.Context, ProviderData) error
	DataSources(context.Context) (map[string]provider.DataSourceType, error)
}
