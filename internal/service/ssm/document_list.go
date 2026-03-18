// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ssm

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_ssm_document")
func newDocumentResourceAsListResource() inttypes.ListResourceForSDK {
	l := documentListResource{}
	l.SetResourceSchema(resourceDocument())

	return &l
}

type documentListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type documentListResourceModel struct {
	framework.WithRegionModel
}

func (l *documentListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.SSMClient(ctx)

	var query documentListResourceModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	tflog.Info(ctx, "Listing SSM documents")

	stream.Results = func(yield func(list.ListResult) bool) {
		input := &ssm.ListDocumentsInput{
			Filters: []awstypes.DocumentKeyValuesFilter{{
				Key:    aws.String("Owner"),
				Values: []string{"Self"},
			}},
		}

		for item, err := range listDocuments(ctx, conn, input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			name := aws.ToString(item.Name)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrName), name)

			result := request.NewListResult(ctx)
			rd := l.ResourceData()
			rd.SetId(name)
			rd.Set(names.AttrName, name)

			if request.IncludeResource {
				doc, err := findDocumentByName(ctx, conn, name)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					result = fwdiag.NewListResultErrorDiagnostic(fmt.Errorf("reading SSM Document (%s): %w", name, err))
					yield(result)
					return
				}

				diags := resourceDocumentFlatten(ctx, conn, rd, awsClient, doc)
				if diags.HasError() {
					result = fwdiag.NewListResultSDKDiagnostics(diags)
					yield(result)
					return
				}
			}

			result.DisplayName = name

			l.SetResult(ctx, awsClient, request.IncludeResource, rd, &result)
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

func listDocuments(ctx context.Context, conn *ssm.Client, input *ssm.ListDocumentsInput) iter.Seq2[awstypes.DocumentIdentifier, error] {
	return func(yield func(awstypes.DocumentIdentifier, error) bool) {
		pages := ssm.NewListDocumentsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.DocumentIdentifier{}, fmt.Errorf("listing SSM Documents: %w", err))
				return
			}

			for _, item := range page.DocumentIdentifiers {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
