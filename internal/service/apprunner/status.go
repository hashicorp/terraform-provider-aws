// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apprunner

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apprunner"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

const (
	AutoScalingConfigurationStatusActive   = "active"
	AutoScalingConfigurationStatusInactive = "inactive"

	CustomDomainAssociationStatusActive                          = "active"
	CustomDomainAssociationStatusCreating                        = "creating"
	CustomDomainAssociationStatusDeleting                        = "deleting"
	CustomDomainAssociationStatusPendingCertificateDNSValidation = "pending_certificate_dns_validation"
	CustomDomainAssociationStatusBindingCertificate              = "binding_certificate"

	ObservabilityConfigurationStatusActive   = "ACTIVE"
	ObservabilityConfigurationStatusInactive = "INACTIVE"

	VPCIngressConnectionstatusActive          = "AVAILABLE"
	VPCIngressConnectionstatusPendingDeletion = "PENDING_DELETION"
	VPCIngressConnectionstatusDeleted         = "DELETED"
)

func StatusAutoScalingConfiguration(ctx context.Context, conn *apprunner.Client, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &apprunner.DescribeAutoScalingConfigurationInput{
			AutoScalingConfigurationArn: aws.String(arn),
		}

		output, err := conn.DescribeAutoScalingConfiguration(ctx, input)

		if err != nil {
			return nil, "", err
		}

		if output == nil || output.AutoScalingConfiguration == nil {
			return nil, "", nil
		}

		return output.AutoScalingConfiguration, string(output.AutoScalingConfiguration.Status), nil
	}
}

func StatusObservabilityConfiguration(ctx context.Context, conn *apprunner.Client, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &apprunner.DescribeObservabilityConfigurationInput{
			ObservabilityConfigurationArn: aws.String(arn),
		}

		output, err := conn.DescribeObservabilityConfiguration(ctx, input)

		if err != nil {
			return nil, "", err
		}

		if output == nil || output.ObservabilityConfiguration == nil {
			return nil, "", nil
		}

		return output.ObservabilityConfiguration, string(output.ObservabilityConfiguration.Status), nil
	}
}

func StatusVPCIngressConnection(ctx context.Context, conn *apprunner.Client, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &apprunner.DescribeVpcIngressConnectionInput{
			VpcIngressConnectionArn: aws.String(arn),
		}

		output, err := conn.DescribeVpcIngressConnection(ctx, input)

		if err != nil {
			return nil, "", err
		}

		if output == nil || output.VpcIngressConnection == nil {
			return nil, "", nil
		}

		return output.VpcIngressConnection, string(output.VpcIngressConnection.Status), nil
	}
}

func StatusConnection(ctx context.Context, conn *apprunner.Client, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		c, err := FindConnectionsummaryByName(ctx, conn, name)

		if err != nil {
			return nil, "", err
		}

		if c == nil {
			return nil, "", nil
		}

		return c, string(c.Status), nil
	}
}

func StatusCustomDomain(ctx context.Context, conn *apprunner.Client, domainName, serviceArn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		customDomain, err := FindCustomDomain(ctx, conn, domainName, serviceArn)

		if err != nil {
			return nil, "", err
		}

		if customDomain == nil {
			return nil, "", nil
		}

		return customDomain, string(customDomain.Status), nil
	}
}

func StatusService(ctx context.Context, conn *apprunner.Client, serviceArn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &apprunner.DescribeServiceInput{
			ServiceArn: aws.String(serviceArn),
		}

		output, err := conn.DescribeService(ctx, input)

		if err != nil {
			return nil, "", err
		}

		if output == nil || output.Service == nil {
			return nil, "", nil
		}

		return output.Service, string(output.Service.Status), nil
	}
}
