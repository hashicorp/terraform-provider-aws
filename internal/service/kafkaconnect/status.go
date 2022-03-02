package kafkaconnect

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafkaconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusCustomPluginState(conn *kafkaconnect.KafkaConnect, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindCustomPluginByARN(conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.CustomPluginState), nil
	}
}
