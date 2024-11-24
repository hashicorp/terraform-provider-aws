// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package memorydb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/memorydb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// statusSnapshot fetches the MemoryDB Snapshot and its status.
func statusSnapshot(ctx context.Context, conn *memorydb.Client, snapshotName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		snapshot, err := FindSnapshotByName(ctx, conn, snapshotName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return snapshot, aws.ToString(snapshot.Status), nil
	}
}

// statusUser fetches the MemoryDB user and its status.
func statusUser(ctx context.Context, conn *memorydb.Client, userName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		user, err := FindUserByName(ctx, conn, userName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return user, aws.ToString(user.Status), nil
	}
}
