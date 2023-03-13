package configservice

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusRule(ctx context.Context, conn *configservice.ConfigService, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindConfigRule(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.ConfigRuleState), nil
	}
}
