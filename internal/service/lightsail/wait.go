package lightsail

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func waitContainerServiceCreated(ctx context.Context, conn *lightsail.Lightsail, serviceName string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{lightsail.ContainerServiceStatePending},
		Target:     []string{lightsail.ContainerServiceStateReady},
		Refresh:    statusContainerService(ctx, conn, serviceName),
		Timeout:    timeout,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*lightsail.ContainerService); ok {
		if detail := output.StateDetail; detail != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(detail.Code), aws.StringValue(detail.Message)))
		}

		return err
	}

	return err
}

func waitContainerServiceDisabled(ctx context.Context, conn *lightsail.Lightsail, serviceName string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{lightsail.ContainerServiceStateUpdating},
		Target:     []string{lightsail.ContainerServiceStateDisabled},
		Refresh:    statusContainerService(ctx, conn, serviceName),
		Timeout:    timeout,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*lightsail.ContainerService); ok {
		if detail := output.StateDetail; detail != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(detail.Code), aws.StringValue(detail.Message)))
		}

		return err
	}

	return err
}

func waitContainerServiceUpdated(ctx context.Context, conn *lightsail.Lightsail, serviceName string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{lightsail.ContainerServiceStateUpdating},
		Target:     []string{lightsail.ContainerServiceStateReady, lightsail.ContainerServiceStateRunning},
		Refresh:    statusContainerService(ctx, conn, serviceName),
		Timeout:    timeout,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*lightsail.ContainerService); ok {
		if detail := output.StateDetail; detail != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(detail.Code), aws.StringValue(detail.Message)))
		}

		return err
	}

	return err
}

func waitContainerServiceDeleted(ctx context.Context, conn *lightsail.Lightsail, serviceName string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{lightsail.ContainerServiceStateDeleting},
		Target:     []string{},
		Refresh:    statusContainerService(ctx, conn, serviceName),
		Timeout:    timeout,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*lightsail.ContainerService); ok {
		if detail := output.StateDetail; detail != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(detail.Code), aws.StringValue(detail.Message)))
		}

		return err
	}

	return err
}
