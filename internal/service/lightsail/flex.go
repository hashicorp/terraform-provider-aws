package lightsail

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// expandOperations provides a uniform approach for handling lightsail operations and errors.
func expandOperations(ctx context.Context, conn *lightsail.Lightsail, operations []*lightsail.Operation, action string, resource string, id string) diag.Diagnostics {
	if len(operations) == 0 {
		return create.DiagError(names.Lightsail, action, resource, id, errors.New("no operations found for request"))
	}

	op := operations[0]

	err := waitOperation(ctx, conn, op.Id)
	if err != nil {
		return create.DiagError(names.Lightsail, action, resource, id, errors.New("error waiting for request operation"))
	}

	return nil
}

// expandOperation provides a uniform approach for handling a single lightsail operation and errors.
func expandOperation(ctx context.Context, conn *lightsail.Lightsail, operation *lightsail.Operation, action string, resource string, id string) diag.Diagnostics {
	diag := expandOperations(ctx, conn, []*lightsail.Operation{operation}, action, resource, id)

	if diag != nil {
		return diag
	}

	return nil
}

func flattenResourceLocation(apiObject *lightsail.ResourceLocation) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.AvailabilityZone; v != nil {
		m["availability_zone"] = aws.StringValue(v)
	}

	if v := apiObject.RegionName; v != nil {
		m["region_name"] = aws.StringValue(v)
	}

	return m
}
