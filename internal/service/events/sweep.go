// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package events

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
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

	resource.AddTestSweepers("aws_cloudwatch_event_permission", &resource.Sweeper{
		Name: "aws_cloudwatch_event_permission",
		F:    sweepPermissions,
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
	conn := client.EventsConn(ctx)

	var sweeperErrs *multierror.Error

	input := &eventbridge.ListApiDestinationsInput{
		Limit: aws.Int64(100),
	}
	var apiDestinations []*eventbridge.ApiDestination
	for {
		output, err := conn.ListApiDestinationsWithContext(ctx, input)

		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EventBridge API Destination sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("Error retrieving EventBridge API Destinations: %w", err)
		}

		apiDestinations = append(apiDestinations, output.ApiDestinations...)

		if aws.StringValue(output.NextToken) == "" {
			break
		}
		input.NextToken = output.NextToken
	}

	for _, apiDestination := range apiDestinations {

		input := &eventbridge.DeleteApiDestinationInput{
			Name: apiDestination.Name,
		}
		_, err := conn.DeleteApiDestinationWithContext(ctx, input)
		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("Error deleting EventBridge Api Destination (%s): %w", *apiDestination.Name, err))
			continue
		}
	}

	log.Printf("[INFO] Deleted %d EventBridge Api Destinations", len(apiDestinations))

	return sweeperErrs.ErrorOrNil()
}

func sweepArchives(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}
	conn := client.EventsConn(ctx)

	input := &eventbridge.ListArchivesInput{}

	for {
		output, err := conn.ListArchivesWithContext(ctx, input)

		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EventBridge archive sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("Error retrieving EventBridge archive: %w", err)
		}

		if len(output.Archives) == 0 {
			log.Print("[DEBUG] No EventBridge archives to sweep")
			return nil
		}

		for _, archive := range output.Archives {
			name := aws.StringValue(archive.ArchiveName)
			if name == "default" {
				continue
			}

			log.Printf("[INFO] Deleting EventBridge archive (%s)", name)
			_, err := conn.DeleteArchiveWithContext(ctx, &eventbridge.DeleteArchiveInput{
				ArchiveName: aws.String(name),
			})
			if err != nil {
				return fmt.Errorf("Error deleting EventBridge archive (%s): %w", name, err)
			}
		}

		if output.NextToken == nil {
			break
		}
		input.NextToken = output.NextToken
	}

	return nil
}

func sweepBuses(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}
	conn := client.EventsConn(ctx)
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	input := &eventbridge.ListEventBusesInput{}
	err = listEventBusesPages(ctx, conn, input, func(page *eventbridge.ListEventBusesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, eventBus := range page.EventBuses {
			name := aws.StringValue(eventBus.Name)
			if name == DefaultEventBusName {
				continue
			}

			r := ResourceBus()
			d := r.Data(nil)
			d.SetId(name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EventBridge event bus sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EventBridge event buses: %w", err))
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping EventBridge Event Buses: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepConnection(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}
	conn := client.EventsConn(ctx)

	var sweeperErrs *multierror.Error

	input := &eventbridge.ListConnectionsInput{
		Limit: aws.Int64(100),
	}
	var connections []*eventbridge.Connection
	for {
		output, err := conn.ListConnectionsWithContext(ctx, input)

		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EventBridge Connection sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("Error retrieving EventBridge Connections: %w", err)
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}
		input.NextToken = output.NextToken
	}

	for _, connection := range connections {
		input := &eventbridge.DeleteConnectionInput{
			Name: connection.Name,
		}
		_, err := conn.DeleteConnectionWithContext(ctx, input)
		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("Error deleting EventBridge Connection (%s): %w", *connection.Name, err))
			continue
		}
	}

	log.Printf("[INFO] Deleted %d EventBridge Connections", len(connections))

	return sweeperErrs.ErrorOrNil()
}

func sweepPermissions(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}
	conn := client.EventsConn(ctx)

	output, err := conn.DescribeEventBusWithContext(ctx, &eventbridge.DescribeEventBusInput{})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EventBridge Permission sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving EventBridge Permissions: %w", err)
	}

	policy := aws.StringValue(output.Policy)

	if policy == "" {
		log.Print("[DEBUG] No EventBridge Permissions to sweep")
		return nil
	}

	var policyDoc PermissionPolicyDoc
	err = json.Unmarshal([]byte(policy), &policyDoc)
	if err != nil {
		return fmt.Errorf("Parsing EventBridge Permissions policy %q failed: %w", policy, err)
	}

	for _, statement := range policyDoc.Statements {
		sid := statement.Sid

		log.Printf("[INFO] Deleting EventBridge Permission %s", sid)
		_, err := conn.RemovePermissionWithContext(ctx, &eventbridge.RemovePermissionInput{
			StatementId: aws.String(sid),
		})
		if err != nil {
			return fmt.Errorf("Error deleting EventBridge Permission %s: %w", sid, err)
		}
	}

	return nil
}

func sweepRules(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.EventsConn(ctx)
	input := &eventbridge.ListEventBusesInput{}
	var sweeperErrs *multierror.Error

	err = listEventBusesPages(ctx, conn, input, func(page *eventbridge.ListEventBusesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, eventBus := range page.EventBuses {
			eventBusName := aws.StringValue(eventBus.Name)

			input := &eventbridge.ListRulesInput{
				EventBusName: aws.String(eventBusName),
			}

			err := listRulesPages(ctx, conn, input, func(page *eventbridge.ListRulesOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, rule := range page.Rules {
					ruleName := aws.StringValue(rule.Name)

					log.Printf("[DEBUG] Deleting EventBridge Rule: %s/%s", eventBusName, ruleName)
					_, err := conn.DeleteRuleWithContext(ctx, &eventbridge.DeleteRuleInput{
						EventBusName: aws.String(eventBusName),
						Force:        aws.Bool(true),
						Name:         aws.String(ruleName),
					})

					if err != nil {
						sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error deleting EventBridge Rule (%s/%s): %w", eventBusName, ruleName, err))
						continue
					}
				}

				return !lastPage
			})

			if sweep.SkipSweepError(err) {
				continue
			}

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EventBridge Rules (%s): %w", region, err))
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EventBridge Rule sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EventBridge event buses (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepTargets(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.EventsConn(ctx)
	input := &eventbridge.ListEventBusesInput{}
	var sweeperErrs *multierror.Error

	err = listEventBusesPages(ctx, conn, input, func(page *eventbridge.ListEventBusesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, eventBus := range page.EventBuses {
			eventBusName := aws.StringValue(eventBus.Name)

			input := &eventbridge.ListRulesInput{
				EventBusName: aws.String(eventBusName),
			}

			err := listRulesPages(ctx, conn, input, func(page *eventbridge.ListRulesOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, rule := range page.Rules {
					ruleName := aws.StringValue(rule.Name)

					input := &eventbridge.ListTargetsByRuleInput{
						EventBusName: aws.String(eventBusName),
						Rule:         aws.String(ruleName),
					}

					err := listTargetsByRulePages(ctx, conn, input, func(page *eventbridge.ListTargetsByRuleOutput, lastPage bool) bool {
						if page == nil {
							return !lastPage
						}

						for _, target := range page.Targets {
							targetID := aws.StringValue(target.Id)

							log.Printf("[DEBUG] Deleting EventBridge Target: %s/%s/%s", eventBusName, ruleName, targetID)
							_, err := conn.RemoveTargetsWithContext(ctx, &eventbridge.RemoveTargetsInput{
								EventBusName: aws.String(eventBusName),
								Force:        aws.Bool(true),
								Ids:          aws.StringSlice([]string{targetID}),
								Rule:         aws.String(ruleName),
							})

							if err != nil {
								sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error deleting EventBridge Target (%s/%s/%s): %w", eventBusName, ruleName, targetID, err))
								continue
							}
						}

						return !lastPage
					})

					if sweep.SkipSweepError(err) {
						continue
					}

					if err != nil {
						sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EventBridge Targets (%s): %w", region, err))
					}
				}

				return !lastPage
			})

			if sweep.SkipSweepError(err) {
				continue
			}

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EventBridge Rules (%s): %w", region, err))
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EventBridge Rule sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EventBridge event buses (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}
