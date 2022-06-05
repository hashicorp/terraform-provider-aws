//go:build sweep
// +build sweep

package codestarconnections

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codestarconnections"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_codestarconnections_connection", &resource.Sweeper{
		Name: "aws_codestarconnections_connection",
		F:    sweepConnections,
	})

	resource.AddTestSweepers("aws_codestarconnections_host", &resource.Sweeper{
		Name: "aws_codestarconnections_host",
		F:    sweepHosts,
		Dependencies: []string{
			"aws_codestarconnections_connection",
		},
	})
}

func sweepConnections(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).CodeStarConnectionsConn
	input := &codestarconnections.ListConnectionsInput{}
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.ListConnectionsPages(input, func(page *codestarconnections.ListConnectionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Connections {
			r := ResourceConnection()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.ConnectionArn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CodeStar Connections Connection sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing CodeStar Connections Connections (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping CodeStar Connections Connections (%s): %w", region, err)
	}

	return nil
}

func sweepHosts(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).CodeStarConnectionsConn
	input := &codestarconnections.ListHostsInput{}
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.ListHostsPages(input, func(page *codestarconnections.ListHostsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Hosts {
			r := ResourceHost()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.HostArn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CodeStar Connections Host sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing CodeStar Connections Hosts (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping CodeStar Connections Hosts (%s): %w", region, err)
	}

	return nil
}
