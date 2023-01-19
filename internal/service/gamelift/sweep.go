//go:build sweep
// +build sweep

package gamelift

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/gamelift"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
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
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).GameLiftConn()

	err = listAliases(ctx, &gamelift.ListAliasesInput{}, conn, func(resp *gamelift.ListAliasesOutput) error {
		if len(resp.Aliases) == 0 {
			log.Print("[DEBUG] No GameLift Aliases to sweep")
			return nil
		}

		log.Printf("[INFO] Found %d GameLift Aliases", len(resp.Aliases))

		for _, alias := range resp.Aliases {
			log.Printf("[INFO] Deleting GameLift Alias %q", *alias.AliasId)
			_, err := conn.DeleteAliasWithContext(ctx, &gamelift.DeleteAliasInput{
				AliasId: alias.AliasId,
			})
			if err != nil {
				return fmt.Errorf("Error deleting GameLift Alias (%s): %s",
					*alias.AliasId, err)
			}
		}
		return nil
	})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping GameLift Alias sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error listing GameLift Aliases: %s", err)
	}

	return nil
}

func sweepBuilds(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).GameLiftConn()

	resp, err := conn.ListBuildsWithContext(ctx, &gamelift.ListBuildsInput{})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Gamelife Build sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error listing GameLift Builds: %s", err)
	}

	if len(resp.Builds) == 0 {
		log.Print("[DEBUG] No GameLift Builds to sweep")
		return nil
	}

	log.Printf("[INFO] Found %d GameLift Builds", len(resp.Builds))

	for _, build := range resp.Builds {
		log.Printf("[INFO] Deleting GameLift Build %q", *build.BuildId)
		_, err := conn.DeleteBuildWithContext(ctx, &gamelift.DeleteBuildInput{
			BuildId: build.BuildId,
		})
		if err != nil {
			return fmt.Errorf("Error deleting GameLift Build (%s): %s",
				*build.BuildId, err)
		}
	}

	return nil
}

func sweepScripts(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).GameLiftConn()

	resp, err := conn.ListScriptsWithContext(ctx, &gamelift.ListScriptsInput{})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Gamelife Script sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error listing GameLift Scripts: %s", err)
	}

	if len(resp.Scripts) == 0 {
		log.Print("[DEBUG] No GameLift Scripts to sweep")
		return nil
	}

	log.Printf("[INFO] Found %d GameLift Scripts", len(resp.Scripts))

	for _, build := range resp.Scripts {
		log.Printf("[INFO] Deleting GameLift Script %q", *build.ScriptId)
		_, err := conn.DeleteScriptWithContext(ctx, &gamelift.DeleteScriptInput{
			ScriptId: build.ScriptId,
		})
		if err != nil {
			return fmt.Errorf("Error deleting GameLift Script (%s): %s",
				*build.ScriptId, err)
		}
	}

	return nil
}

func sweepFleets(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).GameLiftConn()
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &gamelift.ListFleetsInput{}

	for {
		output, err := conn.ListFleetsWithContext(ctx, input)

		for _, fleet := range output.FleetIds {
			r := ResourceFleet()
			d := r.Data(nil)

			id := aws.StringValue(fleet)
			d.SetId(id)

			if err != nil {
				err := fmt.Errorf("reading GameLift Fleet (%s): %w", id, err)
				log.Printf("[ERROR] %s", err)
				errs = multierror.Append(errs, err)
				continue
			}

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("listing GameLift Fleet for %s: %w", region, err))
	}

	if err := sweep.SweepOrchestratorWithContext(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("sweeping GameLift Fleet for %s: %w", region, err))
	}

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping GameLift Fleet sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepGameServerGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).GameLiftConn()
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &gamelift.ListGameServerGroupsInput{}

	for {
		output, err := conn.ListGameServerGroupsWithContext(ctx, input)

		for _, gameServerGroup := range output.GameServerGroups {
			r := ResourceGameServerGroup()
			d := r.Data(nil)

			id := aws.StringValue(gameServerGroup.GameServerGroupName)
			d.SetId(id)

			if err != nil {
				err := fmt.Errorf("reading GameLift Game Server Group (%s): %w", id, err)
				log.Printf("[ERROR] %s", err)
				errs = multierror.Append(errs, err)
				continue
			}

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("listing GameLift Game Server Group for %s: %w", region, err))
	}

	if err := sweep.SweepOrchestratorWithContext(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("sweeping GameLift Game Server Group for %s: %w", region, err))
	}

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping GameLift Game Server Group sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepGameSessionQueue(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).GameLiftConn()

	out, err := conn.DescribeGameSessionQueuesWithContext(ctx, &gamelift.DescribeGameSessionQueuesInput{})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Gamelife Queue sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("listing GameLift Session Queue: %s", err)
	}

	if len(out.GameSessionQueues) == 0 {
		log.Print("[DEBUG] No GameLift Session Queue to sweep")
		return nil
	}

	log.Printf("[INFO] Found %d GameLift Session Queue", len(out.GameSessionQueues))

	for _, queue := range out.GameSessionQueues {
		log.Printf("[INFO] Deleting GameLift Session Queue %q", *queue.Name)
		_, err := conn.DeleteGameSessionQueueWithContext(ctx, &gamelift.DeleteGameSessionQueueInput{
			Name: aws.String(*queue.Name),
		})
		if err != nil {
			return fmt.Errorf("deleting GameLift Session Queue (%s): %s",
				*queue.Name, err)
		}
	}

	return nil
}

func listAliases(ctx context.Context, input *gamelift.ListAliasesInput, conn *gamelift.GameLift, f func(*gamelift.ListAliasesOutput) error) error {
	resp, err := conn.ListAliasesWithContext(ctx, input)
	if err != nil {
		return err
	}
	err = f(resp)
	if err != nil {
		return err
	}

	if resp.NextToken != nil {
		return listAliases(ctx, input, conn, f)
	}
	return nil
}
