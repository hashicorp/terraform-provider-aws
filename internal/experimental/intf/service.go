package intf

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

// ServiceData is data about a service.
type ServiceData interface {
	Configure(context.Context, any) error
	FrameworkDataSources(context.Context) []func(context.Context) (datasource.DataSource, error)
}
