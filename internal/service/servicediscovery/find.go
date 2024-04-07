// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicediscovery

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicediscovery"
	awstypes "github.com/aws/aws-sdk-go-v2/service/servicediscovery/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindInstanceByServiceIDAndInstanceID(ctx context.Context, conn *servicediscovery.Client, serviceID, instanceID string) (*awstypes.Instance, error) {
	input := &servicediscovery.GetInstanceInput{
		InstanceId: aws.String(instanceID),
		ServiceId:  aws.String(serviceID),
	}

	output, err := conn.GetInstance(ctx, input)

	if errs.IsA[*awstypes.InstanceNotFound](err) {
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

func findNamespaces(ctx context.Context, conn *servicediscovery.Client, input *servicediscovery.ListNamespacesInput) ([]awstypes.NamespaceSummary, error) {
	var output []awstypes.NamespaceSummary

	pages := servicediscovery.NewListNamespacesPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.Namespaces...)
	}

	return output, nil
}

func findNamespacesByType(ctx context.Context, conn *servicediscovery.Client, nsType string) ([]awstypes.NamespaceSummary, error) {
	input := &servicediscovery.ListNamespacesInput{
		Filters: []awstypes.NamespaceFilter{{
			Condition: awstypes.FilterConditionEq,
			Name:      awstypes.NamespaceFilterNameType,
			Values:    []string{nsType},
		}},
	}

	return findNamespaces(ctx, conn, input)
}

func findNamespaceByNameAndType(ctx context.Context, conn *servicediscovery.Client, name, nsType string) (awstypes.NamespaceSummary, error) {
	output, err := findNamespacesByType(ctx, conn, nsType)

	if err != nil {
		return awstypes.NamespaceSummary{}, err
	}

	for _, v := range output {
		if aws.ToString(v.Name) == name {
			return v, nil
		}
	}

	return awstypes.NamespaceSummary{}, &retry.NotFoundError{}
}

func FindNamespaceByID(ctx context.Context, conn *servicediscovery.Client, id string) (*awstypes.Namespace, error) {
	input := &servicediscovery.GetNamespaceInput{
		Id: aws.String(id),
	}

	output, err := conn.GetNamespace(ctx, input)

	if errs.IsA[*awstypes.NamespaceNotFound](err) {
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

func FindOperationByID(ctx context.Context, conn *servicediscovery.Client, id string) (*awstypes.Operation, error) {
	input := &servicediscovery.GetOperationInput{
		OperationId: aws.String(id),
	}

	output, err := conn.GetOperation(ctx, input)

	if errs.IsA[*awstypes.OperationNotFound](err) {
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

func findServices(ctx context.Context, conn *servicediscovery.Client, input *servicediscovery.ListServicesInput) ([]awstypes.ServiceSummary, error) {
	var output []awstypes.ServiceSummary

	pages := servicediscovery.NewListServicesPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.Services...)
	}

	return output, nil
}

func findServicesByNamespaceID(ctx context.Context, conn *servicediscovery.Client, namespaceID string) ([]awstypes.ServiceSummary, error) {
	input := &servicediscovery.ListServicesInput{
		Filters: []awstypes.ServiceFilter{{
			Condition: awstypes.FilterConditionEq,
			Name:      awstypes.ServiceFilterNameNamespaceId,
			Values:    []string{namespaceID},
		}},
	}

	return findServices(ctx, conn, input)
}

func findServiceByNameAndNamespaceID(ctx context.Context, conn *servicediscovery.Client, name, namespaceID string) (awstypes.ServiceSummary, error) {
	output, err := findServicesByNamespaceID(ctx, conn, namespaceID)

	if err != nil {
		return awstypes.ServiceSummary{}, err
	}

	for _, v := range output {
		if aws.ToString(v.Name) == name {
			return v, nil
		}
	}

	return awstypes.ServiceSummary{}, &retry.NotFoundError{}
}

func FindServiceByID(ctx context.Context, conn *servicediscovery.Client, id string) (*awstypes.Service, error) {
	input := &servicediscovery.GetServiceInput{
		Id: aws.String(id),
	}

	output, err := conn.GetService(ctx, input)

	if errs.IsA[*awstypes.ServiceNotFound](err) {
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
