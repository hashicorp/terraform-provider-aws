// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_glue_catalog_database", sweepCatalogDatabases)
	awsv2.Register("aws_glue_classifier", sweepClassifiers)
	awsv2.Register("aws_glue_connection", sweepConnections)
	awsv2.Register("aws_glue_crawler", sweepCrawlers)
	awsv2.Register("aws_glue_dev_endpoint", sweepDevEndpoints)

	awsv2.Register("aws_glue_job", sweepJobs)

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

func sweepCatalogDatabases(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.GlueClient(ctx)
	var input glue.GetDatabasesInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := glue.NewGetDatabasesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.DatabaseList {
			r := resourceCatalogDatabase()
			d := r.Data(nil)
			d.SetId("unused")
			d.Set(names.AttrCatalogID, v.CatalogId)
			d.Set(names.AttrName, v.Name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepClassifiers(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.GlueClient(ctx)
	var input glue.GetClassifiersInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := glue.NewGetClassifiersPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Classifiers {
			var name string
			if v.CsvClassifier != nil {
				name = aws.ToString(v.CsvClassifier.Name)
			} else if v.GrokClassifier != nil {
				name = aws.ToString(v.GrokClassifier.Name)
			} else if v.JsonClassifier != nil {
				name = aws.ToString(v.JsonClassifier.Name)
			} else if v.XMLClassifier != nil {
				name = aws.ToString(v.XMLClassifier.Name)
			}
			if name == "" {
				continue
			}

			r := resourceClassifier()
			d := r.Data(nil)
			d.SetId(name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepConnections(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.GlueClient(ctx)
	catalogID := client.AccountID(ctx)
	var input glue.GetConnectionsInput
	input.CatalogId = aws.String(catalogID)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := glue.NewGetConnectionsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.ConnectionList {
			r := resourceConnection()
			d := r.Data(nil)
			d.SetId(fmt.Sprintf("%s:%s", catalogID, aws.ToString(v.Name)))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepCrawlers(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.GlueClient(ctx)
	var input glue.GetCrawlersInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := glue.NewGetCrawlersPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Crawlers {
			r := resourceCrawler()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Name))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepDevEndpoints(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.GlueClient(ctx)
	var input glue.GetDevEndpointsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := glue.NewGetDevEndpointsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.DevEndpoints {
			name := aws.ToString(v.EndpointName)
			if !strings.HasPrefix(name, sweep.ResourcePrefix) {
				log.Printf("[INFO] Skipping Glue Dev Endpoint: %s", name)
				continue
			}

			r := resourceDevEndpoint()
			d := r.Data(nil)
			d.SetId(name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepJobs(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.GlueClient(ctx)
	var input glue.GetJobsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := glue.NewGetJobsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Jobs {
			r := resourceJob()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Name))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
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
