// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iot"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iot/types"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_iot_certificate", &resource.Sweeper{
		Name: "aws_iot_certificate",
		F:    sweepCertificates,
		Dependencies: []string{
			"aws_iot_policy_attachment",
			"aws_iot_thing_principal_attachment",
		},
	})

	resource.AddTestSweepers("aws_iot_policy_attachment", &resource.Sweeper{
		Name: "aws_iot_policy_attachment",
		F:    sweepPolicyAttachments,
	})

	resource.AddTestSweepers("aws_iot_policy", &resource.Sweeper{
		Name: "aws_iot_policy",
		F:    sweepPolicies,
		Dependencies: []string{
			"aws_iot_policy_attachment",
		},
	})

	resource.AddTestSweepers("aws_iot_role_alias", &resource.Sweeper{
		Name: "aws_iot_role_alias",
		F:    sweepRoleAliases,
	})

	resource.AddTestSweepers("aws_iot_thing_principal_attachment", &resource.Sweeper{
		Name: "aws_iot_thing_principal_attachment",
		F:    sweepThingPrincipalAttachments,
	})

	resource.AddTestSweepers("aws_iot_thing", &resource.Sweeper{
		Name: "aws_iot_thing",
		F:    sweepThings,
		Dependencies: []string{
			"aws_iot_thing_principal_attachment",
		},
	})

	resource.AddTestSweepers("aws_iot_thing_group", &resource.Sweeper{
		Name: "aws_iot_thing_group",
		F:    sweepThingGroups,
	})

	resource.AddTestSweepers("aws_iot_thing_type", &resource.Sweeper{
		Name: "aws_iot_thing_type",
		F:    sweepThingTypes,
		Dependencies: []string{
			"aws_iot_thing",
		},
	})

	resource.AddTestSweepers("aws_iot_topic_rule", &resource.Sweeper{
		Name: "aws_iot_topic_rule",
		F:    sweepTopicRules,
		Dependencies: []string{
			"aws_iot_topic_rule_destination",
		},
	})

	resource.AddTestSweepers("aws_iot_topic_rule_destination", &resource.Sweeper{
		Name: "aws_iot_topic_rule_destination",
		F:    sweepTopicRuleDestinations,
	})

	resource.AddTestSweepers("aws_iot_authorizer", &resource.Sweeper{
		Name: "aws_iot_authorizer",
		F:    sweepAuthorizers,
		Dependencies: []string{
			"aws_iot_domain_configuration",
		},
	})

	resource.AddTestSweepers("aws_iot_domain_configuration", &resource.Sweeper{
		Name: "aws_iot_domain_configuration",
		F:    sweepDomainConfigurations,
	})

	resource.AddTestSweepers("aws_iot_ca_certificate", &resource.Sweeper{
		Name: "aws_iot_ca_certificate",
		F:    sweepCACertificates,
		Dependencies: []string{
			"aws_iot_certificate",
		},
	})
}

func sweepCertificates(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.IoTClient(ctx)
	input := &iot.ListCertificatesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := iot.NewListCertificatesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping IoT Certificate sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing IoT Certificates (%s): %w", region, err)
		}

		for _, v := range page.Certificates {
			r := resourceCertificate()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.CertificateId))
			d.Set("active", true)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping IoT Certificates (%s): %w", region, err)
	}

	return nil
}

func sweepPolicyAttachments(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.IoTClient(ctx)
	input := &iot.ListPoliciesInput{}
	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	pages := iot.NewListPoliciesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping IoT Policy Attachment sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing IoT Policies (%s): %w", region, err))
		}

		for _, v := range page.Policies {
			policyName := v.PolicyName
			input := &iot.ListTargetsForPolicyInput{
				PolicyName: policyName,
			}

			pages := iot.NewListTargetsForPolicyPaginator(conn, input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				for _, v := range page.Targets {
					r := resourcePolicyAttachment()
					d := r.Data(nil)
					d.SetId(fmt.Sprintf("%s|%s", aws.ToString(policyName), v))
					d.Set(names.AttrPolicy, policyName)
					d.Set(names.AttrTarget, v)

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}

				if awsv2.SkipSweepError(err) {
					continue
				}

				if err != nil {
					sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing IoT Targets For Policy (%s): %w", region, err))
				}
			}
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping IoT Policy Attachments (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepPolicies(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.IoTClient(ctx)
	input := &iot.ListPoliciesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := iot.NewListPoliciesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping IoT Policy sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing IoT Policies (%s): %w", region, err)
		}

		for _, v := range page.Policies {
			r := resourcePolicy()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.PolicyName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping IoT Policies (%s): %w", region, err)
	}

	return nil
}

func sweepRoleAliases(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.IoTClient(ctx)
	input := &iot.ListRoleAliasesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := iot.NewListRoleAliasesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Role Alias sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Role Aliases (%s): %w", region, err)
		}

		for _, v := range page.RoleAliases {
			r := ResourceRoleAlias()
			d := r.Data(nil)
			d.SetId(v)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping IoT Role Aliases (%s): %w", region, err)
	}

	return nil
}

func sweepThingPrincipalAttachments(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.IoTClient(ctx)
	input := &iot.ListThingsInput{}
	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	pages := iot.NewListThingsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping IoT Thing Principal Attachment sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing IoT Things (%s): %w", region, err))
		}

		for _, v := range page.Things {
			thingName := aws.ToString(v.ThingName)
			input := &iot.ListThingPrincipalsInput{
				ThingName: aws.String(thingName),
			}

			pages := iot.NewListThingPrincipalsPaginator(conn, input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if awsv2.SkipSweepError(err) {
					continue
				}

				if err != nil {
					sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing IoT Thing Principals (%s): %w", region, err))
				}

				for _, v := range page.Principals {
					r := resourceThingPrincipalAttachment()
					d := r.Data(nil)
					d.SetId(fmt.Sprintf("%s|%s", thingName, v))
					d.Set(names.AttrPrincipal, v)
					d.Set("thing", thingName)

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}
			}
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping IoT Thing Principal Attachments (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepThings(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.IoTClient(ctx)
	input := &iot.ListThingsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := iot.NewListThingsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping IoT Thing sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing IoT Things (%s): %w", region, err)
		}

		for _, v := range page.Things {
			r := resourceThing()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.ThingName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping IoT Things (%s): %w", region, err)
	}

	return nil
}

func sweepThingTypes(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.IoTClient(ctx)
	input := &iot.ListThingTypesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := iot.NewListThingTypesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping IoT Thing Type sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing IoT Thing Types (%s): %w", region, err)
		}

		for _, v := range page.ThingTypes {
			r := resourceThingType()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.ThingTypeName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping IoT Thing Types (%s): %w", region, err)
	}

	return nil
}

func sweepTopicRules(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.IoTClient(ctx)
	input := &iot.ListTopicRulesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := iot.NewListTopicRulesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping IoT Topic Rule sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing IoT Topic Rules (%s): %w", region, err)
		}

		for _, v := range page.Rules {
			r := resourceTopicRule()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.RuleName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping IoT Topic Rules (%s): %w", region, err)
	}

	return nil
}

func sweepThingGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.IoTClient(ctx)
	input := &iot.ListThingGroupsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := iot.NewListThingGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping IoT Thing Group sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing IoT Thing Groups (%s): %w", region, err)
		}

		for _, v := range page.ThingGroups {
			r := resourceThingGroup()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.GroupName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping IoT Thing Groups (%s): %w", region, err)
	}

	return nil
}

func sweepTopicRuleDestinations(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.IoTClient(ctx)
	input := &iot.ListTopicRuleDestinationsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := iot.NewListTopicRuleDestinationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping IoT Topic Rule Destination sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing IoT Topic Rule Destinations (%s): %w", region, err)
		}

		for _, v := range page.DestinationSummaries {
			r := resourceTopicRuleDestination()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Arn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping IoT Topic Rule Destinations (%s): %w", region, err)
	}

	return nil
}

func sweepAuthorizers(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.IoTClient(ctx)
	input := &iot.ListAuthorizersInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := iot.NewListAuthorizersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping IoT Authorizer sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing IoT Authorizers (%s): %w", region, err)
		}

		for _, v := range page.Authorizers {
			r := resourceAuthorizer()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.AuthorizerName))
			d.Set(names.AttrStatus, awstypes.AuthorizerStatusActive)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping IoT Authorizers (%s): %w", region, err)
	}

	return nil
}

func sweepDomainConfigurations(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.IoTClient(ctx)
	input := &iot.ListDomainConfigurationsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := iot.NewListDomainConfigurationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping IoT Domain Configuration sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing IoT Domain Configurations (%s): %w", region, err)
		}

		for _, v := range page.DomainConfigurations {
			name := aws.ToString(v.DomainConfigurationName)

			if strings.HasPrefix(name, "iot:") {
				log.Printf("[INFO] Skipping IoT Domain Configuration %s", name)
				continue
			}

			output, err := findDomainConfigurationByName(ctx, conn, name)

			if err != nil {
				log.Printf("[WARN] IoT Domain Configuration (%s): %s", name, err)
				continue
			}

			if output.DomainType == awstypes.DomainTypeAwsManaged && output.DomainConfigurationStatus == awstypes.DomainConfigurationStatusDisabled {
				// AWS Managed Domain Configuration must be disabled for at least 7 days before it can be deleted.
				if output.LastStatusChangeDate.After(time.Now().AddDate(0, 0, -7)) {
					log.Printf("[INFO] Skipping IoT Domain Configuration %s", name)
					continue
				}
			}

			r := resourceDomainConfiguration()
			d := r.Data(nil)
			d.SetId(name)
			d.Set(names.AttrStatus, output.DomainConfigurationStatus)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping IoT Domain Configurations (%s): %w", region, err)
	}

	return nil
}

func sweepCACertificates(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.IoTClient(ctx)
	input := &iot.ListCACertificatesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := iot.NewListCACertificatesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping IoT CA Certificate sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing IoT CA Certificates (%s): %w", region, err)
		}

		for _, v := range page.Certificates {
			r := resourceCACertificate()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.CertificateId))
			d.Set("active", true)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping IoT CA Certificates (%s): %w", region, err)
	}

	return nil
}
