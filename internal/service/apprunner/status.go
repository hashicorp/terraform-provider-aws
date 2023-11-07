// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apprunner

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apprunner"
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

	VPCIngressConnectionStatusActive          = "AVAILABLE"
	VPCIngressConnectionStatusPendingDeletion = "PENDING_DELETION"
	VPCIngressConnectionStatusDeleted         = "DELETED"
)

func StatusAutoScalingConfiguration(ctx context.Context, conn *apprunner.AppRunner, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &apprunner.DescribeAutoScalingConfigurationInput{
			AutoScalingConfigurationArn: aws.String(arn),
		}

		output, err := conn.DescribeAutoScalingConfigurationWithContext(ctx, input)

		if err != nil {
			return nil, "", err
		}

		if output == nil || output.AutoScalingConfiguration == nil {
			return nil, "", nil
		}

		return output.AutoScalingConfiguration, aws.StringValue(output.AutoScalingConfiguration.Status), nil
	}
}

func StatusObservabilityConfiguration(ctx context.Context, conn *apprunner.AppRunner, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &apprunner.DescribeObservabilityConfigurationInput{
			ObservabilityConfigurationArn: aws.String(arn),
		}

		output, err := conn.DescribeObservabilityConfigurationWithContext(ctx, input)

		if err != nil {
			return nil, "", err
		}

		if output == nil || output.ObservabilityConfiguration == nil {
			return nil, "", nil
		}

		return output.ObservabilityConfiguration, aws.StringValue(output.ObservabilityConfiguration.Status), nil
	}
}

func StatusVPCIngressConnection(ctx context.Context, conn *apprunner.AppRunner, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &apprunner.DescribeVpcIngressConnectionInput{
			VpcIngressConnectionArn: aws.String(arn),
		}

		output, err := conn.DescribeVpcIngressConnectionWithContext(ctx, input)

		if err != nil {
			return nil, "", err
		}

		if output == nil || output.VpcIngressConnection == nil {
			return nil, "", nil
		}

		return output.VpcIngressConnection, aws.StringValue(output.VpcIngressConnection.Status), nil
	}
}

func StatusConnection(ctx context.Context, conn *apprunner.AppRunner, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		c, err := FindConnectionSummaryByName(ctx, conn, name)

		if err != nil {
			return nil, "", err
		}

		if c == nil {
			return nil, "", nil
		}

		return c, aws.StringValue(c.Status), nil
	}
}

func StatusCustomDomain(ctx context.Context, conn *apprunner.AppRunner, domainName, serviceArn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		customDomain, err := FindCustomDomain(ctx, conn, domainName, serviceArn)

		if err != nil {
			return nil, "", err
		}

		if customDomain == nil {
			return nil, "", nil
		}

		return customDomain, aws.StringValue(customDomain.Status), nil
	}
}

func StatusService(ctx context.Context, conn *apprunner.AppRunner, serviceArn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &apprunner.DescribeServiceInput{
			ServiceArn: aws.String(serviceArn),
		}

		output, err := conn.DescribeServiceWithContext(ctx, input)

		if err != nil {
			return nil, "", err
		}

		if output == nil || output.Service == nil {
			return nil, "", nil
		}

		return output.Service, aws.StringValue(output.Service.Status), nil
	}
}
