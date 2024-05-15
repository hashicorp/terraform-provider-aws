// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appfabric

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appfabric"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appfabric/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func newResourceAppBundle(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceAppBundle{}
	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)
	return r, nil
}

const (
	ResNameAppBundle = "AppBundle"
)

type resourceAppBundle struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceAppBundle) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_appfabric_app_bundle"
}

func (r *resourceAppBundle) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": framework.ARNAttributeComputedOnly(),
			"customer_managed_key_identifier": schema.StringAttribute{
				Optional: true,
			},
			names.AttrID:      framework.IDAttribute(),
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (r *resourceAppBundle) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().AppFabricClient(ctx)

	var plan resourceAppBundleData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	in := &appfabric.CreateAppBundleInput{
		Tags: getTagsIn(ctx),
	}

	out, err := conn.CreateAppBundle(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AppFabric, create.ErrActionCreating, ResNameAppBundle, plan.ARN.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.AppBundle == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AppFabric, create.ErrActionCreating, ResNameAppBundle, plan.ARN.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.ARN = flex.StringToFramework(ctx, out.AppBundle.Arn)
	plan.CustomerManagedKeyArn = flex.StringToFramework(ctx, out.AppBundle.CustomerManagedKeyArn)
	plan.ID = types.StringValue(string(*out.AppBundle.Arn))

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceAppBundle) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().AppFabricClient(ctx)

	var state resourceAppBundleData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findAppBundleByID(ctx, conn, state.ARN.ValueString())

	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AppFabric, create.ErrActionSetting, ResNameAppBundle, state.ARN.String(), err),
			err.Error(),
		)
		return
	}

	state.ARN = flex.StringToFramework(ctx, out.Arn)
	state.CustomerManagedKeyArn = flex.StringToFramework(ctx, out.CustomerManagedKeyArn)
	state.ID = types.StringValue(string(*out.Arn))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceAppBundle) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *resourceAppBundle) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

	conn := r.Meta().AppFabricClient(ctx)

	var state resourceAppBundleData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &appfabric.DeleteAppBundleInput{
		AppBundleIdentifier: aws.String(state.ARN.ValueString()),
	}

	_, err := conn.DeleteAppBundle(ctx, in)

	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AppFabric, create.ErrActionDeleting, ResNameAppBundle, state.ARN.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceAppBundle) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("arn"), req, resp)
}

const (
	statusChangePending = "Pending"
	statusDeleting      = "Deleting"
	statusNormal        = "Normal"
	statusUpdated       = "Updated"
)

func findAppBundleByID(ctx context.Context, conn *appfabric.Client, arn string) (*awstypes.AppBundle, error) {
	in := &appfabric.GetAppBundleInput{
		AppBundleIdentifier: aws.String(arn),
	}

	out, err := conn.GetAppBundle(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.AppBundle == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.AppBundle, nil
}

type resourceAppBundleData struct {
	ARN                   types.String `tfsdk:"arn"`
	CustomerManagedKeyArn types.String `tfsdk:"customer_managed_key_identifier"`
	ID                    types.String `tfsdk:"id"`
	Tags                  types.Map    `tfsdk:"tags"`
	TagsAll               types.Map    `tfsdk:"tags_all"`
}
