package workmail

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/workmail"
	"github.com/aws/aws-sdk-go-v2/service/workmail/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func findOrganizationByID(ctx context.Context, conn *workmail.Client, id string) (*workmail.DescribeOrganizationOutput, error) {
	in := &workmail.DescribeOrganizationInput{
		OrganizationId: aws.String(id),
	}
	out, err := conn.DescribeOrganization(ctx, in)
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &resource.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.OrganizationId == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}
