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
	awsv2.Register("aws_networksecuritymanager_scope", sweepScopes)
}

func sweepScopes(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	var input networksecuritymanager.ListScopesInput
	conn := client.NetworkSecurityManagerClient(ctx)
	var sweepResources []sweep.Sweepable

	for v, err := range listScopes(ctx, conn, &input) {
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		sweepResources = append(sweepResources, sweepfw.NewSweepResource(newScopeResource, client,
			sweepfw.NewAttribute(names.AttrARN, aws.ToString(v.ScopeArn))),
		)
	}

	return sweepResources, nil
}
