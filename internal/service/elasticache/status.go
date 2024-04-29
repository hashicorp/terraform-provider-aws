// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	cacheClusterStatusAvailable             = "available"
	cacheClusterStatusCreating              = "creating"
	cacheClusterStatusDeleted               = "deleted"
	cacheClusterStatusDeleting              = "deleting"
	cacheClusterStatusIncompatibleNetwork   = "incompatible-network"
	cacheClusterStatusModifying             = "modifying"
	cacheClusterStatusRebootingClusterNodes = "rebooting cluster nodes"
	cacheClusterStatusRestoreFailed         = "restore-failed"
	cacheClusterStatusSnapshotting          = "snapshotting"
)

// StatusCacheCluster fetches the Cache Cluster and its Status
func StatusCacheCluster(ctx context.Context, conn *elasticache.ElastiCache, cacheClusterID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		c, err := FindCacheClusterByID(ctx, conn, cacheClusterID)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}

		return c, aws.StringValue(c.CacheClusterStatus), nil
	}
}
