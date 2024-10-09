// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package gamelift

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/gamelift"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_gamelift_alias", &resource.Sweeper{
		Name: "aws_gamelift_alias",
		Dependencies: []string{
			"aws_gamelift_fleet",
		},
		F: sweepAliases,
	})

	resource.AddTestSweepers("aws_gamelift_build", &resource.Sweeper{
		Name: "aws_gamelift_build",
		F:    sweepBuilds,
	})

	resource.AddTestSweepers("aws_gamelift_script", &resource.Sweeper{
		Name: "aws_gamelift_script",
		F:    sweepScripts,
	})

	resource.AddTestSweepers("aws_gamelift_fleet", &resource.Sweeper{
		Name: "aws_gamelift_fleet",
		Dependencies: []string{
			"aws_gamelift_build",
		},
		F: sweepFleets,
	})

	resource.AddTestSweepers("aws_gamelift_game_server_group", &resource.Sweeper{
		Name: "aws_gamelift_game_server_group",
		F:    sweepGameServerGroups,
	})

	resource.AddTestSweepers("aws_gamelift_game_session_queue", &resource.Sweeper{
		Name: "aws_gamelift_game_session_queue",
		F:    sweepGameSessionQueue,
	})
}

func sweepAliases(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	input := &gamelift.ListAliasesInput{}
	conn := client.GameLiftClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := gamelift.NewListAliasesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping GameLift Alias sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing GameLift Aliases (%s): %w", region, err)
		}

		for _, v := range page.Aliases {
			r := resourceAlias()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.AliasId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping GameLift Aliases (%s): %w", region, err)
	}

	return nil
}

func sweepBuilds(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	input := &gamelift.ListBuildsInput{}
	conn := client.GameLiftClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := gamelift.NewListBuildsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping GameLift Build sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing GameLift Builds (%s): %w", region, err)
		}

		for _, v := range page.Builds {
			r := resourceBuild()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.BuildId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping GameLift Builds (%s): %w", region, err)
	}

	return nil
}

func sweepScripts(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	input := &gamelift.ListScriptsInput{}
	conn := client.GameLiftClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := gamelift.NewListScriptsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping GameLift Script sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing GameLift Scripts (%s): %w", region, err)
		}

		for _, v := range page.Scripts {
			r := resourceScript()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.ScriptId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping GameLift Scripts (%s): %w", region, err)
	}

	return nil
}

func sweepFleets(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	input := &gamelift.ListFleetsInput{}
	conn := client.GameLiftClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := gamelift.NewListFleetsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping GameLift Fleet sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing GameLift Fleets (%s): %w", region, err)
		}

		for _, v := range page.FleetIds {
			r := resourceFleet()
			d := r.Data(nil)
			d.SetId(v)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping GameLift Fleets (%s): %w", region, err)
	}

	return nil
}

func sweepGameServerGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.GameLiftClient(ctx)
	input := &gamelift.ListGameServerGroupsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := gamelift.NewListGameServerGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping GameLift Game Server Group sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing GameLift Game Server Groups (%s): %w", region, err)
		}

		for _, v := range page.GameServerGroups {
			r := resourceGameServerGroup()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.GameServerGroupName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping GameLift Game Server Groups (%s): %w", region, err)
	}

	return nil
}

func sweepGameSessionQueue(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	input := &gamelift.DescribeGameSessionQueuesInput{}
	conn := client.GameLiftClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := gamelift.NewDescribeGameSessionQueuesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping GameLift Game Session Queue sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing GameLift Game Session Queues (%s): %w", region, err)
		}

		for _, v := range page.GameSessionQueues {
			r := resourceGameSessionQueue()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Name))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping GameLift Game Session Queues (%s): %w", region, err)
	}

	return nil
}
