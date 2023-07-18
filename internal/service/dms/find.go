// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dms

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	dms "github.com/aws/aws-sdk-go/service/databasemigrationservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindEndpointByID(ctx context.Context, conn *dms.DatabaseMigrationService, id string) (*dms.Endpoint, error) {
	input := &dms.DescribeEndpointsInput{
		Filters: []*dms.Filter{
			{
				Name:   aws.String("endpoint-id"),
				Values: aws.StringSlice([]string{id}),
			},
		},
	}

	output, err := conn.DescribeEndpointsWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, dms.ErrCodeResourceNotFoundFault) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Endpoints) == 0 || output.Endpoints[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.Endpoints); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.Endpoints[0], nil
}

func FindReplicationTaskByID(ctx context.Context, conn *dms.DatabaseMigrationService, id string) (*dms.ReplicationTask, error) {
	input := &dms.DescribeReplicationTasksInput{
		Filters: []*dms.Filter{
			{
				Name:   aws.String("replication-task-id"),
				Values: []*string{aws.String(id)}, // Must use d.Id() to work with import.
			},
		},
	}

	var results []*dms.ReplicationTask

	err := conn.DescribeReplicationTasksPagesWithContext(ctx, input, func(page *dms.DescribeReplicationTasksOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, task := range page.ReplicationTasks {
			if task == nil {
				continue
			}
			results = append(results, task)
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, dms.ErrCodeResourceNotFoundFault) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(results); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return results[0], nil
}
