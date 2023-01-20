package transfer

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/transfer"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	userStateExists = "exists"
)

func statusServerState(ctx context.Context, conn *transfer.Transfer, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindServerByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func statusUserState(ctx context.Context, conn *transfer.Transfer, serverID, userName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindUserByServerIDAndUserName(ctx, conn, serverID, userName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, userStateExists, nil
	}
}
