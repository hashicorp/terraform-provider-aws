//go:build sweep
// +build sweep

package opsworks

import (
	"fmt"
	"log"
	"strings"
	"time"

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
			"aws_opsworks_custom_layer",
			"aws_opsworks_ecs_cluster_layer",
			"aws_opsworks_ganglia_layer",
			"aws_opsworks_haproxy_layer",
			"aws_opsworks_instance",
			"aws_opsworks_java_app_layer",
			"aws_opsworks_memcached_layer",
			"aws_opsworks_mysql_layer",
			"aws_opsworks_nodejs_app_layer",
			"aws_opsworks_permission",
			"aws_opsworks_php_app_layer",
			"aws_opsworks_rails_app_layer",
			"aws_opsworks_rds_db_instance",
			"aws_opsworks_static_web_layer",
		},
	})

	resource.AddTestSweepers("aws_opsworks_application", &resource.Sweeper{
		Name: "aws_opsworks_application",
		F:    sweepApplication,
	})

	resource.AddTestSweepers("aws_opsworks_instance", &resource.Sweeper{
		Name: "aws_opsworks_instance",
		F:    sweepInstance,
		Dependencies: []string{
			"aws_opsworks_custom_layer",
			"aws_opsworks_ecs_cluster_layer",
			"aws_opsworks_ganglia_layer",
			"aws_opsworks_haproxy_layer",
			"aws_opsworks_java_app_layer",
			"aws_opsworks_memcached_layer",
			"aws_opsworks_mysql_layer",
			"aws_opsworks_nodejs_app_layer",
			"aws_opsworks_php_app_layer",
			"aws_opsworks_rails_app_layer",
			"aws_opsworks_static_web_layer",
		},
	})

	resource.AddTestSweepers("aws_opsworks_permission", &resource.Sweeper{
		Name: "aws_opsworks_permission",
		F:    sweepPermission,
	})

	resource.AddTestSweepers("aws_opsworks_rds_db_instance", &resource.Sweeper{
		Name: "aws_opsworks_rds_db_instance",
		F:    sweepRDSDBInstance,
	})

	resource.AddTestSweepers("aws_opsworks_user_profile", &resource.Sweeper{
		Name: "aws_opsworks_user_profile",
		F:    sweepUserProfiles,
		Dependencies: []string{
			"aws_opsworks_permission",
		},
	})
}

func sweepApplication(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).OpsWorksConn
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.DescribeDBInstancesPages(&opsworks.DescribeDBInstancesInput{}, func(out *opsworks.DescribeDBInstancesOutput, lastPage bool) bool {
		for _, dbi := range out.DBInstances {
			r := ResourceInstance()
			d := r.Data(nil)
			d.SetId(aws.StringValue(dbi.DBInstanceIdentifier))
			d.Set("skip_final_snapshot", true)
			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
		return !lastPage
	})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping RDS DB Instance sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving DB instances: %s", err)
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

	err = conn.DescribeDBInstancesPages(&opsworks.DescribeDBInstancesInput{}, func(out *opsworks.DescribeDBInstancesOutput, lastPage bool) bool {
		for _, dbi := range out.DBInstances {
			r := ResourceInstance()
			d := r.Data(nil)
			d.SetId(aws.StringValue(dbi.DBInstanceIdentifier))
			d.Set("skip_final_snapshot", true)
			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
		return !lastPage
	})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping RDS DB Instance sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving DB instances: %s", err)
	}

	return sweep.SweepOrchestrator(sweepResources)
}

func sweepPermission(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).OpsWorksConn
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.DescribeDBInstancesPages(&opsworks.DescribeDBInstancesInput{}, func(out *opsworks.DescribeDBInstancesOutput, lastPage bool) bool {
		for _, dbi := range out.DBInstances {
			r := ResourceInstance()
			d := r.Data(nil)
			d.SetId(aws.StringValue(dbi.DBInstanceIdentifier))
			d.Set("skip_final_snapshot", true)
			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
		return !lastPage
	})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping RDS DB Instance sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving DB instances: %s", err)
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

	err = conn.DescribeDBInstancesPages(&opsworks.DescribeDBInstancesInput{}, func(out *opsworks.DescribeDBInstancesOutput, lastPage bool) bool {
		for _, dbi := range out.DBInstances {
			r := ResourceInstance()
			d := r.Data(nil)
			d.SetId(aws.StringValue(dbi.DBInstanceIdentifier))
			d.Set("skip_final_snapshot", true)
			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
		return !lastPage
	})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping RDS DB Instance sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving DB instances: %s", err)
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

	var sweeperErrs *multierror.Error

	for _, stack := range output.Stacks {
		if stack == nil {
			continue
		}

		r := ResourceStack()
		d := r.Data(nil)
		d.SetId(aws.StringValue(stack.StackId))

		if aws.StringValue(stack.Region) != region {
			d.Set("stack_endpoint", stack.Region)
		}

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
				if v, ok := layer.Attributes[opsworks.LayerAttributesKeysEcsClusterArn].(string); v != "" {
					r = ResourceEcsClusterLayer()
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

	return sweep.SweepOrchestrator(sweepResources)
}
