// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicediscovery

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindInstanceByServiceIDAndInstanceID(ctx context.Context, conn *servicediscovery.ServiceDiscovery, serviceID, instanceID string) (*servicediscovery.Instance, error) {
	input := &servicediscovery.GetInstanceInput{
		InstanceId: aws.String(instanceID),
		ServiceId:  aws.String(serviceID),
	}

	output, err := conn.GetInstanceWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, servicediscovery.ErrCodeInstanceNotFound) {
		return nil, &retry.NotFoundError{
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

func findNamespaces(ctx context.Context, conn *servicediscovery.ServiceDiscovery, input *servicediscovery.ListNamespacesInput) ([]*servicediscovery.NamespaceSummary, error) {
	var output []*servicediscovery.NamespaceSummary

	err := conn.ListNamespacesPagesWithContext(ctx, input, func(page *servicediscovery.ListNamespacesOutput, lastPage bool) bool {
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

func findNamespacesByType(ctx context.Context, conn *servicediscovery.ServiceDiscovery, nsType string) ([]*servicediscovery.NamespaceSummary, error) {
	input := &servicediscovery.ListNamespacesInput{
		Filters: []*servicediscovery.NamespaceFilter{{
			Condition: aws.String(servicediscovery.FilterConditionEq),
			Name:      aws.String(servicediscovery.NamespaceFilterNameType),
			Values:    aws.StringSlice([]string{nsType}),
		}},
	}

	return findNamespaces(ctx, conn, input)
}

func findNamespaceByNameAndType(ctx context.Context, conn *servicediscovery.ServiceDiscovery, name, nsType string) (*servicediscovery.NamespaceSummary, error) {
	output, err := findNamespacesByType(ctx, conn, nsType)

	if err != nil {
		return nil, err
	}

	for _, v := range output {
		if aws.StringValue(v.Name) == name {
			return v, nil
		}
	}

	return nil, &retry.NotFoundError{}
}

func FindNamespaceByID(ctx context.Context, conn *servicediscovery.ServiceDiscovery, id string) (*servicediscovery.Namespace, error) {
	input := &servicediscovery.GetNamespaceInput{
		Id: aws.String(id),
	}

	output, err := conn.GetNamespaceWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, servicediscovery.ErrCodeNamespaceNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Namespace == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Namespace, nil
}

func FindOperationByID(ctx context.Context, conn *servicediscovery.ServiceDiscovery, id string) (*servicediscovery.Operation, error) {
	input := &servicediscovery.GetOperationInput{
		OperationId: aws.String(id),
	}

	output, err := conn.GetOperationWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, servicediscovery.ErrCodeOperationNotFound) {
		return nil, &retry.NotFoundError{
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

func findServices(ctx context.Context, conn *servicediscovery.ServiceDiscovery, input *servicediscovery.ListServicesInput) ([]*servicediscovery.ServiceSummary, error) {
	var output []*servicediscovery.ServiceSummary

	err := conn.ListServicesPagesWithContext(ctx, input, func(page *servicediscovery.ListServicesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Services {
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

func findServicesByNamespaceID(ctx context.Context, conn *servicediscovery.ServiceDiscovery, namespaceID string) ([]*servicediscovery.ServiceSummary, error) {
	input := &servicediscovery.ListServicesInput{
		Filters: []*servicediscovery.ServiceFilter{{
			Condition: aws.String(servicediscovery.FilterConditionEq),
			Name:      aws.String(servicediscovery.ServiceFilterNameNamespaceId),
			Values:    aws.StringSlice([]string{namespaceID}),
		}},
	}

	return findServices(ctx, conn, input)
}

func findServiceByNameAndNamespaceID(ctx context.Context, conn *servicediscovery.ServiceDiscovery, name, namespaceID string) (*servicediscovery.ServiceSummary, error) {
	output, err := findServicesByNamespaceID(ctx, conn, namespaceID)

	if err != nil {
		return nil, err
	}

	for _, v := range output {
		if aws.StringValue(v.Name) == name {
			return v, nil
		}
	}

	return nil, &retry.NotFoundError{}
}

func FindServiceByID(ctx context.Context, conn *servicediscovery.ServiceDiscovery, id string) (*servicediscovery.Service, error) {
	input := &servicediscovery.GetServiceInput{
		Id: aws.String(id),
	}

	output, err := conn.GetServiceWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, servicediscovery.ErrCodeServiceNotFound) {
		return nil, &retry.NotFoundError{
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
