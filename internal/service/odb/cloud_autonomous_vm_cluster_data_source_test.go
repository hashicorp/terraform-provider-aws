// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/odb"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfodb "github.com/hashicorp/terraform-provider-aws/internal/service/odb"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type autonomousVMClusterDSTest struct {
	exaInfraDisplayNamePrefix            string
	odbNetDisplayNamePrefix              string
	autonomousVmClusterDisplayNamePrefix string
}

var autonomousVMClusterDSTestEntity = autonomousVMClusterDSTest{
	exaInfraDisplayNamePrefix:            "Ofake-exa",
	odbNetDisplayNamePrefix:              "odb-net",
	autonomousVmClusterDisplayNamePrefix: "Ofake-avmc",
}

func TestAccODBCloudAutonomousVmClusterDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	avmcResource := "aws_odb_cloud_autonomous_vm_cluster.test"
	avmcDataSource := "data.aws_odb_cloud_autonomous_vm_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			autonomousVMClusterDSTestEntity.testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             autonomousVMClusterDSTestEntity.testAccCheckCloudAutonomousVmClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: autonomousVMClusterDSTestEntity.avmcBasic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(avmcResource, names.AttrID, avmcDataSource, names.AttrID),
				),
			},
		},
	})
}

func (autonomousVMClusterDSTest) testAccCheckCloudAutonomousVmClusterDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_odb_cloud_autonomous_vm_cluster" {
				continue
			}

			_, err := tfodb.FindCloudAutonomousVmClusterByID(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.ODB, create.ErrActionCheckingDestroyed, tfodb.ResNameCloudAutonomousVmCluster, rs.Primary.ID, err)
			}

			return create.Error(names.ODB, create.ErrActionCheckingDestroyed, tfodb.ResNameCloudAutonomousVmCluster, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}
func (autonomousVMClusterDSTest) testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)
	input := odb.ListCloudAutonomousVmClustersInput{}
	_, err := conn.ListCloudAutonomousVmClusters(ctx, &input)
	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func (autonomousVMClusterDSTest) avmcBasic() string {
	exaInfraDisplayName := sdkacctest.RandomWithPrefix(autonomousVMClusterDSTestEntity.exaInfraDisplayNamePrefix)
	odbNetworkDisplayName := sdkacctest.RandomWithPrefix(autonomousVMClusterDSTestEntity.odbNetDisplayNamePrefix)
	avmcDisplayName := sdkacctest.RandomWithPrefix(autonomousVMClusterDSTestEntity.autonomousVmClusterDisplayNamePrefix)
	domain := acctest.RandomDomainName()
	emailAddress := acctest.RandomEmailAddress(domain)
	exaInfraRes := autonomousVMClusterDSTestEntity.exaInfra(exaInfraDisplayName, emailAddress)
	odbNetRes := autonomousVMClusterDSTestEntity.oracleDBNetwork(odbNetworkDisplayName)
	res := fmt.Sprintf(`
%s

%s

data "aws_odb_db_servers" "test" {
  cloud_exadata_infrastructure_id = aws_odb_cloud_exadata_infrastructure.test.id
}

resource "aws_odb_cloud_autonomous_vm_cluster" "test" {
  cloud_exadata_infrastructure_id       = aws_odb_cloud_exadata_infrastructure.test.id
  odb_network_id                        = aws_odb_network.test.id
  display_name                          = %[3]q
  autonomous_data_storage_size_in_tbs   = 5
  memory_per_oracle_compute_unit_in_gbs = 2
  total_container_databases             = 1
  cpu_core_count_per_node               = 40
  license_model                         = "LICENSE_INCLUDED"
  db_servers                            = [for db_server in data.aws_odb_db_servers.test.db_servers : db_server.id]
  scan_listener_port_tls                = 8561
  scan_listener_port_non_tls            = 1024
  maintenance_window {
    preference = "NO_PREFERENCE"
  }
  tags = {
    "env" = "dev"
  }

}


data "aws_odb_cloud_autonomous_vm_cluster" "test" {
  id = aws_odb_cloud_autonomous_vm_cluster.test.id

}
`, exaInfraRes, odbNetRes, avmcDisplayName)

	return res
}

func (autonomousVMClusterDSTest) oracleDBNetwork(odbNetName string) string {
	networkRes := fmt.Sprintf(`




resource "aws_odb_network" "test" {
  display_name         = %[1]q
  availability_zone_id = "use1-az6"
  client_subnet_cidr   = "10.2.0.0/24"
  backup_subnet_cidr   = "10.2.1.0/24"
  s3_access            = "DISABLED"
  zero_etl_access      = "DISABLED"
}


`, odbNetName)
	return networkRes
}

func (autonomousVMClusterDSTest) exaInfra(exaInfraName, emailAddress string) string {
	exaInfraRes := fmt.Sprintf(`




resource "aws_odb_cloud_exadata_infrastructure" "test" {
  display_name                     = %[1]q
  shape                            = "Exadata.X9M"
  storage_count                    = 3
  compute_count                    = 2
  availability_zone_id             = "use1-az6"
  customer_contacts_to_send_to_oci = ["%[2]s"]
  maintenance_window {
    custom_action_timeout_in_mins    = 16
    is_custom_action_timeout_enabled = true
    patching_mode                    = "ROLLING"
    preference                       = "NO_PREFERENCE"
  }
}


`, exaInfraName, emailAddress)
	return exaInfraRes
}
