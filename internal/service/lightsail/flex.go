package lightsail

import (
	"context"
	"errors"

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
