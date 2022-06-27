package amp

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/prometheusservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	resourceStatusFailed  = "Failed"
	resourceStatusUnknown = "Unknown"
	resourceStatusDeleted = "Deleted"
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

// statusWorkspaceCreated fetches the Workspace and its Status.
func statusWorkspaceCreated(ctx context.Context, conn *prometheusservice.PrometheusService, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &prometheusservice.DescribeWorkspaceInput{
			WorkspaceId: aws.String(id),
		}

		output, err := conn.DescribeWorkspaceWithContext(ctx, input)

		if err != nil {
			return output, resourceStatusFailed, err
		}

		if output == nil || output.Workspace == nil {
			return output, resourceStatusUnknown, nil
		}

		return output.Workspace, aws.StringValue(output.Workspace.Status.StatusCode), nil
	}
}

// statusWorkspaceDeleted fetches the Workspace and its Status
func statusWorkspaceDeleted(ctx context.Context, conn *prometheusservice.PrometheusService, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &prometheusservice.DescribeWorkspaceInput{
			WorkspaceId: aws.String(id),
		}

		output, err := conn.DescribeWorkspaceWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, prometheusservice.ErrCodeResourceNotFoundException) {
			return output, resourceStatusDeleted, nil
		}

		if err != nil {
			return output, resourceStatusUnknown, err
		}

		if output == nil || output.Workspace == nil {
			return output, resourceStatusUnknown, nil
		}

		return output.Workspace, aws.StringValue(output.Workspace.Status.StatusCode), nil
	}
}
