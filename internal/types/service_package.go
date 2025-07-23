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
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
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
	Import   FrameworkImport
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
	Import   SDKv2Import
}

type Identity struct {
	IsGlobalResource       bool   // All
	IsSingleton            bool   // Singleton
	IsARN                  bool   // ARN
	IsGlobalARNFormat      bool   // ARN
	IdentityAttribute      string // ARN
	IDAttrShadowsAttr      string
	Attributes             []IdentityAttribute
	IdentityDuplicateAttrs []string
	IsSingleParameter      bool
	IsMutable              bool
}

func (i Identity) HasInherentRegion() bool {
	if i.IsGlobalResource {
		return false
	}
	if i.IsSingleton {
		return true
	}
	if i.IsARN && !i.IsGlobalARNFormat {
		return true
	}
	return false
}

func RegionalParameterizedIdentity(attributes []IdentityAttribute, opts ...IdentityOptsFunc) Identity {
	baseAttributes := []IdentityAttribute{
		StringIdentityAttribute("account_id", false),
		StringIdentityAttribute("region", false),
	}
	baseAttributes = slices.Grow(baseAttributes, len(attributes))
	identity := Identity{
		Attributes: append(baseAttributes, attributes...),
	}
	if len(attributes) == 1 {
		identity.IDAttrShadowsAttr = attributes[0].Name()
	}

	for _, opt := range opts {
		opt(&identity)
	}

	return identity
}

type IdentityAttribute struct {
	name                  string
	required              bool
	resourceAttributeName string
}

func (ia IdentityAttribute) Name() string {
	return ia.name
}

func (ia IdentityAttribute) Required() bool {
	return ia.required
}

func (ia IdentityAttribute) ResourceAttributeName() string {
	if ia.resourceAttributeName == "" {
		return ia.name
	}
	return ia.resourceAttributeName
}

func StringIdentityAttribute(name string, required bool) IdentityAttribute {
	return IdentityAttribute{
		name:     name,
		required: required,
	}
}

func StringIdentityAttributeWithMappedName(name string, required bool, resourceAttributeName string) IdentityAttribute {
	return IdentityAttribute{
		name:                  name,
		required:              required,
		resourceAttributeName: resourceAttributeName,
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
		IsARN:             true,
		IsGlobalARNFormat: isGlobalResource,
		IdentityAttribute: name,
		Attributes: []IdentityAttribute{
			StringIdentityAttribute(name, true),
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
	identity.Attributes = slices.Insert(identity.Attributes, 0,
		StringIdentityAttribute("region", false),
	)

	return identity
}

func RegionalSingleParameterIdentity(name string, opts ...IdentityOptsFunc) Identity {
	identity := Identity{
		Attributes: []IdentityAttribute{
			StringIdentityAttribute("account_id", false),
			StringIdentityAttribute("region", false),
			StringIdentityAttribute(name, true),
		},
		IsSingleParameter: true,
	}

	for _, opt := range opts {
		opt(&identity)
	}

	return identity
}

func RegionalSingleParameterIdentityWithMappedName(name string, resourceAttributeName string, opts ...IdentityOptsFunc) Identity {
	identity := Identity{
		Attributes: []IdentityAttribute{
			StringIdentityAttribute("account_id", false),
			StringIdentityAttribute("region", false),
			StringIdentityAttributeWithMappedName(name, true, resourceAttributeName),
		},
		IsSingleParameter: true,
	}

	for _, opt := range opts {
		opt(&identity)
	}

	return identity
}

func GlobalSingleParameterIdentity(name string, opts ...IdentityOptsFunc) Identity {
	identity := Identity{
		IsGlobalResource: true,
		Attributes: []IdentityAttribute{
			StringIdentityAttribute("account_id", false),
			StringIdentityAttribute(name, true),
		},
		IsSingleParameter: true,
	}

	for _, opt := range opts {
		opt(&identity)
	}

	return identity
}

func GlobalSingleParameterIdentityWithMappedName(name string, resourceAttributeName string, opts ...IdentityOptsFunc) Identity {
	identity := Identity{
		IsGlobalResource: true,
		Attributes: []IdentityAttribute{
			StringIdentityAttribute("account_id", false),
			StringIdentityAttributeWithMappedName(name, true, resourceAttributeName),
		},
		IsSingleParameter: true,
	}

	for _, opt := range opts {
		opt(&identity)
	}

	return identity
}

func GlobalParameterizedIdentity(attributes []IdentityAttribute, opts ...IdentityOptsFunc) Identity {
	baseAttributes := []IdentityAttribute{
		StringIdentityAttribute("account_id", false),
	}
	baseAttributes = slices.Grow(baseAttributes, len(attributes))
	identity := Identity{
		IsGlobalResource: true,
		Attributes:       append(baseAttributes, attributes...),
	}
	if len(attributes) == 1 {
		identity.IDAttrShadowsAttr = attributes[0].Name()
	}

	for _, opt := range opts {
		opt(&identity)
	}

	return identity
}

func GlobalSingletonIdentity(opts ...IdentityOptsFunc) Identity {
	identity := Identity{
		IsGlobalResource: true,
		IsSingleton:      true,
		Attributes: []IdentityAttribute{
			StringIdentityAttribute("account_id", false),
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
		IsSingleton:      true,
		Attributes: []IdentityAttribute{
			StringIdentityAttribute("account_id", false),
			StringIdentityAttribute("region", false),
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

// WithV6_0SDKv2Fix is for use ONLY for resource types affected by the v6.0 SDKv2 existing resource issue
func WithV6_0SDKv2Fix() IdentityOptsFunc {
	return func(opts *Identity) {
		opts.IsMutable = true
	}
}

// WithIdentityFix is for use ONLY for resource types that must be able to modify Resource Identity due to an error
func WithIdentityFix() IdentityOptsFunc {
	return func(opts *Identity) {
		opts.IsMutable = true
	}
}

type ImportIDParser interface {
	Parse(id string) (string, map[string]string, error)
}

type FrameworkImportIDCreator interface {
	Create(ctx context.Context, state tfsdk.State) string
}

type FrameworkImport struct {
	WrappedImport bool
	ImportID      ImportIDParser // Multi-Parameter
	SetIDAttr     bool
}

type SDKv2ImportID interface {
	Create(d *schema.ResourceData) string
	ImportIDParser
}

type SDKv2Import struct {
	WrappedImport bool
	ImportID      SDKv2ImportID // Multi-Parameter
}
