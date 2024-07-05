// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindCapacityProviderByARN(ctx context.Context, conn *ecs.ECS, arn string) (*ecs.CapacityProvider, error) {
	input := &ecs.DescribeCapacityProvidersInput{
		CapacityProviders: aws.StringSlice([]string{arn}),
		Include:           aws.StringSlice([]string{ecs.CapacityProviderFieldTags}),
	}

	output, err := conn.DescribeCapacityProvidersWithContext(ctx, input)

	// Some partitions (i.e., ISO) may not support tagging, giving error
	if errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
		log.Printf("[WARN] ECS tagging failed describing Capacity Provider (%s) with tags: %s; retrying without tags", arn, err)

		input.Include = nil
		output, err = conn.DescribeCapacityProvidersWithContext(ctx, input)
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.CapacityProviders) == 0 || output.CapacityProviders[0] == nil {
		return nil, &retry.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	capacityProvider := output.CapacityProviders[0]

	if status := aws.StringValue(capacityProvider.Status); status == ecs.CapacityProviderStatusInactive {
		return nil, &retry.NotFoundError{
			Message:     status,
			LastRequest: input,
		}
	}

	return capacityProvider, nil
}

func findServiceByTwoPartKey(ctx context.Context, conn *ecs.ECS, serviceName, clusterNameOrARN string) (*ecs.Service, error) {
	input := &ecs.DescribeServicesInput{
		Cluster:  aws.String(clusterNameOrARN),
		Include:  aws.StringSlice([]string{ecs.ServiceFieldTags}),
		Services: aws.StringSlice([]string{serviceName}),
	}

	return findService(ctx, conn, input)
}

func FindServiceNoTagsByID(ctx context.Context, conn *ecs.ECS, id, cluster string) (*ecs.Service, error) {
	input := &ecs.DescribeServicesInput{
		Services: aws.StringSlice([]string{id}),
	}
	if cluster != "" {
		input.Cluster = aws.String(cluster)
	}

	return findService(ctx, conn, input)
}

type expectActiveError struct {
	status string
}

func newExpectActiveError(status string) *expectActiveError {
	return &expectActiveError{
		status: status,
	}
}

func (e *expectActiveError) Error() string {
	return fmt.Sprintf("expected status %[1]q, was %[2]q", serviceStatusActive, e.status)
}

func FindServiceByIDWaitForActive(ctx context.Context, conn *ecs.ECS, id, cluster string) (*ecs.Service, error) {
	var service *ecs.Service
	// Use the retry.RetryContext function instead of WaitForState() because we don't want the timeout error, if any
	err := retry.RetryContext(ctx, serviceDescribeTimeout, func() *retry.RetryError {
		var err error
		service, err = findServiceByTwoPartKey(ctx, conn, id, cluster)
		if tfresource.NotFound(err) {
			return retry.RetryableError(err)
		}
		if err != nil {
			return retry.NonRetryableError(err)
		}

		if status := aws.StringValue(service.Status); status != serviceStatusActive {
			return retry.RetryableError(newExpectActiveError(status))
		}

		return nil
	})
	if tfresource.TimedOut(err) {
		service, err = findServiceByTwoPartKey(ctx, conn, id, cluster)
	}

	return service, err
}

func findService(ctx context.Context, conn *ecs.ECS, input *ecs.DescribeServicesInput) (*ecs.Service, error) {
	output, err := findServices(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findServices(ctx context.Context, conn *ecs.ECS, input *ecs.DescribeServicesInput) ([]*ecs.Service, error) {
	output, err := conn.DescribeServicesWithContext(ctx, input)

	if errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) && input.Include != nil {
		input.Include = nil
		output, err = conn.DescribeServicesWithContext(ctx, input)
	}

	// As of AWS SDK for Go v1.44.42, DescribeServices does not return the error code ecs.ErrCodeServiceNotFoundException
	// Keep this here in case it ever does
	if tfawserr.ErrCodeEquals(err, ecs.ErrCodeServiceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	// When an ECS Service is not found by DescribeServices(), it will return a Failure struct with Reason = "MISSING"
	for _, v := range output.Failures {
		if aws.StringValue(v.Reason) == "MISSING" {
			return nil, &retry.NotFoundError{
				LastRequest: input,
			}
		}
	}

	return output.Services, nil
}
