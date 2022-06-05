//go:build sweep
// +build sweep

package globalaccelerator

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/globalaccelerator"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_globalaccelerator_accelerator", &resource.Sweeper{
		Name: "aws_globalaccelerator_accelerator",
		F:    sweepAccelerators,
	})
}

func sweepAccelerators(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).GlobalAcceleratorConn

	input := &globalaccelerator.ListAcceleratorsInput{}
	var sweeperErrs *multierror.Error

	for {
		output, err := conn.ListAccelerators(input)

		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Global Accelerator Accelerator sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("Error retrieving Global Accelerator Accelerators: %s", err)
		}

		for _, accelerator := range output.Accelerators {
			arn := aws.StringValue(accelerator.AcceleratorArn)

			errs := sweepListeners(client, accelerator.AcceleratorArn)
			if errs != nil {
				sweeperErrs = multierror.Append(sweeperErrs, errs)
			}

			r := ResourceAccelerator()
			d := r.Data(nil)
			d.SetId(arn)
			err = r.Delete(d, client)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting Global Accelerator Accelerator (%s): %s", arn, err)
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

func sweepEndpointGroups(client interface{}, listenerArn *string) *multierror.Error {
	conn := client.(*conns.AWSClient).GlobalAcceleratorConn
	var sweeperErrs *multierror.Error

	log.Printf("[INFO] deleting Endpoint Groups for Listener %s", *listenerArn)
	input := &globalaccelerator.ListEndpointGroupsInput{
		ListenerArn: listenerArn,
	}
	output, err := conn.ListEndpointGroups(input)
	if err != nil {
		sweeperErr := fmt.Errorf("error listing Global Accelerator Endpoint Groups for Listener (%s): %s", *listenerArn, err)
		log.Printf("[ERROR] %s", sweeperErr)
		sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
	}

	for _, endpoint := range output.EndpointGroups {
		arn := aws.StringValue(endpoint.EndpointGroupArn)

		r := ResourceEndpointGroup()
		d := r.Data(nil)
		d.SetId(arn)
		err = r.Delete(d, client)

		if err != nil {
			sweeperErr := fmt.Errorf("error deleting Global Accelerator endpoint group (%s): %s", arn, err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}
	}

	return sweeperErrs
}

func sweepListeners(client interface{}, acceleratorArn *string) *multierror.Error {
	conn := client.(*conns.AWSClient).GlobalAcceleratorConn
	var sweeperErrs *multierror.Error

	log.Printf("[INFO] deleting Listeners for Accelerator %s", *acceleratorArn)
	listenersInput := &globalaccelerator.ListListenersInput{
		AcceleratorArn: acceleratorArn,
	}
	listenersOutput, err := conn.ListListeners(listenersInput)
	if err != nil {
		sweeperErr := fmt.Errorf("error listing Global Accelerator Listeners for Accelerator (%s): %s", *acceleratorArn, err)
		log.Printf("[ERROR] %s", sweeperErr)
		sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
	}

	for _, listener := range listenersOutput.Listeners {
		errs := sweepEndpointGroups(client, listener.ListenerArn)
		if errs != nil {
			sweeperErrs = multierror.Append(sweeperErrs, errs)
		}

		arn := aws.StringValue(listener.ListenerArn)

		r := ResourceListener()
		d := r.Data(nil)
		d.SetId(arn)
		err = r.Delete(d, client)

		if err != nil {
			sweeperErr := fmt.Errorf("error deleting Global Accelerator listener (%s): %s", arn, err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}
	}

	return sweeperErrs
}
