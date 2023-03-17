package directconnect

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusConnectionState(ctx context.Context, conn *directconnect.DirectConnect, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindConnectionByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.ConnectionState), nil
	}
}

func statusGatewayState(ctx context.Context, conn *directconnect.DirectConnect, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindGatewayByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.DirectConnectGatewayState), nil
	}
}

func statusGatewayAssociationState(ctx context.Context, conn *directconnect.DirectConnect, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindGatewayAssociationByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.AssociationState), nil
	}
}

func statusHostedConnectionState(ctx context.Context, conn *directconnect.DirectConnect, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindHostedConnectionByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.ConnectionState), nil
	}
}

func statusLagState(ctx context.Context, conn *directconnect.DirectConnect, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindLagByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.LagState), nil
	}
}
