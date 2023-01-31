package configservice

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindConfigRule(ctx context.Context, conn *configservice.ConfigService, name string) (*configservice.ConfigRule, error) {
	input := &configservice.DescribeConfigRulesInput{
		ConfigRuleNames: []*string{aws.String(name)},
	}

	output, err := conn.DescribeConfigRulesWithContext(ctx, input)
	if tfawserr.ErrCodeEquals(err, configservice.ErrCodeNoSuchConfigRuleException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if output == nil {
		return nil, nil
	}

	if output == nil || output.ConfigRules == nil || len(output.ConfigRules) == 0 || output.ConfigRules[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ConfigRules[0], nil
}
