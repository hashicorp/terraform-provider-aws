// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apprunner

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/apprunner"
	"github.com/aws/aws-sdk-go-v2/service/apprunner/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
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

func WaitAutoScalingConfigurationActive(ctx context.Context, conn *apprunner.Client, arn string) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{},
		Target:  []string{AutoScalingConfigurationStatusActive},
		Refresh: StatusAutoScalingConfiguration(ctx, conn, arn),
		Timeout: AutoScalingConfigurationCreateTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func WaitAutoScalingConfigurationInactive(ctx context.Context, conn *apprunner.Client, arn string) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{AutoScalingConfigurationStatusActive},
		Target:  []string{AutoScalingConfigurationStatusInactive},
		Refresh: StatusAutoScalingConfiguration(ctx, conn, arn),
		Timeout: AutoScalingConfigurationDeleteTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func WaitConnectionDeleted(ctx context.Context, conn *apprunner.Client, name string) error {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.ConnectionStatusPendingHandshake, types.ConnectionStatusAvailable, types.ConnectionStatusDeleted),
		Target:  []string{},
		Refresh: StatusConnection(ctx, conn, name),
		Timeout: ConnectionDeleteTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func WaitCustomDomainAssociationCreated(ctx context.Context, conn *apprunner.Client, domainName, serviceArn string) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{CustomDomainAssociationStatusCreating},
		Target:  []string{CustomDomainAssociationStatusPendingCertificateDNSValidation, CustomDomainAssociationStatusBindingCertificate},
		Refresh: StatusCustomDomain(ctx, conn, domainName, serviceArn),
		Timeout: CustomDomainAssociationCreateTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func WaitCustomDomainAssociationDeleted(ctx context.Context, conn *apprunner.Client, domainName, serviceArn string) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{CustomDomainAssociationStatusActive, CustomDomainAssociationStatusDeleting},
		Target:  []string{},
		Refresh: StatusCustomDomain(ctx, conn, domainName, serviceArn),
		Timeout: CustomDomainAssociationDeleteTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func WaitObservabilityConfigurationActive(ctx context.Context, conn *apprunner.Client, observabilityConfigurationArn string) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{},
		Target:  []string{ObservabilityConfigurationStatusActive},
		Refresh: StatusObservabilityConfiguration(ctx, conn, observabilityConfigurationArn),
		Timeout: ObservabilityConfigurationCreateTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func WaitObservabilityConfigurationInactive(ctx context.Context, conn *apprunner.Client, observabilityConfigurationArn string) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ObservabilityConfigurationStatusActive},
		Target:  []string{ObservabilityConfigurationStatusInactive},
		Refresh: StatusObservabilityConfiguration(ctx, conn, observabilityConfigurationArn),
		Timeout: ObservabilityConfigurationDeleteTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func WaitVPCIngressConnectionActive(ctx context.Context, conn *apprunner.Client, vpcIngressConnectionArn string) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{},
		Target:  []string{VPCIngressConnectionstatusActive},
		Refresh: StatusVPCIngressConnection(ctx, conn, vpcIngressConnectionArn),
		Timeout: VPCIngressConnectionCreateTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func WaitVPCIngressConnectionDeleted(ctx context.Context, conn *apprunner.Client, vpcIngressConnectionArn string) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{VPCIngressConnectionstatusActive, VPCIngressConnectionstatusPendingDeletion},
		Target:  []string{VPCIngressConnectionstatusDeleted},
		Refresh: StatusVPCIngressConnection(ctx, conn, vpcIngressConnectionArn),
		Timeout: VPCIngressConnectionDeleteTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func WaitServiceCreated(ctx context.Context, conn *apprunner.Client, serviceArn string) error {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.ServiceStatusOperationInProgress),
		Target:  enum.Slice(types.ServiceStatusRunning),
		Refresh: StatusService(ctx, conn, serviceArn),
		Timeout: ServiceCreateTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func WaitServiceUpdated(ctx context.Context, conn *apprunner.Client, serviceArn string) error {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.ServiceStatusOperationInProgress),
		Target:  enum.Slice(types.ServiceStatusRunning),
		Refresh: StatusService(ctx, conn, serviceArn),
		Timeout: ServiceUpdateTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func WaitServiceDeleted(ctx context.Context, conn *apprunner.Client, serviceArn string) error {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.ServiceStatusRunning, types.ServiceStatusOperationInProgress),
		Target:  enum.Slice(types.ServiceStatusDeleted),
		Refresh: StatusService(ctx, conn, serviceArn),
		Timeout: ServiceDeleteTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}
