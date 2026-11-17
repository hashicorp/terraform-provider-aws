// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mockframework

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mockframework"
	awstypes "github.com/aws/aws-sdk-go-v2/service/mockframework/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_mockframework_bar", name="Bar")
func newBarResource(context.Context) (resource.ResourceWithConfigure, error) {
	return &barResource{}, nil
}

type barResource struct {
	framework.ResourceWithModel[barResourceModel]
	framework.WithNoUpdate
}

func (r *barResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data barResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().MockFrameworkClient(ctx)

	input := mockframework.CreateBarInput{
		Name: fwflex.StringFromFramework(ctx, data.Name),
	}

	output, err := conn.CreateBar(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, names.AttrName, data.Name.ValueString())
		return
	}

	data.ID = fwflex.StringToFramework(ctx, output.BarId)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func (r barResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data barResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().MockFrameworkClient(ctx)

	output, err := findBarByID(ctx, conn, data.ID.ValueString())
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, names.AttrID, data.ID.ValueString())
		return
	}

	data.Name = fwflex.StringToFramework(ctx, output.Name)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *barResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data barResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	if err := r.deleteBar(ctx, data.ID.ValueString()); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, names.AttrID, data.ID.ValueString())
		return
	}
}

// ImportState imports a Bar by name.
func (r *barResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	conn := r.Meta().MockFrameworkClient(ctx)

	bar, err := findBarByName(ctx, conn, request.ID)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("importing Bar (%s)", request.ID), err.Error())
		return
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrID), aws.ToString(bar.BarId))...)
}

func (r *barResource) deleteBar(ctx context.Context, id string) error {
	conn := r.Meta().MockFrameworkClient(ctx)

	input := mockframework.DeleteBarInput{
		BarId: aws.String(id),
	}
	_, err := conn.DeleteBar(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil
	}

	return err
}

type barResourceModel struct {
	framework.WithRegionModel
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}
