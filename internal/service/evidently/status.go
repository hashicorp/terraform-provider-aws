package evidently

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchevidently"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusFeature(ctx context.Context, conn *cloudwatchevidently.CloudWatchEvidently, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		featureName, projectNameOrARN, err := FeatureParseID(id)

		if err != nil {
			return nil, "", err
		}

		output, err := FindFeatureWithProjectNameorARN(ctx, conn, featureName, projectNameOrARN)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func statusProject(ctx context.Context, conn *cloudwatchevidently.CloudWatchEvidently, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindProjectByNameOrARN(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}
