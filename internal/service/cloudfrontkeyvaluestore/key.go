// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfrontkeyvaluestore

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfrontkeyvaluestore"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfrontkeyvaluestore/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var (
	ResourceKey = newResourceKey
)

// @FrameworkResource(name="Key")
func newResourceKey(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceKey{}

	return r, nil
}

const (
	ResNameKey        = "Key"
	ResIDKeyPartCount = 2
)

type resourceKey struct {
	framework.ResourceWithConfigure
}

func (r *resourceKey) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_cloudfrontkeyvaluestore_key"
}

func (r *resourceKey) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": framework.IDAttribute(),
			"key": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The key to put.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"key_value_store_arn": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The Amazon Resource Name (ARN) of the Key Value Store.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"total_size_in_bytes": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Total size of the Key Value Store in bytes.",
			},
			"value": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The value to put.",
			},
		},
	}
}

func (r *resourceKey) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().CloudFrontKeyValueStoreClient(ctx)

	var plan resourceKeyData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	etag, err := findEtagByARN(ctx, conn, plan.KeyValueStoreARN.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CloudFrontKeyValueStore, create.ErrActionCreating, ResNameKey, plan.Key.String(), err),
			err.Error(),
		)
		return
	}

	in := &cloudfrontkeyvaluestore.PutKeyInput{
		IfMatch: etag,
		Key:     aws.String(plan.Key.ValueString()),
		KvsARN:  aws.String(plan.KeyValueStoreARN.ValueString()),
		Value:   aws.String(plan.Value.ValueString()),
	}

	out, err := conn.PutKey(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CloudFrontKeyValueStore, create.ErrActionCreating, ResNameKey, plan.Key.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CloudFrontKeyValueStore, create.ErrActionCreating, ResNameKey, plan.Key.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.TotalSizeInBytes = flex.Int64ToFramework(ctx, out.TotalSizeInBytes)
	err = plan.setID(ctx)

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CloudFrontKeyValueStore, create.ErrActionCreating, ResNameKey, plan.Key.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceKey) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().CloudFrontKeyValueStoreClient(ctx)

	var state resourceKeyData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	kvsARN, out, err := FindKeyByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CloudFrontKeyValueStore, create.ErrActionSetting, ResNameKey, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	state.Key = flex.StringToFramework(ctx, out.Key)
	state.KeyValueStoreARN = flex.StringValueToFramework(ctx, kvsARN)
	state.TotalSizeInBytes = flex.Int64ToFramework(ctx, out.TotalSizeInBytes)
	state.Value = flex.StringToFramework(ctx, out.Value)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceKey) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().CloudFrontKeyValueStoreClient(ctx)

	var plan, state resourceKeyData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Value.Equal(state.Value) {
		etag, err := findEtagByARN(ctx, conn, plan.KeyValueStoreARN.ValueString())

		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.CloudFrontKeyValueStore, create.ErrActionUpdating, ResNameKey, plan.ID.String(), err),
				err.Error(),
			)
			return
		}

		in := &cloudfrontkeyvaluestore.PutKeyInput{
			IfMatch: etag,
			Key:     aws.String(plan.Key.ValueString()),
			KvsARN:  aws.String(plan.KeyValueStoreARN.ValueString()),
			Value:   aws.String(plan.Value.ValueString()),
		}

		out, err := conn.PutKey(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.CloudFrontKeyValueStore, create.ErrActionUpdating, ResNameKey, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.CloudFrontKeyValueStore, create.ErrActionUpdating, ResNameKey, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		plan.TotalSizeInBytes = flex.Int64ToFramework(ctx, out.TotalSizeInBytes)
		err = plan.setID(ctx)

		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.CloudFrontKeyValueStore, create.ErrActionUpdating, ResNameKey, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceKey) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().CloudFrontKeyValueStoreClient(ctx)

	var state resourceKeyData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	etag, err := findEtagByARN(ctx, conn, state.KeyValueStoreARN.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CloudFrontKeyValueStore, create.ErrActionDeleting, ResNameKey, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	in := &cloudfrontkeyvaluestore.DeleteKeyInput{
		IfMatch: etag,
		Key:     aws.String(state.Key.ValueString()),
		KvsARN:  aws.String(state.KeyValueStoreARN.ValueString()),
	}

	_, err = conn.DeleteKey(ctx, in)

	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CloudFrontKeyValueStore, create.ErrActionDeleting, ResNameKey, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceKey) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func FindKeyByID(ctx context.Context, conn *cloudfrontkeyvaluestore.Client, id string) (kvsARN string, getKeyOutPut *cloudfrontkeyvaluestore.GetKeyOutput, err error) {
	parts, err := intflex.ExpandResourceId(id, ResIDKeyPartCount, false)

	if err != nil {
		return "", nil, err
	}

	in := &cloudfrontkeyvaluestore.GetKeyInput{
		KvsARN: aws.String(parts[0]),
		Key:    aws.String(parts[1]),
	}

	out, err := conn.GetKey(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return "", nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return "", nil, err
	}

	if out == nil || out.Key == nil {
		return "", nil, tfresource.NewEmptyResultError(in)
	}

	return parts[0], out, nil
}

func findEtagByARN(ctx context.Context, conn *cloudfrontkeyvaluestore.Client, arn string) (*string, error) {
	in := &cloudfrontkeyvaluestore.DescribeKeyValueStoreInput{
		KvsARN: aws.String(arn),
	}

	out, err := conn.DescribeKeyValueStore(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.ETag == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.ETag, nil
}

type resourceKeyData struct {
	ID               types.String `tfsdk:"id"`
	Key              types.String `tfsdk:"key"`
	KeyValueStoreARN types.String `tfsdk:"key_value_store_arn"`
	TotalSizeInBytes types.Int64  `tfsdk:"total_size_in_bytes"`
	Value            types.String `tfsdk:"value"`
}

func (data *resourceKeyData) setID(ctx context.Context) error {
	id, err := intflex.FlattenResourceId([]string{data.KeyValueStoreARN.ValueString(), data.Key.ValueString()}, ResIDKeyPartCount, false)

	if err != nil {
		return err
	}

	data.ID = flex.StringToFramework(ctx, aws.String(id))

	return nil
}
