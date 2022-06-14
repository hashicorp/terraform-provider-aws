package lightsail

import (
	"context"
	"errors"
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

func waitContainerServiceDeploymentVersionActive(ctx context.Context, conn *lightsail.Lightsail, serviceName string, version int, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{lightsail.ContainerServiceDeploymentStateActivating},
		Target:     []string{lightsail.ContainerServiceDeploymentStateActive},
		Refresh:    statusContainerServiceDeploymentVersion(ctx, conn, serviceName, version),
		Timeout:    timeout,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*lightsail.ContainerServiceDeployment); ok {
		if aws.StringValue(output.State) == lightsail.ContainerServiceDeploymentStateFailed {
			tfresource.SetLastError(err, errors.New("The deployment failed. Use the GetContainerLog action to view the log events for the containers in the deployment to try to determine the reason for the failure."))
		}

		return err
	}

	return err
}
