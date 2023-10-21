// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

var (
	_ WithMeta = (*withMeta)(nil)
)

type WithMeta interface {
	Meta() *conns.AWSClient
}

type withMeta struct {
	meta *conns.AWSClient
}

// Meta returns the provider Meta (instance data).
func (w *withMeta) Meta() *conns.AWSClient {
	return w.meta
}

// RegionalARN returns a regional ARN for the specified service namespace and resource.
func (w *withMeta) RegionalARN(service, resource string) string {
	return arn.ARN{
		Partition: w.meta.Partition,
		Service:   service,
		Region:    w.meta.Region,
		AccountID: w.meta.AccountID,
		Resource:  resource,
	}.String()
}

// WithNoUpdate is intended to be embedded in resources which cannot be updated.
type WithNoUpdate struct{}

func (w *WithNoUpdate) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	response.Diagnostics.Append(diag.NewErrorDiagnostic("not supported", "This resource's Update method should not have been called"))
}

// WithNoOpUpdate is intended to be embedded in resources which have no need of a custom Update method.
// For example, resources where only `tags` can be updated and that is handled via transparent tagging.
type WithNoOpUpdate[T any] struct{}

func (w *WithNoOpUpdate[T]) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var t T

	response.Diagnostics.Append(request.Plan.Get(ctx, &t)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &t)...)
}

// WithNoOpUpdate is intended to be embedded in resources which have no need of a custom Delete method.
type WithNoOpDelete struct{}

func (w *WithNoOpDelete) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
}
