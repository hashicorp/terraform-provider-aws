// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mockframework

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/mockframework"
	awstypes "github.com/aws/aws-sdk-go-v2/service/mockframework/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_mockframework_foo", name="Foo")
// @Tags(identifierAttribute="arn")
func newFooResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &fooResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type fooResource struct {
	framework.ResourceWithModel[fooResourceModel]
	framework.WithTimeouts
}

func (r *fooResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data fooResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().MockFrameworkClient(ctx)

	var input mockframework.CreateFooInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}

	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateFoo(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, names.AttrName, data.Name.ValueString())
		return
	}

	id := fwflex.StringValueFromFramework(ctx, fwflex.StringToFramework(ctx, output.FooId))
	if _, err := waitFooCreated(ctx, conn, id, r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, names.AttrID, id)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, output, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func (r *fooResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data fooResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().MockFrameworkClient(ctx)

	output, err := findFooByID(ctx, conn, data.ID.ValueString())
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, names.AttrID, data.ID.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, output, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *fooResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old fooResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &new))
	if response.Diagnostics.HasError() {
		return
	}
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &old))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().MockFrameworkClient(ctx)

	if !new.Size.Equal(old.Size) {
		var input mockframework.UpdateFooInput
		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, new, &input))
		if response.Diagnostics.HasError() {
			return
		}

		if _, err := conn.UpdateFoo(ctx, &input); err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, names.AttrID, new.ID.ValueString())
			return
		}
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &new))
}

func (r *fooResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data fooResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().MockFrameworkClient(ctx)

	id := data.ID.ValueString()
	input := mockframework.DeleteFooInput{
		FooId: fwflex.StringFromFramework(ctx, data.ID),
	}
	_, err := conn.DeleteFoo(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, names.AttrID, id)
		return
	}

	if _, err := waitFooDeleted(ctx, conn, id, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, names.AttrID, id)
		return
	}
}

// ModifyPlan checks the planned size against the account's quota.
func (r *fooResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	if request.Plan.Raw.IsNull() {
		return
	}

	var size types.Int64
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.GetAttribute(ctx, path.Root("size"), &size))
	if response.Diagnostics.HasError() || size.IsUnknown() {
		return
	}

	quota, err := findQuotaByName(ctx, r.Meta().MockFrameworkClient(ctx), "size")
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err)
		return
	}

	if v := size.ValueInt64(); v > int64(quota.Value) {
		response.Diagnostics.AddAttributeError(path.Root("size"), "value exceeds quota", fmt.Sprintf("size (%d) exceeds quota (%d)", v, quota.Value))
	}
}

func (r *fooResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), request, response)
}

type fooResourceModel struct {
	framework.WithRegionModel
	ARN      types.String   `tfsdk:"arn"`
	ID       types.String   `tfsdk:"id"`
	Name     types.String   `tfsdk:"name"`
	Size     types.Int64    `tfsdk:"size"`
	Tags     tftags.Map     `tfsdk:"tags"`
	TagsAll  tftags.Map     `tfsdk:"tags_all"`
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}
