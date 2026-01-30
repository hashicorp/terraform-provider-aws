// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"context"
	"slices"
	"unique"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// ServicePackageResourceRegion represents resource-level Region information.
type ServicePackageResourceRegion struct {
	IsOverrideEnabled             bool // Is per-resource Region override supported?
	IsValidateOverrideInPartition bool // Is the per-resource Region override value validated againt the configured partition?
	IsOverrideDeprecated          bool // Is per-resource Region override deprecated? i.e. the resource type is actually global
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

// ResourceRegionDeprecatedOverride returns the resource region configuration indicating that per-resource Region override is enabled but deprecated.
func ResourceRegionDeprecatedOverride() ServicePackageResourceRegion {
	return ServicePackageResourceRegion{
		IsOverrideEnabled:             true,
		IsValidateOverrideInPartition: true,
		IsOverrideDeprecated:          true,
	}
}

// ServicePackageResourceTags represents resource-level tagging information.
type ServicePackageResourceTags struct {
	IdentifierAttribute string // The attribute for the identifier for UpdateTags etc.
	ResourceType        string // Extra resourceType parameter value for UpdateTags etc.
}

// ServicePackageAction represents a Terraform Plugin Framework action
// implemented by a service package.
type ServicePackageAction struct {
	Factory  func(context.Context) (action.ActionWithConfigure, error)
	TypeName string
	Name     string
	Region   unique.Handle[ServicePackageResourceRegion]
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

type ServicePackageFrameworkListResource struct {
	Factory  func() list.ListResourceWithConfigure
	TypeName string
	Name     string
	Tags     unique.Handle[ServicePackageResourceTags]
	Region   unique.Handle[ServicePackageResourceRegion]
	Identity Identity
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

type ListResourceForSDK interface {
	list.ListResourceWithRawV5Schemas
	list.ListResourceWithConfigure
}

type ServicePackageSDKListResource struct {
	Factory  func() ListResourceForSDK
	TypeName string
	Name     string
	Tags     unique.Handle[ServicePackageResourceTags]
	Region   unique.Handle[ServicePackageResourceRegion]
	Identity Identity
}

type Identity struct {
	IsGlobalResource           bool   // All
	IsSingleton                bool   // Singleton
	IsARN                      bool   // ARN
	IsGlobalARNFormat          bool   // ARN
	IdentityAttribute          string // ARN
	IDAttrShadowsAttr          string
	Attributes                 []IdentityAttribute
	IdentityDuplicateAttrs     []string
	IsSingleParameter          bool
	IsMutable                  bool
	IsSetOnUpdate              bool
	IsCustomInherentRegion     bool
	customInherentRegionParser RegionalCustomInherentRegionIdentityFunc
	version                    int64
	sdkv2IdentityUpgraders     []schema.IdentityUpgrader
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
	if i.IsCustomInherentRegion {
		return true
	}
	return false
}

func (i Identity) Version() int64 {
	return i.version
}

func (i Identity) SDKv2IdentityUpgraders() []schema.IdentityUpgrader {
	return i.sdkv2IdentityUpgraders
}

func (i Identity) CustomInherentRegionParser() RegionalCustomInherentRegionIdentityFunc {
	return i.customInherentRegionParser
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

func RegionalCustomInherentRegionIdentity(name string, parser RegionalCustomInherentRegionIdentityFunc, opts ...IdentityOptsFunc) Identity {
	identity := Identity{
		IsGlobalResource:  false,
		IdentityAttribute: name,
		Attributes: []IdentityAttribute{
			StringIdentityAttribute(name, true),
		},
		IsCustomInherentRegion:     true,
		customInherentRegionParser: parser,
	}

	for _, opt := range opts {
		opt(&identity)
	}

	return identity
}

type BaseIdentity struct {
	AccountID string
	Region    string
}

type RegionalCustomInherentRegionIdentityFunc func(value string) (BaseIdentity, error)

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

// WithMutableIdentity is for use for resource types that normally have a mutable identity
// If Identity must be mutable to fix potential errors, use WithIdentityFix()
func WithMutableIdentity() IdentityOptsFunc {
	return func(opts *Identity) {
		opts.IsMutable = true
		opts.IsSetOnUpdate = true
	}
}

// WithIdentityFix is for use ONLY for resource types that must be able to modify Resource Identity due to an error
func WithIdentityFix() IdentityOptsFunc {
	return func(opts *Identity) {
		opts.IsMutable = true
	}
}

// WithV6_0SDKv2Fix is for use ONLY for resource types affected by the v6.0 SDKv2 existing resource issue
func WithV6_0SDKv2Fix() IdentityOptsFunc {
	return func(opts *Identity) {
		opts.IsMutable = true
	}
}

func WithVersion(version int64) IdentityOptsFunc {
	return func(opts *Identity) {
		opts.version = version
	}
}

func WithSDKv2IdentityUpgraders(identityUpgraders ...schema.IdentityUpgrader) IdentityOptsFunc {
	return func(opts *Identity) {
		opts.sdkv2IdentityUpgraders = identityUpgraders
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
	CustomImport  bool
	ImportID      SDKv2ImportID // Multi-Parameter
}
