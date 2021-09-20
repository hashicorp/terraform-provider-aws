package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datasync"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/datasync/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	agentStatusReady = "ready"
)

func statusAgent(conn *datasync.DataSync, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.FindAgentByARN(conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, agentStatusReady, nil
	}
}

func statusTask(conn *datasync.DataSync, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.FindTaskByARN(conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}
