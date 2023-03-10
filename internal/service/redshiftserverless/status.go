package redshiftserverless

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshiftserverless"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusNamespace(ctx context.Context, conn *redshiftserverless.RedshiftServerless, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindNamespaceByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func statusWorkgroup(ctx context.Context, conn *redshiftserverless.RedshiftServerless, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindWorkgroupByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func statusEndpointAccess(ctx context.Context, conn *redshiftserverless.RedshiftServerless, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindEndpointAccessByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.EndpointStatus), nil
	}
}

func statusSnapshot(ctx context.Context, conn *redshiftserverless.RedshiftServerless, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindSnapshotByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}
