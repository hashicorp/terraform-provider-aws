package apprunner

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/apprunner"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	AutoScalingConfigurationCreateTimeout = 2 * time.Minute
	AutoScalingConfigurationDeleteTimeout = 2 * time.Minute

	ConnectionDeleteTimeout = 5 * time.Minute

	CustomDomainAssociationCreateTimeout = 5 * time.Minute
	CustomDomainAssociationDeleteTimeout = 5 * time.Minute

	vpcConnectorCreateTimeout = 2 * time.Minute
	vpcConnectorDeleteTimeout = 2 * time.Minute

	ServiceCreateTimeout = 20 * time.Minute
	ServiceDeleteTimeout = 20 * time.Minute
	ServiceUpdateTimeout = 20 * time.Minute
)

func WaitAutoScalingConfigurationActive(ctx context.Context, conn *apprunner.AppRunner, arn string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{},
		Target:  []string{AutoScalingConfigurationStatusActive},
		Refresh: StatusAutoScalingConfiguration(ctx, conn, arn),
		Timeout: AutoScalingConfigurationCreateTimeout,
	}

	_, err := stateConf.WaitForState()

	return err
}

func WaitAutoScalingConfigurationInactive(ctx context.Context, conn *apprunner.AppRunner, arn string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{AutoScalingConfigurationStatusActive},
		Target:  []string{AutoScalingConfigurationStatusInactive},
		Refresh: StatusAutoScalingConfiguration(ctx, conn, arn),
		Timeout: AutoScalingConfigurationDeleteTimeout,
	}

	_, err := stateConf.WaitForState()

	return err
}

func WaitConnectionDeleted(ctx context.Context, conn *apprunner.AppRunner, name string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{apprunner.ConnectionStatusPendingHandshake, apprunner.ConnectionStatusAvailable, apprunner.ConnectionStatusDeleted},
		Target:  []string{},
		Refresh: StatusConnection(ctx, conn, name),
		Timeout: ConnectionDeleteTimeout,
	}

	_, err := stateConf.WaitForState()

	return err
}

func WaitCustomDomainAssociationCreated(ctx context.Context, conn *apprunner.AppRunner, domainName, serviceArn string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{CustomDomainAssociationStatusCreating},
		Target:  []string{CustomDomainAssociationStatusPendingCertificateDNSValidation, CustomDomainAssociationStatusBindingCertificate},
		Refresh: StatusCustomDomain(ctx, conn, domainName, serviceArn),
		Timeout: CustomDomainAssociationCreateTimeout,
	}

	_, err := stateConf.WaitForState()

	return err
}

func WaitCustomDomainAssociationDeleted(ctx context.Context, conn *apprunner.AppRunner, domainName, serviceArn string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{CustomDomainAssociationStatusActive, CustomDomainAssociationStatusDeleting},
		Target:  []string{},
		Refresh: StatusCustomDomain(ctx, conn, domainName, serviceArn),
		Timeout: CustomDomainAssociationDeleteTimeout,
	}

	_, err := stateConf.WaitForState()

	return err
}

func WaitServiceCreated(ctx context.Context, conn *apprunner.AppRunner, serviceArn string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{apprunner.ServiceStatusOperationInProgress},
		Target:  []string{apprunner.ServiceStatusRunning},
		Refresh: StatusService(ctx, conn, serviceArn),
		Timeout: ServiceCreateTimeout,
	}

	_, err := stateConf.WaitForState()

	return err
}

func WaitServiceUpdated(ctx context.Context, conn *apprunner.AppRunner, serviceArn string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{apprunner.ServiceStatusOperationInProgress},
		Target:  []string{apprunner.ServiceStatusRunning},
		Refresh: StatusService(ctx, conn, serviceArn),
		Timeout: ServiceUpdateTimeout,
	}

	_, err := stateConf.WaitForState()

	return err
}

func WaitServiceDeleted(ctx context.Context, conn *apprunner.AppRunner, serviceArn string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{apprunner.ServiceStatusRunning, apprunner.ServiceStatusOperationInProgress},
		Target:  []string{apprunner.ServiceStatusDeleted},
		Refresh: StatusService(ctx, conn, serviceArn),
		Timeout: ServiceDeleteTimeout,
	}

	_, err := stateConf.WaitForState()

	return err
}

func waitVPCConnectorActive(ctx context.Context, conn *apprunner.AppRunner, arn string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{},
		Target:  []string{apprunner.VpcConnectorStatusActive},
		Refresh: StatusVPCConnector(ctx, conn, arn),
		Timeout: vpcConnectorCreateTimeout,
	}

	_, err := stateConf.WaitForState()

	return err
}

func waitVPCConnectorInactive(ctx context.Context, conn *apprunner.AppRunner, arn string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{apprunner.VpcConnectorStatusActive},
		Target:  []string{apprunner.VpcConnectorStatusInactive},
		Refresh: StatusVPCConnector(ctx, conn, arn),
		Timeout: vpcConnectorDeleteTimeout,
	}

	_, err := stateConf.WaitForState()

	return err
}
