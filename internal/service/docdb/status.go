// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package docdb

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/docdb"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func statusGlobalClusterRefreshFunc(ctx context.Context, conn *docdb.DocDB, globalClusterID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		globalCluster, err := FindGlobalClusterById(ctx, conn, globalClusterID)

		if tfawserr.ErrCodeEquals(err, docdb.ErrCodeGlobalClusterNotFoundFault) || globalCluster == nil {
			return nil, GlobalClusterStatusDeleted, nil
		}

		if err != nil {
			return nil, "", fmt.Errorf("reading DocumentDB Global Cluster (%s): %w", globalClusterID, err)
		}

		return globalCluster, aws.StringValue(globalCluster.Status), nil
	}
}
