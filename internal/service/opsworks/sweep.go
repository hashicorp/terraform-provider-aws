//go:build sweep
// +build sweep

package opsworks

import (
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
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
	})

	resource.AddTestSweepers("aws_opsworks_user_profile", &resource.Sweeper{
		Name: "aws_opsworks_user_profile",
		F:    sweepUserProfiles,
	})
}

func sweepApplication(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).OpsWorksConn
	sweepResources := make([]*sweep.SweepResource, 0)

	output, err := conn.DescribeStacks(&opsworks.DescribeStacksInput{})

	if err != nil {
		if sweep.SkipSweepError(err) {
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

		appOutput, err := conn.DescribeApps(input)

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

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweep.SweepOrchestrator(sweepResources)
}

func sweepInstance(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).OpsWorksConn
	sweepResources := make([]*sweep.SweepResource, 0)

	output, err := conn.DescribeStacks(&opsworks.DescribeStacksInput{})

	if err != nil {
		if sweep.SkipSweepError(err) {
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

		instanceOutput, err := conn.DescribeInstances(input)

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
			d.Set("status", instance.Status)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweep.SweepOrchestrator(sweepResources)
}

func sweepRDSDBInstance(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).OpsWorksConn
	sweepResources := make([]*sweep.SweepResource, 0)

	output, err := conn.DescribeStacks(&opsworks.DescribeStacksInput{})

	if err != nil {
		if sweep.SkipSweepError(err) {
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

		dbInstOutput, err := conn.DescribeRdsDbInstances(input)

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

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweep.SweepOrchestrator(sweepResources)
}

func sweepStacks(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).OpsWorksConn
	sweepResources := make([]*sweep.SweepResource, 0)

	output, err := conn.DescribeStacks(&opsworks.DescribeStacksInput{})

	if err != nil {
		if sweep.SkipSweepError(err) {
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
			d.Set("vpc_id", stack.VpcId)
		}

		if aws.BoolValue(stack.UseOpsworksSecurityGroups) {
			d.Set("use_opsworks_security_groups", true)
		}

		sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
	}

	return sweep.SweepOrchestrator(sweepResources)
}

func sweepLayers(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).OpsWorksConn
	sweepResources := make([]*sweep.SweepResource, 0)

	output, err := conn.DescribeStacks(&opsworks.DescribeStacksInput{})

	if err != nil {
		if sweep.SkipSweepError(err) {
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

		layerOutput, err := conn.DescribeLayers(input)

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
			r := l.SchemaResource()
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

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweep.SweepOrchestrator(sweepResources)
}

func sweepUserProfiles(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).OpsWorksConn
	sweepResources := make([]*sweep.SweepResource, 0)

	output, err := conn.DescribeUserProfiles(&opsworks.DescribeUserProfilesInput{})

	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping OpsWorks User Profile sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("retrieving OpsWorks User Profiles: %w", err)
	}

	for _, profile := range output.UserProfiles {
		r := ResourceUserProfile()
		d := r.Data(nil)
		d.SetId(aws.StringValue(profile.IamUserArn))
		sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
	}

	err = sweep.SweepOrchestrator(sweepResources)

	var errs *multierror.Error
	if errors.As(err, &errs) {
		var es *multierror.Error
		for _, e := range errs.Errors {
			if tfawserr.ErrMessageContains(err, opsworks.ErrCodeValidationException, "Cannot delete self") {
				log.Printf("[WARN] Ignoring error: %s", e.Error())
			} else {
				es = multierror.Append(es, e)
			}
		}
		return es.ErrorOrNil()
	}

	return err
}
