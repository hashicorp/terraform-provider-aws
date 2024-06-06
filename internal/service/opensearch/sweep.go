// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearch

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/opensearchservice"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv1"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_opensearch_domain", &resource.Sweeper{
		Name: "aws_opensearch_domain",
		F:    sweepDomains,
		Dependencies: []string{
			"aws_opensearch_inbound_connection_accepter",
			"aws_opensearch_outbound_connection",
		},
	})

	resource.AddTestSweepers("aws_opensearch_inbound_connection_accepter", &resource.Sweeper{
		Name: "aws_opensearch_inbound_connection_accepter",
		F:    sweepInboundConnections,
	})

	resource.AddTestSweepers("aws_opensearch_outbound_connection", &resource.Sweeper{
		Name: "aws_opensearch_outbound_connection",
		F:    sweepOutboundConnections,
	})
}

func sweepDomains(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.OpenSearchConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &opensearchservice.ListDomainNamesInput{}

	// ListDomainNames has no pagination support whatsoever
	output, err := conn.ListDomainNamesWithContext(ctx, input)

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping OpenSearch Domain sweep for %s: %s", region, err)
		return errs.ErrorOrNil()
	}

	if err != nil {
		sweeperErr := fmt.Errorf("error listing OpenSearch Domains: %w", err)
		log.Printf("[ERROR] %s", sweeperErr)
		errs = multierror.Append(errs, sweeperErr)
		return errs.ErrorOrNil()
	}

	if output == nil {
		log.Printf("[WARN] Skipping OpenSearch Domain sweep for %s: empty response", region)
		return errs.ErrorOrNil()
	}

	for _, domainInfo := range output.DomainNames {
		if domainInfo == nil {
			continue
		}

		name := aws.StringValue(domainInfo.DomainName)

		if engineType := aws.StringValue(domainInfo.EngineType); engineType != opensearchservice.EngineTypeOpenSearch {
			log.Printf("[INFO] Skipping OpenSearch Domain %s: EngineType = %s", name, engineType)
			continue
		}

		// OpenSearch Domains have regularly gotten stuck in a "being deleted" state
		// e.g. Deleted and Processing are both true for days in the API
		// Filter out domains that are Deleted already.

		output, err := FindDomainByName(ctx, conn, name)
		if err != nil {
			sweeperErr := fmt.Errorf("error describing OpenSearch Domain (%s): %w", name, err)
			log.Printf("[ERROR] %s", sweeperErr)
			errs = multierror.Append(errs, sweeperErr)
			continue
		}

		if output != nil && aws.BoolValue(output.Deleted) {
			log.Printf("[INFO] Skipping OpenSearch Domain (%s) with deleted status", name)
			continue
		}

		r := ResourceDomain()
		d := r.Data(nil)
		d.SetId(name)
		d.Set(names.AttrDomainName, name)

		sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping OpenSearch Domains for %s: %w", region, err))
	}

	if awsv1.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping OpenSearch Domain sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepInboundConnections(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.OpenSearchConn(ctx)
	input := &opensearchservice.DescribeInboundConnectionsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.DescribeInboundConnectionsPagesWithContext(ctx, input, func(page *opensearchservice.DescribeInboundConnectionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Connections {
			id := aws.StringValue(v.ConnectionId)

			status := aws.StringValue(v.ConnectionStatus.StatusCode)
			if status == opensearchservice.InboundConnectionStatusCodeDeleted || status == opensearchservice.InboundConnectionStatusCodeRejected {
				log.Printf("[INFO] Skipping OpenSearch Inbound Connection %s: %s", id, status)
				continue
			}

			r := ResourceInboundConnectionAccepter()
			d := r.Data(nil)
			d.SetId(id)
			d.Set("connection_status", status)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping OpenSearch Inbound Connection sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing OpenSearch Inbound Connections: %w", err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping OpenSearch Inbound Connections (%s): %w", region, err)
	}

	return nil
}

func sweepOutboundConnections(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.OpenSearchConn(ctx)
	input := &opensearchservice.DescribeOutboundConnectionsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.DescribeOutboundConnectionsPagesWithContext(ctx, input, func(page *opensearchservice.DescribeOutboundConnectionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Connections {
			id := aws.StringValue(v.ConnectionId)

			if status := aws.StringValue(v.ConnectionStatus.StatusCode); status == opensearchservice.InboundConnectionStatusCodeDeleted {
				log.Printf("[INFO] Skipping OpenSearch Outbound Connection %s: %s", id, status)
				continue
			}

			r := ResourceOutboundConnection()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping OpenSearch Outbound Connection sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing OpenSearch Outbound Connections: %w", err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping OpenSearch Outbound Connections (%s): %w", region, err)
	}

	return nil
}
