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
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_glue_catalog_database", sweepCatalogDatabases,
		"aws_datazone_environment",
	)

	awsv2.Register("aws_glue_classifier", sweepClassifiers)
	awsv2.Register("aws_glue_connection", sweepConnections)
	awsv2.Register("aws_glue_crawler", sweepCrawlers)
	awsv2.Register("aws_glue_dev_endpoint", sweepDevEndpoints)
	awsv2.Register("aws_glue_job", sweepJobs)
	awsv2.Register("aws_glue_ml_transform", sweepMLTransforms)
	awsv2.Register("aws_glue_registry", sweepRegistries)
	awsv2.Register("aws_glue_schema", sweepSchemas)
	awsv2.Register("aws_glue_security_configuration", sweepSecurityConfigurations)
	awsv2.Register("aws_glue_trigger", sweepTriggers)
	awsv2.Register("aws_glue_workflow", sweepWorkflows)
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

func sweepMLTransforms(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.GlueClient(ctx)
	var input glue.GetMLTransformsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := glue.NewGetMLTransformsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Transforms {
			r := resourceMLTransform()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.TransformId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepRegistries(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.GlueClient(ctx)
	var input glue.ListRegistriesInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := glue.NewListRegistriesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Registries {
			r := resourceRegistry()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.RegistryArn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepSchemas(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.GlueClient(ctx)
	var input glue.ListSchemasInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := glue.NewListSchemasPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, schema := range page.Schemas {
			r := resourceSchema()
			d := r.Data(nil)
			d.SetId(aws.ToString(schema.SchemaArn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepSecurityConfigurations(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.GlueClient(ctx)
	var input glue.GetSecurityConfigurationsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := glue.NewGetSecurityConfigurationsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.SecurityConfigurations {
			r := resourceSecurityConfiguration()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Name))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepTriggers(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.GlueClient(ctx)
	var input glue.GetTriggersInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := glue.NewGetTriggersPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Triggers {
			r := resourceTrigger()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Name))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepWorkflows(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.GlueClient(ctx)
	var input glue.ListWorkflowsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := glue.NewListWorkflowsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Workflows {
			r := resourceWorkflow()
			d := r.Data(nil)
			d.SetId(v)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}
