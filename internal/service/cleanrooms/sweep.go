// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cleanrooms

import (
	"context"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cleanrooms"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/sdk"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_cleanrooms_collaboration", sweepCollaborations)
	awsv2.Register("aws_cleanrooms_configured_table", sweepConfiguredTables)
	awsv2.Register("aws_cleanrooms_membership", sweepMemberships)
}

func sweepCollaborations(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.CleanRoomsClient(ctx)
	input := &cleanrooms.ListCollaborationsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := cleanrooms.NewListCollaborationsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, c := range page.CollaborationList {
			id := aws.ToString(c.Id)
			r := resourceCollaboration()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sdk.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepConfiguredTables(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.CleanRoomsClient(ctx)
	input := &cleanrooms.ListConfiguredTablesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := cleanrooms.NewListConfiguredTablesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, c := range page.ConfiguredTableSummaries {
			id := aws.ToString(c.Id)
			r := resourceConfiguredTable()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sdk.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepMemberships(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.CleanRoomsClient(ctx)
	input := &cleanrooms.ListMembershipsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := cleanrooms.NewListMembershipsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, c := range page.MembershipSummaries {
			id := aws.ToString(c.Id)

			sweepResources = append(sweepResources, framework.NewSweepResource(newMembershipResource, client,
				framework.NewAttribute(names.AttrID, id),
			))
		}
	}

	return sweepResources, nil
}
