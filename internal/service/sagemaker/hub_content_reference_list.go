// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @FrameworkListResource("aws_sagemaker_hub_content_reference")
func newHubContentReferenceResourceAsListResource() list.ListResourceWithConfigure {
	return &hubContentReferenceListResource{}
}

var _ list.ListResource = &hubContentReferenceListResource{}

type hubContentReferenceListResource struct {
	hubContentReferenceResource
	framework.WithList
}

func (l *hubContentReferenceListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.SageMakerClient(ctx)

	tflog.Info(ctx, "Listing SageMaker Hub Content Reference resources")

	stream.Results = func(yield func(list.ListResult) bool) {
		for item, err := range listHubContentReferences(ctx, conn) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			hubName := aws.ToString(item.hubName)
			hubContentName := aws.ToString(item.info.HubContentName)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey("hub_content_name"), hubContentName)

			result := request.NewListResult(ctx)
			result.DisplayName = hubContentName

			var data hubContentReferenceResourceModel
			l.SetResult(ctx, awsClient, request.IncludeResource, &data, &result, func() {
				output, err := findHubContentByName(ctx, conn, hubName, hubContentName, awstypes.HubContentTypeModelReference)
				if retry.NotFound(err) {
					tflog.Warn(ctx, "Resource disappeared during listing, skipping")
					return
				}
				if err != nil {
					result.Diagnostics.AddError(
						"Reading SageMaker Hub Content Reference",
						fmt.Sprintf("Error reading hub content reference (%s): %s", hubContentName, err),
					)
					return
				}

				result.Diagnostics.Append(l.flatten(ctx, output, &data)...)
			})

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

type hubContentReferenceListItem struct {
	hubName *string
	info    awstypes.HubContentInfo
}

func listHubContentReferences(ctx context.Context, conn *sagemaker.Client) iter.Seq2[hubContentReferenceListItem, error] {
	return func(yield func(hubContentReferenceListItem, error) bool) {
		var stopped bool
		hubInput := &sagemaker.ListHubsInput{}
		err := listHubsPages(ctx, conn, hubInput, func(page *sagemaker.ListHubsOutput, lastPage bool) bool {
			for _, hub := range page.HubSummaries {
				hubName := hub.HubName

				contentInput := &sagemaker.ListHubContentsInput{
					HubName:        hubName,
					HubContentType: awstypes.HubContentTypeModelReference,
				}

				err := listHubContentsPages(ctx, conn, contentInput, func(page *sagemaker.ListHubContentsOutput, lastPage bool) bool {
					for _, info := range page.HubContentSummaries {
						item := hubContentReferenceListItem{
							hubName: hubName,
							info:    info,
						}
						if !yield(item, nil) {
							stopped = true
							return false
						}
					}
					return true
				})
				if stopped {
					return false
				}
				if err != nil {
					yield(hubContentReferenceListItem{}, fmt.Errorf("listing SageMaker Hub Content Reference resources for hub (%s): %w", aws.ToString(hubName), err))
					return false
				}
			}
			return true
		})
		if !stopped && err != nil {
			yield(hubContentReferenceListItem{}, fmt.Errorf("listing SageMaker Hub Content Reference resources: %w", err))
		}
	}
}
