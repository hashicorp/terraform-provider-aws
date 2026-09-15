// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package resiliencehub

import (
	"context"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resiliencehub"
	awstypes "github.com/aws/aws-sdk-go-v2/service/resiliencehub/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const appSweepDeleteTimeout = 15 * time.Minute

func RegisterSweepers() {
	awsv2.Register("aws_resiliencehub_app", sweepApps)
	awsv2.Register("aws_resiliencehub_resiliency_policy", sweepResiliencyPolicy)
}

func sweepApps(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.ResilienceHubClient(ctx)

	var sweepResources []sweep.Sweepable

	pages := resiliencehub.NewListAppsPaginator(conn, &resiliencehub.ListAppsInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, app := range page.AppSummaries {
			sweepResources = append(sweepResources, &appSweeper{
				client: client,
				arn:    aws.ToString(app.AppArn),
			})
		}
	}

	return sweepResources, nil
}

// appSweeper force-deletes a Resilience Hub application. The resource's own
// Delete does not force deletion, but a swept application may have published
// versions or imported resources that can only be removed with ForceDelete.
type appSweeper struct {
	client *conns.AWSClient
	arn    string
}

func (s *appSweeper) Delete(ctx context.Context, optFns ...tfresource.OptionsFunc) error {
	conn := s.client.ResilienceHubClient(ctx)

	_, err := conn.DeleteApp(ctx, &resiliencehub.DeleteAppInput{
		AppArn:      aws.String(s.arn),
		ClientToken: aws.String(create.UniqueId(ctx)),
		ForceDelete: aws.Bool(true),
	})
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil
	}
	if err != nil {
		return smarterr.NewError(err)
	}

	if err := waitAppDeleted(ctx, conn, s.arn, appSweepDeleteTimeout); err != nil {
		return smarterr.NewError(err)
	}

	return nil
}

func sweepResiliencyPolicy(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.ResilienceHubClient(ctx)

	var sweepResources []sweep.Sweepable

	pages := resiliencehub.NewListResiliencyPoliciesPaginator(conn, &resiliencehub.ListResiliencyPoliciesInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, policies := range page.ResiliencyPolicies {
			sweepResources = append(sweepResources, framework.NewSweepResource(newResiliencyPolicyResource, client,
				framework.NewAttribute(names.AttrARN, aws.ToString(policies.PolicyArn)),
			))
		}
	}

	return sweepResources, nil
}
