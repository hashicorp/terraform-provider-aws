// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schemas

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/schemas"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_schemas_discoverer", &resource.Sweeper{
		Name: "aws_schemas_discoverer",
		F:    sweepDiscoverers,
	})

	resource.AddTestSweepers("aws_schemas_registry", &resource.Sweeper{
		Name: "aws_schemas_registry",
		F:    sweepRegistries,
		Dependencies: []string{
			"aws_schemas_schema",
		},
	})

	resource.AddTestSweepers("aws_schemas_schema", &resource.Sweeper{
		Name: "aws_schemas_registry",
		F:    sweepSchemas,
	})
}

func sweepDiscoverers(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.SchemasClient(ctx)
	input := &schemas.ListDiscoverersInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := schemas.NewListDiscoverersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EventBridge Schemas Discoverer sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing EventBridge Schemas Discoverers (%s): %w", region, err)
		}

		for _, v := range page.Discoverers {
			r := resourceDiscoverer()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.DiscovererId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EventBridge Schemas Discoverers (%s): %w", region, err)
	}

	return nil
}

func sweepRegistries(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.SchemasClient(ctx)
	input := &schemas.ListRegistriesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := schemas.NewListRegistriesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EventBridge Schemas Registry sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing EventBridge Schemas Registries (%s): %w", region, err)
		}

		for _, v := range page.Registries {
			registryName := aws.ToString(v.RegistryName)
			if strings.HasPrefix(registryName, "aws.") {
				log.Printf("[INFO] Skipping EventBridge Schemas Registry %s", registryName)
				continue
			}

			r := resourceRegistry()
			d := r.Data(nil)
			d.SetId(registryName)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EventBridge Schemas Registries (%s): %w", region, err)
	}

	return nil
}

func sweepSchemas(region string) error { // nosemgrep:ci.schemas-in-func-name
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.SchemasClient(ctx)
	input := &schemas.ListRegistriesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := schemas.NewListRegistriesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EventBridge Schemas Schema sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			continue
		}

		for _, v := range page.Registries {
			registryName := aws.ToString(v.RegistryName)
			input := &schemas.ListSchemasInput{
				RegistryName: aws.String(registryName),
			}

			pages := schemas.NewListSchemasPaginator(conn, input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if awsv2.SkipSweepError(err) {
					log.Printf("[WARN] Skipping EventBridge Schemas Schema sweep for %s: %s", region, err)
					return nil
				}

				if err != nil {
					return fmt.Errorf("error listing EventBridge Schemas Schemas (%s): %w", region, err)
				}

				for _, v := range page.Schemas {
					schemaName := aws.ToString(v.SchemaName)
					if strings.HasPrefix(schemaName, "aws.") {
						log.Printf("[DEBUG] skipping EventBridge Schemas Schema (%s): managed by AWS", schemaName)
						continue
					}

					r := resourceSchema()
					d := r.Data(nil)
					d.SetId(schemaCreateResourceID(schemaName, registryName))

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}
			}
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EventBridge Schemas Schemas (%s): %w", region, err)
	}

	return nil
}
