// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Key Value Store")
func newResourceKeyValueStore(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceKeyValueStore{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameKeyValueStore = "Key Value Store"
)

type resourceKeyValueStore struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceKeyValueStore) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_cloudfront_key_value_store"
}

func (r *resourceKeyValueStore) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": framework.ARNAttributeComputedOnly(),
			"comment": schema.StringAttribute{
				Optional: true,
			},
			"id": framework.IDAttribute(),
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"last_modified_time": schema.StringAttribute{
				Computed: true,
			},
			"etag": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *resourceKeyValueStore) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().CloudFrontClient(ctx)

	var plan resourceKeyValueStoreData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &cloudfront.CreateKeyValueStoreInput{
		Name: aws.String(plan.Name.ValueString()),
	}

	if !plan.Comment.IsNull() {
		in.Comment = aws.String(plan.Comment.ValueString())
	}

	out, err := conn.CreateKeyValueStore(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CloudFront, create.ErrActionCreating, ResNameKeyValueStore, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.KeyValueStore == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CloudFront, create.ErrActionCreating, ResNameKeyValueStore, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.ARN = flex.StringToFramework(ctx, out.KeyValueStore.ARN)
	plan.ID = flex.StringToFramework(ctx, out.KeyValueStore.Name)
	plan.Comment = flex.StringToFramework(ctx, out.KeyValueStore.Comment)
	plan.Name = flex.StringToFramework(ctx, out.KeyValueStore.Name)
	plan.LastModifiedTime = flex.StringToFramework(ctx, aws.String(fmt.Sprintf("%s", out.KeyValueStore.LastModifiedTime)))
	plan.ETag = flex.StringToFramework(ctx, out.ETag)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceKeyValueStore) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().CloudFrontClient(ctx)

	var state resourceKeyValueStoreData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findKeyValueStoreByName(ctx, conn, state.ID.ValueString())

	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CloudFront, create.ErrActionSetting, ResNameKeyValueStore, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	state.ARN = flex.StringToFramework(ctx, out.KeyValueStore.ARN)
	state.ETag = flex.StringToFramework(ctx, out.ETag)
	state.ID = flex.StringToFramework(ctx, out.KeyValueStore.Name)
	state.Comment = flex.StringToFramework(ctx, out.KeyValueStore.Comment)
	state.Name = flex.StringToFramework(ctx, out.KeyValueStore.Name)
	state.LastModifiedTime = flex.StringToFramework(ctx, aws.String(fmt.Sprintf("%s", out.KeyValueStore.LastModifiedTime)))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceKeyValueStore) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().CloudFrontClient(ctx)

	var plan, state resourceKeyValueStoreData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Comment.Equal(state.Comment) {

		in := &cloudfront.UpdateKeyValueStoreInput{
			Name:    aws.String(plan.Name.ValueString()),
			Comment: aws.String(plan.Comment.ValueString()),
			IfMatch: aws.String(plan.ETag.ValueString()),
		}

		out, err := conn.UpdateKeyValueStore(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.CloudFront, create.ErrActionUpdating, ResNameKeyValueStore, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.KeyValueStore == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.CloudFront, create.ErrActionUpdating, ResNameKeyValueStore, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		plan.ARN = flex.StringToFramework(ctx, out.KeyValueStore.ARN)
		plan.ID = flex.StringToFramework(ctx, out.KeyValueStore.Name)
		plan.Comment = flex.StringToFramework(ctx, out.KeyValueStore.Comment)
		plan.Name = flex.StringToFramework(ctx, out.KeyValueStore.Name)
		plan.LastModifiedTime = flex.StringToFramework(ctx, aws.String(fmt.Sprintf("%s", out.KeyValueStore.LastModifiedTime)))
		plan.ETag = flex.StringToFramework(ctx, out.ETag)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceKeyValueStore) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().CloudFrontClient(ctx)

	var state resourceKeyValueStoreData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &cloudfront.DeleteKeyValueStoreInput{
		Name:    aws.String(state.Name.ValueString()),
		IfMatch: aws.String(state.ETag.ValueString()),
	}

	_, err := conn.DeleteKeyValueStore(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.EntityNotFound](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CloudFront, create.ErrActionDeleting, ResNameKeyValueStore, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceKeyValueStore) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func findKeyValueStoreByName(ctx context.Context, conn *cloudfront.Client, name string) (*cloudfront.DescribeKeyValueStoreOutput, error) {
	in := &cloudfront.DescribeKeyValueStoreInput{
		Name: aws.String(name),
	}

	out, err := conn.DescribeKeyValueStore(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.EntityNotFound](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.KeyValueStore == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

type resourceKeyValueStoreData struct {
	ARN              types.String `tfsdk:"arn"`
	Comment          types.String `tfsdk:"comment"`
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	LastModifiedTime types.String `tfsdk:"last_modified_time"`
	ETag             types.String `tfsdk:"etag"`
}
