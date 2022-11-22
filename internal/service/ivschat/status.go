package ivschat

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ivschat"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusLoggingConfiguration(ctx context.Context, conn *ivschat.Client, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findLoggingConfigurationByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.State), nil
	}
}
