package servicediscovery

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
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

func findNamespaces(conn *servicediscovery.ServiceDiscovery, input *servicediscovery.ListNamespacesInput) ([]*servicediscovery.NamespaceSummary, error) {
	var output []*servicediscovery.NamespaceSummary

	err := conn.ListNamespacesPages(input, func(page *servicediscovery.ListNamespacesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Namespaces {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}

func findNamespacesByType(conn *servicediscovery.ServiceDiscovery, nsType string) ([]*servicediscovery.NamespaceSummary, error) {
	input := &servicediscovery.ListNamespacesInput{
		Filters: []*servicediscovery.NamespaceFilter{{
			Condition: aws.String(servicediscovery.FilterConditionEq),
			Name:      aws.String(servicediscovery.NamespaceFilterNameType),
			Values:    aws.StringSlice([]string{nsType}),
		}},
	}

	return findNamespaces(conn, input)
}

func findNamespaceByNameAndType(conn *servicediscovery.ServiceDiscovery, name, nsType string) (*servicediscovery.NamespaceSummary, error) {
	output, err := findNamespacesByType(conn, nsType)

	if err != nil {
		return nil, err
	}

	for _, v := range output {
		if aws.StringValue(v.Name) == name {
			return v, nil
		}
	}

	return nil, &resource.NotFoundError{}
}

func FindNamespaceByID(conn *servicediscovery.ServiceDiscovery, id string) (*servicediscovery.Namespace, error) {
	input := &servicediscovery.GetNamespaceInput{
		Id: aws.String(id),
	}

	output, err := conn.GetNamespace(input)

	if tfawserr.ErrCodeEquals(err, servicediscovery.ErrCodeNamespaceNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Namespace == nil || output.Namespace.Properties == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Namespace, nil
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
