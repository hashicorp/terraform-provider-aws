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
	FrameworkResources(context.Context) []func(context.Context) (ResourceWithConfigureAndImportState, error)
	ServicePackageName() string
}

type ResourceWithConfigureAndImportState interface {
	resource.ResourceWithConfigure
	ImportState(context.Context, resource.ImportStateRequest, *resource.ImportStateResponse)
}
