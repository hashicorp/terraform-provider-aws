package intf

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// ServicePackageData is data about a service package.
type ServicePackageData interface {
	Configure(context.Context, any) error
	FrameworkDataSources(context.Context) []func(context.Context) (datasource.DataSourceWithConfigure, error)
	FrameworkResources(context.Context) []func(context.Context) (resource.ResourceWithConfigure, error)
	ServicePackageName() string
}
