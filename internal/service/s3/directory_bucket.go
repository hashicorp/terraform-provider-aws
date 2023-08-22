// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Directory Bucket")
// @Tags(identifierAttribute="id")
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
		names.AttrID: framework.IDAttribute(),
	}
}

func (r *resourceDirectoryBucket) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
}

func (r *resourceDirectoryBucket) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
}

func (r *resourceDirectoryBucket) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
}

func (r *resourceDirectoryBucket) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
}
