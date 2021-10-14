package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func AuthorizerByName(conn *iot.IoT, name string) (*iot.AuthorizerDescription, error) {
	input := &iot.DescribeAuthorizerInput{
		AuthorizerName: aws.String(name),
	}

	output, err := conn.DescribeAuthorizer(input)

	if tfawserr.ErrCodeEquals(err, iot.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.AuthorizerDescription == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.AuthorizerDescription, nil
}
