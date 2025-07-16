//Copyright © 2025, Oracle and/or its affiliates. All rights reserved.

package odb_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/odb"
	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfodb "github.com/hashicorp/terraform-provider-aws/internal/service/odb"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
	"testing"
)

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type cloudVmClusterDSTest struct {
	vmClusterDisplayNamePrefix string
	exaInfraDisplayNamePrefix  string
	odbNetDisplayNamePrefix    string
}

var vmClusterTestDS = cloudVmClusterDSTest{
	vmClusterDisplayNamePrefix: "Ofake-vmc",
	exaInfraDisplayNamePrefix:  "Ofake-exa-infra",
	odbNetDisplayNamePrefix:    "odb-net",
}

func TestAccODBCloudVmClusterDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var cloudvmcluster odbtypes.CloudVmCluster
	vmcDisplayName := sdkacctest.RandomWithPrefix(vmClusterTestDS.vmClusterDisplayNamePrefix)
	dataSourceName := "data.aws_odb_cloud_vm_cluster.test"
	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatal(err)
		return
	}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			//	testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             vmClusterTestDS.testAccCheckCloudVmClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: vmClusterTestDS.cloudVmClusterWithHardcoded("odbnet_c91byo6y6m", "exa_ji5quxxzn9", vmcDisplayName, publicKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					vmClusterTestDS.testAccCheckCloudVmClusterExists(ctx, dataSourceName, &cloudvmcluster),
					resource.TestCheckResourceAttr(dataSourceName, "display_name", vmcDisplayName),
				),
			},
		},
	})
}

func (cloudVmClusterDSTest) testAccCheckCloudVmClusterDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_odb_cloud_vm_cluster" {
				continue
			}

			_, err := tfodb.FindCloudVmClusterForResourceByID(ctx, conn, rs.Primary.ID)
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.ODB, create.ErrActionCheckingDestroyed, tfodb.ResNameCloudVmCluster, rs.Primary.ID, err)
			}

			return create.Error(names.ODB, create.ErrActionCheckingDestroyed, tfodb.ResNameCloudVmCluster, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func (cloudVmClusterDSTest) testAccCheckCloudVmClusterExists(ctx context.Context, name string, cloudvmcluster *odbtypes.CloudVmCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.ResNameCloudVmCluster, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.ResNameCloudVmCluster, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)

		resp, err := tfodb.FindCloudVmClusterForResourceByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.ResNameCloudVmCluster, rs.Primary.ID, err)
		}

		*cloudvmcluster = *resp

		return nil
	}
}

func (cloudVmClusterDSTest) testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)

	input := &odb.ListCloudVmClustersInput{}

	_, err := conn.ListCloudVmClusters(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func (cloudVmClusterDSTest) cloudVmClusterWithHardcoded(odbNet, exaInfra, displayName, sshKey string) string {
	dsTfCodeVmCluster := fmt.Sprintf(`
resource "aws_odb_cloud_vm_cluster" "test" {
  odb_network_id                  	= %[1]q
  cloud_exadata_infrastructure_id 	= %[2]q
  display_name             			= %[3]q
  ssh_public_keys                 	= [%[4]q]
  cpu_core_count                  = 6
  gi_version                	  = "23.0.0.0"
  hostname_prefix                 = "apollo12"
  is_local_backup_enabled         = true
  is_sparse_diskgroup_enabled     = true
  license_model                   = "LICENSE_INCLUDED"
  data_storage_size_in_tbs        = 20.0
  db_servers					  = [ "dbs_7ecm4wbjxy","dbs_uy5wmaqk6s"]
  db_node_storage_size_in_gbs     = 120.0
  memory_size_in_gbs              = 60
  tags = {
  	  "env"= "dev"
  }

}

data "aws_odb_cloud_vm_cluster" "test" {
  id             = aws_odb_cloud_vm_cluster.test.id
}
`, odbNet, exaInfra, displayName, sshKey)
	//fmt.Println(dsTfCodeVmCluster)
	return dsTfCodeVmCluster
}
