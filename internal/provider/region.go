// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func validateRegionInPartition(ctx context.Context, c *conns.AWSClient, region string) error {
	if got, want := names.PartitionForRegion(region).ID(), c.Partition(ctx); got != want {
		return fmt.Errorf("partition (%s) for per-resource Region (%s) is not the provider's configured partition (%s)", got, region, want)
	}

	return nil
}

type verifyRegionValueInConfiguredPartitionInterceptor struct{}

func verifyRegionValueInConfiguredPartition() customizeDiffInterceptor {
	return &verifyRegionValueInConfiguredPartitionInterceptor{}
}

func (r verifyRegionValueInConfiguredPartitionInterceptor) run(ctx context.Context, opts customizeDiffInterceptorOptions) error {
	c := opts.c

	switch d, when, why := opts.d, opts.when, opts.why; when {
	case Before:
		switch why {
		case CustomizeDiff:
			// Verify that the value of the top-level `region` attribute is in the configured AWS partition.
			if v, ok := d.GetOk(names.AttrRegion); ok {
				if err := validateRegionInPartition(ctx, c, v.(string)); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

type defaultRegionValueInterceptor struct{}

func defaultRegionValue() customizeDiffInterceptor {
	return &defaultRegionValueInterceptor{}
}

func (r defaultRegionValueInterceptor) run(ctx context.Context, opts customizeDiffInterceptorOptions) error {
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
}

type newRegionValueInterceptor struct{}

func newRegionValue() crudInterceptor {
	return &newRegionValueInterceptor{}
}

// TODO REGION This is the same logic as in the data source interceptor.
func (r newRegionValueInterceptor) run(ctx context.Context, opts crudInterceptorOptions) diag.Diagnostics {
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
}

type forceNewIfRegionValueChangesInterceptor struct{}

func forceNewIfRegionValueChanges() customizeDiffInterceptor {
	return &forceNewIfRegionValueChangesInterceptor{}
}

func (r forceNewIfRegionValueChangesInterceptor) run(ctx context.Context, opts customizeDiffInterceptorOptions) error {
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
}

type importRegionInterceptor struct{}

func importRegion() importInterceptor {
	return &importRegionInterceptor{}
}

func (r importRegionInterceptor) run(ctx context.Context, opts importInterceptorOptions) ([]*schema.ResourceData, error) {
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
}

// regionDataSourceCRUDInterceptor implements per-resource Region override functionality on CRUD operations for data sources.
type regionDataSourceCRUDInterceptor struct {
	validateRegionInPartition bool
}

func newRegionDataSourceCRUDInterceptor(validateRegionInPartition bool) crudInterceptor {
	return &regionDataSourceCRUDInterceptor{
		validateRegionInPartition: validateRegionInPartition,
	}
}

func (r regionDataSourceCRUDInterceptor) run(ctx context.Context, opts crudInterceptorOptions) diag.Diagnostics {
	c := opts.c
	var diags diag.Diagnostics

	switch d, when, why := opts.d, opts.when, opts.why; when {
	case Before:
		switch why {
		case Read:
			// As data sources have no CustomizeDiff functionality we validate the per-resource Region override value here.
			if r.validateRegionInPartition {
				if inContext, ok := conns.FromContext(ctx); ok {
					if v := inContext.OverrideRegion(); v != "" {
						if err := validateRegionInPartition(ctx, c, v); err != nil {
							return sdkdiag.AppendFromErr(diags, err)
						}
					}
				}
			}
		}
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
}
