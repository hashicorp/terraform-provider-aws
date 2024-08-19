// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lightsail

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/aws/aws-sdk-go-v2/service/lightsail/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// expandOperations provides a uniform approach for handling lightsail operations and errors.
func expandOperations(ctx context.Context, conn *lightsail.Client, operations []types.Operation, action types.OperationType, resource string, id string) diag.Diagnostics {
	var diags diag.Diagnostics

	if len(operations) == 0 {
		return create.AppendDiagError(diags, names.Lightsail, string(action), resource, id, errors.New("no operations found for request"))
	}

	op := operations[0]

	err := waitOperation(ctx, conn, op.Id)
	if err != nil {
		return create.AppendDiagError(diags, names.Lightsail, string(action), resource, id, errors.New("error waiting for request operation"))
	}

	return diags
}

// expandOperation provides a uniform approach for handling a single lightsail operation and errors.
func expandOperation(ctx context.Context, conn *lightsail.Client, operation types.Operation, action types.OperationType, resource string, id string) diag.Diagnostics {
	diags := expandOperations(ctx, conn, []types.Operation{operation}, action, resource, id)

	if diags != nil {
		return diags
	}

	return diags
}

func flattenResourceLocation(apiObject *types.ResourceLocation) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.AvailabilityZone; v != nil {
		m[names.AttrAvailabilityZone] = aws.ToString(v)
	}

	if v := apiObject.RegionName; string(v) != "" {
		m["region_name"] = string(v)
	}

	return m
}
