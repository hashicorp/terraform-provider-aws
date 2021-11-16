//go:build sweep
// +build sweep

package gamelift

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/gamelift"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
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

	resource.AddTestSweepers("aws_gamelift_fleet", &resource.Sweeper{
		Name: "aws_gamelift_fleet",
		Dependencies: []string{
			"aws_gamelift_build",
		},
		F: sweepFleets,
	})

	resource.AddTestSweepers("aws_gamelift_game_session_queue", &resource.Sweeper{
		Name: "aws_gamelift_game_session_queue",
		F:    sweepGameSessionQueue,
	})
}

func sweepAliases(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).GameLiftConn

	err = listAliases(&gamelift.ListAliasesInput{}, conn, func(resp *gamelift.ListAliasesOutput) error {
		if len(resp.Aliases) == 0 {
			log.Print("[DEBUG] No Gamelift Aliases to sweep")
			return nil
		}

		log.Printf("[INFO] Found %d Gamelift Aliases", len(resp.Aliases))

		for _, alias := range resp.Aliases {
			log.Printf("[INFO] Deleting Gamelift Alias %q", *alias.AliasId)
			_, err := conn.DeleteAlias(&gamelift.DeleteAliasInput{
				AliasId: alias.AliasId,
			})
			if err != nil {
				return fmt.Errorf("Error deleting Gamelift Alias (%s): %s",
					*alias.AliasId, err)
			}
		}
		return nil
	})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Gamelift Alias sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error listing Gamelift Aliases: %s", err)
	}

	return nil
}

func sweepBuilds(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).GameLiftConn

	resp, err := conn.ListBuilds(&gamelift.ListBuildsInput{})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Gamelife Build sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error listing Gamelift Builds: %s", err)
	}

	if len(resp.Builds) == 0 {
		log.Print("[DEBUG] No Gamelift Builds to sweep")
		return nil
	}

	log.Printf("[INFO] Found %d Gamelift Builds", len(resp.Builds))

	for _, build := range resp.Builds {
		log.Printf("[INFO] Deleting Gamelift Build %q", *build.BuildId)
		_, err := conn.DeleteBuild(&gamelift.DeleteBuildInput{
			BuildId: build.BuildId,
		})
		if err != nil {
			return fmt.Errorf("Error deleting Gamelift Build (%s): %s",
				*build.BuildId, err)
		}
	}

	return nil
}

func sweepFleets(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).GameLiftConn

	return listFleets(conn, nil, region, func(fleetIds []*string) error {
		if len(fleetIds) == 0 {
			log.Print("[DEBUG] No Gamelift Fleets to sweep")
			return nil
		}

		out, err := conn.DescribeFleetAttributes(&gamelift.DescribeFleetAttributesInput{
			FleetIds: fleetIds,
		})
		if err != nil {
			return fmt.Errorf("Error describing Gamelift Fleet attributes: %s", err)
		}

		log.Printf("[INFO] Found %d Gamelift Fleets", len(out.FleetAttributes))

		for _, attr := range out.FleetAttributes {
			log.Printf("[INFO] Deleting Gamelift Fleet %q", *attr.FleetId)
			err := resource.Retry(60*time.Minute, func() *resource.RetryError {
				_, err := conn.DeleteFleet(&gamelift.DeleteFleetInput{
					FleetId: attr.FleetId,
				})
				if err != nil {
					msg := fmt.Sprintf("Cannot delete fleet %s that is in status of ", *attr.FleetId)
					if tfawserr.ErrMessageContains(err, gamelift.ErrCodeInvalidRequestException, msg) {
						return resource.RetryableError(err)
					}
					return resource.NonRetryableError(err)
				}
				return nil
			})
			if err != nil {
				return fmt.Errorf("Error deleting Gamelift Fleet (%s): %s",
					*attr.FleetId, err)
			}

			err = WaitForFleetToBeDeleted(conn, *attr.FleetId, FleetDeletedDefaultTimeout)
			if err != nil {
				return fmt.Errorf("Error waiting for Gamelift Fleet (%s) to be deleted: %s",
					*attr.FleetId, err)
			}
		}
		return nil
	})
}

func sweepGameSessionQueue(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).GameLiftConn

	out, err := conn.DescribeGameSessionQueues(&gamelift.DescribeGameSessionQueuesInput{})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Gamelife Queue sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Gamelift Session Queue: %s", err)
	}

	if len(out.GameSessionQueues) == 0 {
		log.Print("[DEBUG] No Gamelift Session Queue to sweep")
		return nil
	}

	log.Printf("[INFO] Found %d Gamelift Session Queue", len(out.GameSessionQueues))

	for _, queue := range out.GameSessionQueues {
		log.Printf("[INFO] Deleting Gamelift Session Queue %q", *queue.Name)
		_, err := conn.DeleteGameSessionQueue(&gamelift.DeleteGameSessionQueueInput{
			Name: aws.String(*queue.Name),
		})
		if err != nil {
			return fmt.Errorf("error deleting Gamelift Session Queue (%s): %s",
				*queue.Name, err)
		}
	}

	return nil
}

func listAliases(input *gamelift.ListAliasesInput, conn *gamelift.GameLift, f func(*gamelift.ListAliasesOutput) error) error {
	resp, err := conn.ListAliases(input)
	if err != nil {
		return err
	}
	err = f(resp)
	if err != nil {
		return err
	}

	if resp.NextToken != nil {
		return listAliases(input, conn, f)
	}
	return nil
}

func listFleets(conn *gamelift.GameLift, nextToken *string, region string, f func([]*string) error) error {
	resp, err := conn.ListFleets(&gamelift.ListFleetsInput{
		NextToken: nextToken,
	})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Gamelift Fleet sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error listing Gamelift Fleets: %s", err)
	}

	err = f(resp.FleetIds)
	if err != nil {
		return err
	}
	if nextToken != nil {
		return listFleets(conn, nextToken, region, f)
	}
	return nil
}
