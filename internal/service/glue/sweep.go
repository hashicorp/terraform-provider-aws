// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package glue

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
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
	conn := client.GlueConn(ctx)

	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	input := &glue.GetDatabasesInput{}
	err = conn.GetDatabasesPagesWithContext(ctx, input, func(page *glue.GetDatabasesOutput, lastPage bool) bool {
		if len(page.DatabaseList) == 0 {
			log.Printf("[INFO] No Glue Catalog Databases to sweep")
			return false
		}
		for _, database := range page.DatabaseList {
			name := aws.StringValue(database.Name)

			log.Printf("[INFO] Deleting Glue Catalog Database: %s", name)

			r := ResourceCatalogDatabase()
			d := r.Data(nil)
			d.SetId("unused")
			d.Set("name", name)
			d.Set("catalog_id", database.CatalogId)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
		return !lastPage
	})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Glue Catalog Database sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving Glue Catalog Databases: %s", err)
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
	conn := client.GlueConn(ctx)

	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	input := &glue.GetClassifiersInput{}
	err = conn.GetClassifiersPagesWithContext(ctx, input, func(page *glue.GetClassifiersOutput, lastPage bool) bool {
		if len(page.Classifiers) == 0 {
			log.Printf("[INFO] No Glue Classifiers to sweep")
			return false
		}
		for _, classifier := range page.Classifiers {
			var name string
			if classifier.CsvClassifier != nil {
				name = aws.StringValue(classifier.CsvClassifier.Name)
			} else if classifier.GrokClassifier != nil {
				name = aws.StringValue(classifier.GrokClassifier.Name)
			} else if classifier.JsonClassifier != nil {
				name = aws.StringValue(classifier.JsonClassifier.Name)
			} else if classifier.XMLClassifier != nil {
				name = aws.StringValue(classifier.XMLClassifier.Name)
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
		return !lastPage
	})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Glue Classifier sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving Glue Classifiers: %s", err)
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
	conn := client.GlueConn(ctx)
	catalogID := client.AccountID

	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	input := &glue.GetConnectionsInput{
		CatalogId: aws.String(catalogID),
	}
	err = conn.GetConnectionsPagesWithContext(ctx, input, func(page *glue.GetConnectionsOutput, lastPage bool) bool {
		if len(page.ConnectionList) == 0 {
			log.Printf("[INFO] No Glue Connections to sweep")
			return false
		}
		for _, connection := range page.ConnectionList {
			id := fmt.Sprintf("%s:%s", catalogID, aws.StringValue(connection.Name))

			log.Printf("[INFO] Deleting Glue Connection: %s", id)
			r := ResourceConnection()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
		return !lastPage
	})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Glue Connection sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving Glue Connections: %s", err)
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
	conn := client.GlueConn(ctx)

	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	input := &glue.GetCrawlersInput{}
	err = conn.GetCrawlersPagesWithContext(ctx, input, func(page *glue.GetCrawlersOutput, lastPage bool) bool {
		if len(page.Crawlers) == 0 {
			log.Printf("[INFO] No Glue Crawlers to sweep")
			return false
		}
		for _, crawler := range page.Crawlers {
			name := aws.StringValue(crawler.Name)

			r := ResourceCrawler()
			d := r.Data(nil)
			d.SetId(name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
		return !lastPage
	})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Glue Crawler sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving Glue Crawlers: %s", err)
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
	conn := client.GlueConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.GetDevEndpointsPagesWithContext(ctx, input, func(page *glue.GetDevEndpointsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DevEndpoints {
			name := aws.StringValue(v.EndpointName)
			if !strings.HasPrefix(name, sweep.ResourcePrefix) {
				log.Printf("[INFO] Skipping Glue Dev Endpoint: %s", name)
				continue
			}

			r := ResourceDevEndpoint()
			d := r.Data(nil)
			d.SetId(name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Glue Dev Endpoint sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Glue Dev Endpoints (%s): %w", region, err)
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
	conn := client.GlueConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.GetJobsPagesWithContext(ctx, input, func(page *glue.GetJobsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Jobs {
			r := ResourceJob()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.Name))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Glue Job sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Glue Jobs (%s): %w", region, err)
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
	conn := client.GlueConn(ctx)

	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	input := &glue.GetMLTransformsInput{}
	err = conn.GetMLTransformsPagesWithContext(ctx, input, func(page *glue.GetMLTransformsOutput, lastPage bool) bool {
		if len(page.Transforms) == 0 {
			log.Printf("[INFO] No Glue ML Transforms to sweep")
			return false
		}
		for _, transforms := range page.Transforms {
			id := aws.StringValue(transforms.TransformId)

			log.Printf("[INFO] Deleting Glue ML Transform: %s", id)
			r := ResourceMLTransform()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
		return !lastPage
	})
	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Glue ML Transforms sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving Glue ML Transforms: %w", err))
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
	conn := client.GlueConn(ctx)

	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	listOutput, err := conn.ListRegistriesWithContext(ctx, &glue.ListRegistriesInput{})
	if err != nil {
		// Some endpoints that do not support Glue Registrys return InternalFailure
		if sweep.SkipSweepError(err) || tfawserr.ErrCodeEquals(err, "InternalFailure") {
			log.Printf("[WARN] Skipping Glue Registry sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving Glue Registry: %s", err)
	}
	for _, registry := range listOutput.Registries {
		arn := aws.StringValue(registry.RegistryArn)
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
	conn := client.GlueConn(ctx)

	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	listOutput, err := conn.ListSchemasWithContext(ctx, &glue.ListSchemasInput{})
	if err != nil {
		// Some endpoints that do not support Glue Schemas return InternalFailure
		if sweep.SkipSweepError(err) || tfawserr.ErrCodeEquals(err, "InternalFailure") {
			log.Printf("[WARN] Skipping Glue Schema sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving Glue Schema: %s", err)
	}
	for _, schema := range listOutput.Schemas {
		arn := aws.StringValue(schema.SchemaArn)
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
	conn := client.GlueConn(ctx)

	input := &glue.GetSecurityConfigurationsInput{}

	for {
		output, err := conn.GetSecurityConfigurationsWithContext(ctx, input)

		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Glue Security Configuration sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("Error retrieving Glue Security Configurations: %s", err)
		}

		for _, securityConfiguration := range output.SecurityConfigurations {
			name := aws.StringValue(securityConfiguration.Name)

			log.Printf("[INFO] Deleting Glue Security Configuration: %s", name)
			err := DeleteSecurityConfiguration(ctx, conn, name)
			if err != nil {
				log.Printf("[ERROR] Failed to delete Glue Security Configuration %s: %s", name, err)
			}
		}

		if aws.StringValue(output.NextToken) == "" {
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
	conn := client.GlueConn(ctx)

	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	input := &glue.GetTriggersInput{}
	err = conn.GetTriggersPagesWithContext(ctx, input, func(page *glue.GetTriggersOutput, lastPage bool) bool {
		if page == nil || len(page.Triggers) == 0 {
			log.Printf("[INFO] No Glue Triggers to sweep")
			return false
		}
		for _, trigger := range page.Triggers {
			name := aws.StringValue(trigger.Name)

			log.Printf("[INFO] Deleting Glue Trigger: %s", name)
			r := ResourceTrigger()
			d := r.Data(nil)
			d.SetId(name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
		return !lastPage
	})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Glue Trigger sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving Glue Triggers: %s", err)
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
	conn := client.GlueConn(ctx)

	listOutput, err := conn.ListWorkflowsWithContext(ctx, &glue.ListWorkflowsInput{})
	if err != nil {
		// Some endpoints that do not support Glue Workflows return InternalFailure
		if sweep.SkipSweepError(err) || tfawserr.ErrCodeEquals(err, "InternalFailure") {
			log.Printf("[WARN] Skipping Glue Workflow sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving Glue Workflow: %s", err)
	}
	for _, workflowName := range listOutput.Workflows {
		err := DeleteWorkflow(ctx, conn, *workflowName)
		if err != nil {
			log.Printf("[ERROR] Failed to delete Glue Workflow %s: %s", *workflowName, err)
		}
	}
	return nil
}
