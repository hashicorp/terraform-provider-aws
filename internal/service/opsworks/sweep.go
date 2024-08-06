// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opsworks

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv1"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/sdk"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_opsworks_stack", &resource.Sweeper{
		Name: "aws_opsworks_stack",
		F:    sweepStacks,
		Dependencies: []string{
			"aws_opsworks_application",
			"aws_opsworks_layer",
			"aws_opsworks_instance",
			"aws_opsworks_rds_db_instance",
		},
	})

	resource.AddTestSweepers("aws_opsworks_application", &resource.Sweeper{
		Name: "aws_opsworks_application",
		F:    sweepApplication,
	})

	resource.AddTestSweepers("aws_opsworks_instance", &resource.Sweeper{
		Name: "aws_opsworks_instance",
		F:    sweepInstance,
	})

	// This sweep all the custom, ecs, ganglia, etc. layers
	resource.AddTestSweepers("aws_opsworks_layer", &resource.Sweeper{
		Name: "aws_opsworks_layer",
		F:    sweepLayers,
		Dependencies: []string{
			"aws_opsworks_instance",
			"aws_opsworks_rds_db_instance",
		},
	})

	resource.AddTestSweepers("aws_opsworks_rds_db_instance", &resource.Sweeper{
		Name: "aws_opsworks_rds_db_instance",
		F:    sweepRDSDBInstance,
		Dependencies: []string{
			"aws_db_instance",
		},
	})

	resource.AddTestSweepers("aws_opsworks_user_profile", &resource.Sweeper{
		Name: "aws_opsworks_user_profile",
		F:    sweepUserProfiles,
	})
}

func sweepApplication(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.OpsWorksConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	output, err := conn.DescribeStacksWithContext(ctx, &opsworks.DescribeStacksInput{})

	if err != nil {
		if awsv1.SkipSweepError(err) {
			log.Printf("[WARN] Skipping OpsWorks Application sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("retrieving OpsWorks Stacks (Application sweep): %s", err)
	}

	var sweeperErrs *multierror.Error

	for _, stack := range output.Stacks {
		input := &opsworks.DescribeAppsInput{
			StackId: stack.StackId,
		}

		appOutput, err := conn.DescribeAppsWithContext(ctx, input)

		if err != nil {
			sweeperErr := fmt.Errorf("describing OpsWorks Applications for Stack (%s): %w", aws.StringValue(stack.StackId), err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}

		for _, app := range appOutput.Apps {
			if app == nil {
				continue
			}

			r := ResourceApplication()
			d := r.Data(nil)
			d.SetId(aws.StringValue(app.AppId))

			sweepResources = append(sweepResources, sdk.NewSweepResource(r, d, client))
		}
	}

	return sweep.SweepOrchestrator(ctx, sweepResources)
}

func sweepInstance(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.OpsWorksConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	output, err := conn.DescribeStacksWithContext(ctx, &opsworks.DescribeStacksInput{})

	if err != nil {
		if awsv1.SkipSweepError(err) {
			log.Printf("[WARN] Skipping OpsWorks Instance sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("retrieving OpsWorks Stacks (Instance sweep): %s", err)
	}

	var sweeperErrs *multierror.Error

	for _, stack := range output.Stacks {
		input := &opsworks.DescribeInstancesInput{
			StackId: stack.StackId,
		}

		instanceOutput, err := conn.DescribeInstancesWithContext(ctx, input)

		if err != nil {
			sweeperErr := fmt.Errorf("describing OpsWorks Instances for Stack (%s): %w", aws.StringValue(stack.StackId), err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}

		for _, instance := range instanceOutput.Instances {
			if instance == nil {
				continue
			}

			r := ResourceInstance()
			d := r.Data(nil)
			d.SetId(aws.StringValue(instance.InstanceId))
			d.Set(names.AttrStatus, instance.Status)

			sweepResources = append(sweepResources, sdk.NewSweepResource(r, d, client))
		}
	}

	return sweep.SweepOrchestrator(ctx, sweepResources)
}

func sweepRDSDBInstance(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.OpsWorksConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	output, err := conn.DescribeStacksWithContext(ctx, &opsworks.DescribeStacksInput{})

	if err != nil {
		if awsv1.SkipSweepError(err) {
			log.Printf("[WARN] Skipping OpsWorks RDS DB Instance sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("retrieving OpsWorks Stacks (RDS DB Instance sweep): %s", err)
	}

	var sweeperErrs *multierror.Error

	for _, stack := range output.Stacks {
		input := &opsworks.DescribeRdsDbInstancesInput{
			StackId: stack.StackId,
		}

		dbInstOutput, err := conn.DescribeRdsDbInstancesWithContext(ctx, input)

		if err != nil {
			sweeperErr := fmt.Errorf("describing OpsWorks RDS DB Instances for Stack (%s): %w", aws.StringValue(stack.StackId), err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}

		for _, dbInstance := range dbInstOutput.RdsDbInstances {
			if dbInstance == nil {
				continue
			}

			r := ResourceRDSDBInstance()
			d := r.Data(nil)
			d.SetId(aws.StringValue(dbInstance.DbInstanceIdentifier))
			d.Set("rds_db_instance_arn", dbInstance.RdsDbInstanceArn)

			sweepResources = append(sweepResources, sdk.NewSweepResource(r, d, client))
		}
	}

	return sweep.SweepOrchestrator(ctx, sweepResources)
}

func sweepStacks(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.OpsWorksConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	output, err := conn.DescribeStacksWithContext(ctx, &opsworks.DescribeStacksInput{})

	if err != nil {
		if awsv1.SkipSweepError(err) {
			log.Printf("[WARN] Skipping OpsWorks Stack sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("retrieving OpsWorks Stacks: %s", err)
	}

	for _, stack := range output.Stacks {
		if stack == nil {
			continue
		}

		r := ResourceStack()
		d := r.Data(nil)
		d.SetId(aws.StringValue(stack.StackId))

		if aws.StringValue(stack.VpcId) != "" {
			d.Set(names.AttrVPCID, stack.VpcId)
		}

		if aws.BoolValue(stack.UseOpsworksSecurityGroups) {
			d.Set("use_opsworks_security_groups", true)
		}

		sweepResources = append(sweepResources, sdk.NewSweepResource(r, d, client))
	}

	return sweep.SweepOrchestrator(ctx, sweepResources)
}

func sweepLayers(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.OpsWorksConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	output, err := conn.DescribeStacksWithContext(ctx, &opsworks.DescribeStacksInput{})

	if err != nil {
		if awsv1.SkipSweepError(err) {
			log.Printf("[WARN] Skipping OpsWorks Layer sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("retrieving OpsWorks Stacks (Layer sweep): %s", err)
	}

	var sweeperErrs *multierror.Error

	for _, stack := range output.Stacks {
		input := &opsworks.DescribeLayersInput{
			StackId: stack.StackId,
		}

		layerOutput, err := conn.DescribeLayersWithContext(ctx, input)

		if err != nil {
			sweeperErr := fmt.Errorf("describing OpsWorks Layers for Stack (%s): %w", aws.StringValue(stack.StackId), err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}

		for _, layer := range layerOutput.Layers {
			if layer == nil {
				continue
			}

			l := &opsworksLayerType{}
			r := l.resourceSchema()
			d := r.Data(nil)
			d.SetId(aws.StringValue(layer.LayerId))

			if layer.Attributes != nil {
				if v, ok := layer.Attributes[opsworks.LayerAttributesKeysEcsClusterArn]; ok && aws.StringValue(v) != "" {
					r = ResourceECSClusterLayer()
					d = r.Data(nil)
					d.SetId(aws.StringValue(layer.LayerId))
					d.Set("ecs_cluster_arn", v)
				}
			}

			sweepResources = append(sweepResources, sdk.NewSweepResource(r, d, client))
		}
	}

	return sweep.SweepOrchestrator(ctx, sweepResources)
}

func sweepUserProfiles(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.OpsWorksConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	output, err := conn.DescribeUserProfilesWithContext(ctx, &opsworks.DescribeUserProfilesInput{})

	if err != nil {
		if awsv1.SkipSweepError(err) {
			log.Printf("[WARN] Skipping OpsWorks User Profile sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("retrieving OpsWorks User Profiles: %w", err)
	}

	for _, profile := range output.UserProfiles {
		r := ResourceUserProfile()
		d := r.Data(nil)
		d.SetId(aws.StringValue(profile.IamUserArn))
		sweepResources = append(sweepResources, newUserProfileSweeper(r, d, client))
	}

	return sweep.SweepOrchestrator(ctx, sweepResources)
}

type userProfileSweeper struct {
	d         *schema.ResourceData
	sweepable sweep.Sweepable
}

func newUserProfileSweeper(resource *schema.Resource, d *schema.ResourceData, client *conns.AWSClient) *userProfileSweeper {
	return &userProfileSweeper{
		d:         d,
		sweepable: sdk.NewSweepResource(resource, d, client),
	}
}

func (ups userProfileSweeper) Delete(ctx context.Context, timeout time.Duration, optFns ...tfresource.OptionsFunc) error {
	err := ups.sweepable.Delete(ctx, timeout, optFns...)
	if err != nil && strings.Contains(err.Error(), "Cannot delete self") {
		log.Printf("[WARN] Skipping OpsWorks User Profile (%s): %s", ups.d.Id(), err)
		return nil
	}
	return err
}
