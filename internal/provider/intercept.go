// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// schemaResourceData is an interface that implements functions from schema.ResourceData
type schemaResourceData interface {
	Get(key string) any
	GetChange(key string) (any, any)
	GetRawConfig() cty.Value
	GetRawPlan() cty.Value
	GetRawState() cty.Value
	HasChange(key string) bool
	Id() string
	Set(string, any) error
}

// An interceptor is functionality invoked during the CRUD request lifecycle.
// If a Before interceptor returns Diagnostics indicating an error occurred then
// no further interceptors in the chain are run and neither is the schema's method.
// In other cases all interceptors in the chain are run.
type interceptor interface {
	run(context.Context, schemaResourceData, any, when, why, diag.Diagnostics) (context.Context, diag.Diagnostics)
}

type interceptorFunc func(context.Context, schemaResourceData, any, when, why, diag.Diagnostics) (context.Context, diag.Diagnostics)

func (f interceptorFunc) run(ctx context.Context, d schemaResourceData, meta any, when when, why why, diags diag.Diagnostics) (context.Context, diag.Diagnostics) {
	return f(ctx, d, meta, when, why, diags)
}

// interceptorItem represents a single interceptor invocation.
type interceptorItem struct {
	when        when
	why         why
	interceptor interceptor
}

// when represents the point in the CRUD request lifecycle that an interceptor is run.
// Multiple values can be ORed together.
type when uint16

const (
	Before  when = 1 << iota // Interceptor is invoked before call to method in schema
	After                    // Interceptor is invoked after successful call to method in schema
	OnError                  // Interceptor is invoked after unsuccessful call to method in schema
	Finally                  // Interceptor is invoked after After or OnError
)

// why represents the CRUD operation(s) that an interceptor is run.
// Multiple values can be ORed together.
type why uint16

const (
	Create why = 1 << iota // Interceptor is invoked for a Create call
	Read                   // Interceptor is invoked for a Read call
	Update                 // Interceptor is invoked for an Update call
	Delete                 // Interceptor is invoked for a Delete call

	AllOps = Create | Read | Update | Delete // Interceptor is invoked for all calls
)

type interceptorItems []interceptorItem

// why returns a slice of interceptors that run for the specified CRUD operation.
func (s interceptorItems) why(why why) interceptorItems {
	return slices.Filter(s, func(e interceptorItem) bool {
		return e.why&why != 0
	})
}

// interceptedHandler returns a handler that invokes the specified CRUD handler, running any interceptors.
func interceptedHandler[F ~func(context.Context, *schema.ResourceData, any) diag.Diagnostics](bootstrapContext contextFunc, interceptors interceptorItems, f F, why why) F {
	return func(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
		var diags diag.Diagnostics
		ctx = bootstrapContext(ctx, meta)
		// Before interceptors are run first to last.
		forward := interceptors.why(why)

		when := Before
		for _, v := range forward {
			if v.when&when != 0 {
				ctx, diags = v.interceptor.run(ctx, d, meta, when, why, diags)

				// Short circuit if any Before interceptor errors.
				if diags.HasError() {
					return diags
				}
			}
		}

		// All other interceptors are run last to first.
		reverse := slices.Reverse(forward)
		diags = f(ctx, d, meta)

		if diags.HasError() {
			when = OnError
		} else {
			when = After
		}
		for _, v := range reverse {
			if v.when&when != 0 {
				ctx, diags = v.interceptor.run(ctx, d, meta, when, why, diags)
			}
		}

		when = Finally
		for _, v := range reverse {
			if v.when&when != 0 {
				ctx, diags = v.interceptor.run(ctx, d, meta, when, why, diags)
			}
		}

		return diags
	}
}

// contextFunc augments Context.
type contextFunc func(context.Context, any) context.Context

// wrappedDataSource represents an interceptor dispatcher for a Plugin SDK v2 data source.
type wrappedDataSource struct {
	// bootstrapContext is run on all wrapped methods before any interceptors.
	bootstrapContext contextFunc
	interceptors     interceptorItems
}

func (ds *wrappedDataSource) Read(f schema.ReadContextFunc) schema.ReadContextFunc {
	return interceptedHandler(ds.bootstrapContext, ds.interceptors, f, Read)
}

// wrappedResource represents an interceptor dispatcher for a Plugin SDK v2 resource.
type wrappedResource struct {
	// bootstrapContext is run on all wrapped methods before any interceptors.
	bootstrapContext contextFunc
	interceptors     interceptorItems
}

func (r *wrappedResource) Create(f schema.CreateContextFunc) schema.CreateContextFunc {
	return interceptedHandler(r.bootstrapContext, r.interceptors, f, Create)
}

func (r *wrappedResource) Read(f schema.ReadContextFunc) schema.ReadContextFunc {
	return interceptedHandler(r.bootstrapContext, r.interceptors, f, Read)
}

func (r *wrappedResource) Update(f schema.UpdateContextFunc) schema.UpdateContextFunc {
	return interceptedHandler(r.bootstrapContext, r.interceptors, f, Update)
}

func (r *wrappedResource) Delete(f schema.DeleteContextFunc) schema.DeleteContextFunc {
	return interceptedHandler(r.bootstrapContext, r.interceptors, f, Delete)
}

func (r *wrappedResource) State(f schema.StateContextFunc) schema.StateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
		ctx = r.bootstrapContext(ctx, meta)

		return f(ctx, d, meta)
	}
}

func (r *wrappedResource) CustomizeDiff(f schema.CustomizeDiffFunc) schema.CustomizeDiffFunc {
	return func(ctx context.Context, d *schema.ResourceDiff, meta any) error {
		ctx = r.bootstrapContext(ctx, meta)

		return f(ctx, d, meta)
	}
}

func (r *wrappedResource) StateUpgrade(f schema.StateUpgradeFunc) schema.StateUpgradeFunc {
	return func(ctx context.Context, rawState map[string]interface{}, meta any) (map[string]interface{}, error) {
		ctx = r.bootstrapContext(ctx, meta)

		return f(ctx, rawState, meta)
	}
}

type tagsCRUDFunc func(context.Context, schemaResourceData, conns.ServicePackage, *types.ServicePackageResourceTags, string, string, any, diag.Diagnostics) (context.Context, diag.Diagnostics)

// tagsResourceInterceptor implements transparent tagging for resources.
type tagsResourceInterceptor struct {
	tags       *types.ServicePackageResourceTags
	updateFunc tagsCRUDFunc
	readFunc   tagsCRUDFunc
}

func (r tagsResourceInterceptor) run(ctx context.Context, d schemaResourceData, meta any, when when, why why, diags diag.Diagnostics) (context.Context, diag.Diagnostics) {
	if r.tags == nil {
		return ctx, diags
	}

	inContext, ok := conns.FromContext(ctx)
	if !ok {
		return ctx, diags
	}

	sp, ok := meta.(*conns.AWSClient).ServicePackages[inContext.ServicePackageName]
	if !ok {
		return ctx, diags
	}

	serviceName, err := names.HumanFriendly(inContext.ServicePackageName)
	if err != nil {
		serviceName = "<service>"
	}

	resourceName := inContext.ResourceName
	if resourceName == "" {
		resourceName = "<thing>"
	}

	tagsInContext, ok := tftags.FromContext(ctx)
	if !ok {
		return ctx, diags
	}

	switch when {
	case Before:
		switch why {
		case Create, Update:
			// Merge the resource's configured tags with any provider configured default_tags.
			tags := tagsInContext.DefaultConfig.MergeTags(tftags.New(ctx, d.Get(names.AttrTags).(map[string]interface{})))
			// Remove system tags.
			tags = tags.IgnoreSystem(inContext.ServicePackageName)

			tagsInContext.TagsIn = types.Some(tags)

			if why == Create {
				break
			}

			if d.GetRawPlan().GetAttr("tags_all").IsWhollyKnown() {
				if d.HasChange(names.AttrTagsAll) {
					if identifierAttribute := r.tags.IdentifierAttribute; identifierAttribute != "" {
						var identifier string
						if identifierAttribute == "id" {
							identifier = d.Id()
						} else {
							identifier = d.Get(identifierAttribute).(string)
						}

						// Some old resources may not have the required attribute set after Read:
						// https://github.com/hashicorp/terraform-provider-aws/issues/31180
						if identifier != "" {
							o, n := d.GetChange(names.AttrTagsAll)

							// If the service package has a generic resource update tags methods, call it.
							var err error

							if v, ok := sp.(interface {
								UpdateTags(context.Context, any, string, any, any) error
							}); ok {
								err = v.UpdateTags(ctx, meta, identifier, o, n)
							} else if v, ok := sp.(interface {
								UpdateTags(context.Context, any, string, string, any, any) error
							}); ok && r.tags.ResourceType != "" {
								err = v.UpdateTags(ctx, meta, identifier, r.tags.ResourceType, o, n)
							}

							// ISO partitions may not support tagging, giving error.
							if errs.IsUnsupportedOperationInPartitionError(meta.(*conns.AWSClient).Partition, err) {
								return ctx, diags
							}

							if err != nil {
								return ctx, sdkdiag.AppendErrorf(diags, "updating tags for %s %s (%s): %s", serviceName, resourceName, identifier, err)
							}
						}
						// TODO If the only change was to tags it would be nice to not call the resource's U handler.
					}
				}
			}
		}
	case After:
		// Set tags and tags_all in state after CRU.
		// C & U handlers are assumed to tail call the R handler.
		switch why {
		case Read:
			// Will occur on a refresh when the resource does not exist in AWS and needs to be recreated, e.g. "_disappears" tests.
			if d.Id() == "" {
				return ctx, diags
			}

			fallthrough
		case Create, Update:
			// If the R handler didn't set tags, try and read them from the service API.
			if tagsInContext.TagsOut.IsNone() {
				if identifierAttribute := r.tags.IdentifierAttribute; identifierAttribute != "" {
					var identifier string
					if identifierAttribute == "id" {
						identifier = d.Id()
					} else {
						identifier = d.Get(identifierAttribute).(string)
					}

					// Some old resources may not have the required attribute set after Read:
					// https://github.com/hashicorp/terraform-provider-aws/issues/31180
					if identifier != "" {
						// If the service package has a generic resource list tags methods, call it.
						var err error

						if v, ok := sp.(interface {
							ListTags(context.Context, any, string) error
						}); ok {
							err = v.ListTags(ctx, meta, identifier) // Sets tags in Context
						} else if v, ok := sp.(interface {
							ListTags(context.Context, any, string, string) error
						}); ok && r.tags.ResourceType != "" {
							err = v.ListTags(ctx, meta, identifier, r.tags.ResourceType) // Sets tags in Context
						}

						// ISO partitions may not support tagging, giving error.
						if errs.IsUnsupportedOperationInPartitionError(meta.(*conns.AWSClient).Partition, err) {
							return ctx, diags
						}

						if inContext.ServicePackageName == names.DynamoDB && err != nil {
							// When a DynamoDB Table is `ARCHIVED`, ListTags returns `ResourceNotFoundException`.
							if tfresource.NotFound(err) || tfawserr.ErrMessageContains(err, "UnknownOperationException", "Tagging is not currently supported in DynamoDB Local.") {
								err = nil
							}
						}

						if err != nil {
							return ctx, sdkdiag.AppendErrorf(diags, "listing tags for %s %s (%s): %s", serviceName, resourceName, identifier, err)
						}
					}
				}
			}

			// Remove any provider configured ignore_tags and system tags from those returned from the service API.
			tags := tagsInContext.TagsOut.UnwrapOrDefault().IgnoreSystem(inContext.ServicePackageName).IgnoreConfig(tagsInContext.IgnoreConfig)

			// The resource's configured tags can now include duplicate tags that have been configured on the provider.
			if err := d.Set(names.AttrTags, tags.ResolveDuplicates(ctx, tagsInContext.DefaultConfig, tagsInContext.IgnoreConfig, d).Map()); err != nil {
				return ctx, sdkdiag.AppendErrorf(diags, "setting %s: %s", names.AttrTags, err)
			}

			// Computed tags_all do.
			if err := d.Set(names.AttrTagsAll, tags.Map()); err != nil {
				return ctx, sdkdiag.AppendErrorf(diags, "setting %s: %s", names.AttrTagsAll, err)
			}
		}
	case Finally:
		switch why {
		case Update:
			if r.tags.IdentifierAttribute != "" && !d.GetRawPlan().GetAttr(names.AttrTagsAll).IsWhollyKnown() {
				ctx, diags = r.updateFunc(ctx, d, sp, r.tags, serviceName, resourceName, meta, diags)
				ctx, diags = r.readFunc(ctx, d, sp, r.tags, serviceName, resourceName, meta, diags)
			}
		}
	}

	return ctx, diags
}

// tagsResourceInterceptor implements transparent tagging for data sources.
type tagsDataSourceInterceptor struct {
	tags *types.ServicePackageResourceTags
}

func (r tagsDataSourceInterceptor) run(ctx context.Context, d schemaResourceData, meta any, when when, why why, diags diag.Diagnostics) (context.Context, diag.Diagnostics) {
	if r.tags == nil {
		return ctx, diags
	}

	inContext, ok := conns.FromContext(ctx)
	if !ok {
		return ctx, diags
	}

	// sp, ok := meta.(*conns.AWSClient).ServicePackages[inContext.ServicePackageName]
	// if !ok {
	// 	return ctx, diags
	// }

	// serviceName, err := names.HumanFriendly(inContext.ServicePackageName)
	// if err != nil {
	// 	serviceName = "<service>"
	// }

	// resourceName := inContext.ResourceName
	// if resourceName == "" {
	// 	resourceName = "<thing>"
	// }

	tagsInContext, ok := tftags.FromContext(ctx)
	if !ok {
		return ctx, diags
	}

	switch when {
	case Before:
		switch why {
		case Read:
			// Get the data source's configured tags.
			tags := tftags.New(ctx, d.Get(names.AttrTags).(map[string]interface{}))
			tagsInContext.TagsIn = types.Some(tags)
		}
	case After:
		// Set tags and tags_all in state after CRU.
		// C & U handlers are assumed to tail call the R handler.
		switch why {
		case Read:
			// Will occur on a refresh when the resource does not exist in AWS and needs to be recreated, e.g. "_disappears" tests.
			if d.Id() == "" {
				return ctx, diags
			}

			fallthrough
		case Create, Update:
			// If the R handler didn't set tags, try and read them from the service API.
			// TODO.
			// if tagsInContext.TagsOut.IsNone() {
			// }

			// Remove any provider configured ignore_tags and system tags from those returned from the service API.
			tags := tagsInContext.TagsOut.UnwrapOrDefault().IgnoreSystem(inContext.ServicePackageName).IgnoreConfig(tagsInContext.IgnoreConfig)
			if err := d.Set(names.AttrTags, tags.Map()); err != nil {
				return ctx, sdkdiag.AppendErrorf(diags, "setting %s: %s", names.AttrTags, err)
			}
		}
	}

	return ctx, diags
}
