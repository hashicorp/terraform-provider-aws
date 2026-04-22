// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb_test

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/odb"
	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfodb "github.com/hashicorp/terraform-provider-aws/internal/service/odb"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type dbNodesListDataSourceTest struct {
	exadataInfraDisplayNamePrefix    string
	oracleDBNetworkDisplayNamePrefix string
	vmClusterDisplayNamePrefix       string
}

var dbNodesListDataSourceTestEntity = dbNodesListDataSourceTest{
	exadataInfraDisplayNamePrefix:    "Ofake-exa",
	oracleDBNetworkDisplayNamePrefix: "odbn",
	vmClusterDisplayNamePrefix:       "Ofake-vmc",
}

// Acceptance test access AWS and cost money to run.
func TestAccODBDBNodesListDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatal(err)
		return
	}
	var dbNodesList odb.ListDbNodesOutput
	dbNodesListsDataSourceName := "data.aws_odb_db_nodes.test"
	vmClusterListsResourceName := "aws_odb_cloud_vm_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             dbNodesListDataSourceTestEntity.testAccCheckDBNodesDestroyed(ctx),
		Steps: []resource.TestStep{
			{
				Config: dbNodesListDataSourceTestEntity.basicDBNodesListDataSource(publicKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					dbNodesListDataSourceTestEntity.testAccCheckDBNodesListExists(ctx, vmClusterListsResourceName, &dbNodesList),
					resource.TestCheckResourceAttr(dbNodesListsDataSourceName, "aws_odb_db_nodes.db_nodes.#", strconv.Itoa(len(dbNodesList.DbNodes))),
				),
			},
		},
	})
}

func (dbNodesListDataSourceTest) testAccCheckDBNodesListExists(ctx context.Context, name string, output *odb.ListDbNodesOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.DSNameDBNodesList, name, errors.New("not found"))
		}
		conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)
		var vmClusterId = &rs.Primary.ID
		input := odb.ListDbNodesInput{
			CloudVmClusterId: vmClusterId,
		}
		lisOfDbNodes := odb.ListDbNodesOutput{}
		paginator := odb.NewListDbNodesPaginator(conn, &input)
		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				return err
			}
			lisOfDbNodes.DbNodes = append(lisOfDbNodes.DbNodes, page.DbNodes...)
		}
		*output = lisOfDbNodes
		return nil
	}
}

func (dbNodesListDataSourceTest) testAccCheckDBNodesDestroyed(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_odb_cloud_vm_cluster" {
				continue
			}
			_, err := dbNodesListDataSourceTestEntity.findVmCluster(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.ODB, create.ErrActionCheckingDestroyed, tfodb.DSNameDBServersList, rs.Primary.ID, err)
			}
			return create.Error(names.ODB, create.ErrActionCheckingDestroyed, tfodb.DSNameDBServersList, rs.Primary.ID, errors.New("not destroyed"))
		}
		return nil
	}
}

func (dbNodesListDataSourceTest) findVmCluster(ctx context.Context, conn *odb.Client, id string) (*odbtypes.CloudVmCluster, error) {
	input := odb.GetCloudVmClusterInput{
		CloudVmClusterId: aws.String(id),
	}
	output, err := conn.GetCloudVmCluster(ctx, &input)
	if err != nil {
		if errs.IsA[*odbtypes.ResourceNotFoundException](err) {
			return nil, &sdkretry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			}
		}
		return nil, err
	}
	if output == nil || output.CloudVmCluster == nil {
		return nil, tfresource.NewEmptyResultError()
	}
	return output.CloudVmCluster, nil
}

func (dbNodesListDataSourceTest) basicDBNodesListDataSource(publicKey string) string {
	vmCluster := dbNodesListDataSourceTestEntity.vmClusterBasic(publicKey)
	return fmt.Sprintf(`

	%s

data "aws_odb_db_nodes" "test" {
  cloud_vm_cluster_id = aws_odb_cloud_vm_cluster.test.id
}
`, vmCluster)
}

func (dbNodesListDataSourceTest) vmClusterBasic(publicKey string) string {
	odbNetRName := sdkacctest.RandomWithPrefix(dbNodesListDataSourceTestEntity.oracleDBNetworkDisplayNamePrefix)
	exaInfraRName := sdkacctest.RandomWithPrefix(dbNodesListDataSourceTestEntity.exadataInfraDisplayNamePrefix)
	vmcDisplayName := sdkacctest.RandomWithPrefix(dbNodesListDataSourceTestEntity.vmClusterDisplayNamePrefix)
	return fmt.Sprintf(`

resource "aws_odb_network" "test" {
  display_name         = %[1]q
  availability_zone_id = "use1-az6"
  client_subnet_cidr   = "10.2.0.0/24"
  backup_subnet_cidr   = "10.2.1.0/24"
  s3_access            = "DISABLED"
  zero_etl_access      = "DISABLED"
}

resource "aws_odb_cloud_exadata_infrastructure" "test" {
  display_name         = %[2]q
  shape                = "Exadata.X9M"
  storage_count        = 3
  compute_count        = 2
  availability_zone_id = "use1-az6"
  maintenance_window {
    custom_action_timeout_in_mins    = 16
    is_custom_action_timeout_enabled = true
    patching_mode                    = "ROLLING"
    preference                       = "NO_PREFERENCE"
  }
}

data "aws_odb_db_servers" "test" {
  cloud_exadata_infrastructure_id = aws_odb_cloud_exadata_infrastructure.test.id
}

resource "aws_odb_cloud_vm_cluster" "test" {
  display_name                    = %[3]q
  cloud_exadata_infrastructure_id = aws_odb_cloud_exadata_infrastructure.test.id
  cpu_core_count                  = 6
  gi_version                      = "23.0.0.0"
  hostname_prefix                 = "apollo12"
  ssh_public_keys                 = ["%[4]s"]
  odb_network_id                  = aws_odb_network.test.id
  is_local_backup_enabled         = true
  is_sparse_diskgroup_enabled     = true
  license_model                   = "LICENSE_INCLUDED"
  data_storage_size_in_tbs        = 20.0
  db_servers                      = [for db_server in data.aws_odb_db_servers.test.db_servers : db_server.id]
  db_node_storage_size_in_gbs     = 120.0
  memory_size_in_gbs              = 60
  data_collection_options {
    is_diagnostics_events_enabled = false
    is_health_monitoring_enabled  = false
    is_incident_logs_enabled      = false
  }
  tags = {
    "env" = "dev"
  }

}
`, odbNetRName, exaInfraRName, vmcDisplayName, publicKey)
}
