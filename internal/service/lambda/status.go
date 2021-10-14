package lambda

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/lambda/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	tflambda "github.com/hashicorp/terraform-provider-aws/internal/service/lambda"
	tflambda "github.com/hashicorp/terraform-provider-aws/internal/service/lambda"
)

const (
	eventSourceMappingStateCreating  = "Creating"
	eventSourceMappingStateDeleting  = "Deleting"
	eventSourceMappingStateDisabled  = "Disabled"
	eventSourceMappingStateDisabling = "Disabling"
	eventSourceMappingStateEnabled   = "Enabled"
	eventSourceMappingStateEnabling  = "Enabling"
	eventSourceMappingStateUpdating  = "Updating"
)

func statusEventSourceMappingState(conn *lambda.Lambda, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		eventSourceMappingConfiguration, err := tflambda.FindEventSourceMappingConfigurationByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return eventSourceMappingConfiguration, aws.StringValue(eventSourceMappingConfiguration.State), nil
	}
}
