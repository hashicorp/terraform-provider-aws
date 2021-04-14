package waiter

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

const (
	// Maximum amount of time to wait for a Capacity Provider to return INACTIVE
	CapacityProviderInactiveTimeout = 20 * time.Minute

	ServiceCreateTimeout      = 2 * time.Minute
	ServiceInactiveTimeout    = 10 * time.Minute
	ServiceInactiveTimeoutMin = 1 * time.Second
	ServiceDescribeTimeout    = 2 * time.Minute
	ServiceUpdateTimeout      = 2 * time.Minute
	ServiceStableTimeout      = 10 * time.Minute
)

// CapacityProviderInactive waits for a Capacity Provider to return INACTIVE
func CapacityProviderInactive(conn *ecs.ECS, capacityProvider string) (*ecs.CapacityProvider, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ecs.CapacityProviderStatusActive},
		Target:  []string{ecs.CapacityProviderStatusInactive},
		Refresh: CapacityProviderStatus(conn, capacityProvider),
		Timeout: CapacityProviderInactiveTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*ecs.CapacityProvider); ok {
		return v, err
	}

	return nil, err
}

// ServiceStable waits for a Service to be stable
func ServiceStable(conn *ecs.ECS, id, cluster string, timeout *int64) error {
	input := &ecs.DescribeServicesInput{
		Services: aws.StringSlice([]string{id}),
	}

	if cluster != "" {
		input.Cluster = aws.String(cluster)
	}

	stableTimeout := ServiceStableTimeout
	if timeout != nil {
		stableTimeout = time.Duration(*timeout) * time.Minute
	}

	log.Printf("[DEBUG] Waiting until services are stable. ECS Service (%s): %s", id, input)

	err := resource.Retry(stableTimeout, func() *resource.RetryError {
		if err := conn.WaitUntilServicesStable(input); err != nil {
			if tfawserr.ErrMessageContains(err, request.WaiterResourceNotReadyErrorCode, "error waiting for service") {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		err = conn.WaitUntilServicesStable(input)
	}

	if err != nil {
		return fmt.Errorf("error waiting for service to be stable. ECS Service (%s): %w", id, err)
	}

	return nil
}

func ServiceInactive(conn *ecs.ECS, id, cluster string) error {
	input := &ecs.DescribeServicesInput{
		Services: aws.StringSlice([]string{id}),
	}

	if cluster != "" {
		input.Cluster = aws.String(cluster)
	}

	if err := conn.WaitUntilServicesInactive(input); err != nil {
		return err
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{ServiceStatusActive, ServiceStatusDraining},
		Target:     []string{ServiceStatusInactive, ServiceStatusNone},
		Refresh:    ServiceStatus(conn, id, cluster),
		Timeout:    ServiceInactiveTimeout,
		MinTimeout: ServiceInactiveTimeoutMin,
	}

	_, err := stateConf.WaitForState()

	if err != nil {
		return err
	}

	return nil
}

func ServiceDescribeReady(conn *ecs.ECS, id, cluster string) (*ecs.DescribeServicesOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ServiceStatusInactive, ServiceStatusDraining, ServiceStatusNone},
		Target:  []string{ServiceStatusActive},
		Refresh: ServiceStatus(conn, id, cluster),
		Timeout: ServiceDescribeTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*ecs.DescribeServicesOutput); ok {
		return v, err
	}

	return nil, err
}
