package types

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// ServicePackageTags represents resource-level tagging information.
type ServicePackageTags struct {
	IDAttribute string // The attribute used in UpdateTags etc.
}

// ServicePackageFrameworkDataSource represents a Terraform Plugin Framework data source
// implemented by a service package.
type ServicePackageFrameworkDataSource struct {
	Factory func(context.Context) (datasource.DataSourceWithConfigure, error)
	Tags    ServicePackageTags
}

// ServicePackageFrameworkResource represents a Terraform Plugin Framework resource
// implemented by a service package.
type ServicePackageFrameworkResource struct {
	Factory func(context.Context) (resource.ResourceWithConfigure, error)
	Tags    ServicePackageTags
}

// ServicePackageSDKDataSource represents a Terraform Plugin SDK data source
// implemented by a service package.
type ServicePackageSDKDataSource struct {
	Factory  func() *schema.Resource
	TypeName string
	Tags     ServicePackageTags
}

// ServicePackageSDKResource represents a Terraform Plugin SDK resource
// implemented by a service package.
type ServicePackageSDKResource struct {
	Factory  func() *schema.Resource
	TypeName string
	Tags     ServicePackageTags
}
