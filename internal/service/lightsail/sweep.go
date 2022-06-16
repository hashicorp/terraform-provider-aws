//go:build sweep
// +build sweep

package lightsail

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_lightsail_container_service", &resource.Sweeper{
		Name: "aws_lightsail_container_service",
		F:    sweepContainerServices,
	})

	resource.AddTestSweepers("aws_lightsail_instance", &resource.Sweeper{
		Name: "aws_lightsail_instance",
		F:    sweepInstances,
	})

	resource.AddTestSweepers("aws_lightsail_static_ip", &resource.Sweeper{
		Name: "aws_lightsail_static_ip",
		F:    sweepStaticIPs,
	})
}

func sweepContainerServices(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).LightsailConn

	input := &lightsail.GetContainerServicesInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]*sweep.SweepResource, 0)

	output, err := conn.GetContainerServicesWithContext(context.TODO(), input)

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Lightsail Container Service sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error retrieving Lightsail Container Services: %s", err)
	}

	for _, service := range output.ContainerServices {
		if service == nil {
			continue
		}

		r := ResourceContainerService()
		d := r.Data(nil)
		d.SetId(aws.StringValue(service.ContainerServiceName))

		sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
	}

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Lightsail Container Services sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Lightsail Container Services  for %s: %w", region, err))
	}

	if err := sweep.SweepOrchestrator(sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Lightsail Container Services for %s: %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepInstances(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).LightsailConn

	input := &lightsail.GetInstancesInput{}
	var sweeperErrs *multierror.Error

	for {
		output, err := conn.GetInstances(input)

		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Lightsail Instance sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("Error retrieving Lightsail Instances: %s", err)
		}

		for _, instance := range output.Instances {
			name := aws.StringValue(instance.Name)
			input := &lightsail.DeleteInstanceInput{
				InstanceName: instance.Name,
			}

			log.Printf("[INFO] Deleting Lightsail Instance: %s", name)
			_, err := conn.DeleteInstance(input)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting Lightsail Instance (%s): %s", name, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			}
		}

		if aws.StringValue(output.NextPageToken) == "" {
			break
		}

		input.PageToken = output.NextPageToken
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepStaticIPs(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).LightsailConn

	input := &lightsail.GetStaticIpsInput{}

	for {
		output, err := conn.GetStaticIps(input)
		if err != nil {
			if sweep.SkipSweepError(err) {
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
			name := aws.StringValue(staticIp.Name)

			log.Printf("[INFO] Deleting Lightsail Static IP %s", name)
			_, err := conn.ReleaseStaticIp(&lightsail.ReleaseStaticIpInput{
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
