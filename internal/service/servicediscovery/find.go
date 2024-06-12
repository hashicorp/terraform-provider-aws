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
