// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// ServicePackageResourceTags represents resource-level tagging information.
type ServicePackageResourceTags struct {
	IdentifierAttribute string // The attribute for the identifier for UpdateTags etc.
	ResourceType        string // Extra resourceType parameter value for UpdateTags etc.
}

// ServicePackageEphemeralResource represents a Terraform Plugin Framework ephemeral resource
// implemented by a service package.
type ServicePackageEphemeralResource struct {
	Factory                 func(context.Context) (ephemeral.EphemeralResourceWithConfigure, error)
	TypeName                string
	Name                    string
	IsRegionOverrideEnabled bool // Does the ephemeral resource support per-resource Region override?
}

// ServicePackageFrameworkDataSource represents a Terraform Plugin Framework data source
// implemented by a service package.
type ServicePackageFrameworkDataSource struct {
	Factory                 func(context.Context) (datasource.DataSourceWithConfigure, error)
	TypeName                string
	Name                    string
	Tags                    *ServicePackageResourceTags
	IsRegionOverrideEnabled bool // Does the data source support per-resource Region override?
}

// ServicePackageFrameworkResource represents a Terraform Plugin Framework resource
// implemented by a service package.
type ServicePackageFrameworkResource struct {
	Factory                 func(context.Context) (resource.ResourceWithConfigure, error)
	TypeName                string
	Name                    string
	Tags                    *ServicePackageResourceTags
	IsRegionOverrideEnabled bool // Does the resource support per-resource Region override?
}

// ServicePackageSDKDataSource represents a Terraform Plugin SDK data source
// implemented by a service package.
type ServicePackageSDKDataSource struct {
	Factory                 func() *schema.Resource
	TypeName                string
	Name                    string
	Tags                    *ServicePackageResourceTags
	IsRegionOverrideEnabled bool // Does the data source support per-resource Region override?
}

// ServicePackageSDKResource represents a Terraform Plugin SDK resource
// implemented by a service package.
type ServicePackageSDKResource struct {
	Factory                 func() *schema.Resource
	TypeName                string
	Name                    string
	Tags                    *ServicePackageResourceTags
	IsRegionOverrideEnabled bool // Does the resource support per-resource Region override?
}
