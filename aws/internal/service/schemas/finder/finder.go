package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/schemas"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func DiscovererByID(conn *schemas.Schemas, id string) (*schemas.DescribeDiscovererOutput, error) {
	input := &schemas.DescribeDiscovererInput{
		DiscovererId: aws.String(id),
	}

	output, err := conn.DescribeDiscoverer(input)

	if tfawserr.ErrCodeEquals(err, schemas.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return output, nil
}
