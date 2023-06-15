package sweep

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

type SweepResource struct {
	d        *schema.ResourceData
	meta     *conns.AWSClient
	resource *schema.Resource
}

func NewSweepResource(resource *schema.Resource, d *schema.ResourceData, meta *conns.AWSClient) *SweepResource {
	return &SweepResource{
		d:        d,
		meta:     meta,
		resource: resource,
	}
}

func (sr *SweepResource) Delete(ctx context.Context, timeout time.Duration, optFns ...tfresource.OptionsFunc) error {
	err := tfresource.Retry(ctx, timeout, func() *retry.RetryError {
		err := DeleteResource(ctx, sr.resource, sr.d, sr.meta)

		if err != nil {
			if strings.Contains(err.Error(), "Throttling") {
				log.Printf("[INFO] While sweeping resource (%s), encountered throttling error (%s). Retrying...", sr.d.Id(), err)
				return retry.RetryableError(err)
			}

			return retry.NonRetryableError(err)
		}

		return nil
	}, optFns...)

	if tfresource.TimedOut(err) {
		err = DeleteResource(ctx, sr.resource, sr.d, sr.meta)
	}

	return err
}

func DeleteResource(ctx context.Context, resource *schema.Resource, d *schema.ResourceData, meta any) error {
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

func ReadResource(ctx context.Context, resource *schema.Resource, d *schema.ResourceData, meta any) error {
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
