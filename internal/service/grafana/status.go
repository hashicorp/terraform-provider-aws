package grafana

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/managedgrafana"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusWorkspaceStatus(ctx context.Context, conn *managedgrafana.ManagedGrafana, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindWorkspaceByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func statusWorkspaceSAMLConfiguration(ctx context.Context, conn *managedgrafana.ManagedGrafana, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindSamlConfigurationByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}
