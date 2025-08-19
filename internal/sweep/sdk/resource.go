// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sdk

import (
	"context"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

type sweepResource struct {
	d        *schema.ResourceData
	meta     *conns.AWSClient
	resource *schema.Resource
}

func NewSweepResource(resource *schema.Resource, d *schema.ResourceData, meta *conns.AWSClient) *sweepResource {
	s := newSweepResource(resource, d, meta)
	return &s
}

func newSweepResource(resource *schema.Resource, d *schema.ResourceData, meta *conns.AWSClient) sweepResource {
	return sweepResource{
		d:        d,
		meta:     meta,
		resource: resource,
	}
}

func (sr *sweepResource) Delete(ctx context.Context, optFns ...tfresource.OptionsFunc) error {
	ctx = tflog.SetField(ctx, "id", sr.d.Id())

	return deleteResource(ctx, sr.resource, sr.d, sr.meta)
}

type readerSweepResource struct {
	sweepResource
}

func NewReaderSweepResource(resource *schema.Resource, d *schema.ResourceData, meta *conns.AWSClient) *readerSweepResource {
	return &readerSweepResource{
		sweepResource: newSweepResource(resource, d, meta),
	}
}

func (rsr *readerSweepResource) Read(ctx context.Context) error {
	ctx = tflog.SetField(ctx, "id", rsr.d.Id())

	return ReadResource(ctx, rsr.resource, rsr.d, rsr.meta)
}

func deleteResource(ctx context.Context, resource *schema.Resource, d *schema.ResourceData, meta *conns.AWSClient) error {
	if resource.DeleteContext != nil || resource.DeleteWithoutTimeout != nil {
		var diags diag.Diagnostics

		if resource.DeleteContext != nil {
			diags = resource.DeleteContext(ctx, d, meta)
		} else {
			diags = resource.DeleteWithoutTimeout(ctx, d, meta)
		}

		return sdkdiag.DiagnosticsError(diags)
	}

	return resource.Delete(d, meta)
}

func ReadResource(ctx context.Context, resource *schema.Resource, d *schema.ResourceData, meta *conns.AWSClient) error {
	if resource.ReadContext != nil || resource.ReadWithoutTimeout != nil {
		var diags diag.Diagnostics

		if resource.ReadContext != nil {
			diags = resource.ReadContext(ctx, d, meta)
		} else {
			diags = resource.ReadWithoutTimeout(ctx, d, meta)
		}

		return sdkdiag.DiagnosticsError(diags)
	}

	return resource.Read(d, meta)
}
