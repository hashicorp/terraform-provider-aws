// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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

	for k, kV := range oldParametersMap {
		keyValueUnchanged := false
		operation := apigateway.PatchOperation{
			Op:   aws.String(apigateway.OpRemove),
			Path: aws.String(fmt.Sprintf("/%s/%s", prefix, k)),
		}

		for nK, nV := range newParametersMap {
			b, ok := nV.(bool)
			if !ok {
				value, _ := strconv.ParseBool(nV.(string))
				b = value
			}

			if (nK == k) && (nV != kV) {
				operation.Op = aws.String(apigateway.OpReplace)
				operation.Value = aws.String(strconv.FormatBool(b))
			} else if (nK == k) && (nV == kV) {
				keyValueUnchanged = true
			}
		}

		if !keyValueUnchanged {
			operations = append(operations, &operation)
		}
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
				Op:    aws.String(apigateway.OpAdd),
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
			Op:   aws.String(apigateway.OpRemove),
			Path: aws.String(fmt.Sprintf("/%s/%s", prefix, strings.Replace(k, "/", "~1", -1))),
		}

		for nK, nV := range newModelMap {
			if nK == k {
				operation.Op = aws.String(apigateway.OpReplace)
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
				Op:    aws.String(apigateway.OpAdd),
				Path:  aws.String(fmt.Sprintf("/%s/%s", prefix, strings.Replace(nK, "/", "~1", -1))),
				Value: aws.String(nV.(string)),
			}
			operations = append(operations, &operation)
		}
	}

	return operations
}
