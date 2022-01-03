package servicediscovery

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindInstanceByServiceIDAndInstanceID(conn *servicediscovery.ServiceDiscovery, serviceID, instanceID string) (*servicediscovery.Instance, error) {
	input := &servicediscovery.GetInstanceInput{
		InstanceId: aws.String(instanceID),
		ServiceId:  aws.String(serviceID),
	}

	output, err := conn.GetInstance(input)

	if tfawserr.ErrCodeEquals(err, servicediscovery.ErrCodeInstanceNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Instance == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Instance, nil
}

func FindOperationByID(conn *servicediscovery.ServiceDiscovery, id string) (*servicediscovery.Operation, error) {
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

func FindServiceByID(conn *servicediscovery.ServiceDiscovery, id string) (*servicediscovery.Service, error) {
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
