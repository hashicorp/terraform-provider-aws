// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package opensearch

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/opensearch"
	awstypes "github.com/aws/aws-sdk-go-v2/service/opensearch/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_opensearch_application", sweepApplications)
	awsv2.Register("aws_opensearch_domain", sweepDomains, "aws_opensearch_inbound_connection_accepter", "aws_opensearch_outbound_connection")
	awsv2.Register("aws_opensearch_inbound_connection_accepter", sweepInboundConnections)
	awsv2.Register("aws_opensearch_outbound_connection", sweepOutboundConnections)
}

func sweepApplications(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.OpenSearchClient(ctx)
	var input opensearch.ListApplicationsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := opensearch.NewListApplicationsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.ApplicationSummaries {
			sweepResources = append(sweepResources, framework.NewSweepResource(newResourceApplication, client,
				framework.NewAttribute(names.AttrID, aws.ToString(v.Id))),
			)
		}
	}

	return sweepResources, nil
}

func sweepDomains(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.OpenSearchClient(ctx)
	var input opensearch.ListDomainNamesInput
	sweepResources := make([]sweep.Sweepable, 0)

	output, err := conn.ListDomainNames(ctx, &input)
	if err != nil {
		return nil, err
	}

	for _, v := range output.DomainNames {
		name := aws.ToString(v.DomainName)

		if engineType := v.EngineType; engineType != awstypes.EngineTypeOpenSearch {
			log.Printf("[INFO] Skipping OpenSearch Domain %s: EngineType=%s", name, engineType)
			continue
		}

		// OpenSearch Domains have regularly gotten stuck in a "being deleted" state
		// e.g. Deleted and Processing are both true for days in the API
		// Filter out domains that are Deleted already.

		output, err := findDomainByName(ctx, conn, name)
		if err != nil {
			continue
		}

		if output != nil && aws.ToBool(output.Deleted) {
			log.Printf("[INFO] Skipping OpenSearch Domain %s: Deleted", name)
			continue
		}

		r := resourceDomain()
		d := r.Data(nil)
		d.SetId(name)
		d.Set(names.AttrDomainName, name)

		sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
	}

	return sweepResources, nil
}

func sweepInboundConnections(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.OpenSearchClient(ctx)
	var input opensearch.DescribeInboundConnectionsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := opensearch.NewDescribeInboundConnectionsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.Connections {
			id := aws.ToString(v.ConnectionId)

			status := v.ConnectionStatus.StatusCode
			if status == awstypes.InboundConnectionStatusCodeDeleted || status == awstypes.InboundConnectionStatusCodeRejected {
				log.Printf("[INFO] Skipping OpenSearch Inbound Connection %s: Status=%s", id, status)
				continue
			}

			r := resourceInboundConnectionAccepter()
			d := r.Data(nil)
			d.SetId(id)
			d.Set("connection_status", status)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepOutboundConnections(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.OpenSearchClient(ctx)
	var input opensearch.DescribeOutboundConnectionsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := opensearch.NewDescribeOutboundConnectionsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.Connections {
			id := aws.ToString(v.ConnectionId)

			if status := v.ConnectionStatus.StatusCode; status == awstypes.OutboundConnectionStatusCodeDeleted {
				log.Printf("[INFO] Skipping OpenSearch Outbound Connection %s: Status=%s", id, status)
				continue
			}

			r := resourceOutboundConnection()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}
