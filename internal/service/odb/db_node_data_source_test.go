// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb_test

import (
	"context"
	"errors"
	"fmt"
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

type testDbNodeDataSourceTest struct {
	exaDisplayNamePrefix             string
	oracleDBNetworkDisplayNamePrefix string
	vmClusterDisplayNamePrefix       string
}

var dbNodeDataSourceTestEntity = testDbNodeDataSourceTest{
	exaDisplayNamePrefix:             "Ofake-exa",
	oracleDBNetworkDisplayNamePrefix: "odb-net",
	vmClusterDisplayNamePrefix:       "Ofake-vmc",
}

// Acceptance test access AWS and cost money to run.
func TestAccODBDBNodeDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatal(err)
		return
	}
	var dbNode odb.GetDbNodeOutput
	dataSourceName := "data.aws_odb_db_node.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             dbNodeDataSourceTestEntity.testAccCheckDBNodeDestroyed(ctx),
		Steps: []resource.TestStep{
			{
				Config: dbNodeDataSourceTestEntity.dbNodeDataSourceBasicConfig(publicKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					dbNodeDataSourceTestEntity.testAccCheckDBNodeExists(ctx, dataSourceName, &dbNode),
				),
			},
		},
	})
}

func (testDbNodeDataSourceTest) testAccCheckDBNodeExists(ctx context.Context, name string, output *odb.GetDbNodeOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.DSNameDBServer, name, errors.New("not found"))
		}
		conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)
		var dbNodeId = rs.Primary.ID
		var attributes = rs.Primary.Attributes
		cloudVmClusterId := attributes["cloud_vm_cluster_id"]
		input := odb.GetDbNodeInput{
			CloudVmClusterId: &cloudVmClusterId,
			DbNodeId:         &dbNodeId,
		}
		resp, err := conn.GetDbNode(ctx, &input)
		if err != nil {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.DSNameDBNode, rs.Primary.ID, err)
		}
		*output = *resp
		return nil
	}
}

func (testDbNodeDataSourceTest) testAccCheckDBNodeDestroyed(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_odb_cloud_vm_cluster" {
				continue
			}
			err := dbNodeDataSourceTestEntity.findVmCluster(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.ODB, create.ErrActionCheckingDestroyed, tfodb.DSNameDBServer, rs.Primary.ID, err)
			}
			return create.Error(names.ODB, create.ErrActionCheckingDestroyed, tfodb.DSNameDBServer, rs.Primary.ID, errors.New("not destroyed"))
		}
		return nil
	}
}

func (testDbNodeDataSourceTest) findVmCluster(ctx context.Context, conn *odb.Client, id string) error {
	input := odb.GetCloudVmClusterInput{
		CloudVmClusterId: aws.String(id),
	}
	output, err := conn.GetCloudVmCluster(ctx, &input)
	if err != nil {
		if errs.IsA[*odbtypes.ResourceNotFoundException](err) {
			return &sdkretry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			}
		}
		return err
	}
	if output == nil || output.CloudVmCluster == nil {
		return tfresource.NewEmptyResultError()
	}
	return nil
}

func (testDbNodeDataSourceTest) dbNodeDataSourceBasicConfig(publicKey string) string {
	vmClusterConfig := dbNodeDataSourceTestEntity.vmClusterBasicConfig(publicKey)

	return fmt.Sprintf(`
%s

data "aws_odb_db_nodes" "test" {
  cloud_vm_cluster_id = aws_odb_cloud_vm_cluster.test.id
}

data "aws_odb_db_node" "test" {
  id                  = data.aws_odb_db_nodes.test.db_nodes[0].id
  cloud_vm_cluster_id = aws_odb_cloud_vm_cluster.test.id
}

`, vmClusterConfig)
}

func (testDbNodeDataSourceTest) vmClusterBasicConfig(publicKey string) string {
	exaInfraDisplayName := sdkacctest.RandomWithPrefix(dbNodeDataSourceTestEntity.exaDisplayNamePrefix)
	oracleDBNetDisplayName := sdkacctest.RandomWithPrefix(dbNodeDataSourceTestEntity.oracleDBNetworkDisplayNamePrefix)
	vmcDisplayName := sdkacctest.RandomWithPrefix(dbNodeDataSourceTestEntity.vmClusterDisplayNamePrefix)
	dsTfCodeVmCluster := fmt.Sprintf(`


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

`, oracleDBNetDisplayName, exaInfraDisplayName, vmcDisplayName, publicKey)
	return dsTfCodeVmCluster
}
