// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package odb_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/odb"
	"github.com/aws/aws-sdk-go-v2/service/odb/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfodb "github.com/hashicorp/terraform-provider-aws/internal/service/odb"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type dbNodesListDataSourceTest struct {
	exadataInfraDisplayNamePrefix    string
	oracleDBNetworkDisplayNamePrefix string
	vmClusterDisplayNamePrefix       string
}

var dbNodesListDataSourceTestEntity = dbNodesListDataSourceTest{
	exadataInfraDisplayNamePrefix:    "Ofake",
	oracleDBNetworkDisplayNamePrefix: "odbn",
	vmClusterDisplayNamePrefix:       "Ofake",
}

// Acceptance test access AWS and cost money to run.
func TestAccODBDbNodesListDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbnodeslist odb.DescribeDbNodesListResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_odb_db_nodes_list.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ODBEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDbNodesListDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: dbNodesListDataSourceTestEntity.testAccDbNodesListDataSourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					dbNodesListDataSourceTestEntity.testAccCheckDbNodesListExists(ctx, dataSourceName, &dbnodeslist),
					resource.TestCheckResourceAttr(dataSourceName, "auto_minor_version_upgrade", "false"),
					resource.TestCheckResourceAttrSet(dataSourceName, "maintenance_window_start_time.0.day_of_week"),
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "user.*", map[string]string{
						"console_access": "false",
						"groups.#":       "0",
						"username":       "Test",
						"password":       "TestTest1234",
					}),
					acctest.MatchResourceAttrRegionalARN(ctx, dataSourceName, names.AttrARN, "odb", regexache.MustCompile(`dbnodeslist:.+$`)),
				),
			},
		},
	})
}

func (dbNodesListDataSourceTest) testAccDbNodesListDataSourceConfig_basic() string {
	vmCluster := dbNodeDataSourceTestEntity.vmClusterBasicConfig()
	return fmt.Sprintf(`

	%s

data "aws_odb_db_nodes_list" "test" {
  cloud_vm_cluster_id = aws_odb_cloud_vm_cluster.test.id
}
`, vmCluster)
}

func (dbNodesListDataSourceTest) vmClusterBasic() string {

	odbNetRName := sdkacctest.RandomWithPrefix(dbNodeDataSourceTestEntity.oracleDBNetworkDisplayNamePrefix)
	exaInfraRName := sdkacctest.RandomWithPrefix(dbNodeDataSourceTestEntity.exaDisplayNamePrefix)
	vmcDisplayName := sdkacctest.RandomWithPrefix(dbNodeDataSourceTestEntity.vmClusterDisplayNamePrefix)

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
  display_name         = %[1]q
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

data "aws_odb_db_servers_list" "test" {
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
  db_servers                      = [for db_server in data.aws_odb_db_servers_list.test.db_servers : db_server.id]
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
`, odbNetRName, exaInfraRName, vmcDisplayName)
}
