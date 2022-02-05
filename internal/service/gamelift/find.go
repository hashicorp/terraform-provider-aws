package gamelift

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/gamelift"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindBuildByID(conn *gamelift.GameLift, id string) (*gamelift.Build, error) {
	input := &gamelift.DescribeBuildInput{
		BuildId: aws.String(id),
	}

	output, err := conn.DescribeBuild(input)

	if tfawserr.ErrCodeEquals(err, gamelift.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Build == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Build, nil
}
