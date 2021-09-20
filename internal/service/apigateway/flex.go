package apigateway

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func expandMethodParametersOperations(d *schema.ResourceData, key string, prefix string) []*apigateway.PatchOperation {
	operations := make([]*apigateway.PatchOperation, 0)

	oldParameters, newParameters := d.GetChange(key)
	oldParametersMap := oldParameters.(map[string]interface{})
	newParametersMap := newParameters.(map[string]interface{})

	for k := range oldParametersMap {
		operation := apigateway.PatchOperation{
			Op:   aws.String("remove"),
			Path: aws.String(fmt.Sprintf("/%s/%s", prefix, k)),
		}

		for nK, nV := range newParametersMap {
			b, ok := nV.(bool)
			if !ok {
				value, _ := strconv.ParseBool(nV.(string))
				b = value
			}
			if nK == k {
				operation.Op = aws.String("replace")
				operation.Value = aws.String(strconv.FormatBool(b))
			}
		}

		operations = append(operations, &operation)
	}

	for nK, nV := range newParametersMap {
		exists := false
		for k := range oldParametersMap {
			if k == nK {
				exists = true
			}
		}
		if !exists {
			b, ok := nV.(bool)
			if !ok {
				value, _ := strconv.ParseBool(nV.(string))
				b = value
			}
			operation := apigateway.PatchOperation{
				Op:    aws.String("add"),
				Path:  aws.String(fmt.Sprintf("/%s/%s", prefix, nK)),
				Value: aws.String(strconv.FormatBool(b)),
			}
			operations = append(operations, &operation)
		}
	}

	return operations
}

func expandRequestResponseModelOperations(d *schema.ResourceData, key string, prefix string) []*apigateway.PatchOperation {
	operations := make([]*apigateway.PatchOperation, 0)

	oldModels, newModels := d.GetChange(key)
	oldModelMap := oldModels.(map[string]interface{})
	newModelMap := newModels.(map[string]interface{})

	for k := range oldModelMap {
		operation := apigateway.PatchOperation{
			Op:   aws.String("remove"),
			Path: aws.String(fmt.Sprintf("/%s/%s", prefix, strings.Replace(k, "/", "~1", -1))),
		}

		for nK, nV := range newModelMap {
			if nK == k {
				operation.Op = aws.String("replace")
				operation.Value = aws.String(nV.(string))
			}
		}

		operations = append(operations, &operation)
	}

	for nK, nV := range newModelMap {
		exists := false
		for k := range oldModelMap {
			if k == nK {
				exists = true
			}
		}
		if !exists {
			operation := apigateway.PatchOperation{
				Op:    aws.String("add"),
				Path:  aws.String(fmt.Sprintf("/%s/%s", prefix, strings.Replace(nK, "/", "~1", -1))),
				Value: aws.String(nV.(string)),
			}
			operations = append(operations, &operation)
		}
	}

	return operations
}

func flattenAPIGatewayThrottleSettings(settings *apigateway.ThrottleSettings) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, 1)

	if settings != nil {
		r := make(map[string]interface{})
		if settings.BurstLimit != nil {
			r["burst_limit"] = *settings.BurstLimit
		}

		if settings.RateLimit != nil {
			r["rate_limit"] = *settings.RateLimit
		}

		result = append(result, r)
	}

	return result
}
