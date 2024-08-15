// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfrontkeyvaluestore

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/cloudfrontkeyvaluestore"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfrontkeyvaluestore/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Key")
func newKeyResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &keyResource{}

	return r, nil
}

type keyResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
}

func (*keyResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_cloudfrontkeyvaluestore_key"
}

func (r *keyResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			names.AttrKey: schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The key to put.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"key_value_store_arn": schema.StringAttribute{
				CustomType:          fwtypes.ARNType,
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
			names.AttrValue: schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The value to put.",
			},
		},
	}
}

func (r *keyResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data keyResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudFrontKeyValueStoreClient(ctx)

	kvsARN := data.KvsARN.ValueString()

	// Adding a key changes the etag of the key value store.
	// Use a mutex serialize actions
	mutexKey := kvsARN
	conns.GlobalMutexKV.Lock(mutexKey)
	defer conns.GlobalMutexKV.Unlock(mutexKey)

	etag, err := findETagByARN(ctx, conn, kvsARN)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading CloudFront KeyValueStore ETag (%s)", kvsARN), err.Error())

		return
	}

	input := &cloudfrontkeyvaluestore.PutKeyInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.IfMatch = etag

	output, err := conn.PutKey(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating CloudFront KeyValueStore (%s) Key (%s)", kvsARN, data.Key.ValueString()), err.Error())

		return
	}

	// Set values for unknowns.
	data.TotalSizeInBytes = fwflex.Int64ToFramework(ctx, output.TotalSizeInBytes)
	data.setID()

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *keyResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data keyResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().CloudFrontKeyValueStoreClient(ctx)

	output, err := findKeyByTwoPartKey(ctx, conn, data.KvsARN.ValueString(), data.Key.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading CloudFront KeyValueStore Key (%s)", data.ID.ValueString()), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *keyResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new keyResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudFrontKeyValueStoreClient(ctx)

	if !new.Value.Equal(old.Value) {
		kvsARN := new.KvsARN.ValueString()

		// Updating a key changes the etag of the key value store.
		// Use a mutex serialize actions
		mutexKey := kvsARN
		conns.GlobalMutexKV.Lock(mutexKey)
		defer conns.GlobalMutexKV.Unlock(mutexKey)

		etag, err := findETagByARN(ctx, conn, kvsARN)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("reading CloudFront KeyValueStore ETag (%s)", kvsARN), err.Error())

			return
		}

		input := &cloudfrontkeyvaluestore.PutKeyInput{}
		response.Diagnostics.Append(fwflex.Expand(ctx, new, input)...)
		if response.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		input.IfMatch = etag

		output, err := conn.PutKey(ctx, input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating CloudFront KeyValueStore (%s) Key (%s)", kvsARN, new.Key.ValueString()), err.Error())

			return
		}

		// Set values for unknowns.
		new.TotalSizeInBytes = fwflex.Int64ToFramework(ctx, output.TotalSizeInBytes)
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *keyResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data keyResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudFrontKeyValueStoreClient(ctx)

	kvsARN := data.KvsARN.ValueString()

	// Deleting a key changes the etag of the key value store.
	// Use a mutex serialize actions
	mutexKey := kvsARN
	conns.GlobalMutexKV.Lock(mutexKey)
	defer conns.GlobalMutexKV.Unlock(mutexKey)

	etag, err := findETagByARN(ctx, conn, kvsARN)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading CloudFront KeyValueStore ETag (%s)", kvsARN), err.Error())

		return
	}

	_, err = conn.DeleteKey(ctx, &cloudfrontkeyvaluestore.DeleteKeyInput{
		IfMatch: etag,
		Key:     fwflex.StringFromFramework(ctx, data.Key),
		KvsARN:  fwflex.StringFromFramework(ctx, data.KvsARN),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting CloudFront KeyValueStore Key (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

func findKeyByTwoPartKey(ctx context.Context, conn *cloudfrontkeyvaluestore.Client, kvsARN, key string) (*cloudfrontkeyvaluestore.GetKeyOutput, error) {
	input := &cloudfrontkeyvaluestore.GetKeyInput{
		Key:    aws.String(key),
		KvsARN: aws.String(kvsARN),
	}

	output, err := conn.GetKey(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Key == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func findETagByARN(ctx context.Context, conn *cloudfrontkeyvaluestore.Client, arn string) (*string, error) {
	input := &cloudfrontkeyvaluestore.DescribeKeyValueStoreInput{
		KvsARN: aws.String(arn),
	}

	output, err := conn.DescribeKeyValueStore(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ETag == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ETag, nil
}

type keyResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Key              types.String `tfsdk:"key"`
	KvsARN           fwtypes.ARN  `tfsdk:"key_value_store_arn"`
	TotalSizeInBytes types.Int64  `tfsdk:"total_size_in_bytes"`
	Value            types.String `tfsdk:"value"`
}

const (
	keyResourceIDPartCount = 2
)

func (data *keyResourceModel) InitFromID() error {
	id := data.ID.ValueString()
	parts, err := flex.ExpandResourceId(id, keyResourceIDPartCount, false)
	if err != nil {
		return err
	}

	_, err = arn.Parse(parts[0])
	if err != nil {
		return err
	}

	data.KvsARN = fwtypes.ARNValue(parts[0])
	data.Key = types.StringValue(parts[1])

	return nil
}

func (data *keyResourceModel) setID() {
	data.ID = types.StringValue(errs.Must(flex.FlattenResourceId([]string{data.KvsARN.ValueString(), data.Key.ValueString()}, keyResourceIDPartCount, false)))
}
