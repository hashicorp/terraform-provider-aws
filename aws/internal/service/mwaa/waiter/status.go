package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mwaa"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/mwaa/finder"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	environmentStatusNotFound = "NotFound"
	environmentStatusUnknown  = "Unknown"
)

// statusEnvironment fetches the Environment and its Status
func statusEnvironment(conn *mwaa.MWAA, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		environment, err := finder.findEnvironmentByName(conn, name)

		if tfawserr.ErrCodeEquals(err, mwaa.ErrCodeResourceNotFoundException) {
			return nil, environmentStatusNotFound, nil
		}

		if err != nil {
			return nil, environmentStatusUnknown, err
		}

		if environment == nil {
			return nil, environmentStatusNotFound, nil
		}

		return environment, aws.StringValue(environment.Status), nil
	}
}
