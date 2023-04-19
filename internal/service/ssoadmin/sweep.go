//go:build sweep
// +build sweep

package ssoadmin

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_ssoadmin_account_assignment", &resource.Sweeper{
		Name: "aws_ssoadmin_account_assignment",
		F:    sweepAccountAssignments,
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
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).SSOAdminConn()

	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	// Need to Read the SSO Instance first; assumes the first instance returned
	// is where the permission sets exist as AWS SSO currently supports only 1 instance
	ds := DataSourceInstances()
	dsData := ds.Data(nil)

	err = sweep.ReadResource(ctx, ds, dsData, client)

	if tfawserr.ErrCodeContains(err, "AccessDenied") {
		log.Printf("[WARN] Skipping SSO Account Assignment sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return err
	}

	instanceArn := dsData.Get("arns").(*schema.Set).List()[0].(string)

	// To sweep account assignments, we need to first determine which Permission Sets
	// are available and then search for their respective assignments
	input := &ssoadmin.ListPermissionSetsInput{
		InstanceArn: aws.String(instanceArn),
	}

	err = conn.ListPermissionSetsPagesWithContext(ctx, input, func(page *ssoadmin.ListPermissionSetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, permissionSet := range page.PermissionSets {
			if permissionSet == nil {
				continue
			}

			permissionSetArn := aws.StringValue(permissionSet)

			input := &ssoadmin.ListAccountAssignmentsInput{
				AccountId:        aws.String(client.(*conns.AWSClient).AccountID),
				InstanceArn:      aws.String(instanceArn),
				PermissionSetArn: permissionSet,
			}

			err := conn.ListAccountAssignmentsPagesWithContext(ctx, input, func(page *ssoadmin.ListAccountAssignmentsOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, a := range page.AccountAssignments {
					if a == nil {
						continue
					}

					principalID := aws.StringValue(a.PrincipalId)
					principalType := aws.StringValue(a.PrincipalType)
					targetID := aws.StringValue(a.AccountId)
					targetType := ssoadmin.TargetTypeAwsAccount // only valid value currently accepted by API

					r := ResourceAccountAssignment()
					d := r.Data(nil)
					d.SetId(fmt.Sprintf("%s,%s,%s,%s,%s,%s", principalID, principalType, targetID, targetType, permissionSetArn, instanceArn))

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}

				return !lastPage
			})

			if sweep.SkipSweepError(err) {
				log.Printf("[WARN] Skipping SSO Account Assignment sweep (PermissionSet %s) for %s: %s", permissionSetArn, region, err)
				continue
			}
			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving SSO Account Assignments for Permission Set (%s): %w", permissionSetArn, err))
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping SSO Account Assignment sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving SSO Permission Sets for Account Assignment sweep: %w", err))
	}

	if err := sweep.SweepOrchestratorWithContext(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping SSO Account Assignments: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepPermissionSets(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).SSOAdminConn()

	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	// Need to Read the SSO Instance first; assumes the first instance returned
	// is where the permission sets exist as AWS SSO currently supports only 1 instance
	ds := DataSourceInstances()
	dsData := ds.Data(nil)

	err = sweep.ReadResource(ctx, ds, dsData, client)

	if tfawserr.ErrCodeContains(err, "AccessDenied") {
		log.Printf("[WARN] Skipping SSO Permission Set sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return err
	}

	instanceArn := dsData.Get("arns").(*schema.Set).List()[0].(string)

	input := &ssoadmin.ListPermissionSetsInput{
		InstanceArn: aws.String(instanceArn),
	}

	err = conn.ListPermissionSetsPagesWithContext(ctx, input, func(page *ssoadmin.ListPermissionSetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, permissionSet := range page.PermissionSets {
			if permissionSet == nil {
				continue
			}

			arn := aws.StringValue(permissionSet)

			log.Printf("[INFO] Deleting SSO Permission Set: %s", arn)

			r := ResourcePermissionSet()
			d := r.Data(nil)
			d.SetId(fmt.Sprintf("%s,%s", arn, instanceArn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping SSO Permission Set sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving SSO Permission Sets: %w", err))
	}

	if err := sweep.SweepOrchestratorWithContext(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping SSO Permission Sets: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}
