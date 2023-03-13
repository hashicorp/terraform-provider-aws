package amp

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/prometheusservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusAlertManagerDefinition(ctx context.Context, conn *prometheusservice.PrometheusService, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindAlertManagerDefinitionByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status.StatusCode), nil
	}
}

func statusRuleGroupNamespace(ctx context.Context, conn *prometheusservice.PrometheusService, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindRuleGroupNamespaceByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status.StatusCode), nil
	}
}

func statusWorkspace(ctx context.Context, conn *prometheusservice.PrometheusService, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindWorkspaceByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status.StatusCode), nil
	}
}

func statusLoggingConfiguration(ctx context.Context, conn *prometheusservice.PrometheusService, workspaceID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindLoggingConfigurationByWorkspaceID(ctx, conn, workspaceID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status.StatusCode), nil
	}
}
