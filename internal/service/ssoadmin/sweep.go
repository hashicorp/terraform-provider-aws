// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssoadmin

import (
	"fmt"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/sdk"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_ssoadmin_account_assignment", &resource.Sweeper{
		Name: "aws_ssoadmin_account_assignment",
		F:    sweepAccountAssignments,
	})
	resource.AddTestSweepers("aws_ssoadmin_application", &resource.Sweeper{
		Name: "aws_ssoadmin_application",
		F:    sweepApplications,
	})
	resource.AddTestSweepers("aws_ssoadmin_permission_set", &resource.Sweeper{
		Name: "aws_ssoadmin_permission_set",
		F:    sweepPermissionSets,
		Dependencies: []string{
			"aws_ssoadmin_account_assignment",
		},
	})
}

func sweepAccountAssignments(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.SSOAdminClient(ctx)

	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	accessDenied := regexache.MustCompile(`AccessDeniedException: .+ is not authorized to perform:`)

	// Need to Read the SSO Instance first; assumes the first instance returned
	// is where the permission sets exist as AWS SSO currently supports only 1 instance
	ds := DataSourceInstances()
	dsData := ds.Data(nil)

	if err := sdk.ReadResource(ctx, ds, dsData, client); err != nil {
		if accessDenied.MatchString(err.Error()) {
			log.Printf("[WARN] Skipping SSO Account Assignment sweep for %s: %s", region, err)
			return nil
		}
		return err
	}

	if v, ok := dsData.GetOk(names.AttrARNs); ok && len(v.([]interface{})) > 0 {
		instanceArn := v.([]interface{})[0].(string)

		// To sweep account assignments, we need to first determine which Permission Sets
		// are available and then search for their respective assignments
		input := &ssoadmin.ListPermissionSetsInput{
			InstanceArn: aws.String(instanceArn),
		}

		var permissionSetArns []string
		paginator := ssoadmin.NewListPermissionSetsPaginator(conn, input)
		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if awsv2.SkipSweepError(err) {
				log.Printf("[WARN] Skipping SSO Account Assignment sweep for %s: %s", region, err)
				return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
			}
			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving SSO Permission Sets for Account Assignment sweep: %w", err))
			}

			if page != nil {
				permissionSetArns = append(permissionSetArns, page.PermissionSets...)
			}
		}

		for _, permissionSetArn := range permissionSetArns {
			input := &ssoadmin.ListAccountAssignmentsInput{
				AccountId:        aws.String(client.AccountID),
				InstanceArn:      aws.String(instanceArn),
				PermissionSetArn: aws.String(permissionSetArn),
			}

			paginator := ssoadmin.NewListAccountAssignmentsPaginator(conn, input)
			for paginator.HasMorePages() {
				page, err := paginator.NextPage(ctx)
				if awsv2.SkipSweepError(err) {
					log.Printf("[WARN] Skipping SSO Account Assignment sweep (PermissionSet %s) for %s: %s", permissionSetArn, region, err)
					continue
				}
				if err != nil {
					sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving SSO Account Assignments for Permission Set (%s): %w", permissionSetArn, err))
				}

				for _, a := range page.AccountAssignments {
					principalID := aws.ToString(a.PrincipalId)
					principalType := string(a.PrincipalType)
					targetID := aws.ToString(a.AccountId)
					targetType := awstypes.TargetTypeAwsAccount // only valid value currently accepted by API

					r := ResourceAccountAssignment()
					d := r.Data(nil)
					d.SetId(fmt.Sprintf("%s,%s,%s,%s,%s,%s", principalID, principalType, targetID, targetType, permissionSetArn, instanceArn))

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}
			}
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping SSO Account Assignments: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepApplications(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.SSOAdminClient(ctx)

	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	accessDenied := regexache.MustCompile(`AccessDeniedException: .+ is not authorized to perform:`)

	// Need to Read the SSO Instance first; assumes the first instance returned
	// is where the permission sets exist as AWS SSO currently supports only 1 instance
	ds := DataSourceInstances()
	dsData := ds.Data(nil)

	if err := sdk.ReadResource(ctx, ds, dsData, client); err != nil {
		if accessDenied.MatchString(err.Error()) {
			log.Printf("[WARN] Skipping SSO Application sweep for %s: %s", region, err)
			return nil
		}
		return err
	}

	if v, ok := dsData.GetOk(names.AttrARNs); ok && len(v.([]interface{})) > 0 {
		instanceArn := v.([]interface{})[0].(string)

		input := &ssoadmin.ListApplicationsInput{
			InstanceArn: aws.String(instanceArn),
		}

		paginator := ssoadmin.NewListApplicationsPaginator(conn, input)
		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if awsv2.SkipSweepError(err) {
				log.Printf("[WARN] Skipping SSO Applications sweep for %s: %s", region, err)
				return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
			}
			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving SSO Applications: %w", err))
			}

			for _, application := range page.Applications {
				applicationARN := aws.ToString(application.ApplicationArn)
				log.Printf("[INFO] Deleting SSO Application: %s", applicationARN)

				sweepResources = append(sweepResources, framework.NewSweepResource(newResourceApplication, client, framework.NewAttribute("application_arn", applicationARN)))
			}
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping SSO Applications: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepPermissionSets(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.SSOAdminClient(ctx)

	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	accessDenied := regexache.MustCompile(`AccessDeniedException: .+ is not authorized to perform:`)

	// Need to Read the SSO Instance first; assumes the first instance returned
	// is where the permission sets exist as AWS SSO currently supports only 1 instance
	ds := DataSourceInstances()
	dsData := ds.Data(nil)

	if err := sdk.ReadResource(ctx, ds, dsData, client); err != nil {
		if accessDenied.MatchString(err.Error()) {
			log.Printf("[WARN] Skipping SSO Permission Set sweep for %s: %s", region, err)
			return nil
		}
		return err
	}

	if v, ok := dsData.GetOk(names.AttrARNs); ok && len(v.([]interface{})) > 0 {
		instanceArn := v.([]interface{})[0].(string)

		input := &ssoadmin.ListPermissionSetsInput{
			InstanceArn: aws.String(instanceArn),
		}

		paginator := ssoadmin.NewListPermissionSetsPaginator(conn, input)
		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if awsv2.SkipSweepError(err) {
				log.Printf("[WARN] Skipping SSO Permission Set sweep for %s: %s", region, err)
				return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
			}
			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving SSO Permission Sets: %w", err))
			}

			for _, permissionSetArn := range page.PermissionSets {
				log.Printf("[INFO] Deleting SSO Permission Set: %s", permissionSetArn)

				r := ResourcePermissionSet()
				d := r.Data(nil)
				d.SetId(fmt.Sprintf("%s,%s", permissionSetArn, instanceArn))

				sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
			}
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping SSO Permission Sets: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}
