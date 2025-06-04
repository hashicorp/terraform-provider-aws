// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sdkv2

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func resourceValidateRegion() customizeDiffInterceptor {
	return interceptorFunc1[*schema.ResourceDiff, error](func(ctx context.Context, opts customizeDiffInterceptorOptions) error {
		c := opts.c

		switch when, why := opts.when, opts.why; when {
		case Before:
			switch why {
			case CustomizeDiff:
				return c.ValidateInContextRegionInPartition(ctx)
			}
		}

		return nil
	})
}

func dataSourceValidateRegion() crudInterceptor {
	return interceptorFunc1[schemaResourceData, diag.Diagnostics](func(ctx context.Context, opts crudInterceptorOptions) diag.Diagnostics {
		c := opts.c
		var diags diag.Diagnostics

		switch when, why := opts.when, opts.why; when {
		case Before:
			switch why {
			case Read:
				// As data sources have no CustomizeDiff functionality, we validate the per-resource Region override value here.
				if err := c.ValidateInContextRegionInPartition(ctx); err != nil {
					return sdkdiag.AppendFromErr(diags, err)
				}
			}
		}

		return diags
	})
}

func defaultRegion() customizeDiffInterceptor {
	return interceptorFunc1[*schema.ResourceDiff, error](func(ctx context.Context, opts customizeDiffInterceptorOptions) error {
		c := opts.c

		switch d, when, why := opts.d, opts.when, opts.why; when {
		case Before:
			switch why {
			case CustomizeDiff:
				// Set the value of the top-level `region` attribute to the provider's configured Region if it is not set.
				if _, ok := d.GetOk(names.AttrRegion); !ok {
					return d.SetNew(names.AttrRegion, c.AwsConfig(ctx).Region)
				}
			}
		}

		return nil
	})
}

func setRegionInState() crudInterceptor {
	return interceptorFunc1[schemaResourceData, diag.Diagnostics](func(ctx context.Context, opts crudInterceptorOptions) diag.Diagnostics {
		c := opts.c
		var diags diag.Diagnostics

		switch d, when, why := opts.d, opts.when, opts.why; when {
		case After:
			// Set region in state after R.
			switch why {
			case Read:
				if err := d.Set(names.AttrRegion, c.Region(ctx)); err != nil {
					return sdkdiag.AppendErrorf(diags, "setting %s: %s", names.AttrRegion, err)
				}
			}
		}

		return diags
	})
}

func forceNewIfRegionChanges() customizeDiffInterceptor {
	return interceptorFunc1[*schema.ResourceDiff, error](func(ctx context.Context, opts customizeDiffInterceptorOptions) error {
		c := opts.c

		switch d, when, why := opts.d, opts.when, opts.why; when {
		case Before:
			switch why {
			case CustomizeDiff:
				// Force resource replacement if the value of the top-level `region` attribute changes.
				if d.Id() != "" && d.HasChange(names.AttrRegion) {
					providerRegion := c.AwsConfig(ctx).Region
					o, n := d.GetChange(names.AttrRegion)
					if o, n := o.(string), n.(string); (o == "" && n == providerRegion) || (o == providerRegion && n == "") {
						return nil
					}
					return d.ForceNew(names.AttrRegion)
				}
			}
		}

		return nil
	})
}

func importRegion() importInterceptor {
	return interceptorFunc2[*schema.ResourceData, []*schema.ResourceData, error](func(ctx context.Context, opts importInterceptorOptions) ([]*schema.ResourceData, error) {
		c, d := opts.c, opts.d

		switch when, why := opts.when, opts.why; when {
		case Before:
			switch why {
			case Import:
				// Import ID optionally ends with "@<region>".
				if matches := regexache.MustCompile(`^(.+)@([a-z]{2}(?:-[a-z]+)+-\d{1,2})$`).FindStringSubmatch(d.Id()); len(matches) == 3 {
					d.SetId(matches[1])
					d.Set(names.AttrRegion, matches[2])
				} else {
					d.Set(names.AttrRegion, c.AwsConfig(ctx).Region)
				}
			}
		}

		return []*schema.ResourceData{d}, nil
	})
}

func resourceImportRegion() interceptorInvocation {
	return interceptorInvocation{
		when:        Before,
		why:         Import,
		interceptor: importRegion(),
	}
}

// importRegionNoDefault does not provide a default value for `region`. This should be used when the import ID is or contains a region.
func importRegionNoDefault() importInterceptor {
	return interceptorFunc2[*schema.ResourceData, []*schema.ResourceData, error](func(ctx context.Context, opts importInterceptorOptions) ([]*schema.ResourceData, error) {
		d := opts.d

		switch when, why := opts.when, opts.why; when {
		case Before:
			switch why {
			case Import:
				// Import ID optionally ends with "@<region>".
				if matches := regexache.MustCompile(`^(.+)@([a-z]{2}(?:-[a-z]+)+-\d{1,2})$`).FindStringSubmatch(d.Id()); len(matches) == 3 {
					d.SetId(matches[1])
					d.Set(names.AttrRegion, matches[2])
				}
			}
		}

		return []*schema.ResourceData{d}, nil
	})
}

func resourceImportRegionNoDefault() interceptorInvocation {
	return interceptorInvocation{
		when:        Before,
		why:         Import,
		interceptor: importRegionNoDefault(),
	}
}
