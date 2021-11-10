package docdb

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/service/docdb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	GlobalClusterStatusAvailable = "available"
	GlobalClusterStatusDeleted   = "deleted"
	GlobalClusterStatusDeleting  = "deleting"
	GlobalClusterStatusModifying = "modifying"
	GlobalClusterStatusUpgrading = "upgrading"
)

func WaitForGlobalClusterDeletion(ctx context.Context, conn *docdb.DocDB, globalClusterID string) error {
	stateConf := &resource.StateChangeConf{
		Pending:        []string{GlobalClusterStatusAvailable, GlobalClusterStatusDeleting},
		Target:         []string{GlobalClusterStatusDeleted},
		Refresh:        statusGlobalClusterRefreshFunc(ctx, conn, globalClusterID),
		Timeout:        10 * time.Minute,
		NotFoundChecks: 1,
	}

	log.Printf("[DEBUG] Waiting for DocDB Global Cluster (%s) deletion", globalClusterID)
	_, err := stateConf.WaitForStateContext(ctx)

	if tfresource.NotFound(err) {
		return nil
	}

	return err
}