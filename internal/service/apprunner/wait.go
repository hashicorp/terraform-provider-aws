// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apprunner

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/apprunner"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

const (
	AutoScalingConfigurationCreateTimeout = 2 * time.Minute
	AutoScalingConfigurationDeleteTimeout = 2 * time.Minute

	ConnectionDeleteTimeout = 5 * time.Minute

	CustomDomainAssociationCreateTimeout = 5 * time.Minute
	CustomDomainAssociationDeleteTimeout = 5 * time.Minute

	ServiceCreateTimeout = 20 * time.Minute
	ServiceDeleteTimeout = 20 * time.Minute
	ServiceUpdateTimeout = 20 * time.Minute

	ObservabilityConfigurationCreateTimeout = 2 * time.Minute
	ObservabilityConfigurationDeleteTimeout = 2 * time.Minute

	VPCIngressConnectionCreateTimeout = 2 * time.Minute
	VPCIngressConnectionDeleteTimeout = 2 * time.Minute
)

func WaitAutoScalingConfigurationActive(ctx context.Context, conn *apprunner.AppRunner, arn string) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{},
		Target:  []string{AutoScalingConfigurationStatusActive},
		Refresh: StatusAutoScalingConfiguration(ctx, conn, arn),
		Timeout: AutoScalingConfigurationCreateTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func WaitAutoScalingConfigurationInactive(ctx context.Context, conn *apprunner.AppRunner, arn string) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{AutoScalingConfigurationStatusActive},
		Target:  []string{AutoScalingConfigurationStatusInactive},
		Refresh: StatusAutoScalingConfiguration(ctx, conn, arn),
		Timeout: AutoScalingConfigurationDeleteTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func WaitConnectionDeleted(ctx context.Context, conn *apprunner.AppRunner, name string) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{apprunner.ConnectionStatusPendingHandshake, apprunner.ConnectionStatusAvailable, apprunner.ConnectionStatusDeleted},
		Target:  []string{},
		Refresh: StatusConnection(ctx, conn, name),
		Timeout: ConnectionDeleteTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func WaitCustomDomainAssociationCreated(ctx context.Context, conn *apprunner.AppRunner, domainName, serviceArn string) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{CustomDomainAssociationStatusCreating},
		Target:  []string{CustomDomainAssociationStatusPendingCertificateDNSValidation, CustomDomainAssociationStatusBindingCertificate},
		Refresh: StatusCustomDomain(ctx, conn, domainName, serviceArn),
		Timeout: CustomDomainAssociationCreateTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func WaitCustomDomainAssociationDeleted(ctx context.Context, conn *apprunner.AppRunner, domainName, serviceArn string) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{CustomDomainAssociationStatusActive, CustomDomainAssociationStatusDeleting},
		Target:  []string{},
		Refresh: StatusCustomDomain(ctx, conn, domainName, serviceArn),
		Timeout: CustomDomainAssociationDeleteTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func WaitObservabilityConfigurationActive(ctx context.Context, conn *apprunner.AppRunner, observabilityConfigurationArn string) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{},
		Target:  []string{ObservabilityConfigurationStatusActive},
		Refresh: StatusObservabilityConfiguration(ctx, conn, observabilityConfigurationArn),
		Timeout: ObservabilityConfigurationCreateTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func WaitObservabilityConfigurationInactive(ctx context.Context, conn *apprunner.AppRunner, observabilityConfigurationArn string) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ObservabilityConfigurationStatusActive},
		Target:  []string{ObservabilityConfigurationStatusInactive},
		Refresh: StatusObservabilityConfiguration(ctx, conn, observabilityConfigurationArn),
		Timeout: ObservabilityConfigurationDeleteTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func WaitVPCIngressConnectionActive(ctx context.Context, conn *apprunner.AppRunner, vpcIngressConnectionArn string) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{},
		Target:  []string{VPCIngressConnectionStatusActive},
		Refresh: StatusVPCIngressConnection(ctx, conn, vpcIngressConnectionArn),
		Timeout: VPCIngressConnectionCreateTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func WaitVPCIngressConnectionDeleted(ctx context.Context, conn *apprunner.AppRunner, vpcIngressConnectionArn string) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{VPCIngressConnectionStatusActive, VPCIngressConnectionStatusPendingDeletion},
		Target:  []string{VPCIngressConnectionStatusDeleted},
		Refresh: StatusVPCIngressConnection(ctx, conn, vpcIngressConnectionArn),
		Timeout: VPCIngressConnectionDeleteTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func WaitServiceCreated(ctx context.Context, conn *apprunner.AppRunner, serviceArn string) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{apprunner.ServiceStatusOperationInProgress},
		Target:  []string{apprunner.ServiceStatusRunning},
		Refresh: StatusService(ctx, conn, serviceArn),
		Timeout: ServiceCreateTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func WaitServiceUpdated(ctx context.Context, conn *apprunner.AppRunner, serviceArn string) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{apprunner.ServiceStatusOperationInProgress},
		Target:  []string{apprunner.ServiceStatusRunning},
		Refresh: StatusService(ctx, conn, serviceArn),
		Timeout: ServiceUpdateTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func WaitServiceDeleted(ctx context.Context, conn *apprunner.AppRunner, serviceArn string) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{apprunner.ServiceStatusRunning, apprunner.ServiceStatusOperationInProgress},
		Target:  []string{apprunner.ServiceStatusDeleted},
		Refresh: StatusService(ctx, conn, serviceArn),
		Timeout: ServiceDeleteTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}
