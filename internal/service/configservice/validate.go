package configservice

import (
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func validExecutionFrequency() schema.SchemaValidateFunc {
	return validation.StringInSlice(configservice.MaximumExecutionFrequency_Values(), false)
}
