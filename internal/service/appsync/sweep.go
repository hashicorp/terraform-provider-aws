//go:build sweep
// +build sweep

package appsync

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_appsync_graphql_api", &resource.Sweeper{
		Name: "aws_appsync_graphql_api",
		F:    sweepGraphQLAPIs,
	})

	resource.AddTestSweepers("aws_appsync_domain_name", &resource.Sweeper{
		Name: "aws_appsync_domain_name",
		F:    sweepDomainNames,
		Dependencies: []string{
			"aws_appsync_domain_name_api_association",
		},
	})

	resource.AddTestSweepers("aws_appsync_domain_name_api_association", &resource.Sweeper{
		Name: "aws_appsync_domain_name_api_association",
		F:    sweepDomainNameAssociations,
	})
}

func sweepGraphQLAPIs(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).AppSyncConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	input := &appsync.ListGraphqlApisInput{}

	for {
		output, err := conn.ListGraphqlApis(input)
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping AppSync GraphQL API sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			err := fmt.Errorf("error reading AppSync GraphQL API: %w", err)
			log.Printf("[ERROR] %s", err)
			errs = multierror.Append(errs, err)
			break
		}

		for _, graphAPI := range output.GraphqlApis {

			r := ResourceGraphQLAPI()
			d := r.Data(nil)

			id := aws.StringValue(graphAPI.ApiId)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	if err := sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping AppSync GraphQL API %s: %w", region, err))
	}

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping AppSync GraphQL API sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepDomainNames(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).AppSyncConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	input := &appsync.ListDomainNamesInput{}

	for {
		output, err := conn.ListDomainNames(input)
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping AppSync Domain Name sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			err := fmt.Errorf("error reading AppSync Domain Name: %w", err)
			log.Printf("[ERROR] %s", err)
			errs = multierror.Append(errs, err)
			break
		}

		for _, dm := range output.DomainNameConfigs {

			r := ResourceDomainName()
			d := r.Data(nil)

			id := aws.StringValue(dm.DomainName)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	if err := sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping AppSync Domain Name %s: %w", region, err))
	}

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping AppSync Domain Name sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepDomainNameAssociations(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).AppSyncConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	input := &appsync.ListDomainNamesInput{}

	for {
		output, err := conn.ListDomainNames(input)
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping AppSync Domain Name Association sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			err := fmt.Errorf("error reading AppSync Domain Name Association: %w", err)
			log.Printf("[ERROR] %s", err)
			errs = multierror.Append(errs, err)
			break
		}

		for _, dm := range output.DomainNameConfigs {

			r := ResourceDomainNameAPIAssociation()
			d := r.Data(nil)

			id := aws.StringValue(dm.DomainName)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	if err := sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping AppSync Domain Name Association %s: %w", region, err))
	}

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping AppSync Domain Name Association sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}
