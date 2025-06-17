// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"context"
	"slices"
	"unique"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// ServicePackageResourceRegion represents resource-level Region information.
type ServicePackageResourceRegion struct {
	IsOverrideEnabled             bool // Is per-resource Region override supported?
	IsValidateOverrideInPartition bool // Is the per-resource Region override value validated againt the configured partition?
}

// ResourceRegionDefault returns the default resource region configuration.
// The default is to enable per-resource Region override and validate the override value.
func ResourceRegionDefault() ServicePackageResourceRegion {
	return ServicePackageResourceRegion{
		IsOverrideEnabled:             true,
		IsValidateOverrideInPartition: true,
	}
}

// ResourceRegionDisabled returns the resource region configuration indicating that there is no per-resource Region override.
func ResourceRegionDisabled() ServicePackageResourceRegion {
	return ServicePackageResourceRegion{}
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
	Region   unique.Handle[ServicePackageResourceRegion]
}

// ServicePackageFrameworkDataSource represents a Terraform Plugin Framework data source
// implemented by a service package.
type ServicePackageFrameworkDataSource struct {
	Factory  func(context.Context) (datasource.DataSourceWithConfigure, error)
	TypeName string
	Name     string
	Tags     unique.Handle[ServicePackageResourceTags]
	Region   unique.Handle[ServicePackageResourceRegion]
}

// ServicePackageFrameworkResource represents a Terraform Plugin Framework resource
// implemented by a service package.
type ServicePackageFrameworkResource struct {
	Factory  func(context.Context) (resource.ResourceWithConfigure, error)
	TypeName string
	Name     string
	Tags     unique.Handle[ServicePackageResourceTags]
	Region   unique.Handle[ServicePackageResourceRegion]
	Identity Identity
	Import   Import
}

// ServicePackageSDKDataSource represents a Terraform Plugin SDK data source
// implemented by a service package.
type ServicePackageSDKDataSource struct {
	Factory  func() *schema.Resource
	TypeName string
	Name     string
	Tags     unique.Handle[ServicePackageResourceTags]
	Region   unique.Handle[ServicePackageResourceRegion]
}

// ServicePackageSDKResource represents a Terraform Plugin SDK resource
// implemented by a service package.
type ServicePackageSDKResource struct {
	Factory  func() *schema.Resource
	TypeName string
	Name     string
	Tags     unique.Handle[ServicePackageResourceTags]
	Region   unique.Handle[ServicePackageResourceRegion]
	Identity Identity
	Import   Import
}

type Identity struct {
	IsGlobalResource       bool   // All
	Singleton              bool   // Singleton
	ARN                    bool   // ARN
	IsGlobalARNFormat      bool   // ARN
	IdentityAttribute      string // ARN
	IDAttrShadowsAttr      string
	Attributes             []IdentityAttribute
	IdentityDuplicateAttrs []string
	IsSingleParameter      bool
}

func (i Identity) HasInherentRegion() bool {
	if i.IsGlobalResource {
		return false
	}
	if i.Singleton {
		return true
	}
	if i.ARN && !i.IsGlobalARNFormat {
		return true
	}
	return false
}

func RegionalParameterizedIdentity(attributes ...IdentityAttribute) Identity {
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
	identity := Identity{
		Attributes: append(baseAttributes, attributes...),
	}
	if len(attributes) == 1 {
		identity.IDAttrShadowsAttr = attributes[0].Name
	}
	return identity
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

func GlobalARNIdentity(opts ...IdentityOptsFunc) Identity {
	return GlobalARNIdentityNamed(names.AttrARN, opts...)
}

func GlobalARNIdentityNamed(name string, opts ...IdentityOptsFunc) Identity {
	return arnIdentity(true, name, opts)
}

func RegionalARNIdentity(opts ...IdentityOptsFunc) Identity {
	return RegionalARNIdentityNamed(names.AttrARN, opts...)
}

func RegionalARNIdentityNamed(name string, opts ...IdentityOptsFunc) Identity {
	return arnIdentity(false, name, opts)
}

func arnIdentity(isGlobalResource bool, name string, opts []IdentityOptsFunc) Identity {
	identity := Identity{
		IsGlobalResource:  isGlobalResource,
		ARN:               true,
		IsGlobalARNFormat: isGlobalResource,
		IdentityAttribute: name,
		Attributes: []IdentityAttribute{
			{
				Name:     name,
				Required: true,
			},
		},
	}

	for _, opt := range opts {
		opt(&identity)
	}

	return identity
}

func RegionalResourceWithGlobalARNFormat(opts ...IdentityOptsFunc) Identity {
	return RegionalResourceWithGlobalARNFormatNamed(names.AttrARN, opts...)
}

func RegionalResourceWithGlobalARNFormatNamed(name string, opts ...IdentityOptsFunc) Identity {
	identity := RegionalARNIdentityNamed(name, opts...)

	identity.IsGlobalARNFormat = true
	identity.Attributes = slices.Insert(identity.Attributes, 0, IdentityAttribute{
		Name:     "region",
		Required: false,
	})

	return identity
}

func RegionalSingleParameterIdentity(name string) Identity {
	return Identity{
		IdentityAttribute: name,
		Attributes: []IdentityAttribute{
			{
				Name:     "account_id",
				Required: false,
			},
			{
				Name:     "region",
				Required: false,
			},
			{
				Name:     name,
				Required: true,
			},
		},
		IsSingleParameter: true,
	}
}

func GlobalSingleParameterIdentity(name string) Identity {
	return Identity{
		IsGlobalResource:  true,
		IdentityAttribute: name,
		Attributes: []IdentityAttribute{
			{
				Name:     "account_id",
				Required: false,
			},
			{
				Name:     name,
				Required: true,
			},
		},
		IsSingleParameter: true,
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
	identity := Identity{
		Attributes: append(baseAttributes, attributes...),
	}
	if len(attributes) == 1 {
		identity.IDAttrShadowsAttr = attributes[0].Name
	}
	return identity
}

func GlobalSingletonIdentity(opts ...IdentityOptsFunc) Identity {
	identity := Identity{
		IsGlobalResource: true,
		Singleton:        true,
		Attributes: []IdentityAttribute{
			{
				Name:     "account_id",
				Required: false,
			},
		},
	}

	for _, opt := range opts {
		opt(&identity)
	}

	return identity
}

func RegionalSingletonIdentity(opts ...IdentityOptsFunc) Identity {
	identity := Identity{
		IsGlobalResource: false,
		Singleton:        true,
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

	for _, opt := range opts {
		opt(&identity)
	}

	return identity
}

type IdentityOptsFunc func(opts *Identity)

func WithIdentityDuplicateAttrs(attrs ...string) IdentityOptsFunc {
	return func(opts *Identity) {
		opts.IdentityDuplicateAttrs = attrs
	}
}

type Import struct {
	WrappedImport bool
}
