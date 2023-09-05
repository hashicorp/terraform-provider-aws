// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Directory Bucket")
func newResourceDirectoryBucket(context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceDirectoryBucket{}

	return r, nil
}

type resourceDirectoryBucket struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
}

func (r *resourceDirectoryBucket) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_s3_directory_bucket"
}

func (r *resourceDirectoryBucket) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"bucket": schema.StringAttribute{
				Required: true,
			},
			names.AttrID: framework.IDAttribute(),
		},
	}
}

func (r *resourceDirectoryBucket) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resourceDirectoryBucketData

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3Client(ctx)

	input := &s3.CreateBucketInput{
		Bucket: flex.StringFromFramework(ctx, data.Bucket),
	}

	_, err := conn.CreateBucket(ctx, input)

	if err != nil {
		response.Diagnostics.AddError("creating S3 Directory Bucket", err.Error())

		return
	}

	// Set values for unknowns.
	arn := arn.ARN{
		Partition: r.Meta().Partition,
		Service:   "s3beta2022a",
		Region:    r.Meta().Region,
		AccountID: r.Meta().AccountID,
		Resource:  data.Bucket.ValueString(),
	}.String()
	data.ARN = types.StringValue(arn)
	data.ID = data.Bucket

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceDirectoryBucket) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resourceDirectoryBucketData

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3Client(ctx)

	input := &s3.HeadBucketInput{
		Bucket: flex.StringFromFramework(ctx, data.ID),
	}

	_, err := conn.HeadBucket(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading S3 Directory Bucket (%s)", data.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceDirectoryBucket) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new resourceDirectoryBucketData

	response.Diagnostics.Append(request.State.Get(ctx, &old)...)

	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)

	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *resourceDirectoryBucket) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resourceDirectoryBucketData

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3Client(ctx)

	_, err := conn.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: flex.StringFromFramework(ctx, data.ID),
	})

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting S3 Directory Bucket (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

type resourceDirectoryBucketData struct {
	ARN    types.String `tfsdk:"arn"`
	Bucket types.String `tfsdk:"bucket"`
	ID     types.String `tfsdk:"id"`
}
