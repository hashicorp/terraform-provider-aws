// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package odb_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/odb"
	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfodb "github.com/hashicorp/terraform-provider-aws/internal/service/odb"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
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
	var avmc1 odbtypes.CloudAutonomousVmCluster
	dataSourceName := "data.aws_odb_cloud_autonomous_vm_cluster.test"
	avmcDisplayName := sdkacctest.RandomWithPrefix(autonomousVMClusterDSTestEntity.autonomousVmClusterDisplayNamePrefix)
	//odbDisplayNamePrefix := sdkacctest.RandomWithPrefix(autonomousVMClusterDSTestEntity.odbNetDisplayNamePrefix)
	//exaDisplayName := sdkacctest.RandomWithPrefix(autonomousVMClusterDSTestEntity.exaInfraDisplayNamePrefix)
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
				Config: autonomousVMClusterDSTestEntity.basicHardcodedAVmCluster("exa_ji5quxxzn9", "odbnet_c91byo6y6m", avmcDisplayName),
				Check: resource.ComposeAggregateTestCheckFunc(
					autonomousVMClusterDSTestEntity.checkCloudAutonomousVmClusterExists(ctx, dataSourceName, &avmc1),
					resource.TestCheckResourceAttr(dataSourceName, "display_name", avmcDisplayName),
				),
			},
		},
	})
}
func (autonomousVMClusterDSTest) checkCloudAutonomousVmClusterExists(ctx context.Context, name string, cloudAutonomousVMCluster *odbtypes.CloudAutonomousVmCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.ResNameCloudAutonomousVmCluster, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.ResNameCloudAutonomousVmCluster, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)
		fmt.Println("")
		resp, err := autonomousVMClusterResourceTest{}.findVMC(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.ResNameCloudAutonomousVmCluster, rs.Primary.ID, err)
		}

		*cloudAutonomousVMCluster = *resp

		return nil
	}
}
func (autonomousVMClusterDSTest) testAccCheckCloudAutonomousVmClusterDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_odb_cloud_autonomous_vm_cluster" {
				continue
			}

			_, err := tfodb.FindCloudAutonomousVmClusterByID(ctx, conn, rs.Primary.ID)
			if tfresource.NotFound(err) {
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

	input := &odb.ListCloudAutonomousVmClustersInput{}

	_, err := conn.ListCloudAutonomousVmClusters(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func (autonomousVMClusterDSTest) basicHardcodedAVmCluster(exaInfra, odbNetwork, avmcDisplayName string) string {
	avmcDataSource := fmt.Sprintf(`
resource "aws_odb_cloud_autonomous_vm_cluster" "test" {
    cloud_exadata_infrastructure_id         = %[1]q
  	odb_network_id                          = %[2]q
	display_name             				= %[3]q
  	autonomous_data_storage_size_in_tbs     = 5
  	memory_per_oracle_compute_unit_in_gbs   = 2
  	total_container_databases               = 1
  	cpu_core_count_per_node                 = 4
    db_servers								   = ["dbs_7ecm4wbjxy","dbs_uy5wmaqk6s"]
    scan_listener_port_tls = 1521
    scan_listener_port_non_tls = 2484
    maintenance_window = {
		 preference = "NO_PREFERENCE"
    }

}
data "aws_odb_cloud_autonomous_vm_cluster" "test" {
  id             = aws_odb_cloud_autonomous_vm_cluster.test.id

}
`, exaInfra, odbNetwork, avmcDisplayName)
	return avmcDataSource
}
func (autonomousVMClusterDSTest) avmcBasic(exaInfra, odbNetwork, avmcDisplayName string) string {
	res := fmt.Sprintf(`
resource "aws_odb_cloud_exadata_infrastructure" "test" {
  display_name          = %[1]q
  shape             	= "Exadata.X9M"
  storage_count      	= 3
  compute_count         = 2
  availability_zone_id 	= "use1-az6"
  customer_contacts_to_send_to_oci = ["abc@example.com"]
  
}

resource "aws_odb_network" "test" {
  		display_name          = %[2]q
  		availability_zone_id = "use1-az6"
  		client_subnet_cidr   = "10.2.0.0/24"
  		backup_subnet_cidr   = "10.2.1.0/24"
  		s3_access = "DISABLED"
  		zero_etl_access = "DISABLED"
	}

resource "aws_odb_cloud_autonomous_vm_cluster" "test" {
  	display_name             				= %[3]q
    cloud_exadata_infrastructure_id         = aws_odb_cloud_exadata_infrastructure.test.id
  	odb_network_id                          = aws_odb_network.test.id
  	autonomous_data_storage_size_in_tbs     = 5
  	memory_per_oracle_compute_unit_in_gbs   = 2
  	total_container_databases               = 1
  	cpu_core_count_per_node                 = 4

}
`, exaInfra, odbNetwork, avmcDisplayName)

	return res
}
