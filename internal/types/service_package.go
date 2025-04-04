// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"context"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// ServicePackageResourceRegion represents resource-level Region information.
type ServicePackageResourceRegion struct {
	IsGlobal                      bool // Is the resource global?
	IsOverrideEnabled             bool // Is per-resource Region override supported?
	IsValidateOverrideInPartition bool // Is the per-resource Region override value validated againt the configured partition?
}

// ServicePackageResourceTags represents resource-level tagging information.
type ServicePackageResourceTags struct {
	IdentifierAttribute string // The attribute for the identifier for UpdateTags etc.
	ResourceType        string // Extra resourceType parameter value for UpdateTags etc.
}

// ServicePackageEphemeralResource represents a Terraform Plugin Framework ephemeral resource
// implemented by a service package.
type ServicePackageEphemeralResource struct {
	Factory  func(context.Context) (ephemeral.EphemeralResourceWithConfigure, error)
	TypeName string
	Name     string
	Region   *ServicePackageResourceRegion
}

// ServicePackageFrameworkDataSource represents a Terraform Plugin Framework data source
// implemented by a service package.
type ServicePackageFrameworkDataSource struct {
	Factory  func(context.Context) (datasource.DataSourceWithConfigure, error)
	TypeName string
	Name     string
	Tags     *ServicePackageResourceTags
	Region   *ServicePackageResourceRegion
}

// ServicePackageFrameworkResource represents a Terraform Plugin Framework resource
// implemented by a service package.
type ServicePackageFrameworkResource struct {
	Factory  func(context.Context) (resource.ResourceWithConfigure, error)
	TypeName string
	Name     string
	Tags     *ServicePackageResourceTags
	Region   *ServicePackageResourceRegion
}

// ServicePackageSDKDataSource represents a Terraform Plugin SDK data source
// implemented by a service package.
type ServicePackageSDKDataSource struct {
	Factory  func() *schema.Resource
	TypeName string
	Name     string
	Tags     *ServicePackageResourceTags
	Region   *ServicePackageResourceRegion
}

// ServicePackageSDKResource represents a Terraform Plugin SDK resource
// implemented by a service package.
type ServicePackageSDKResource struct {
	Factory  func() *schema.Resource
	TypeName string
	Name     string
	Tags     *ServicePackageResourceTags
	Region   *ServicePackageResourceRegion
	Identity Identity
}

type Identity struct {
	Attributes []IdentityAttribute
}

func ParameterizedIdentity(attributes ...IdentityAttribute) Identity {
	baseAttributes := []IdentityAttribute{
		{
			Name:     "account_id",
			Required: false,
		},
		{
			Name:     "region",
			Required: false,
		},
	}
	baseAttributes = slices.Grow(baseAttributes, len(attributes))
	return Identity{
		Attributes: append(baseAttributes, attributes...),
	}
}

type IdentityAttribute struct {
	Name     string
	Required bool
}

func StringIdentityAttribute(name string, required bool) IdentityAttribute {
	return IdentityAttribute{
		Name:     name,
		Required: required,
	}
}

func ARNIdentity() Identity {
	return Identity{
		Attributes: []IdentityAttribute{
			{
				Name:     "arn",
				Required: true,
			},
		},
	}
}

func GlobalParameterizedIdentity(attributes ...IdentityAttribute) Identity {
	baseAttributes := []IdentityAttribute{
		{
			Name:     "account_id",
			Required: false,
		},
	}
	baseAttributes = slices.Grow(baseAttributes, len(attributes))
	return Identity{
		Attributes: append(baseAttributes, attributes...),
	}
}

func GlobalSingletonIdentity() Identity {
	return Identity{
		Attributes: []IdentityAttribute{
			{
				Name:     "account_id",
				Required: false,
			},
		},
	}
}

func RegionalSingletonIdentity() Identity {
	return Identity{
		Attributes: []IdentityAttribute{
			{
				Name:     "account_id",
				Required: false,
			},
			{
				Name:     "region",
				Required: false,
			},
		},
	}
}
