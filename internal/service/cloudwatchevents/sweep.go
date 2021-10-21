//go:build sweep
// +build sweep

package cloudwatchevents

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	events "github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).CloudWatchEventsConn

	var sweeperErrs *multierror.Error

	input := &events.ListApiDestinationsInput{
		Limit: aws.Int64(100),
	}
	var apiDestinations []*events.ApiDestination
	for {
		output, err := conn.ListApiDestinations(input)

		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping CloudWatch Events Api Destination sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("Error retrieving CloudWatch Events Api Destinations: %w", err)
		}

		apiDestinations = append(apiDestinations, output.ApiDestinations...)

		if aws.StringValue(output.NextToken) == "" {
			break
		}
		input.NextToken = output.NextToken
	}

	for _, apiDestination := range apiDestinations {

		input := &events.DeleteApiDestinationInput{
			Name: apiDestination.Name,
		}
		_, err := conn.DeleteApiDestination(input)
		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("Error deleting CloudWatch Event Api Destination (%s): %w", *apiDestination.Name, err))
			continue
		}
	}

	log.Printf("[INFO] Deleted %d CloudWatch Event Api Destinations", len(apiDestinations))

	return sweeperErrs.ErrorOrNil()
}

func sweepArchives(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).CloudWatchEventsConn

	input := &events.ListArchivesInput{}

	for {
		output, err := conn.ListArchives(input)

		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping CloudWatch Events archive sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("Error retrieving CloudWatch Events archive: %w", err)
		}

		if len(output.Archives) == 0 {
			log.Print("[DEBUG] No CloudWatch Events archives to sweep")
			return nil
		}

		for _, archive := range output.Archives {
			name := aws.StringValue(archive.ArchiveName)
			if name == "default" {
				continue
			}

			log.Printf("[INFO] Deleting CloudWatch Events archive (%s)", name)
			_, err := conn.DeleteArchive(&events.DeleteArchiveInput{
				ArchiveName: aws.String(name),
			})
			if err != nil {
				return fmt.Errorf("Error deleting CloudWatch Events archive (%s): %w", name, err)
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
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).CloudWatchEventsConn
	input := &events.ListEventBusesInput{}
	var sweeperErrs *multierror.Error

	err = listEventBusesPages(conn, input, func(page *events.ListEventBusesOutput, lastPage bool) bool {
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
			err = r.Delete(d, client)

			if err != nil {
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CloudWatch Events event bus sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing CloudWatch Events event buses: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepConnection(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).CloudWatchEventsConn

	var sweeperErrs *multierror.Error

	input := &events.ListConnectionsInput{
		Limit: aws.Int64(100),
	}
	var connections []*events.Connection
	for {
		output, err := conn.ListConnections(input)

		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping CloudWatch Events Connection sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("Error retrieving CloudWatch Events Connections: %w", err)
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}
		input.NextToken = output.NextToken
	}

	for _, connection := range connections {
		input := &events.DeleteConnectionInput{
			Name: connection.Name,
		}
		_, err := conn.DeleteConnection(input)
		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("Error deleting CloudWatch Event Connection (%s): %w", *connection.Name, err))
			continue
		}
	}

	log.Printf("[INFO] Deleted %d CloudWatch Event Connections", len(connections))

	return sweeperErrs.ErrorOrNil()
}

func sweepPermissions(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).CloudWatchEventsConn

	output, err := conn.DescribeEventBus(&events.DescribeEventBusInput{})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping CloudWatch Event Permission sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving CloudWatch Event Permissions: %w", err)
	}

	policy := aws.StringValue(output.Policy)

	if policy == "" {
		log.Print("[DEBUG] No CloudWatch Event Permissions to sweep")
		return nil
	}

	var policyDoc PermissionPolicyDoc
	err = json.Unmarshal([]byte(policy), &policyDoc)
	if err != nil {
		return fmt.Errorf("Parsing CloudWatch Event Permissions policy %q failed: %w", policy, err)
	}

	for _, statement := range policyDoc.Statements {
		sid := statement.Sid

		log.Printf("[INFO] Deleting CloudWatch Event Permission %s", sid)
		_, err := conn.RemovePermission(&events.RemovePermissionInput{
			StatementId: aws.String(sid),
		})
		if err != nil {
			return fmt.Errorf("Error deleting CloudWatch Event Permission %s: %w", sid, err)
		}
	}

	return nil
}

func sweepRules(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).CloudWatchEventsConn

	var sweeperErrs *multierror.Error
	var count int

	rulesInput := &events.ListRulesInput{}

	err = listRulesPages(conn, rulesInput, func(rulesPage *events.ListRulesOutput, lastPage bool) bool {
		if rulesPage == nil {
			return !lastPage
		}

		for _, rule := range rulesPage.Rules {
			count++
			name := aws.StringValue(rule.Name)

			log.Printf("[INFO] Deleting CloudWatch Events rule (%s)", name)
			_, err := conn.DeleteRule(&events.DeleteRuleInput{
				Name:  aws.String(name),
				Force: aws.Bool(true), // Required for AWS-managed rules, ignored otherwise
			})
			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error deleting CloudWatch Events rule (%s): %w", name, err))
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CloudWatch Events rule sweeper for %q: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing CloudWatch Events rules: %w", err))
	}

	log.Printf("[INFO] Deleted %d CloudWatch Events rules", count)

	return sweeperErrs.ErrorOrNil()
}

func sweepTargets(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).CloudWatchEventsConn

	var sweeperErrs *multierror.Error
	var rulesCount, targetsCount int

	rulesInput := &events.ListRulesInput{}

	err = listRulesPages(conn, rulesInput, func(rulesPage *events.ListRulesOutput, lastPage bool) bool {
		if rulesPage == nil {
			return !lastPage
		}

		for _, rule := range rulesPage.Rules {
			rulesCount++
			ruleName := aws.StringValue(rule.Name)

			log.Printf("[INFO] Deleting CloudWatch Events targets for rule (%s)", ruleName)
			targetsInput := &events.ListTargetsByRuleInput{
				Rule:  rule.Name,
				Limit: aws.Int64(100), // Set limit to allowed maximum to prevent API throttling
			}

			err := listTargetsByRulePages(conn, targetsInput, func(targetsPage *events.ListTargetsByRuleOutput, lastPage bool) bool {
				if targetsPage == nil {
					return !lastPage
				}

				for _, target := range targetsPage.Targets {
					targetsCount++
					removeTargetsInput := &events.RemoveTargetsInput{
						Ids:   []*string{target.Id},
						Rule:  rule.Name,
						Force: aws.Bool(true), // Required for AWS-managed rules, ignored otherwise
					}
					targetID := aws.StringValue(target.Id)

					log.Printf("[INFO] Deleting CloudWatch Events target (%s/%s)", ruleName, targetID)
					_, err := conn.RemoveTargets(removeTargetsInput)

					if err != nil {
						sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("Error deleting CloudWatch Events target (%s/%s): %w", ruleName, targetID, err))
						continue
					}
				}

				return !lastPage
			})

			if sweep.SkipSweepError(err) {
				log.Printf("[WARN] Skipping CloudWatch Events target sweeper for %q: %s", region, err)
				return false
			}
			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing CloudWatch Events targets for rule (%s): %w", ruleName, err))
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CloudWatch Events rule target sweeper for %q: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing CloudWatch Events rules: %w", err))
	}

	log.Printf("[INFO] Deleted %d CloudWatch Events targets across %d CloudWatch Events rules", targetsCount, rulesCount)

	return sweeperErrs.ErrorOrNil()
}
