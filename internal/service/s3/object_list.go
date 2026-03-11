// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @SDKListResource("aws_s3_object")
func newObjectResourceAsListResource() inttypes.ListResourceForSDK {
	l := listResourceObject{}
	l.SetResourceSchema(resourceObject())
	return &l
}

var _ list.ListResource = &listResourceObject{}
var _ list.ListResourceWithRawV5Schemas = &listResourceObject{}

type listResourceObject struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *listResourceObject) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			names.AttrBucket: listschema.StringAttribute{
				Required:    true,
				Description: "Name of the S3 bucket to list objects from.",
			},
			names.AttrPrefix: listschema.StringAttribute{
				Optional:    true,
				Description: "Limits the response to keys that begin with the specified prefix.",
			},
		},
	}
}

func (l *listResourceObject) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().S3Client(ctx)

	var query listObjectModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	bucket := query.Bucket.ValueString()
	if isDirectoryBucket(bucket) {
		conn = l.Meta().S3ExpressClient(ctx)
	}

	tflog.Info(ctx, "Listing S3 Object", map[string]any{
		names.AttrBucket: bucket,
	})
	stream.Results = func(yield func(list.ListResult) bool) {
		input := s3.ListObjectsV2Input{
			Bucket: aws.String(bucket),
		}
		if !query.Prefix.IsNull() {
			input.Prefix = query.Prefix.ValueStringPointer()
		}
		for item, err := range listObjects(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			key := aws.ToString(item.Key)
			id := fmt.Sprintf("%s/%s", bucket, key)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)

			result := request.NewListResult(ctx)
			rd := l.ResourceData()
			rd.SetId(id)
			rd.Set(names.AttrBucket, bucket)
			rd.Set(names.AttrKey, key)

			tflog.Info(ctx, "Reading S3 Object")
			diags := resourceObjectRead(ctx, rd, l.Meta())
			if diags.HasError() {
				tflog.Error(ctx, "Reading S3 Object", map[string]any{
					names.AttrID: id,
					"diags":      sdkdiag.DiagnosticsString(diags),
				})
				continue
			}
			if rd.Id() == "" || rd.Id() == "/" {
				// Resource is logically deleted
				continue
			}

			result.DisplayName = fmt.Sprintf("%s/%s", bucket, key)

			l.SetResult(ctx, l.Meta(), request.IncludeResource, &result, rd)
			if result.Diagnostics.HasError() {
				yield(result)
				return
			}

			if !yield(result) {
				return
			}
		}
	}
}

type listObjectModel struct {
	framework.WithRegionModel
	Bucket types.String `tfsdk:"bucket"`
	Prefix types.String `tfsdk:"prefix"`
}

func listObjects(ctx context.Context, conn *s3.Client, input *s3.ListObjectsV2Input) iter.Seq2[awstypes.Object, error] {
	return func(yield func(awstypes.Object, error) bool) {
		pages := s3.NewListObjectsV2Paginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.Object{}, fmt.Errorf("listing S3 Object resources: %w", err))
				return
			}

			for _, item := range page.Contents {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
