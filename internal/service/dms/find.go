package dms

import (
	"github.com/aws/aws-sdk-go/aws"
	dms "github.com/aws/aws-sdk-go/service/databasemigrationservice"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindEndpointByID(conn *dms.DatabaseMigrationService, id string) (*dms.Endpoint, error) {
	input := &dms.DescribeEndpointsInput{
		Filters: []*dms.Filter{
			{
				Name:   aws.String("endpoint-id"),
				Values: aws.StringSlice([]string{id}),
			},
		},
	}

	output, err := conn.DescribeEndpoints(input)

	if tfawserr.ErrCodeEquals(err, dms.ErrCodeResourceNotFoundFault) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Endpoints) == 0 || output.Endpoints[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.Endpoints); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.Endpoints[0], nil
}
