package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type DataSource struct {
	interceptors Interceptors
}

func (ds *DataSource) Read(f schema.ReadContextFunc) schema.ReadContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
		return InvokeHandler(ctx, d, meta, ds.interceptors.Why(Read), f)
	}
}

type Resource struct {
	interceptors Interceptors
}

func (r *Resource) Create(f schema.CreateContextFunc) schema.CreateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
		return InvokeHandler(ctx, d, meta, r.interceptors.Why(Create), f)
	}
}

func (r *Resource) Read(f schema.ReadContextFunc) schema.ReadContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
		return InvokeHandler(ctx, d, meta, r.interceptors.Why(Read), f)
	}
}

func (r *Resource) Update(f schema.UpdateContextFunc) schema.UpdateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
		return InvokeHandler(ctx, d, meta, r.interceptors.Why(Update), f)
	}
}

func (r *Resource) Delete(f schema.DeleteContextFunc) schema.DeleteContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
		return InvokeHandler(ctx, d, meta, r.interceptors.Why(Delete), f)
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
