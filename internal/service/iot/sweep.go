// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv1"
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
		Name:         "aws_iot_thing",
		F:            sweepThings,
		Dependencies: []string{"aws_iot_thing_principal_attachment"},
	})

	resource.AddTestSweepers("aws_iot_thing_group", &resource.Sweeper{
		Name: "aws_iot_policy_attachment",
		F:    sweepThingGroups,
	})

	resource.AddTestSweepers("aws_iot_thing_type", &resource.Sweeper{
		Name:         "aws_iot_thing_type",
		F:            sweepThingTypes,
		Dependencies: []string{"aws_iot_thing"},
	})

	resource.AddTestSweepers("aws_iot_topic_rule", &resource.Sweeper{
		Name:         "aws_iot_topic_rule",
		F:            sweepTopicRules,
		Dependencies: []string{"aws_iot_topic_rule_destination"},
	})

	resource.AddTestSweepers("aws_iot_topic_rule_destination", &resource.Sweeper{
		Name: "aws_iot_topic_rule_destination",
		F:    sweepTopicRuleDestinations,
	})

	resource.AddTestSweepers("aws_iot_authorizer", &resource.Sweeper{
		Name:         "aws_iot_authorizer",
		F:            sweepAuthorizers,
		Dependencies: []string{"aws_iot_domain_configuration"},
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
	conn := client.IoTConn(ctx)
	input := &iot.ListCertificatesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListCertificatesPagesWithContext(ctx, input, func(page *iot.ListCertificatesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Certificates {
			r := ResourceCertificate()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.CertificateId))
			d.Set("active", true)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping IoT Certificate sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing IoT Certificates (%s): %w", region, err)
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
	conn := client.IoTConn(ctx)
	input := &iot.ListPoliciesInput{}
	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	err = conn.ListPoliciesPagesWithContext(ctx, input, func(page *iot.ListPoliciesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Policies {
			policyName := aws.StringValue(v.PolicyName)
			input := &iot.ListTargetsForPolicyInput{
				PolicyName: aws.String(policyName),
			}

			err := conn.ListTargetsForPolicyPagesWithContext(ctx, input, func(page *iot.ListTargetsForPolicyOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, v := range page.Targets {
					r := ResourcePolicyAttachment()
					d := r.Data(nil)
					d.SetId(fmt.Sprintf("%s|%s", policyName, aws.StringValue(v)))
					d.Set(names.AttrPolicy, policyName)
					d.Set("target", v)

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}

				return !lastPage
			})

			if awsv1.SkipSweepError(err) {
				continue
			}

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing IoT Targets For Policy (%s): %w", region, err))
			}
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping IoT Policy Attachment sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing IoT Policies (%s): %w", region, err))
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
	conn := client.IoTConn(ctx)
	input := &iot.ListPoliciesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListPoliciesPagesWithContext(ctx, input, func(page *iot.ListPoliciesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Policies {
			r := ResourcePolicy()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.PolicyName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping IoT Policy sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing IoT Policies (%s): %w", region, err)
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
	conn := client.IoTConn(ctx)
	input := &iot.ListRoleAliasesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListRoleAliasesPagesWithContext(ctx, input, func(page *iot.ListRoleAliasesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.RoleAliases {
			r := ResourceRoleAlias()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Role Alias sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Role Aliases (%s): %w", region, err)
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
	conn := client.IoTConn(ctx)
	input := &iot.ListThingsInput{}
	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	err = conn.ListThingsPagesWithContext(ctx, input, func(page *iot.ListThingsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Things {
			thingName := aws.StringValue(v.ThingName)
			input := &iot.ListThingPrincipalsInput{
				ThingName: aws.String(thingName),
			}

			err := conn.ListThingPrincipalsPagesWithContext(ctx, input, func(page *iot.ListThingPrincipalsOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, v := range page.Principals {
					r := ResourceThingPrincipalAttachment()
					d := r.Data(nil)
					d.SetId(fmt.Sprintf("%s|%s", thingName, aws.StringValue(v)))
					d.Set("principal", v)
					d.Set("thing", thingName)

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}

				return !lastPage
			})

			if awsv1.SkipSweepError(err) {
				continue
			}

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing IoT Thing Principals (%s): %w", region, err))
			}
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping IoT Thing Principal Attachment sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing IoT Things (%s): %w", region, err))
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
	conn := client.IoTConn(ctx)
	input := &iot.ListThingsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListThingsPagesWithContext(ctx, input, func(page *iot.ListThingsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Things {
			r := ResourceThing()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.ThingName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping IoT Thing sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing IoT Things (%s): %w", region, err)
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
	conn := client.IoTConn(ctx)
	input := &iot.ListThingTypesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListThingTypesPagesWithContext(ctx, input, func(page *iot.ListThingTypesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ThingTypes {
			r := ResourceThingType()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.ThingTypeName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping IoT Thing Type sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing IoT Thing Types (%s): %w", region, err)
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
	conn := client.IoTConn(ctx)
	input := &iot.ListTopicRulesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListTopicRulesPagesWithContext(ctx, input, func(page *iot.ListTopicRulesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Rules {
			r := ResourceTopicRule()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.RuleName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping IoT Topic Rule sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing IoT Topic Rules (%s): %w", region, err)
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
	conn := client.IoTConn(ctx)
	input := &iot.ListThingGroupsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListThingGroupsPagesWithContext(ctx, input, func(page *iot.ListThingGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ThingGroups {
			r := ResourceThingGroup()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.GroupName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping IoT Thing Group sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing IoT Thing Groups (%s): %w", region, err)
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
	conn := client.IoTConn(ctx)
	input := &iot.ListTopicRuleDestinationsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListTopicRuleDestinationsPagesWithContext(ctx, input, func(page *iot.ListTopicRuleDestinationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DestinationSummaries {
			r := ResourceTopicRuleDestination()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.Arn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping IoT Topic Rule Destination sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing IoT Topic Rule Destinations (%s): %w", region, err)
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
	conn := client.IoTConn(ctx)
	input := &iot.ListAuthorizersInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListAuthorizersPagesWithContext(ctx, input, func(page *iot.ListAuthorizersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Authorizers {
			r := ResourceAuthorizer()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.AuthorizerName))
			d.Set(names.AttrStatus, iot.AuthorizerStatusActive)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping IoT Authorizer sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing IoT Authorizers (%s): %w", region, err)
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
	conn := client.IoTConn(ctx)
	input := &iot.ListDomainConfigurationsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListDomainConfigurationsPagesWithContext(ctx, input, func(page *iot.ListDomainConfigurationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DomainConfigurations {
			name := aws.StringValue(v.DomainConfigurationName)

			if strings.HasPrefix(name, "iot:") {
				log.Printf("[INFO] Skipping IoT Domain Configuration %s", name)
				continue
			}

			output, err := FindDomainConfigurationByName(ctx, conn, name)

			if err != nil {
				log.Printf("[WARN] IoT Domain Configuration (%s): %s", name, err)
				continue
			}

			if aws.StringValue(output.DomainType) == iot.DomainTypeAwsManaged && aws.StringValue(output.DomainConfigurationStatus) == iot.DomainConfigurationStatusDisabled {
				// AWS Managed Domain Configuration must be disabled for at least 7 days before it can be deleted.
				if output.LastStatusChangeDate.After(time.Now().AddDate(0, 0, -7)) {
					log.Printf("[INFO] Skipping IoT Domain Configuration %s", name)
					continue
				}
			}

			r := ResourceDomainConfiguration()
			d := r.Data(nil)
			d.SetId(name)
			d.Set(names.AttrStatus, output.DomainConfigurationStatus)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping IoT Domain Configuration sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing IoT Domain Configurations (%s): %w", region, err)
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
	conn := client.IoTConn(ctx)
	input := &iot.ListCACertificatesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListCACertificatesPagesWithContext(ctx, input, func(page *iot.ListCACertificatesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Certificates {
			r := ResourceCACertificate()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.CertificateId))
			d.Set("active", true)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping IoT CA Certificate sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing IoT CA Certificates (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping IoT CA Certificates (%s): %w", region, err)
	}

	return nil
}
