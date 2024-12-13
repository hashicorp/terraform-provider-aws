// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sdk

import (
	"context"
	"math/rand"
	"strings"
	"time"

	retry_sdkv2 "github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/maps"
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

func (sr *sweepResource) Delete(ctx context.Context, timeout time.Duration, optFns ...tfresource.OptionsFunc) error {
	ctx = tflog.SetField(ctx, "id", sr.d.Id())

	// TODO
	// TODO Once all services have moved to AWS SDK for Go v2 I _think_ we can remove this
	// TODO custom retry logic as the API clients have been configured to use Adaptive retry.
	// TODO

	jitter := time.Duration(rand.Int63n(int64(1*time.Second))) - 1*time.Second/2
	defaultOpts := []tfresource.OptionsFunc{
		tfresource.WithMinPollInterval(2*time.Second + jitter),
	}
	// Put defaults first so subsequent optFns will override them
	optFns = append(defaultOpts, optFns...)

	err := tfresource.Retry(ctx, timeout, func() *retry.RetryError {
		err := deleteResource(ctx, sr.resource, sr.d, sr.meta)

		if err != nil {
			var throttled bool
			// The throttling error codes defined by the AWS SDK for Go v2 are a superset of the
			// codes defined by v1, so use the v2 codes here.
			for _, code := range maps.Keys(retry_sdkv2.DefaultThrottleErrorCodes) {
				// The resource delete operation returns a diag.Diagnostics, so we have to do a
				// string comparison instead of checking the error code of an actual error
				if strings.Contains(err.Error(), code) {
					throttled = true
					break
				}
			}
			if throttled {
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
