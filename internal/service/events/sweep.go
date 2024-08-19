// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package events

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_cloudwatch_event_api_destination", &resource.Sweeper{
		Name: "aws_cloudwatch_event_api_destination",
		F:    sweepAPIDestination,
		Dependencies: []string{
			"aws_cloudwatch_event_connection",
		},
	})

	resource.AddTestSweepers("aws_cloudwatch_event_archive", &resource.Sweeper{
		Name: "aws_cloudwatch_event_archive",
		F:    sweepArchives,
		Dependencies: []string{
			"aws_cloudwatch_event_bus",
		},
	})

	resource.AddTestSweepers("aws_cloudwatch_event_bus", &resource.Sweeper{
		Name: "aws_cloudwatch_event_bus",
		F:    sweepBuses,
		Dependencies: []string{
			"aws_cloudwatch_event_rule",
			"aws_cloudwatch_event_target",
			"aws_schemas_discoverer",
		},
	})

	resource.AddTestSweepers("aws_cloudwatch_event_connection", &resource.Sweeper{
		Name: "aws_cloudwatch_event_connection",
		F:    sweepConnection,
	})

	resource.AddTestSweepers("aws_cloudwatch_event_rule", &resource.Sweeper{
		Name: "aws_cloudwatch_event_rule",
		F:    sweepRules,
		Dependencies: []string{
			"aws_cloudwatch_event_target",
		},
	})

	resource.AddTestSweepers("aws_cloudwatch_event_target", &resource.Sweeper{
		Name: "aws_cloudwatch_event_target",
		F:    sweepTargets,
	})
}

func sweepAPIDestination(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}
	conn := client.EventsClient(ctx)
	input := &eventbridge.ListApiDestinationsInput{
		Limit: aws.Int32(100),
	}
	sweepResources := make([]sweep.Sweepable, 0)

	err = listAPIDestinationsPages(ctx, conn, input, func(page *eventbridge.ListApiDestinationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ApiDestinations {
			r := resourceAPIDestination()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Name))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EventBridge API Destination sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing EventBridge API Destinations (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EventBridge API Destinations (%s): %w", region, err)
	}

	return nil
}

func sweepArchives(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}
	conn := client.EventsClient(ctx)
	input := &eventbridge.ListArchivesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = listArchivesPages(ctx, conn, input, func(page *eventbridge.ListArchivesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Archives {
			r := resourceArchive()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.ArchiveName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EventBridge Archive sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing EventBridge Archives (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EventBridge Archives (%s): %w", region, err)
	}

	return nil
}

func sweepBuses(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}
	conn := client.EventsClient(ctx)
	input := &eventbridge.ListEventBusesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = listEventBusesPages(ctx, conn, input, func(page *eventbridge.ListEventBusesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, eventBus := range page.EventBuses {
			name := aws.ToString(eventBus.Name)

			if name == DefaultEventBusName {
				log.Printf("[INFO] Skipping EventBridge Event Bus %s", name)
				continue
			}

			r := resourceBus()
			d := r.Data(nil)
			d.SetId(name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EventBridge Event Bus sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing EventBridge Event Buses (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EventBridge Event Buses (%s): %w", region, err)
	}

	return nil
}

func sweepConnection(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}
	conn := client.EventsClient(ctx)
	input := &eventbridge.ListConnectionsInput{
		Limit: aws.Int32(100),
	}
	sweepResources := make([]sweep.Sweepable, 0)

	err = listConnectionsPages(ctx, conn, input, func(page *eventbridge.ListConnectionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Connections {
			r := resourceConnection()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Name))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EventBridge Connection sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing EventBridge Connections (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EventBridge Connections (%s): %w", region, err)
	}

	return nil
}

func sweepRules(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.EventsClient(ctx)
	input := &eventbridge.ListEventBusesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = listEventBusesPages(ctx, conn, input, func(page *eventbridge.ListEventBusesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.EventBuses {
			eventBusName := aws.ToString(v.Name)
			input := &eventbridge.ListRulesInput{
				EventBusName: aws.String(eventBusName),
			}

			err := listRulesPages(ctx, conn, input, func(page *eventbridge.ListRulesOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, v := range page.Rules {
					ruleName := aws.ToString(v.Name)

					r := resourceRule()
					d := r.Data(nil)
					d.SetId(ruleCreateResourceID(eventBusName, ruleName))
					d.Set(names.AttrForceDestroy, true)

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}

				return !lastPage
			})

			if err != nil {
				continue
			}
		}

		return !lastPage
	})

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EventBridge Rule sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing EventBridge Rules (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EventBridge Rules (%s): %w", region, err)
	}

	return nil
}

func sweepTargets(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.EventsClient(ctx)
	input := &eventbridge.ListEventBusesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = listEventBusesPages(ctx, conn, input, func(page *eventbridge.ListEventBusesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.EventBuses {
			eventBusName := aws.ToString(v.Name)
			input := &eventbridge.ListRulesInput{
				EventBusName: aws.String(eventBusName),
			}

			err := listRulesPages(ctx, conn, input, func(page *eventbridge.ListRulesOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, v := range page.Rules {
					ruleName := aws.ToString(v.Name)
					input := &eventbridge.ListTargetsByRuleInput{
						EventBusName: aws.String(eventBusName),
						Rule:         aws.String(ruleName),
					}

					err := listTargetsByRulePages(ctx, conn, input, func(page *eventbridge.ListTargetsByRuleOutput, lastPage bool) bool {
						if page == nil {
							return !lastPage
						}

						for _, v := range page.Targets {
							targetID := aws.ToString(v.Id)

							r := resourceTarget()
							d := r.Data(nil)
							d.SetId(targetCreateResourceID(eventBusName, ruleName, targetID))
							d.Set("event_bus_name", eventBusName)
							d.Set(names.AttrForceDestroy, true)
							d.Set(names.AttrRule, ruleName)
							d.Set("target_id", targetID)

							sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
						}

						return !lastPage
					})

					if err != nil {
						continue
					}
				}

				return !lastPage
			})

			if err != nil {
				continue
			}
		}

		return !lastPage
	})

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EventBridge Target sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing EventBridge Targets (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EventBridge Targets (%s): %w", region, err)
	}

	return nil
}
