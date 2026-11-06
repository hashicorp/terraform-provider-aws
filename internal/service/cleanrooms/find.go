package cleanrooms

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cleanrooms"
	"github.com/aws/aws-sdk-go-v2/service/cleanrooms/types"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"

	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func findAnalysisRuleByID(ctx context.Context, conn *cleanrooms.Client, id string, analysisRule string) (*cleanrooms.GetConfiguredTableAnalysisRuleOutput, error) {
	analysisRuleType, _ := expandAnalysisRuleType(analysisRule)

	in := &cleanrooms.GetConfiguredTableAnalysisRuleInput{
		ConfiguredTableIdentifier: aws.String(id),
		AnalysisRuleType:          analysisRuleType,
	}
	out, err := conn.GetConfiguredTableAnalysisRule(ctx, in)

	if errs.IsA[*types.AccessDeniedException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.AnalysisRule == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}
