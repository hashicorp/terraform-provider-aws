// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfrontkeyvaluestore

import (
	"context"
	"slices"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfrontkeyvaluestore"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfrontkeyvaluestore/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
	maxBatchSizeDefault  = 50
)

// @FrameworkResource("aws_cloudfrontkeyvaluestore_keys_exclusive", name="Keys  Exclusive")
func newKeysExclusiveResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &keysExclusiveResource{}, nil
}

type keysExclusiveResource struct {
	framework.ResourceWithModel[keysExclusiveResourceModel]
	framework.WithNoOpDelete
}

func (r *keysExclusiveResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
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
			"max_batch_size": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(maxBatchSizeDefault),
				Validators: []validator.Int64{
					int64validator.Between(1, 50),
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "Maximum resource key values pairs that you wills update in a single API request. AWS has a default quota of 50 keys or a 3 MB payload, whichever is reached first",
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

func (r *keysExclusiveResource) syncKeyValuePairs(ctx context.Context, plan *keysExclusiveResourceModel) diag.Diagnostics {
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

	// We need to perform a batched operation in the event of many Key Value Pairs
	// to stay within AWS service limits
	//
	// https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/cloudfront-limits.html#limits-keyvaluestores
	batchSize := int(plan.MaximumBatchSize.ValueInt64())
	etag := kvs.ETag
	totalSizeInBytes := kvs.TotalSizeInBytes

	for chunk := range slices.Chunk(expandPutKeyRequestListItem(put), batchSize) {
		input := cloudfrontkeyvaluestore.UpdateKeysInput{
			KvsARN:  aws.String(kvsARN),
			IfMatch: etag,
			Puts:    chunk,
		}

		out, err := conn.UpdateKeys(ctx, &input)
		if err != nil {
			diags.AddError(
				create.ProblemStandardMessage(names.CloudFrontKeyValueStore, create.ErrActionSynchronizing, ResNameKeysExclusive, kvsARN, err),
				err.Error(),
			)
			return diags
		}
		etag = out.ETag
		totalSizeInBytes = out.TotalSizeInBytes
	}

	for chunk := range slices.Chunk(expandDeleteKeyRequestListItem(del), batchSize) {
		input := cloudfrontkeyvaluestore.UpdateKeysInput{
			KvsARN:  aws.String(kvsARN),
			IfMatch: etag,
			Deletes: chunk,
		}

		out, err := conn.UpdateKeys(ctx, &input)
		if err != nil {
			diags.AddError(
				create.ProblemStandardMessage(names.CloudFrontKeyValueStore, create.ErrActionSynchronizing, ResNameKeysExclusive, kvsARN, err),
				err.Error(),
			)
			return diags
		}
		etag = out.ETag
		totalSizeInBytes = out.TotalSizeInBytes
	}

	plan.TotalSizeInBytes = flex.Int64ToFramework(ctx, totalSizeInBytes)

	return diags
}

func (r *keysExclusiveResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan keysExclusiveResourceModel
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

func (r *keysExclusiveResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data keysExclusiveResourceModel
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

	if data.MaximumBatchSize.IsNull() || data.MaximumBatchSize.ValueInt64() == 0 {
		data.MaximumBatchSize = types.Int64Value(maxBatchSizeDefault)
	}

	response.Diagnostics.Append(flex.Flatten(ctx, keyPairs, &data.ResourceKeyValuePair)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *keysExclusiveResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan keysExclusiveResourceModel
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

func (r *keysExclusiveResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
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

type keysExclusiveResourceModel struct {
	ResourceKeyValuePair fwtypes.SetNestedObjectValueOf[resourceKeyValuePairModel] `tfsdk:"resource_key_value_pair"`
	KvsARN               fwtypes.ARN                                               `tfsdk:"key_value_store_arn"`
	MaximumBatchSize     types.Int64                                               `tfsdk:"max_batch_size"`
	TotalSizeInBytes     types.Int64                                               `tfsdk:"total_size_in_bytes"`
}
