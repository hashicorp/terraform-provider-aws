// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/odb"
	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
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

type cloudVmClusterResourceTest struct {
	vmClusterDisplayNamePrefix string
	exaInfraDisplayNamePrefix  string
	odbNetDisplayNamePrefix    string
}

var vmClusterTestEntity = cloudVmClusterResourceTest{
	vmClusterDisplayNamePrefix: "Ofake-vmc",
	exaInfraDisplayNamePrefix:  "Ofake-exa-infra",
	odbNetDisplayNamePrefix:    "odb-net",
}

func TestAccODBCloudVmCluster_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	var cloudvmcluster odbtypes.CloudVmCluster
	vmcDisplayName := sdkacctest.RandomWithPrefix(vmClusterTestEntity.vmClusterDisplayNamePrefix)
	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatal(err)
		return
	}
	resourceName := "aws_odb_cloud_vm_cluster.test"
	basicConfig, _ := vmClusterTestEntity.testAccCloudVmClusterConfigBasic(vmcDisplayName, publicKey)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			vmClusterTestEntity.testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             vmClusterTestEntity.testAccCheckCloudVmClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: basicConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					vmClusterTestEntity.testAccCheckCloudVmClusterExists(ctx, resourceName, &cloudvmcluster),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccODBCloudVmCluster_allParams(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	var cloudvmcluster odbtypes.CloudVmCluster
	vmcClusterDisplayName := sdkacctest.RandomWithPrefix(vmClusterTestEntity.vmClusterDisplayNamePrefix)
	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatal(err)
		return
	}
	resourceName := "aws_odb_cloud_vm_cluster.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			vmClusterTestEntity.testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             vmClusterTestEntity.testAccCheckCloudVmClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: vmClusterTestEntity.cloudVmClusterWithAllParameters(vmcClusterDisplayName, publicKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					vmClusterTestEntity.testAccCheckCloudVmClusterExists(ctx, resourceName, &cloudvmcluster),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccODBCloudVmCluster_taggingTest(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	var cloudvmcluster1 odbtypes.CloudVmCluster
	var cloudvmcluster2 odbtypes.CloudVmCluster
	vmcDisplayName := sdkacctest.RandomWithPrefix(vmClusterTestEntity.vmClusterDisplayNamePrefix)
	resourceName := "aws_odb_cloud_vm_cluster.test"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatal(err)
		return
	}
	vmcNoTag, vmcWithTag := vmClusterTestEntity.testAccCloudVmClusterConfigBasic(vmcDisplayName, publicKey)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			vmClusterTestEntity.testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             vmClusterTestEntity.testAccCheckCloudVmClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: vmcNoTag,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.ComposeTestCheckFunc(func(state *terraform.State) error {
						return nil
					}),
					vmClusterTestEntity.testAccCheckCloudVmClusterExists(ctx, resourceName, &cloudvmcluster1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: vmcWithTag,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.env", "dev"),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "bar"),
					vmClusterTestEntity.testAccCheckCloudVmClusterExists(ctx, resourceName, &cloudvmcluster2),
					resource.ComposeTestCheckFunc(func(state *terraform.State) error {
						if strings.Compare(*(cloudvmcluster1.CloudVmClusterId), *(cloudvmcluster2.CloudVmClusterId)) != 0 {
							return errors.New("Should  not create a new cloud vm cluster for tag update")
						}
						return nil
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccODBCloudVmCluster_giVersionTag(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	var cloudvmcluster1 odbtypes.CloudVmCluster
	vmcDisplayName := sdkacctest.RandomWithPrefix(vmClusterTestEntity.vmClusterDisplayNamePrefix)
	resourceName := "aws_odb_cloud_vm_cluster.test"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatal(err)
		return
	}
	vmcWithGiVersionTag := vmClusterTestEntity.cloudVmClusterConfigWithGiVersionTag(vmcDisplayName, publicKey)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			vmClusterTestEntity.testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             vmClusterTestEntity.testAccCheckCloudVmClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: vmcWithGiVersionTag,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.ComposeTestCheckFunc(func(state *terraform.State) error {
						return nil
					}),
					resource.TestCheckResourceAttr(resourceName, "gi_version_computed", "26.0.0.0"),
					vmClusterTestEntity.testAccCheckCloudVmClusterExists(ctx, resourceName, &cloudvmcluster1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccODBCloudVmCluster_real(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	var cloudvmcluster1 odbtypes.CloudVmCluster
	var cloudvmcluster2 odbtypes.CloudVmCluster
	vmcDisplayName := sdkacctest.RandomWithPrefix("tf-real")
	resourceName := "aws_odb_cloud_vm_cluster.test"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatal(err)
		return
	}
	vmcWithoutTag, vmcWithTag := vmClusterTestEntity.cloudVmClusterReal(vmcDisplayName, publicKey)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			vmClusterTestEntity.testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             vmClusterTestEntity.testAccCheckCloudVmClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: vmcWithoutTag,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.ComposeTestCheckFunc(func(state *terraform.State) error {
						return nil
					}),
					vmClusterTestEntity.testAccCheckCloudVmClusterExists(ctx, resourceName, &cloudvmcluster1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: vmcWithTag,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.env", "dev"),
					vmClusterTestEntity.testAccCheckCloudVmClusterExists(ctx, resourceName, &cloudvmcluster2),
					resource.ComposeTestCheckFunc(func(state *terraform.State) error {
						if strings.Compare(*(cloudvmcluster1.CloudVmClusterId), *(cloudvmcluster2.CloudVmClusterId)) != 0 {
							return errors.New("Should  not create a new cloud vm cluster for tag update")
						}
						return nil
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccODBCloudVmCluster_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	var cloudvmcluster odbtypes.CloudVmCluster
	vmClusterDisplayName := sdkacctest.RandomWithPrefix(vmClusterTestEntity.vmClusterDisplayNamePrefix)
	resourceName := "aws_odb_cloud_vm_cluster.test"
	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatal(err)
		return
	}
	vmcBasicConfig, _ := vmClusterTestEntity.testAccCloudVmClusterConfigBasic(vmClusterDisplayName, publicKey)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			vmClusterTestEntity.testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             vmClusterTestEntity.testAccCheckCloudVmClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: vmcBasicConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					vmClusterTestEntity.testAccCheckCloudVmClusterExists(ctx, resourceName, &cloudvmcluster),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfodb.ResourceCloudVmCluster, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccODBCloudVmCluster_usingARN(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	var cloudvmcluster1 odbtypes.CloudVmCluster
	var cloudvmcluster2 odbtypes.CloudVmCluster
	vmcDisplayName := sdkacctest.RandomWithPrefix("Ofake")
	resourceName := "aws_odb_cloud_vm_cluster.test"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatal(err)
		return
	}
	vmcWithoutTag, vmcWithTag := vmClusterTestEntity.cloudVmClusterByARN(vmcDisplayName, publicKey)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			vmClusterTestEntity.testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             vmClusterTestEntity.testAccCheckCloudVmClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: vmcWithoutTag,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.ComposeTestCheckFunc(func(state *terraform.State) error {
						return nil
					}),
					vmClusterTestEntity.testAccCheckCloudVmClusterExists(ctx, resourceName, &cloudvmcluster1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: vmcWithTag,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.env", "dev"),
					vmClusterTestEntity.testAccCheckCloudVmClusterExists(ctx, resourceName, &cloudvmcluster2),
					resource.ComposeTestCheckFunc(func(state *terraform.State) error {
						if strings.Compare(*(cloudvmcluster1.CloudVmClusterId), *(cloudvmcluster2.CloudVmClusterId)) != 0 {
							return errors.New("Should  not create a new cloud vm cluster for tag update")
						}
						return nil
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccODBCloudVmCluster_variables(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	vmcDisplayName := sdkacctest.RandomWithPrefix(vmClusterTestEntity.vmClusterDisplayNamePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			vmClusterTestEntity.testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             vmClusterTestEntity.testAccCheckCloudVmClusterDestroy(ctx),
		Steps: []resource.TestStep{
			// nosemgrep:ci.semgrep.acctest.checks.replace-planonly-checks
			{
				Config:             testAccCloudVmClusterConfig_useVariables(vmcDisplayName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func (cloudVmClusterResourceTest) testAccCheckCloudVmClusterDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_odb_cloud_vm_cluster" {
				continue
			}
			_, err := tfodb.FindCloudVmClusterForResourceByID(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
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

func (cloudVmClusterResourceTest) testAccCheckCloudVmClusterExists(ctx context.Context, name string, cloudvmcluster *odbtypes.CloudVmCluster) resource.TestCheckFunc {
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

func (cloudVmClusterResourceTest) testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)
	input := odb.ListCloudVmClustersInput{}
	_, err := conn.ListCloudVmClusters(ctx, &input)
	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCloudVmClusterConfig_useVariables(rName string) string {
	return fmt.Sprintf(`
variable cloud_exadata_infrastructure_id {
  default     = "exa_gjrmtxl4qk"
  type        = string
  description = "ODB Exadata Infrastructure Resource ID"

}
variable odb_network_id {
  default     = "odbnet_3l9st3litg"
  type        = string
  description = "ODB Network"
}

resource "aws_odb_cloud_vm_cluster" "test" {
  display_name                    = %[1]q
  cloud_exadata_infrastructure_id = var.cloud_exadata_infrastructure_id
  cpu_core_count                  = 6
  gi_version                      = "23.0.0.0"
  hostname_prefix                 = "apollo12"
  ssh_public_keys                 = ["public-ssh-key"]
  odb_network_id                  = var.odb_network_id
  is_local_backup_enabled         = true
  is_sparse_diskgroup_enabled     = true
  license_model                   = "LICENSE_INCLUDED"
  data_storage_size_in_tbs        = 20.0
  db_servers                      = ["db-server-1", "db-server-2"]
  db_node_storage_size_in_gbs     = 120.0
  memory_size_in_gbs              = 60
  data_collection_options {
    is_diagnostics_events_enabled = false
    is_health_monitoring_enabled  = false
    is_incident_logs_enabled      = false
  }
}
`, rName)
}
func (cloudVmClusterResourceTest) testAccCloudVmClusterConfigBasic(vmClusterDisplayName, sshKey string) (string, string) {
	exaInfraDisplayName := sdkacctest.RandomWithPrefix(vmClusterTestEntity.exaInfraDisplayNamePrefix)
	odbNetDisplayName := sdkacctest.RandomWithPrefix(vmClusterTestEntity.odbNetDisplayNamePrefix)
	exaInfra := vmClusterTestEntity.exaInfra(exaInfraDisplayName)
	odbNet := vmClusterTestEntity.oracleDBNetwork(odbNetDisplayName)
	vmcNoTag := fmt.Sprintf(`

%s

%s

data "aws_odb_db_servers" "test" {
  cloud_exadata_infrastructure_id = aws_odb_cloud_exadata_infrastructure.test.id
}

resource "aws_odb_cloud_vm_cluster" "test" {
  display_name                    = %[3]q
  cloud_exadata_infrastructure_id = aws_odb_cloud_exadata_infrastructure.test.id
  cpu_core_count                  = 16
  gi_version                      = "26.0.0.0"
  hostname_prefix                 = "apollo-12"
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

}
`, exaInfra, odbNet, vmClusterDisplayName, sshKey)

	vmcWithTag := fmt.Sprintf(`

%s

%s

data "aws_odb_db_servers" "test" {
  cloud_exadata_infrastructure_id = aws_odb_cloud_exadata_infrastructure.test.id
}

resource "aws_odb_cloud_vm_cluster" "test" {
  display_name                    = %[3]q
  cloud_exadata_infrastructure_id = aws_odb_cloud_exadata_infrastructure.test.id
  cpu_core_count                  = 16
  gi_version                      = "26.0.0.0"
  hostname_prefix                 = "apollo-12"
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
    "foo" = "bar"
  }

}
`, exaInfra, odbNet, vmClusterDisplayName, sshKey)
	return vmcNoTag, vmcWithTag
}

func (cloudVmClusterResourceTest) cloudVmClusterWithAllParameters(vmClusterDisplayName, sshKey string) string {
	exaInfraDisplayName := sdkacctest.RandomWithPrefix(vmClusterTestEntity.exaInfraDisplayNamePrefix)
	odbNetDisplayName := sdkacctest.RandomWithPrefix(vmClusterTestEntity.odbNetDisplayNamePrefix)
	exaInfra := vmClusterTestEntity.exaInfra(exaInfraDisplayName)
	odbNet := vmClusterTestEntity.oracleDBNetwork(odbNetDisplayName)

	res := fmt.Sprintf(`

%s

%s


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
  cluster_name                    = "julia-13"
  timezone                        = "UTC"
  scan_listener_port_tcp          = 1521
  tags = {
    "env" = "dev"
  }
  data_collection_options {
    is_diagnostics_events_enabled = true
    is_health_monitoring_enabled  = true
    is_incident_logs_enabled      = true
  }
}
`, exaInfra, odbNet, vmClusterDisplayName, sshKey)
	return res
}

func (cloudVmClusterResourceTest) exaInfra(rName string) string {
	resource := fmt.Sprintf(`
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
`, rName)
	return resource
}

func (cloudVmClusterResourceTest) oracleDBNetwork(rName string) string {
	resource := fmt.Sprintf(`
resource "aws_odb_network" "test" {
  display_name         = %[1]q
  availability_zone_id = "use1-az6"
  client_subnet_cidr   = "10.2.0.0/24"
  backup_subnet_cidr   = "10.2.1.0/24"
  s3_access            = "DISABLED"
  zero_etl_access      = "DISABLED"
}
`, rName)
	return resource
}

func (cloudVmClusterResourceTest) cloudVmClusterReal(vmClusterDisplayName, sshKey string) (string, string) {
	exaInfraDisplayName := sdkacctest.RandomWithPrefix("tf-real")
	odbNetDisplayName := sdkacctest.RandomWithPrefix(vmClusterTestEntity.odbNetDisplayNamePrefix)
	exaInfra := vmClusterTestEntity.exaInfra(exaInfraDisplayName)
	odbNet := vmClusterTestEntity.oracleDBNetwork(odbNetDisplayName)
	vmClusterResourceNoTag := fmt.Sprintf(`

%s

%s

data "aws_odb_db_servers" "test" {
  cloud_exadata_infrastructure_id = aws_odb_cloud_exadata_infrastructure.test.id
}

resource "aws_odb_cloud_vm_cluster" "test" {
  display_name                    = %[3]q
  cloud_exadata_infrastructure_id = aws_odb_cloud_exadata_infrastructure.test.id
  cpu_core_count                  = 16
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

}
`, exaInfra, odbNet, vmClusterDisplayName, sshKey)

	vmClusterResourceWithTag := fmt.Sprintf(`

%s

%s

data "aws_odb_db_servers" "test" {
  cloud_exadata_infrastructure_id = aws_odb_cloud_exadata_infrastructure.test.id
}

resource "aws_odb_cloud_vm_cluster" "test" {
  display_name                    = %[3]q
  cloud_exadata_infrastructure_id = aws_odb_cloud_exadata_infrastructure.test.id
  cpu_core_count                  = 16
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
`, exaInfra, odbNet, vmClusterDisplayName, sshKey)

	return vmClusterResourceNoTag, vmClusterResourceWithTag
}

func (cloudVmClusterResourceTest) cloudVmClusterByARN(vmClusterDisplayName, sshKey string) (string, string) {
	exaInfraDisplayName := sdkacctest.RandomWithPrefix("Ofake-exa")
	odbNetDisplayName := sdkacctest.RandomWithPrefix(vmClusterTestEntity.odbNetDisplayNamePrefix)
	exaInfra := vmClusterTestEntity.exaInfra(exaInfraDisplayName)
	odbNet := vmClusterTestEntity.oracleDBNetwork(odbNetDisplayName)
	vmClusterResourceNoTag := fmt.Sprintf(`




%s

%s



data "aws_odb_db_servers" "test" {
  cloud_exadata_infrastructure_id = aws_odb_cloud_exadata_infrastructure.test.arn
}

resource "aws_odb_cloud_vm_cluster" "test" {
  display_name                     = %[3]q
  cloud_exadata_infrastructure_arn = aws_odb_cloud_exadata_infrastructure.test.arn
  cpu_core_count                   = 16
  gi_version                       = "26.0.0.0"
  hostname_prefix                  = "apollo12"
  ssh_public_keys                  = ["%[4]s"]
  odb_network_arn                  = aws_odb_network.test.arn
  is_local_backup_enabled          = true
  is_sparse_diskgroup_enabled      = true
  license_model                    = "LICENSE_INCLUDED"
  data_storage_size_in_tbs         = 20.0
  db_servers                       = [for db_server in data.aws_odb_db_servers.test.db_servers : db_server.id]
  db_node_storage_size_in_gbs      = 120.0
  memory_size_in_gbs               = 60
  data_collection_options {
    is_diagnostics_events_enabled = false
    is_health_monitoring_enabled  = false
    is_incident_logs_enabled      = false
  }

}
`, exaInfra, odbNet, vmClusterDisplayName, sshKey)

	vmClusterResourceWithTag := fmt.Sprintf(`


%s

%s




data "aws_odb_db_servers" "test" {
  cloud_exadata_infrastructure_id = aws_odb_cloud_exadata_infrastructure.test.arn
}

resource "aws_odb_cloud_vm_cluster" "test" {
  display_name                     = %[3]q
  cloud_exadata_infrastructure_arn = aws_odb_cloud_exadata_infrastructure.test.arn
  cpu_core_count                   = 16
  gi_version                       = "26.0.0.0"
  hostname_prefix                  = "apollo12"
  ssh_public_keys                  = ["%[4]s"]
  odb_network_arn                  = aws_odb_network.test.arn
  is_local_backup_enabled          = true
  is_sparse_diskgroup_enabled      = true
  license_model                    = "LICENSE_INCLUDED"
  data_storage_size_in_tbs         = 20.0
  db_servers                       = [for db_server in data.aws_odb_db_servers.test.db_servers : db_server.id]
  db_node_storage_size_in_gbs      = 120.0
  memory_size_in_gbs               = 60
  data_collection_options {
    is_diagnostics_events_enabled = false
    is_health_monitoring_enabled  = false
    is_incident_logs_enabled      = false
  }
  tags = {
    "env" = "dev"
  }

}
`, exaInfra, odbNet, vmClusterDisplayName, sshKey)

	return vmClusterResourceNoTag, vmClusterResourceWithTag
}

func (cloudVmClusterResourceTest) cloudVmClusterConfigWithGiVersionTag(vmClusterDisplayName, sshKey string) string {
	exaInfraDisplayName := sdkacctest.RandomWithPrefix(vmClusterTestEntity.exaInfraDisplayNamePrefix)
	odbNetDisplayName := sdkacctest.RandomWithPrefix(vmClusterTestEntity.odbNetDisplayNamePrefix)
	exaInfra := vmClusterTestEntity.exaInfra(exaInfraDisplayName)
	odbNet := vmClusterTestEntity.oracleDBNetwork(odbNetDisplayName)
	vmcWithGiVersionTag := fmt.Sprintf(`


%s

%s



data "aws_odb_db_servers" "test" {
  cloud_exadata_infrastructure_id = aws_odb_cloud_exadata_infrastructure.test.id
}

resource "aws_odb_cloud_vm_cluster" "test" {
  display_name                    = %[3]q
  cloud_exadata_infrastructure_id = aws_odb_cloud_exadata_infrastructure.test.id
  cpu_core_count                  = 16
  gi_version                      = "23.0.0.0"
  hostname_prefix                 = "apollo-12"
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
    "odb:input_gi_version" = "23.0.0.0"
    "foo"                  = "bar"
  }

}
`, exaInfra, odbNet, vmClusterDisplayName, sshKey)

	return vmcWithGiVersionTag
}
