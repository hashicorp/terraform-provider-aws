// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package datazone

import (
	"context"

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

		for _, da := range page.Items {
			id := aws.ToString(da.Id)

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

	var domainsInput datazone.ListDomainsInput
	pages := datazone.NewListDomainsPaginator(conn, &domainsInput)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, domain := range page.Items {
			projectsInput := datazone.ListProjectsInput{
				DomainIdentifier: domain.Id,
			}
			pages := datazone.NewListProjectsPaginator(conn, &projectsInput)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)
				if err != nil {
					return nil, err
				}

				for _, project := range page.Items {
					environmentsInput := datazone.ListEnvironmentsInput{
						DomainIdentifier:  domain.Id,
						ProjectIdentifier: project.Id,
					}
					pages := datazone.NewListEnvironmentsPaginator(conn, &environmentsInput)
					for pages.HasMorePages() {
						page, err := pages.NextPage(ctx)
						if err != nil {
							return nil, err
						}

						for _, environment := range page.Items {
							sweepResources = append(sweepResources, framework.NewSweepResource(newEnvironmentResource, client,
								framework.NewAttribute(names.AttrID, environment.Id),
								framework.NewAttribute("domain_identifier", domain.Id),
							))
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

	var domainsInput datazone.ListDomainsInput
	pages := datazone.NewListDomainsPaginator(conn, &domainsInput)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, domain := range page.Items {
			environmentProfilesInput := datazone.ListEnvironmentProfilesInput{
				DomainIdentifier: domain.Id,
			}
			pages := datazone.NewListEnvironmentProfilesPaginator(conn, &environmentProfilesInput)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)
				if err != nil {
					return nil, err
				}

				for _, profile := range page.Items {
					sweepResources = append(sweepResources, framework.NewSweepResource(newEnvironmentProfileResource, client,
						framework.NewAttribute(names.AttrID, profile.Id),
						framework.NewAttribute("domain_identifier", domain.Id),
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

	var domainsInput datazone.ListDomainsInput
	pages := datazone.NewListDomainsPaginator(conn, &domainsInput)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, domain := range page.Items {
			projectsInput := datazone.ListProjectsInput{
				DomainIdentifier: domain.Id,
			}
			pages := datazone.NewListProjectsPaginator(conn, &projectsInput)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)
				if err != nil {
					return nil, err
				}

				for _, project := range page.Items {
					sweepResources = append(sweepResources, framework.NewSweepResource(newProjectResource, client,
						framework.NewAttribute(names.AttrID, project.Id),
						framework.NewAttribute("domain_identifier", domain.Id),
						framework.NewAttribute("skip_deletion_check", true), // Automatically delete associated Glossaries
					))
				}
			}
		}
	}

	return sweepResources, nil
}
