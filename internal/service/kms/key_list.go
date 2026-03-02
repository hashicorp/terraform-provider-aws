// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package kms

import (
	"context"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_kms_key")
func newKeyResourceAsListResource() inttypes.ListResourceForSDK {
	l := keyListResource{}
	l.SetResourceSchema(resourceKey())
	return &l
}

type keyListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type keyListResourceModel struct {
	framework.WithRegionModel
}

func (l *keyListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	var query keyListResourceModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	awsClient := l.Meta()
	conn := awsClient.KMSClient(ctx)

	tflog.Info(ctx, "Listing KMS keys")
	stream.Results = func(yield func(list.ListResult) bool) {
		var input kms.ListKeysInput
		for item, err := range listKeys(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			id := aws.ToString(item.KeyId)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)

			result := request.NewListResult(ctx)
			rd := l.ResourceData()
			rd.SetId(id)

			key, err := findKeyInfo(ctx, conn, id, false)
			if err != nil {
				tflog.Error(ctx, "Reading KMS key", map[string]any{
					names.AttrID: id,
					"err":        err.Error(),
				})
				continue
			}

			diags := resourceKeyFlatten(ctx, rd, key)
			if diags.HasError() || rd.Id() == "" {
				// Resource can't be read or is logically deleted.
				// Log and continue.
				tflog.Error(ctx, "Reading KMS key", map[string]any{
					names.AttrID: id,
					"diags":      sdkdiag.DiagnosticsString(diags),
				})
				continue
			}

			result.DisplayName = id

			l.SetResult(ctx, awsClient, request.IncludeResource, &result, rd)
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

func listKeys(ctx context.Context, conn *kms.Client, input *kms.ListKeysInput) iter.Seq2[awstypes.KeyListEntry, error] {
	return func(yield func(awstypes.KeyListEntry, error) bool) {
		pages := kms.NewListKeysPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.KeyListEntry{}, err)
				return
			}

			for _, key := range page.Keys {
				if !yield(key, nil) {
					return
				}
			}
		}
	}
}
