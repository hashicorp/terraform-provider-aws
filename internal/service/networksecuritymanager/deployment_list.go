// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networksecuritymanager

import (
	"context"
	"iter"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networksecuritymanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networksecuritymanager/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tfiter "github.com/hashicorp/terraform-provider-aws/internal/iter"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_networksecuritymanager_deployment")
func newDeploymentResourceAsListResource() list.ListResourceWithConfigure {
	return &deploymentListResource{}
}

var _ list.ListResource = &deploymentListResource{}

type deploymentListResource struct {
	deploymentResource
	framework.WithList
}

func (l *deploymentListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().NetworkSecurityManagerClient(ctx)

	stream.Results = func(yield func(list.ListResult) bool) {
		var input networksecuritymanager.ListDeploymentsInput
		seen := make(map[string]struct{})
		for item, err := range listDeployments(ctx, conn, &input) {
			if err != nil {
				yield(smerr.NewListResultError(ctx, err))
				return
			}

			// A draft is listed with its DRAFT-qualified ARN.
			arn := deploymentBaseARN(aws.ToString(item.DeploymentArn))
			if _, ok := seen[arn]; ok {
				continue
			}
			seen[arn] = struct{}{}
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)

			var deployment *networksecuritymanager.GetDeploymentOutput
			if request.IncludeResource {
				var err error
				deployment, err = findDeploymentByARN(ctx, conn, arn)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					yield(smerr.NewListResultError(ctx, err, smerr.ID, arn))
					return
				}
			}

			result := request.NewListResult(ctx)

			var data deploymentResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.ARN = fwflex.StringValueToFramework(ctx, arn)

				if request.IncludeResource {
					smerr.AddEnrich(ctx, &result.Diagnostics, l.flatten(ctx, deployment, &data), smerr.ID, arn)
					if result.Diagnostics.HasError() {
						return
					}
				}

				result.DisplayName = aws.ToString(item.DeploymentName)
			})

			if result.Diagnostics.HasError() {
				// The deployment can be deleted after it is listed, which fails reading its tags.
				if request.IncludeResource {
					if _, err := findDeploymentByARN(ctx, conn, arn); retry.NotFound(err) {
						continue
					}
				}

				yield(list.ListResult{Diagnostics: result.Diagnostics})
				return
			}
			if !yield(result) {
				return
			}
		}
	}
}

func listDeployments(ctx context.Context, conn *networksecuritymanager.Client, input *networksecuritymanager.ListDeploymentsInput) iter.Seq2[awstypes.DeploymentSummary, error] {
	return tfiter.ConcatValuesWithError(listDeploymentPages(ctx, conn, input))
}

func listDeploymentPages(ctx context.Context, conn *networksecuritymanager.Client, input *networksecuritymanager.ListDeploymentsInput) iter.Seq2[[]awstypes.DeploymentSummary, error] {
	return func(yield func([]awstypes.DeploymentSummary, error) bool) {
		pages := networksecuritymanager.NewListDeploymentsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(nil, smarterr.NewError(err))
				return
			}

			if !yield(page.Deployments, nil) {
				return
			}
		}
	}
}
