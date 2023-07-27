// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sdk

import (
	"context"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
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
	return &sweepResource{
		d:        d,
		meta:     meta,
		resource: resource,
	}
}

func (sr *sweepResource) Delete(ctx context.Context, timeout time.Duration, optFns ...tfresource.OptionsFunc) error {
	ctx = tflog.SetField(ctx, "id", sr.d.Id())

	err := tfresource.Retry(ctx, timeout, func() *retry.RetryError {
		err := deleteResource(ctx, sr.resource, sr.d, sr.meta)

		if err != nil {
			if strings.Contains(err.Error(), "Throttling") {
				tflog.Info(ctx, "Retrying throttling error", map[string]any{
					"err": err.Error(),
				})
				return retry.RetryableError(err)
			}

			return retry.NonRetryableError(err)
		}

		return nil
	}, optFns...)

	if tfresource.TimedOut(err) {
		err = deleteResource(ctx, sr.resource, sr.d, sr.meta)
	}

	return err
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

// Deprecated: Create a list of Sweepables and pass them to SweepOrchestrator instead
func DeleteResource(ctx context.Context, resource *schema.Resource, d *schema.ResourceData, meta *conns.AWSClient) error {
	return deleteResource(ctx, resource, d, meta)
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
