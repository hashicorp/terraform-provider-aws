// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_glue_catalog_database", &resource.Sweeper{
		Name: "aws_glue_catalog_database",
		F:    sweepCatalogDatabases,
	})

	resource.AddTestSweepers("aws_glue_classifier", &resource.Sweeper{
		Name: "aws_glue_classifier",
		F:    sweepClassifiers,
	})

	resource.AddTestSweepers("aws_glue_connection", &resource.Sweeper{
		Name: "aws_glue_connection",
		F:    sweepConnections,
	})

	resource.AddTestSweepers("aws_glue_crawler", &resource.Sweeper{
		Name: "aws_glue_crawler",
		F:    sweepCrawlers,
	})

	resource.AddTestSweepers("aws_glue_dev_endpoint", &resource.Sweeper{
		Name: "aws_glue_dev_endpoint",
		F:    sweepDevEndpoints,
	})

	resource.AddTestSweepers("aws_glue_job", &resource.Sweeper{
		Name: "aws_glue_job",
		F:    sweepJobs,
	})

	resource.AddTestSweepers("aws_glue_ml_transform", &resource.Sweeper{
		Name: "aws_glue_ml_transform",
		F:    sweepMLTransforms,
	})

	resource.AddTestSweepers("aws_glue_registry", &resource.Sweeper{
		Name: "aws_glue_registry",
		F:    sweepRegistry,
	})

	resource.AddTestSweepers("aws_glue_schema", &resource.Sweeper{
		Name: "aws_glue_schema",
		F:    sweepSchema,
	})

	resource.AddTestSweepers("aws_glue_security_configuration", &resource.Sweeper{
		Name: "aws_glue_security_configuration",
		F:    sweepSecurityConfigurations,
	})

	resource.AddTestSweepers("aws_glue_trigger", &resource.Sweeper{
		Name: "aws_glue_trigger",
		F:    sweepTriggers,
	})

	resource.AddTestSweepers("aws_glue_workflow", &resource.Sweeper{
		Name: "aws_glue_workflow",
		F:    sweepWorkflow,
	})
}

func sweepCatalogDatabases(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.GlueClient(ctx)

	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	input := &glue.GetDatabasesInput{}

	pages := glue.NewGetDatabasesPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			if awsv2.SkipSweepError(err) {
				log.Printf("[WARN] Skipping Glue Catalog Database sweep for %s: %s", region, err)
				return nil
			}
			return fmt.Errorf("Error retrieving Glue Catalog Databases: %s", err)
		}

		for _, database := range page.DatabaseList {
			name := aws.ToString(database.Name)

			log.Printf("[INFO] Deleting Glue Catalog Database: %s", name)

			r := ResourceCatalogDatabase()
			d := r.Data(nil)
			d.SetId("unused")
			d.Set(names.AttrName, name)
			d.Set(names.AttrCatalogID, database.CatalogId)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Glue Catalog Databases: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepClassifiers(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.GlueClient(ctx)

	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	input := &glue.GetClassifiersInput{}

	pages := glue.NewGetClassifiersPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			if awsv2.SkipSweepError(err) {
				log.Printf("[WARN] Skipping Glue Classifier sweep for %s: %s", region, err)
				return nil
			}
			return fmt.Errorf("Error retrieving Glue Classifiers: %s", err)
		}

		for _, classifier := range page.Classifiers {
			var name string
			if classifier.CsvClassifier != nil {
				name = aws.ToString(classifier.CsvClassifier.Name)
			} else if classifier.GrokClassifier != nil {
				name = aws.ToString(classifier.GrokClassifier.Name)
			} else if classifier.JsonClassifier != nil {
				name = aws.ToString(classifier.JsonClassifier.Name)
			} else if classifier.XMLClassifier != nil {
				name = aws.ToString(classifier.XMLClassifier.Name)
			}
			if name == "" {
				log.Printf("[WARN] Unable to determine Glue Classifier name: %#v", classifier)
				continue
			}

			log.Printf("[INFO] Deleting Glue Classifier: %s", name)
			r := ResourceClassifier()
			d := r.Data(nil)
			d.SetId(name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Glue Classifiers: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepConnections(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.GlueClient(ctx)
	catalogID := client.AccountID

	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	input := &glue.GetConnectionsInput{
		CatalogId: aws.String(catalogID),
	}

	pages := glue.NewGetConnectionsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			if awsv2.SkipSweepError(err) {
				log.Printf("[WARN] Skipping Glue Connection sweep for %s: %s", region, err)
				return nil
			}
			return fmt.Errorf("Error retrieving Glue Connections: %s", err)
		}

		for _, connection := range page.ConnectionList {
			id := fmt.Sprintf("%s:%s", catalogID, aws.ToString(connection.Name))

			log.Printf("[INFO] Deleting Glue Connection: %s", id)
			r := ResourceConnection()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping API Gateway VPC Links: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepCrawlers(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.GlueClient(ctx)

	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	input := &glue.GetCrawlersInput{}

	pages := glue.NewGetCrawlersPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			if awsv2.SkipSweepError(err) {
				log.Printf("[WARN] Skipping Glue Crawler sweep for %s: %s", region, err)
				return nil
			}
			return fmt.Errorf("Error retrieving Glue Crawlers: %s", err)
		}

		for _, crawler := range page.Crawlers {
			name := aws.ToString(crawler.Name)

			r := ResourceCrawler()
			d := r.Data(nil)
			d.SetId(name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping API Gateway VPC Links: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepDevEndpoints(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	input := &glue.GetDevEndpointsInput{}
	conn := client.GlueClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := glue.NewGetDevEndpointsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Glue Dev Endpoint sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Glue Dev Endpoints (%s): %w", region, err)
		}

		for _, v := range page.DevEndpoints {
			name := aws.ToString(v.EndpointName)
			if !strings.HasPrefix(name, sweep.ResourcePrefix) {
				log.Printf("[INFO] Skipping Glue Dev Endpoint: %s", name)
				continue
			}

			r := ResourceDevEndpoint()
			d := r.Data(nil)
			d.SetId(name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Glue Dev Endpoints (%s): %w", region, err)
	}

	return nil
}

func sweepJobs(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	input := &glue.GetJobsInput{}
	conn := client.GlueClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := glue.NewGetJobsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Glue Job sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Glue Jobs (%s): %w", region, err)
		}

		for _, v := range page.Jobs {
			r := ResourceJob()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Name))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Glue Jobs (%s): %w", region, err)
	}

	return nil
}

func sweepMLTransforms(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.GlueClient(ctx)

	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	input := &glue.GetMLTransformsInput{}

	pages := glue.NewGetMLTransformsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Glue ML Transforms sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
		}
		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving Glue ML Transforms: %w", err))
		}

		for _, transforms := range page.Transforms {
			id := aws.ToString(transforms.TransformId)

			log.Printf("[INFO] Deleting Glue ML Transform: %s", id)
			r := ResourceMLTransform()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Glue ML Transforms: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepRegistry(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.GlueClient(ctx)

	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	listOutput, err := conn.ListRegistries(ctx, &glue.ListRegistriesInput{})
	if err != nil {
		// Some endpoints that do not support Glue Registrys return InternalFailure
		if awsv2.SkipSweepError(err) || tfawserr.ErrCodeEquals(err, "InternalFailure") {
			log.Printf("[WARN] Skipping Glue Registry sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving Glue Registry: %s", err)
	}
	for _, registry := range listOutput.Registries {
		arn := aws.ToString(registry.RegistryArn)
		r := ResourceRegistry()
		d := r.Data(nil)
		d.SetId(arn)

		sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Glue Registry: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepSchema(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.GlueClient(ctx)

	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	listOutput, err := conn.ListSchemas(ctx, &glue.ListSchemasInput{})
	if err != nil {
		// Some endpoints that do not support Glue Schemas return InternalFailure
		if awsv2.SkipSweepError(err) || tfawserr.ErrCodeEquals(err, "InternalFailure") {
			log.Printf("[WARN] Skipping Glue Schema sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving Glue Schema: %s", err)
	}
	for _, schema := range listOutput.Schemas {
		arn := aws.ToString(schema.SchemaArn)
		r := ResourceSchema()
		d := r.Data(nil)
		d.SetId(arn)

		sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping API Gateway VPC Links: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepSecurityConfigurations(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.GlueClient(ctx)

	input := &glue.GetSecurityConfigurationsInput{}

	for {
		output, err := conn.GetSecurityConfigurations(ctx, input)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Glue Security Configuration sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("Error retrieving Glue Security Configurations: %s", err)
		}

		for _, securityConfiguration := range output.SecurityConfigurations {
			name := aws.ToString(securityConfiguration.Name)

			log.Printf("[INFO] Deleting Glue Security Configuration: %s", name)
			err := DeleteSecurityConfiguration(ctx, conn, name)
			if err != nil {
				log.Printf("[ERROR] Failed to delete Glue Security Configuration %s: %s", name, err)
			}
		}

		if aws.ToString(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil
}

func sweepTriggers(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.GlueClient(ctx)

	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	input := &glue.GetTriggersInput{}

	pages := glue.NewGetTriggersPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			if awsv2.SkipSweepError(err) {
				log.Printf("[WARN] Skipping Glue Trigger sweep for %s: %s", region, err)
				return nil
			}
			return fmt.Errorf("Error retrieving Glue Triggers: %s", err)
		}

		for _, trigger := range page.Triggers {
			name := aws.ToString(trigger.Name)

			log.Printf("[INFO] Deleting Glue Trigger: %s", name)
			r := ResourceTrigger()
			d := r.Data(nil)
			d.SetId(name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Glue Triggers: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepWorkflow(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.GlueClient(ctx)

	listOutput, err := conn.ListWorkflows(ctx, &glue.ListWorkflowsInput{})
	if err != nil {
		// Some endpoints that do not support Glue Workflows return InternalFailure
		if awsv2.SkipSweepError(err) || tfawserr.ErrCodeEquals(err, "InternalFailure") {
			log.Printf("[WARN] Skipping Glue Workflow sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving Glue Workflow: %s", err)
	}
	for _, workflowName := range listOutput.Workflows {
		err := DeleteWorkflow(ctx, conn, workflowName)
		if err != nil {
			log.Printf("[ERROR] Failed to delete Glue Workflow %s: %s", workflowName, err)
		}
	}
	return nil
}
