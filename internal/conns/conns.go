// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package conns

import (
	"context"

	"github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/vcr"
)

// ServicePackage is the minimal interface exported from each AWS service package.
// Its methods return the Plugin SDK and Framework resources and data sources implemented in the package.
type ServicePackage interface {
	FrameworkDataSources(context.Context) []*types.ServicePackageFrameworkDataSource
	FrameworkResources(context.Context) []*types.ServicePackageFrameworkResource
	SDKDataSources(context.Context) []*types.ServicePackageSDKDataSource
	SDKResources(context.Context) []*types.ServicePackageSDKResource
	ServicePackageName() string
}

// ServicePackageWithEphemeralResources is an interface that extends ServicePackage with ephemeral resources.
// Ephemeral resources are resources that are not part of the Terraform state, but are used to create other resources.
type ServicePackageWithEphemeralResources interface {
	ServicePackage
	EphemeralResources(context.Context) []*types.ServicePackageEphemeralResource
}

type (
	contextKeyType int
)

var (
	contextKey contextKeyType
)

// InContext represents the resource information kept in Context.
type InContext struct {
	overrideRegion     string // Any currently in effect per-resource Region override.
	resourceName       string // Friendly resource name, e.g. "Subnet"
	servicePackageName string // Canonical name defined as a constant in names package
	vcrEnabled         bool   // Whether VCR testing is enabled
}

// OverrideRegion returns any currently in effect per-resource Region override.
func (c *InContext) OverrideRegion() string {
	return c.overrideRegion
}

// ResourceName returns the friendly resource name, e.g. "Subnet".
func (c *InContext) ResourceName() string {
	return c.resourceName
}

// ServicePackageName returns the canonical service name defined as a constant in the `names` package.
func (c *InContext) ServicePackageName() string {
	return c.servicePackageName
}

// VCREnabled indicates whether VCR testing is enabled.
func (c *InContext) VCREnabled() bool {
	return c.vcrEnabled
}

func NewResourceContext(ctx context.Context, servicePackageName, resourceName, overrideRegion string) context.Context {
	v := InContext{
		overrideRegion:     overrideRegion,
		resourceName:       resourceName,
		servicePackageName: servicePackageName,
		vcrEnabled:         vcr.IsEnabled(),
	}

	return context.WithValue(ctx, contextKey, &v)
}

func FromContext(ctx context.Context) (*InContext, bool) {
	v, ok := ctx.Value(contextKey).(*InContext)
	return v, ok
}
