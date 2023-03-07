package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

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
