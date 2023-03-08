package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// InterceptorFunc is functionality invoked during the CRUD request lifecycle.
// If a Before interceptor returns Diagnostics indicating an error occurred then
// no further interceptors in the chain are run and neither is the schema's method.
// In other cases all interceptors in the chain are run.
type InterceptorFunc func(context.Context, *schema.ResourceData, any, When, Why, diag.Diagnostics) (context.Context, diag.Diagnostics)

// Interceptor represents a single interceptor.
type Interceptor struct {
	When When
	Why  Why
	Func InterceptorFunc
}

// When represents the point in the CRUD request lifecycle that an interceptor is run.
// Multiple values can be ORed together.
type When uint16

const (
	Before  When = 1 << iota // Interceptor is invoked before call to method in schema
	After                    // Interceptor is invoked after successful call to method in schema
	OnError                  // Interceptor is invoked after unsuccessful call to method in schema
	Finally                  // Interceptor is invoked after After or OnError
)

// Why represents the CRUD operation(s) that an interceptor is run.
// Multiple values can be ORed together.
type Why uint16

const (
	Create Why = 1 << iota // Interceptor is invoked for a Create call
	Read                   // Interceptor is invoked for a Read call
	Update                 // Interceptor is invoked for a Update call
	Delete                 // Interceptor is invoked for a Delete call
)

type Interceptors []Interceptor

func (v *Interceptors) Append(when When, why Why, f InterceptorFunc) {
	interceptor := Interceptor{
		When: when,
		Why:  why,
		Func: f,
	}

	if v == nil {
		*v = Interceptors{interceptor}
	} else {
		*v = append(*v, interceptor)
	}
}

// Why returns a slice of Interceptors that run for the specified CRUD operation.
func (v *Interceptors) Why(why Why) Interceptors {
	var interceptors Interceptors

	for _, v := range *v {
		if v.Why&why != 0 {
			interceptors = append(interceptors, v)
		}
	}

	return interceptors
}

// InvokeHandler invokes the specified CRUD handler, running any interceptors.
func InvokeHandler[F ~func(context.Context, *schema.ResourceData, any) diag.Diagnostics](ctx context.Context, d *schema.ResourceData, meta any, interceptors []Interceptor, f F, why Why) diag.Diagnostics {
	var diags diag.Diagnostics

	when := Before
	for _, v := range interceptors {
		if v.When&when != 0 {
			ctx, diags = v.Func(ctx, d, meta, when, why, diags)

			// Short circuit if any Before interceptor errors.
			if diags.HasError() {
				return diags
			}
		}
	}

	reversed := slices.Reverse(interceptors)
	diags = f(ctx, d, meta)

	if diags.HasError() {
		when = OnError
		for _, v := range reversed {
			if v.When&when != 0 {
				ctx, diags = v.Func(ctx, d, meta, when, why, diags)
			}
		}
	} else {
		when = After
		for _, v := range reversed {
			if v.When&when != 0 {
				ctx, diags = v.Func(ctx, d, meta, when, why, diags)
			}
		}
	}

	for _, v := range reversed {
		when = Finally
		if v.When&when != 0 {
			ctx, diags = v.Func(ctx, d, meta, when, why, diags)
		}
	}

	return diags
}

type DataSource struct {
	interceptors Interceptors
}

func (ds *DataSource) Read(f schema.ReadContextFunc) schema.ReadContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
		why := Read
		return InvokeHandler(ctx, d, meta, ds.interceptors.Why(why), f, why)
	}
}

type Resource struct {
	interceptors Interceptors
	tags         *types.ServicePackageResourceTags
}

func (r *Resource) Create(f schema.CreateContextFunc) schema.CreateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
		why := Create
		return InvokeHandler(ctx, d, meta, r.interceptors.Why(why), f, why)
	}
}

func (r *Resource) Read(f schema.ReadContextFunc) schema.ReadContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
		why := Read
		return InvokeHandler(ctx, d, meta, r.interceptors.Why(why), f, why)
	}
}

func (r *Resource) Update(f schema.UpdateContextFunc) schema.UpdateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
		why := Update
		return InvokeHandler(ctx, d, meta, r.interceptors.Why(why), f, why)
	}
}

func (r *Resource) Delete(f schema.DeleteContextFunc) schema.DeleteContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
		why := Delete
		return InvokeHandler(ctx, d, meta, r.interceptors.Why(why), f, why)
	}
}

func (r *Resource) State(f schema.StateContextFunc) schema.StateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
		return f(ctx, d, meta)
	}
}

func (r *Resource) CustomizeDiff(f schema.CustomizeDiffFunc) schema.CustomizeDiffFunc {
	return func(ctx context.Context, d *schema.ResourceDiff, meta any) error {
		return f(ctx, d, meta)
	}
}

func (r *Resource) StateUpgrade(f schema.StateUpgradeFunc) schema.StateUpgradeFunc {
	return func(ctx context.Context, rawState map[string]interface{}, meta any) (map[string]interface{}, error) {
		return f(ctx, rawState, meta)
	}
}

func (r *Resource) tagsInterceptor(ctx context.Context, d *schema.ResourceData, meta any, when When, why Why, diags diag.Diagnostics) (context.Context, diag.Diagnostics) {
	if r.tags == nil {
		return ctx, diags
	}

	spName, ok := conns.ServicePackageNameFromContext(ctx)

	if !ok {
		return ctx, diags
	}

	sp, ok := meta.(*conns.AWSClient).ServicePackages[spName]

	if !ok {
		return ctx, diags
	}

	switch when {
	case Before:
		v := tftags.InContext{
			DefaultConfig: meta.(*conns.AWSClient).DefaultTagsConfig,
			IgnoreConfig:  meta.(*conns.AWSClient).IgnoreTagsConfig,
		}

		ctx = context.WithValue(ctx, tftags.TagKey, &v)

		switch why {
		case Create:
			t, ok := tftags.FromContext(ctx)
			if !ok {
				return ctx, diags
			}

			tags := t.DefaultConfig.MergeTags(tftags.New(ctx, d.Get("tags").(map[string]interface{})))
			tags = tags.IgnoreAWS()

			ctx = context.WithValue(ctx, tftags.MergedTagsKey, &tags)
		case Update:
			if v, ok := sp.(conns.ServicePackageWithUpdateTags); ok {
				var identifier string

				if key := r.tags.IdentifierAttribute; key == "id" {
					identifier = d.Id()
				} else {
					identifier = d.Get(key).(string)
				}

				if d.HasChange("tags_all") {
					o, n := d.GetChange("tags_all")
					err := v.UpdateTags(ctx, meta, identifier, o, n)

					if verify.ErrorISOUnsupported(meta.(*conns.AWSClient).Partition, err) {
						// ISO partitions may not support tagging, giving error
						tflog.Warn(ctx, "failed updating tags for resource", map[string]interface{}{
							r.tags.IdentifierAttribute: d.Id(),
							"error":                    err.Error(),
						})
						return ctx, diags
					}

					if err != nil {
						return ctx, sdkdiag.AppendFromErr(diags, err)
					}
				}

				// TODO It would be nice to be able to skip calling the CRUD hanelder if only tags had changed.
				// if d.HasChangesExcept("tags", "tags_all") {
				// }
			}
		}
	case After:
		switch why {
		case Read:
			if v, ok := sp.(conns.ServicePackageWithListTags); ok {
				var identifier string

				if key := r.tags.IdentifierAttribute; key == "id" {
					identifier = d.Id()
				} else {
					identifier = d.Get(key).(string)
				}

				t, ok := tftags.FromContext(ctx)
				if !ok {
					return ctx, diags
				}

				tags, err := v.ListTags(ctx, meta, identifier)

				if verify.ErrorISOUnsupported(meta.(*conns.AWSClient).Partition, err) {
					// ISO partitions may not support tagging, giving error
					tflog.Warn(ctx, "failed listing tags for resource", map[string]interface{}{
						r.tags.IdentifierAttribute: d.Id(),
						"error":                    err.Error(),
					})
					return ctx, diags
				}

				if err != nil {
					return ctx, sdkdiag.AppendErrorf(diags, "listing tags for resource(%s): %s", d.Id(), err)
				}

				tags = tags.IgnoreAWS().IgnoreConfig(t.IgnoreConfig)

				if err := d.Set("tags", tags.RemoveDefaultConfig(t.DefaultConfig).Map()); err != nil {
					return ctx, sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
				}

				if err := d.Set("tags_all", tags.Map()); err != nil {
					return ctx, sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
				}
			}

		}
	}

	return ctx, diags
}
