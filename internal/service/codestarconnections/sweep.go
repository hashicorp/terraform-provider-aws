// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package codestarconnections

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codestarconnections"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.CodeStarConnectionsConn(ctx)
	input := &codestarconnections.ListConnectionsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListConnectionsPagesWithContext(ctx, input, func(page *codestarconnections.ListConnectionsOutput, lastPage bool) bool {
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

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping CodeStar Connections Connections (%s): %w", region, err)
	}

	return nil
}

func sweepHosts(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.CodeStarConnectionsConn(ctx)
	input := &codestarconnections.ListHostsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListHostsPagesWithContext(ctx, input, func(page *codestarconnections.ListHostsOutput, lastPage bool) bool {
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

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping CodeStar Connections Hosts (%s): %w", region, err)
	}

	return nil
}
