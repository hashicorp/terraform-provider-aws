package workmail

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/workmail"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	statusRequested = "Requested"
	statusCreating  = "Creating"
	statusActive    = "Active"
	statusDeleting  = "Deleting"
	statusDeleted   = "Deleted"
)

func statusOrganization(ctx context.Context, conn *workmail.Client, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findOrganizationByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, aws.ToString(out.State), nil
	}
}
