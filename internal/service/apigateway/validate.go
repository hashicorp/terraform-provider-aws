package apigateway

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func validHTTPMethod() schema.SchemaValidateFunc {
	return validation.StringInSlice([]string{
		"ANY",
		"DELETE",
		"GET",
		"HEAD",
		"OPTIONS",
		"PATCH",
		"POST",
		"PUT",
	}, false)
}

func validIntegrationContentHandling() schema.SchemaValidateFunc {
	return validation.StringInSlice([]string{
		apigateway.ContentHandlingStrategyConvertToBinary,
		apigateway.ContentHandlingStrategyConvertToText,
	}, false)
}

func validUsagePlanQuotaSettings(v map[string]interface{}) (errors []error) {
	period := v["period"].(string)
	offset := v["offset"].(int)

	if period == apigateway.QuotaPeriodTypeDay && offset != 0 {
		errors = append(errors, fmt.Errorf("Usage Plan quota offset must be zero in the DAY period"))
	}

	if period == apigateway.QuotaPeriodTypeWeek && (offset < 0 || offset > 6) {
		errors = append(errors, fmt.Errorf("Usage Plan quota offset must be between 0 and 6 inclusive in the WEEK period"))
	}

	if period == apigateway.QuotaPeriodTypeMonth && (offset < 0 || offset > 27) {
		errors = append(errors, fmt.Errorf("Usage Plan quota offset must be between 0 and 27 inclusive in the MONTH period"))
	}

	return
}
