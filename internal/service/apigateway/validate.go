// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
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

func validIntegrationContentHandling() schema.SchemaValidateDiagFunc {
	return enum.Validate[types.ContentHandlingStrategy]()
}

func validUsagePlanQuotaSettings(v map[string]interface{}) (errors []error) {
	period := v["period"].(string)
	offset := v["offset"].(int)

	if period == string(types.QuotaPeriodTypeDay) && offset != 0 {
		errors = append(errors, fmt.Errorf("Usage Plan quota offset must be zero in the DAY period"))
	}

	if period == string(types.QuotaPeriodTypeWeek) && (offset < 0 || offset > 6) {
		errors = append(errors, fmt.Errorf("Usage Plan quota offset must be between 0 and 6 inclusive in the WEEK period"))
	}

	if period == string(types.QuotaPeriodTypeMonth) && (offset < 0 || offset > 27) {
		errors = append(errors, fmt.Errorf("Usage Plan quota offset must be between 0 and 27 inclusive in the MONTH period"))
	}

	return
}
