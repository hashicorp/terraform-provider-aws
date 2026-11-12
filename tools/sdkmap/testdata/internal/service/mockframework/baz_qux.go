// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mockframework

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// bazQux is implemented by each resource type that embeds bazQuxResource.
type bazQux interface {
	create(context.Context, *bazQuxResourceModel) (string, error)
	delete(context.Context, *bazQuxResourceModel) error
	findByID(context.Context, string) (any, error)
}

// bazQuxResource implements the methods that Baz and Qux share, which call each one's bazQux methods.
type bazQuxResource struct {
	bazQux
	framework.ResourceWithModel[bazQuxResourceModel]
	framework.WithNoUpdate
	framework.WithImportByID
}

func (r *bazQuxResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data bazQuxResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	id, err := r.create(ctx, &data)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, names.AttrName, data.Name.ValueString())
		return
	}

	data.ID = types.StringValue(id)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func (r *bazQuxResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data bazQuxResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	_, err := r.findByID(ctx, data.ID.ValueString())
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, names.AttrID, data.ID.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *bazQuxResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data bazQuxResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	if err := r.delete(ctx, &data); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, names.AttrID, data.ID.ValueString())
		return
	}
}

// ModifyPlan marks the description as unknown when the name changes. It calls no AWS SDK methods.
func (r *bazQuxResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	if request.State.Raw.IsNull() || request.Plan.Raw.IsNull() {
		return
	}

	var plan, state bazQuxResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	if !plan.Name.Equal(state.Name) {
		plan.Description = types.StringUnknown()
		smerr.AddEnrich(ctx, &response.Diagnostics, response.Plan.Set(ctx, &plan))
	}
}

type bazQuxResourceModel struct {
	framework.WithRegionModel
	Description types.String `tfsdk:"description"`
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
}
