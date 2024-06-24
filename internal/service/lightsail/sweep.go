// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lightsail

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_lightsail_container_service", &resource.Sweeper{
		Name: "aws_lightsail_container_service",
		F:    sweepContainerServices,
	})

	resource.AddTestSweepers("aws_lightsail_database", &resource.Sweeper{
		Name: "aws_lightsail_database",
		F:    sweepDatabases,
	})

	resource.AddTestSweepers("aws_lightsail_disk", &resource.Sweeper{
		Name: "aws_lightsail_disk",
		F:    sweepDisks,
	})

	resource.AddTestSweepers("aws_lightsail_distribution", &resource.Sweeper{
		Name: "aws_lightsail_distribution",
		F:    sweepDistributions,
	})

	resource.AddTestSweepers("aws_lightsail_domain", &resource.Sweeper{
		Name: "aws_lightsail_domain",
		F:    sweepDomains,
	})

	resource.AddTestSweepers("aws_lightsail_instance", &resource.Sweeper{
		Name: "aws_lightsail_instance",
		F:    sweepInstances,
	})

	resource.AddTestSweepers("aws_lightsail_lb", &resource.Sweeper{
		Name: "aws_lightsail_lb",
		F:    sweepLoadBalancers,
	})

	resource.AddTestSweepers("aws_lightsail_static_ip", &resource.Sweeper{
		Name: "aws_lightsail_static_ip",
		F:    sweepStaticIPs,
	})
}

func sweepContainerServices(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}
	conn := client.LightsailClient(ctx)

	input := &lightsail.GetContainerServicesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	output, err := conn.GetContainerServices(ctx, input)

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Lightsail Container Service sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error retrieving Lightsail Container Services: %s", err)
	}

	for _, service := range output.ContainerServices {
		r := ResourceContainerService()
		d := r.Data(nil)
		d.SetId(aws.ToString(service.ContainerServiceName))

		sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping Lightsail Container Services for %s: %w", region, err)
	}

	return nil
}

func sweepDatabases(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}
	conn := client.LightsailClient(ctx)

	input := &lightsail.GetRelationalDatabasesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = getRelationalDatabasesPages(ctx, conn, input, func(page *lightsail.GetRelationalDatabasesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.RelationalDatabases {
			r := ResourceDatabase()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Name))
			d.Set("skip_final_snapshot", true)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Lightsail Databases sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("retrieving Lightsail Databases (%s): %w", region, err)
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping Lightsail Databases for %s: %w", region, err)
	}

	return nil
}

func sweepDisks(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}
	conn := client.LightsailClient(ctx)

	input := &lightsail.GetDisksInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = getDisksPages(ctx, conn, input, func(page *lightsail.GetDisksOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Disks {
			r := ResourceDisk()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Name))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Lightsail Disks sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("retrieving Lightsail Disks (%s): %w", region, err)
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping Lightsail Disks for %s: %w", region, err)
	}

	return nil
}

func sweepDistributions(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}
	conn := client.LightsailClient(ctx)

	input := &lightsail.GetDistributionsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = getDistributionsPages(ctx, conn, input, func(page *lightsail.GetDistributionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Distributions {
			r := ResourceDistribution()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Name))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Lightsail Distributions sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("retrieving Lightsail Distributions (%s): %w", region, err)
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping Lightsail Distributions for %s: %w", region, err)
	}

	return nil
}

func sweepDomains(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}
	conn := client.LightsailClient(ctx)

	input := &lightsail.GetDomainsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = getDomainsPages(ctx, conn, input, func(page *lightsail.GetDomainsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Domains {
			r := ResourceDomain()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Name))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Lightsail Domain sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("retrieving Lightsail Domain (%s): %w", region, err)
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping Lightsail Domain for %s: %w", region, err)
	}

	return nil
}

func sweepInstances(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}
	conn := client.LightsailClient(ctx)

	input := &lightsail.GetInstancesInput{}
	var sweeperErrs *multierror.Error

	for {
		output, err := conn.GetInstances(ctx, input)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Lightsail Instance sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("Error retrieving Lightsail Instances: %s", err)
		}

		for _, instance := range output.Instances {
			name := aws.ToString(instance.Name)
			input := &lightsail.DeleteInstanceInput{
				InstanceName: instance.Name,
			}

			log.Printf("[INFO] Deleting Lightsail Instance: %s", name)
			_, err := conn.DeleteInstance(ctx, input)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting Lightsail Instance (%s): %s", name, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			}
		}

		if aws.ToString(output.NextPageToken) == "" {
			break
		}

		input.PageToken = output.NextPageToken
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepLoadBalancers(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}
	conn := client.LightsailClient(ctx)

	input := &lightsail.GetLoadBalancersInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = getLoadBalancersPages(ctx, conn, input, func(page *lightsail.GetLoadBalancersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.LoadBalancers {
			r := ResourceLoadBalancer()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Name))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Lightsail Load Balanders sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("retrieving Lightsail Load Balanders (%s): %w", region, err)
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping Lightsail Load Balanders for %s: %w", region, err)
	}

	return nil
}

func sweepStaticIPs(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}
	conn := client.LightsailClient(ctx)

	input := &lightsail.GetStaticIpsInput{}

	for {
		output, err := conn.GetStaticIps(ctx, input)
		if err != nil {
			if awsv2.SkipSweepError(err) {
				log.Printf("[WARN] Skipping Lightsail Static IP sweep for %s: %s", region, err)
				return nil
			}
			return fmt.Errorf("Error retrieving Lightsail Static IPs: %s", err)
		}

		if len(output.StaticIps) == 0 {
			log.Print("[DEBUG] No Lightsail Static IPs to sweep")
			return nil
		}

		for _, staticIp := range output.StaticIps {
			name := aws.ToString(staticIp.Name)

			log.Printf("[INFO] Deleting Lightsail Static IP %s", name)
			_, err := conn.ReleaseStaticIp(ctx, &lightsail.ReleaseStaticIpInput{
				StaticIpName: aws.String(name),
			})
			if err != nil {
				return fmt.Errorf("Error deleting Lightsail Static IP %s: %s", name, err)
			}
		}

		if output.NextPageToken == nil {
			break
		}
		input.PageToken = output.NextPageToken
	}

	return nil
}
