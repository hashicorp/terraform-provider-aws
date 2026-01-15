// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sdkv2

import (
	"context"
	"errors"
	"slices"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

type awsClient interface {
	AccountID(ctx context.Context) string
	Region(ctx context.Context) string
	DefaultTagsConfig(ctx context.Context) *tftags.DefaultConfig
	IgnoreTagsConfig(ctx context.Context) *tftags.IgnoreConfig
	Partition(context.Context) string
	ServicePackage(_ context.Context, name string) conns.ServicePackage
	TagPolicyConfig(context.Context) *tftags.TagPolicyConfig
	ValidateInContextRegionInPartition(ctx context.Context) error
	AwsConfig(context.Context) aws.Config
}

// schemaResourceData is an interface that implements a subset of schema.ResourceData's public methods.
type schemaResourceData interface {
	sdkv2.ResourceDiffer
	Set(string, any) error
	Identity() (*schema.IdentityData, error)
}

type interceptorOptions[D any] struct {
	c    awsClient
	d    D
	when when
	why  why
}

type (
	crudInterceptorOptions          = interceptorOptions[schemaResourceData]
	customizeDiffInterceptorOptions = interceptorOptions[*schema.ResourceDiff]
	importInterceptorOptions        = interceptorOptions[*schema.ResourceData]
)

// An interceptor is functionality invoked during a request's lifecycle.
// If a Before interceptor returns Diagnostics indicating an error occurred then
// no further interceptors in the chain are run and neither is the schema's method.
// In other cases all interceptors in the chain are run.

type interceptor1[D, E any] interface {
	run(context.Context, interceptorOptions[D]) E
}

type (
	// crudInterceptor is functionality invoked during a CRUD request lifecycle.
	crudInterceptor = interceptor1[schemaResourceData, diag.Diagnostics]
	// customizeDiffInterceptor is functionality invoked during a CustomizeDiff request lifecycle.
	customizeDiffInterceptor = interceptor1[*schema.ResourceDiff, error]
	// importInterceptor is functionality invoked during an Import request lifecycle.
	importInterceptor = interceptor1[*schema.ResourceData, error]
)

type interceptorFunc1[D, E any] func(context.Context, interceptorOptions[D]) E

func (f interceptorFunc1[D, E]) run(ctx context.Context, opts interceptorOptions[D]) E { //nolint:unused // used via crudInterceptor/customizeDiffInterceptor/importInterceptor
	return f(ctx, opts)
}

// interceptorInvocation represents a single interceptor invocation.
type interceptorInvocation struct {
	when        when
	why         why
	interceptor any
}

type typedInterceptorInvocation[D, E any] struct {
	when        when
	why         why
	interceptor interceptor1[D, E]
}

type (
	crudInterceptorInvocation          = typedInterceptorInvocation[schemaResourceData, diag.Diagnostics]
	customizeDiffInterceptorInvocation = typedInterceptorInvocation[*schema.ResourceDiff, error]
	importInterceptorInvocation        = typedInterceptorInvocation[*schema.ResourceData, error]
)

// Only generate strings for use in tests
//go:generate stringer -type=when -output=when_string_test.go

// when represents the point in the request lifecycle that an interceptor is run.
// Multiple values can be ORed together.
type when uint16

const (
	Before  when = 1 << iota // Interceptor is invoked before call to method in schema
	After                    // Interceptor is invoked after successful call to method in schema
	OnError                  // Interceptor is invoked after unsuccessful call to method in schema
	Finally                  // Interceptor is invoked after After or OnError
)

// why represents the operation(s) that an interceptor is run.
// Multiple values can be ORed together.
type why uint16

const (
	Create        why = 1 << iota // Interceptor is invoked for a Create call
	Read                          // Interceptor is invoked for a Read call
	Update                        // Interceptor is invoked for an Update call
	Delete                        // Interceptor is invoked for a Delete call
	CustomizeDiff                 // Interceptor is invoked for a CustomizeDiff call
	Import                        // Interceptor is invoked for an Import call

	AllCRUDOps = Create | Read | Update | Delete // Interceptor is invoked for all CRUD calls
)

type interceptorInvocations []interceptorInvocation

func (s interceptorInvocations) why(why why) interceptorInvocations {
	return tfslices.Filter(s, func(e interceptorInvocation) bool {
		return e.why&why != 0
	})
}

// interceptedCRUDHandler returns a handler that invokes the specified CRUD handler, running any interceptors.
func interceptedCRUDHandler[F ~func(context.Context, *schema.ResourceData, any) diag.Diagnostics](bootstrapContext contextFunc, interceptorInvocations interceptorInvocations, f F, why why) F {
	// We don't run CRUD interceptors if the resource has not defined a corresponding handler function.
	if f == nil {
		return nil
	}

	return func(ctx context.Context, rd *schema.ResourceData, meta any) diag.Diagnostics {
		var diags diag.Diagnostics

		ctx, err := bootstrapContext(ctx, rd.GetOk, rd.GetProviderMeta, meta)
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		var interceptors []crudInterceptorInvocation
		for _, v := range interceptorInvocations.why(why) {
			if interceptor, ok := v.interceptor.(crudInterceptor); ok {
				interceptors = append(interceptors, crudInterceptorInvocation{
					when:        v.when,
					why:         v.why,
					interceptor: interceptor,
				})
			}
		}

		opts := crudInterceptorOptions{
			c:   meta.(awsClient),
			d:   rd,
			why: why,
		}

		// Before interceptors are run first to last.
		opts.when = Before
		for v := range slices.Values(interceptors) {
			if v.when&opts.when != 0 {
				diags = append(diags, v.interceptor.run(ctx, opts)...)

				// Short circuit if any Before interceptor errors.
				if diags.HasError() {
					return diags
				}
			}
		}

		d := f(ctx, rd, meta)
		diags = append(diags, d...)

		// All other interceptors are run last to first.
		if d.HasError() {
			opts.when = OnError
		} else {
			opts.when = After
		}
		for v := range tfslices.BackwardValues(interceptors) {
			if v.when&opts.when != 0 {
				diags = append(diags, v.interceptor.run(ctx, opts)...)
			}
		}

		opts.when = Finally
		for v := range tfslices.BackwardValues(interceptors) {
			if v.when&opts.when != 0 {
				diags = append(diags, v.interceptor.run(ctx, opts)...)
			}
		}

		return diags
	}
}

// interceptedCustomizeDiffHandler returns a handler that invokes the specified CustomizeDiff handler, running any interceptors.
func interceptedCustomizeDiffHandler(bootstrapContext contextFunc, interceptorInvocations interceptorInvocations, f schema.CustomizeDiffFunc) schema.CustomizeDiffFunc {
	// We run CustomizeDiff interceptors even if the resource has not defined a CustomizeDiff function.
	return func(ctx context.Context, d *schema.ResourceDiff, meta any) error {
		ctx, err := bootstrapContext(ctx, d.GetOk, nil, meta)
		if err != nil {
			return err
		}

		why := CustomizeDiff

		var interceptors []customizeDiffInterceptorInvocation
		for _, v := range interceptorInvocations.why(why) {
			if interceptor, ok := v.interceptor.(customizeDiffInterceptor); ok {
				interceptors = append(interceptors, customizeDiffInterceptorInvocation{
					when:        v.when,
					why:         v.why,
					interceptor: interceptor,
				})
			}
		}

		opts := customizeDiffInterceptorOptions{
			c:   meta.(awsClient),
			d:   d,
			why: why,
		}

		// Before interceptors are run first to last.
		opts.when = Before
		for v := range slices.Values(interceptors) {
			if v.when&opts.when != 0 {
				// Short circuit if any Before interceptor errors.
				if err := v.interceptor.run(ctx, opts); err != nil {
					return err
				}
			}
		}

		var errs []error

		opts.when = After
		if f != nil {
			if err := f(ctx, d, meta); err != nil {
				opts.when = OnError
				errs = append(errs, err)
			}
		}

		// All other interceptors are run last to first.
		for v := range tfslices.BackwardValues(interceptors) {
			if v.when&opts.when != 0 {
				if err := v.interceptor.run(ctx, opts); err != nil {
					errs = append(errs, err)
				}
			}
		}

		opts.when = Finally
		for v := range tfslices.BackwardValues(interceptors) {
			if v.when&opts.when != 0 {
				if err := v.interceptor.run(ctx, opts); err != nil {
					errs = append(errs, err)
				}
			}
		}

		return errors.Join(errs...)
	}
}

// interceptedImportHandler returns a handler that invokes the specified Imort handler, running any interceptors.
func interceptedImportHandler(bootstrapContext contextFunc, interceptorInvocations interceptorInvocations, f schema.StateContextFunc) schema.StateContextFunc {
	// We don't run Import interceptors if the resource has not defined a corresponding handler function.
	if f == nil {
		return nil
	}

	return func(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
		ctx, err := bootstrapContext(ctx, d.GetOk, nil, meta)
		if err != nil {
			return nil, err
		}

		why := Import

		var interceptors []importInterceptorInvocation
		for _, v := range interceptorInvocations.why(why) {
			if interceptor, ok := v.interceptor.(importInterceptor); ok {
				interceptors = append(interceptors, importInterceptorInvocation{
					when:        v.when,
					why:         v.why,
					interceptor: interceptor,
				})
			}
		}

		opts := importInterceptorOptions{
			c:   meta.(awsClient),
			d:   d,
			why: why,
		}

		// Before interceptors are run first to last.
		opts.when = Before
		for v := range slices.Values(interceptors) {
			if v.when&opts.when != 0 {
				// Short circuit if any Before interceptor errors.
				if err := v.interceptor.run(ctx, opts); err != nil {
					return nil, err
				}
			}
		}

		var errs []error

		r, err := f(ctx, d, meta)
		if err != nil {
			opts.when = OnError
			errs = append(errs, err)
		} else {
			opts.when = After
		}

		// All other interceptors are run last to first.
		for v := range tfslices.BackwardValues(interceptors) {
			if v.when&opts.when != 0 {
				if err := v.interceptor.run(ctx, opts); err != nil {
					errs = append(errs, err)
				}
			}
		}

		opts.when = Finally
		for v := range tfslices.BackwardValues(interceptors) {
			if v.when&opts.when != 0 {
				if err := v.interceptor.run(ctx, opts); err != nil {
					errs = append(errs, err)
				}
			}
		}

		return r, errors.Join(errs...)
	}
}
