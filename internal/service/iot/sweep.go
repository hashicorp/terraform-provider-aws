//go:build sweep
// +build sweep

package iot

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_iot_certificate", &resource.Sweeper{
		Name: "aws_iot_certificate",
		F:    sweepCertifcates,
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
}

func sweepCertifcates(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).IoTConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	input := &iot.ListCertificatesInput{}

	err = conn.ListCertificatesPages(input, func(page *iot.ListCertificatesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, certificate := range page.Certificates {
			r := ResourceCertificate()
			d := r.Data(nil)

			d.SetId(aws.StringValue(certificate.CertificateId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing IoT Certificate for %s: %w", region, err))
	}

	if err := sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping IoT Certificate for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping IoT Certificate sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepPolicyAttachments(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).IoTConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	input := &iot.ListPoliciesInput{}

	err = conn.ListPoliciesPages(input, func(page *iot.ListPoliciesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, policy := range page.Policies {
			input := &iot.ListTargetsForPolicyInput{
				PolicyName: policy.PolicyName,
			}

			err := conn.ListTargetsForPolicyPages(input, func(page *iot.ListTargetsForPolicyOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, target := range page.Targets {
					r := ResourcePolicyAttachment()
					d := r.Data(nil)

					d.SetId(fmt.Sprintf("%s|%s", aws.StringValue(policy.PolicyName), aws.StringValue(target)))
					d.Set("policy", policy.PolicyName)
					d.Set("target", target)

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}

				return !lastPage
			})

			if err != nil {
				errs = multierror.Append(errs, fmt.Errorf("error listing IoT Policy Attachment for %s: %w", region, err))
			}
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing IoT Policy Attachment for %s: %w", region, err))
	}

	if err := sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping IoT Policy Attachment for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping IoT Policy Attachment sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepPolicies(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).IoTConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	input := &iot.ListPoliciesInput{}

	err = conn.ListPoliciesPages(input, func(page *iot.ListPoliciesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, policy := range page.Policies {
			r := ResourcePolicy()
			d := r.Data(nil)

			d.SetId(aws.StringValue(policy.PolicyName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing IoT Policy for %s: %w", region, err))
	}

	if err := sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping IoT Policy for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping IoT Policy sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepRoleAliases(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).IoTConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	input := &iot.ListRoleAliasesInput{}

	err = conn.ListRoleAliasesPages(input, func(page *iot.ListRoleAliasesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, roleAlias := range page.RoleAliases {
			r := ResourceRoleAlias()
			d := r.Data(nil)

			d.SetId(aws.StringValue(roleAlias))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing IoT Role Alias for %s: %w", region, err))
	}

	if err := sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping IoT Role Alias for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping IoT Role Alias sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepThingPrincipalAttachments(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).IoTConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	input := &iot.ListThingsInput{}

	err = conn.ListThingsPages(input, func(page *iot.ListThingsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, thing := range page.Things {
			input := &iot.ListThingPrincipalsInput{
				ThingName: thing.ThingName,
			}

			err := conn.ListThingPrincipalsPages(input, func(page *iot.ListThingPrincipalsOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, principal := range page.Principals {
					r := ResourceThingPrincipalAttachment()
					d := r.Data(nil)

					d.SetId(fmt.Sprintf("%s|%s", aws.StringValue(thing.ThingName), aws.StringValue(principal)))
					d.Set("principal", principal)
					d.Set("thing", thing.ThingName)

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}

				return !lastPage
			})

			if err != nil {
				errs = multierror.Append(errs, fmt.Errorf("error listing IoT Thing Principal Attachment for %s: %w", region, err))
			}
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing IoT Thing Principal Attachment for %s: %w", region, err))
	}

	if err := sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping IoT Thing Principal Attachment for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping IoT Thing Principal Attachment sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepThings(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).IoTConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	input := &iot.ListThingsInput{}

	err = conn.ListThingsPages(input, func(page *iot.ListThingsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, thing := range page.Things {
			r := ResourceThing()
			d := r.Data(nil)

			d.SetId(aws.StringValue(thing.ThingName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing IoT Thing for %s: %w", region, err))
	}

	if err := sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping IoT Thing for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping IoT Thing sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepThingTypes(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).IoTConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	input := &iot.ListThingTypesInput{}

	err = conn.ListThingTypesPages(input, func(page *iot.ListThingTypesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, thingTypes := range page.ThingTypes {
			r := ResourceThingType()
			d := r.Data(nil)

			d.SetId(aws.StringValue(thingTypes.ThingTypeName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing IoT Thing Type for %s: %w", region, err))
	}

	if err := sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping IoT Thing Type for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping IoT Thing Type sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepTopicRules(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).IoTConn
	input := &iot.ListTopicRulesInput{}
	var sweeperErrs *multierror.Error

	for {
		output, err := conn.ListTopicRules(input)
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping IoT Topic Rules sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
		}
		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving IoT Topic Rules: %w", err))
			return sweeperErrs
		}

		for _, rule := range output.Rules {
			name := aws.StringValue(rule.RuleName)

			log.Printf("[INFO] Deleting IoT Topic Rule: %s", name)
			_, err := conn.DeleteTopicRule(&iot.DeleteTopicRuleInput{
				RuleName: aws.String(name),
			})
			if tfawserr.ErrCodeEquals(err, iot.ErrCodeUnauthorizedException) {
				continue
			}
			if err != nil {
				sweeperErr := fmt.Errorf("error deleting IoT Topic Rule (%s): %w", name, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}
		input.NextToken = output.NextToken
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepThingGroups(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).IoTConn
	input := &iot.ListThingGroupsInput{}
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.ListThingGroupsPages(input, func(page *iot.ListThingGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, group := range page.ThingGroups {
			r := ResourceThingGroup()
			d := r.Data(nil)
			d.SetId(aws.StringValue(group.GroupName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping IoT Thing Group sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing IoT Thing Groups (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping IoT Thing Groups (%s): %w", region, err)
	}

	return nil
}

func sweepTopicRuleDestinations(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).IoTConn
	input := &iot.ListTopicRuleDestinationsInput{}
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.ListTopicRuleDestinationsPages(input, func(page *iot.ListTopicRuleDestinationsOutput, lastPage bool) bool {
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

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping IoT Topic Rule Destination sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing IoT Topic Rule Destinations (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping IoT Topic Rule Destinations (%s): %w", region, err)
	}

	return nil
}
