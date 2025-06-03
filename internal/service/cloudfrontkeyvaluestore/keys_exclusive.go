// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfrontkeyvaluestore

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfrontkeyvaluestore"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfrontkeyvaluestore/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ResNameKeysExclusive = "Keys Exclusive"
	ResNameKeyValueStore = "Key Value Store"
)

// @FrameworkResource("aws_cloudfrontkeyvaluestore_keys_exclusive", name="Keys  Exclusive")
func newResourceKeysExclusive(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceKeysExclusive{}, nil
}

type resourceKeysExclusive struct {
	framework.ResourceWithConfigure
	framework.WithNoOpDelete
}

func (r *resourceKeysExclusive) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
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
		},
		Blocks: map[string]schema.Block{
			"resource_key_value_pair": schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[resourceKeyValuePairModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrKey: schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "The key to put.",
						},
						names.AttrValue: schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "The value to put.",
						},
					},
				},
			},
		},
	}
}

func (r *resourceKeysExclusive) syncKeyValuePairs(ctx context.Context, plan *resourceKeysExclusiveModel) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := r.Meta().CloudFrontKeyValueStoreClient(ctx)
	kvsARN := plan.KvsARN.ValueString()

	// Making key changes the etag of the key value store.
	// Use a mutex serialize actions
	mutexKey := kvsARN
	conns.GlobalMutexKV.Lock(mutexKey)
	defer conns.GlobalMutexKV.Unlock(mutexKey)

	kvs, have, err := FindResourceKeyValuePairsForKeyValueStore(ctx, conn, kvsARN)
	if err != nil {
		diags.AddError(
			create.ProblemStandardMessage(names.CloudFrontKeyValueStore, create.ErrActionReading, ResNameKeyValueStore, kvsARN, err),
			err.Error(),
		)
		return diags
	}

	var want []awstypes.ListKeysResponseListItem
	diags.Append(flex.Expand(ctx, plan.ResourceKeyValuePair, &want)...)
	if diags.HasError() {
		return diags
	}

	put, del, _ := intflex.DiffSlices(have, want, resourceKeyValuePairEqual)

	putRequired := len(put) > 0
	deleteRequired := len(del) > 0

	if putRequired || deleteRequired {
		input := cloudfrontkeyvaluestore.UpdateKeysInput{
			KvsARN:  aws.String(kvsARN),
			IfMatch: kvs.ETag,
		}

		if putRequired {
			input.Puts = expandPutKeyRequestListItem(put)
		}

		if deleteRequired {
			input.Deletes = expandDeleteKeyRequestListItem(del)
		}

		out, err := conn.UpdateKeys(ctx, &input)

		if err != nil {
			diags.AddError(
				create.ProblemStandardMessage(names.CloudFrontKeyValueStore, create.ErrActionSynchronizing, ResNameKeysExclusive, kvsARN, err),
				err.Error(),
			)
			return diags
		}

		plan.TotalSizeInBytes = flex.Int64ToFramework(ctx, out.TotalSizeInBytes)
	} else {
		plan.TotalSizeInBytes = flex.Int64ToFramework(ctx, kvs.TotalSizeInBytes)
	}

	return diags
}

func (r *resourceKeysExclusive) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan resourceKeysExclusiveModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(r.syncKeyValuePairs(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, plan)...)
}

func (r *resourceKeysExclusive) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resourceKeysExclusiveModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudFrontKeyValueStoreClient(ctx)
	kvs, keyPairs, err := FindResourceKeyValuePairsForKeyValueStore(ctx, conn, data.KvsARN.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CloudFrontKeyValueStore, create.ErrActionReading, ResNameKeysExclusive, data.KvsARN.String(), err),
			err.Error(),
		)
		return
	}

	data.KvsARN = fwtypes.ARNValue(aws.ToString(kvs.KvsARN))
	data.TotalSizeInBytes = types.Int64Value(aws.ToInt64(kvs.TotalSizeInBytes))

	response.Diagnostics.Append(flex.Flatten(ctx, keyPairs, &data.ResourceKeyValuePair)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceKeysExclusive) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan resourceKeysExclusiveModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(r.syncKeyValuePairs(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, plan)...)
}

func (r *resourceKeysExclusive) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("key_value_store_arn"), request, response)
}

func FindKeyValueStoreByARN(ctx context.Context, conn *cloudfrontkeyvaluestore.Client, kvsARN string) (*cloudfrontkeyvaluestore.DescribeKeyValueStoreOutput, error) {
	input := &cloudfrontkeyvaluestore.DescribeKeyValueStoreInput{
		KvsARN: aws.String(kvsARN),
	}

	output, err := conn.DescribeKeyValueStore(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) ||
		// Attempting to describe a deleted keyvaluestore produces a ConflictException
		// rather than a ResourceNotFoundError. e.g.
		//
		// ConflictException: Key-Value-Store was not in expected state
		errs.IsAErrorMessageContains[*awstypes.ConflictException](err, "was not in expected state") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.KvsARN == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func FindResourceKeyValuePairsForKeyValueStore(ctx context.Context, conn *cloudfrontkeyvaluestore.Client, kvsARN string) (*cloudfrontkeyvaluestore.DescribeKeyValueStoreOutput, []awstypes.ListKeysResponseListItem, error) {
	kvs, err := FindKeyValueStoreByARN(ctx, conn, kvsARN)
	if err != nil {
		return nil, nil, err
	}

	input := &cloudfrontkeyvaluestore.ListKeysInput{
		KvsARN: aws.String(kvsARN),
	}

	var keyValuePairs []awstypes.ListKeysResponseListItem

	paginator := cloudfrontkeyvaluestore.NewListKeysPaginator(conn, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)

		if err != nil {
			return nil, nil, err
		}

		if page != nil {
			keyValuePairs = append(keyValuePairs, page.Items...)
		}
	}

	return kvs, keyValuePairs, nil
}

func expandPutKeyRequestListItem(put []awstypes.ListKeysResponseListItem) []awstypes.PutKeyRequestListItem {
	out := []awstypes.PutKeyRequestListItem{}
	for _, r := range put {
		out = append(out, awstypes.PutKeyRequestListItem{
			Key:   r.Key,
			Value: r.Value,
		})
	}

	return out
}

func expandDeleteKeyRequestListItem(delete []awstypes.ListKeysResponseListItem) []awstypes.DeleteKeyRequestListItem {
	out := []awstypes.DeleteKeyRequestListItem{}
	for _, r := range delete {
		out = append(out, awstypes.DeleteKeyRequestListItem{
			Key: r.Key})
	}

	return out
}

type resourceKeyValuePairModel struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

type resourceKeysExclusiveModel struct {
	ResourceKeyValuePair fwtypes.SetNestedObjectValueOf[resourceKeyValuePairModel] `tfsdk:"resource_key_value_pair"`
	KvsARN               fwtypes.ARN                                               `tfsdk:"key_value_store_arn"`
	TotalSizeInBytes     types.Int64                                               `tfsdk:"total_size_in_bytes"`
}
