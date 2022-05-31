package apprunner

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apprunner"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	AutoScalingConfigurationStatusActive   = "active"
	AutoScalingConfigurationStatusInactive = "inactive"

	CustomDomainAssociationStatusActive                          = "active"
	CustomDomainAssociationStatusCreating                        = "creating"
	CustomDomainAssociationStatusDeleting                        = "deleting"
	CustomDomainAssociationStatusPendingCertificateDNSValidation = "pending_certificate_dns_validation"
	CustomDomainAssociationStatusBindingCertificate              = "binding_certificate"
)

func StatusAutoScalingConfiguration(ctx context.Context, conn *apprunner.AppRunner, arn string) resource.StateRefreshFunc {
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

func StatusVPCConnector(ctx context.Context, conn *apprunner.AppRunner, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &apprunner.DescribeVpcConnectorInput{
			VpcConnectorArn: aws.String(arn),
		}

		output, err := conn.DescribeVpcConnectorWithContext(ctx, input)

		if err != nil {
			return nil, "", err
		}

		if output == nil || output.VpcConnector == nil {
			return nil, "", nil
		}

		return output.VpcConnector, aws.StringValue(output.VpcConnector.Status), nil
	}
}

func StatusConnection(ctx context.Context, conn *apprunner.AppRunner, name string) resource.StateRefreshFunc {
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

func StatusCustomDomain(ctx context.Context, conn *apprunner.AppRunner, domainName, serviceArn string) resource.StateRefreshFunc {
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

func StatusService(ctx context.Context, conn *apprunner.AppRunner, serviceArn string) resource.StateRefreshFunc {
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
