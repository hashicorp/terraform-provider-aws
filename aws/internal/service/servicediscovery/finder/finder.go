package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func OperationByID(conn *servicediscovery.ServiceDiscovery, id string) (*servicediscovery.Operation, error) {
	input := &servicediscovery.GetOperationInput{
		OperationId: aws.String(id),
	}

	output, err := conn.GetOperation(input)

	if tfawserr.ErrCodeEquals(err, servicediscovery.ErrCodeOperationNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Operation == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Operation, nil
}

func ServiceByID(conn *servicediscovery.ServiceDiscovery, id string) (*servicediscovery.Service, error) {
	input := &servicediscovery.GetServiceInput{
		Id: aws.String(id),
	}

	output, err := conn.GetService(input)

	if tfawserr.ErrCodeEquals(err, servicediscovery.ErrCodeServiceNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Service == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Service, nil
}
