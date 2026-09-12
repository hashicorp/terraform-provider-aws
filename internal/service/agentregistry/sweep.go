// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package agentregistry

import (
	"context"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/agentregistrycontrol"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
)

func RegisterSweepers() {
	awsv2.Register("aws_agentregistry_registry", sweepRegistries, "aws_agentregistry_registry_record")
	awsv2.Register("aws_agentregistry_registry_record", sweepRegistryRecords)
}

func sweepRegistries(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	var input agentregistrycontrol.ListRegistriesInput
	conn := client.AgentRegistryClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := agentregistrycontrol.NewListRegistriesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.Registries {
			sweepResources = append(sweepResources, framework.NewSweepResource(newRegistryResource, client,
				framework.NewAttribute("registry_id", aws.ToString(v.RegistryId))),
			)
		}
	}

	return sweepResources, nil
}

func sweepRegistryRecords(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	var input agentregistrycontrol.ListRegistriesInput
	conn := client.AgentRegistryClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := agentregistrycontrol.NewListRegistriesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, registry := range page.Registries {
			registryID := aws.ToString(registry.RegistryId)
			input := agentregistrycontrol.ListRegistryRecordsInput{
				RegistryId: registry.RegistryId,
			}
			pages := agentregistrycontrol.NewListRegistryRecordsPaginator(conn, &input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)
				if err != nil {
					return nil, smarterr.NewError(err)
				}

				for _, v := range page.RegistryRecords {
					sweepResources = append(sweepResources, framework.NewSweepResource(newRegistryRecordResource, client,
						framework.NewAttribute("registry_id", registryID),
						framework.NewAttribute("record_id", aws.ToString(v.RecordId))),
					)
				}
			}
		}
	}

	return sweepResources, nil
}
