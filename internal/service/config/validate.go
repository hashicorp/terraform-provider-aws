package config

import (
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func validExecutionFrequency() schema.SchemaValidateFunc {
	return validation.StringInSlice([]string{
		configservice.MaximumExecutionFrequencyOneHour,
		configservice.MaximumExecutionFrequencyThreeHours,
		configservice.MaximumExecutionFrequencySixHours,
		configservice.MaximumExecutionFrequencyTwelveHours,
		configservice.MaximumExecutionFrequencyTwentyFourHours,
	}, false)
}
