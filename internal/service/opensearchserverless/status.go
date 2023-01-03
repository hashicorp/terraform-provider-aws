package opensearchserverless

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusVPCEndpoint(ctx context.Context, conn *opensearchserverless.Client, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findVPCEndpointByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}
