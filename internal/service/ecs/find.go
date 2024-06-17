// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindCapacityProviderByARN(ctx context.Context, conn *ecs.Client, partition string, arn string) (*awstypes.CapacityProvider, error) {
	input := &ecs.DescribeCapacityProvidersInput{
		CapacityProviders: []string{arn},
		Include:           []awstypes.CapacityProviderField{awstypes.CapacityProviderFieldTags},
	}

	output, err := conn.DescribeCapacityProviders(ctx, input)

	// Some partitions (i.e., ISO) may not support tagging, giving error
	if errs.IsUnsupportedOperationInPartitionError(partition, err) {
		log.Printf("[WARN] ECS tagging failed describing Capacity Provider (%s) with tags: %s; retrying without tags", arn, err)

		input.Include = nil
		output, err = conn.DescribeCapacityProviders(ctx, input)
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.CapacityProviders) == 0 {
		return nil, &retry.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	capacityProvider := output.CapacityProviders[0]

	if capacityProvider.Status == awstypes.CapacityProviderStatusInactive {
		return nil, &retry.NotFoundError{
			Message:     string(capacityProvider.Status),
			LastRequest: input,
		}
	}

	return &capacityProvider, nil
}

func FindServiceByID(ctx context.Context, conn *ecs.Client, partition, id, cluster string) (*awstypes.Service, error) {
	input := &ecs.DescribeServicesInput{
		Cluster:  aws.String(cluster),
		Include:  []awstypes.ServiceField{awstypes.ServiceFieldTags},
		Services: []string{id},
	}

	return FindService(ctx, conn, partition, input)
}

func FindServiceNoTagsByID(ctx context.Context, conn *ecs.Client, partition, id, cluster string) (*awstypes.Service, error) {
	input := &ecs.DescribeServicesInput{
		Services: []string{id},
	}
	if cluster != "" {
		input.Cluster = aws.String(cluster)
	}

	return FindService(ctx, conn, partition, input)
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

func FindServiceByIDWaitForActive(ctx context.Context, conn *ecs.Client, partition, id, cluster string) (*awstypes.Service, error) {
	var service *awstypes.Service
	// Use the retry.RetryContext function instead of WaitForState() because we don't want the timeout error, if any
	err := retry.RetryContext(ctx, serviceDescribeTimeout, func() *retry.RetryError {
		var err error
		service, err = FindServiceByID(ctx, conn, id, partition, cluster)
		if tfresource.NotFound(err) {
			return retry.RetryableError(err)
		}
		if err != nil {
			return retry.NonRetryableError(err)
		}

		if status := aws.ToString(service.Status); status != serviceStatusActive {
			return retry.RetryableError(newExpectActiveError(status))
		}

		return nil
	})
	if tfresource.TimedOut(err) {
		service, err = FindServiceByID(ctx, conn, id, partition, cluster)
	}

	return service, err
}

func FindService(ctx context.Context, conn *ecs.Client, partition string, input *ecs.DescribeServicesInput) (*awstypes.Service, error) {
	output, err := conn.DescribeServices(ctx, input)

	if errs.IsUnsupportedOperationInPartitionError(partition, err) && input.Include != nil {
		id := input.Services[0]
		log.Printf("[WARN] failed describing ECS Service (%s) with tags: %s; retrying without tags", id, err)

		input.Include = nil
		output, err = conn.DescribeServices(ctx, input)
	}

	// As of AWS SDK for Go v1.44.42, DescribeServices does not return the error code ecs.ErrCodeServiceNotFoundException
	// Keep this here in case it ever does
	if errs.IsA[*awstypes.ServiceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}
	if err != nil {
		return nil, err
	}

	// When an ECS Service is not found by DescribeServices(), it will return a Failure struct with Reason = "MISSING"
	for _, v := range output.Failures {
		if aws.ToString(v.Reason) == "MISSING" {
			return nil, &retry.NotFoundError{
				LastRequest: input,
			}
		}
	}

	if len(output.Services) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}
	if n := len(output.Services); n > 1 {
		return nil, tfresource.NewTooManyResultsError(n, input)
	}

	return &output.Services[0], nil
}
