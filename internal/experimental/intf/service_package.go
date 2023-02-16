package intf

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type ServicePackage interface {
	Configure(context.Context, any) error
	FrameworkDataSources(context.Context) []func(context.Context) (datasource.DataSourceWithConfigure, error)
	FrameworkResources(context.Context) []func(context.Context) (resource.ResourceWithConfigure, error)
	SDKDataSources(context.Context) []struct {
		TypeName string
		Factory  func() *schema.Resource
	}
	SDKResources(context.Context) []struct {
		TypeName string
		Factory  func() *schema.Resource
	}
	ServicePackageName() string
}
