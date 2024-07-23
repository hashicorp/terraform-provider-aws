// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package conns

import (
	"context"

	"github.com/hashicorp/terraform-provider-aws/internal/types"
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

type (
	contextKeyType int
)

var (
	contextKey contextKeyType
)

// InContext represents the resource information kept in Context.
type InContext struct {
	IsDataSource       bool   // Data source?
	ResourceName       string // Friendly resource name, e.g. "Subnet"
	ServicePackageName string // Canonical name defined as a constant in names package
}

func NewDataSourceContext(ctx context.Context, servicePackageName, resourceName string) context.Context {
	v := InContext{
		IsDataSource:       true,
		ResourceName:       resourceName,
		ServicePackageName: servicePackageName,
	}

	return context.WithValue(ctx, contextKey, &v)
}

func NewResourceContext(ctx context.Context, servicePackageName, resourceName string) context.Context {
	v := InContext{
		ResourceName:       resourceName,
		ServicePackageName: servicePackageName,
	}

	return context.WithValue(ctx, contextKey, &v)
}

func FromContext(ctx context.Context) (*InContext, bool) {
	v, ok := ctx.Value(contextKey).(*InContext)
	return v, ok
}
