// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package datazone

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/datazone"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_datazone_domain", sweepDomains,
		"aws_datazone_project",
		"aws_datazone_environment_profile",
	)
	awsv2.Register("aws_datazone_environment", sweepEnvironments)
	awsv2.Register("aws_datazone_environment_profile", sweepEnvironmentProfiles)
	awsv2.Register("aws_datazone_project", sweepProjects,
		"aws_datazone_environment",
	)
}

func sweepDomains(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.DataZoneClient(ctx)
	var sweepResources []sweep.Sweepable
	var input datazone.ListDomainsInput

	pages := datazone.NewListDomainsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.Items {
			id := aws.ToString(v.Id)

			sweepResources = append(sweepResources, framework.NewSweepResource(newDomainResource, client,
				framework.NewAttribute(names.AttrID, id),
			))
		}
	}

	return sweepResources, nil
}

func sweepEnvironments(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.DataZoneClient(ctx)
	var sweepResources []sweep.Sweepable
	var input datazone.ListDomainsInput

	pages := datazone.NewListDomainsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.Items {
			domainID := aws.ToString(v.Id)
			input := datazone.ListProjectsInput{
				DomainIdentifier: aws.String(domainID),
			}

			pages := datazone.NewListProjectsPaginator(conn, &input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)
				if err != nil {
					return nil, err
				}

				for _, v := range page.Items {
					projectID := aws.ToString(v.Id)
					input := datazone.ListEnvironmentsInput{
						DomainIdentifier:  aws.String(domainID),
						ProjectIdentifier: aws.String(projectID),
					}

					pages := datazone.NewListEnvironmentsPaginator(conn, &input)
					for pages.HasMorePages() {
						page, err := pages.NextPage(ctx)
						if err != nil {
							return nil, err
						}

						for _, v := range page.Items {
							id := aws.ToString(v.Id)

							log.Printf("[INFO] Skipping Data Zone Environment: %s", id)
							// sweepResources = append(sweepResources, framework.NewSweepResource(newEnvironmentResource, client,
							// 	framework.NewAttribute(names.AttrID, id),
							// 	framework.NewAttribute("domain_identifier", domainID),
							// ))
						}
					}
				}
			}
		}
	}

	return sweepResources, nil
}

func sweepEnvironmentProfiles(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.DataZoneClient(ctx)
	var sweepResources []sweep.Sweepable
	var input datazone.ListDomainsInput

	pages := datazone.NewListDomainsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.Items {
			domainID := aws.ToString(v.Id)
			input := datazone.ListEnvironmentProfilesInput{
				DomainIdentifier: aws.String(domainID),
			}

			pages := datazone.NewListEnvironmentProfilesPaginator(conn, &input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)
				if err != nil {
					return nil, err
				}

				for _, v := range page.Items {
					sweepResources = append(sweepResources, framework.NewSweepResource(newEnvironmentProfileResource, client,
						framework.NewAttribute(names.AttrID, v.Id),
						framework.NewAttribute("domain_identifier", domainID),
					))
				}
			}
		}
	}

	return sweepResources, nil
}

func sweepProjects(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.DataZoneClient(ctx)
	var sweepResources []sweep.Sweepable
	var input datazone.ListDomainsInput

	pages := datazone.NewListDomainsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.Items {
			domainID := aws.ToString(v.Id)
			input := datazone.ListProjectsInput{
				DomainIdentifier: aws.String(domainID),
			}

			pages := datazone.NewListProjectsPaginator(conn, &input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)
				if err != nil {
					return nil, err
				}

				for _, v := range page.Items {
					sweepResources = append(sweepResources, framework.NewSweepResource(newProjectResource, client,
						framework.NewAttribute(names.AttrID, v.Id),
						framework.NewAttribute("domain_identifier", domainID),
						framework.NewAttribute("skip_deletion_check", true), // Automatically delete associated Glossaries
					))
				}
			}
		}
	}

	return sweepResources, nil
}
