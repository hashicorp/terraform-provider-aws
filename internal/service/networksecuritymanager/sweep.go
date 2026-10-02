// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networksecuritymanager

import (
	"context"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networksecuritymanager"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_networksecuritymanager_deployment", sweepDeployments)
	awsv2.Register("aws_networksecuritymanager_scope", sweepScopes, "aws_networksecuritymanager_deployment")
}

func sweepDeployments(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := networksecuritymanager.ListDeploymentsInput{}
	conn := client.NetworkSecurityManagerClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := networksecuritymanager.NewListDeploymentsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.Deployments {
			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newDeploymentResource, client,
				sweepfw.NewAttribute(names.AttrARN, aws.ToString(v.DeploymentArn))),
			)
		}
	}

	return sweepResources, nil
}

func sweepScopes(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := networksecuritymanager.ListScopesInput{}
	conn := client.NetworkSecurityManagerClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := networksecuritymanager.NewListScopesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.Scopes {
			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newScopeResource, client,
				sweepfw.NewAttribute(names.AttrARN, aws.ToString(v.ScopeArn))),
			)
		}
	}

	return sweepResources, nil
}
