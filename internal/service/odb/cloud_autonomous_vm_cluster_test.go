// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/odb"
	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfodb "github.com/hashicorp/terraform-provider-aws/internal/service/odb"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type autonomousVMClusterResourceTest struct {
	exaInfraDisplayNamePrefix            string
	odbNetDisplayNamePrefix              string
	autonomousVmClusterDisplayNamePrefix string
}

var autonomousVMClusterResourceTestEntity = autonomousVMClusterResourceTest{
	exaInfraDisplayNamePrefix:            "Ofake-exa",
	odbNetDisplayNamePrefix:              "oracleDB-net",
	autonomousVmClusterDisplayNamePrefix: "Ofake-avmc",
}

func TestAccODBCloudAutonomousVmCluster_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var cloudAVMC odbtypes.CloudAutonomousVmCluster

	resourceName := "aws_odb_cloud_autonomous_vm_cluster.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			autonomousVMClusterResourceTestEntity.testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             autonomousVMClusterResourceTestEntity.testAccCheckCloudAutonomousVmClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: autonomousVMClusterResourceTestEntity.avmcBasic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					autonomousVMClusterResourceTestEntity.checkCloudAutonomousVmClusterExists(ctx, resourceName, &cloudAVMC),
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

func TestAccODBCloudAutonomousVmCluster_variables(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	avmcDisplayName := sdkacctest.RandomWithPrefix(autonomousVMClusterResourceTestEntity.autonomousVmClusterDisplayNamePrefix)
	resourceName := "oci_database_cloud_autonomous_vm_cluster.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			autonomousVMClusterResourceTestEntity.testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             autonomousVMClusterResourceTestEntity.testAccCheckCloudAutonomousVmClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: autonomousVmClusterConfig_useVariables(avmcDisplayName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("cloud_exadata_infrastructure_id"), knownvalue.StringExact("exa_gjrmtxl4qk")),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("odb_network_id"), knownvalue.StringExact("odbnet_3l9st3litg")),
					},
				},
			},
		},
	})
}

func TestAccODBCloudAutonomousVmCluster_usingARN(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	var avmc1, avmc2 odbtypes.CloudAutonomousVmCluster
	resourceName := "aws_odb_cloud_autonomous_vm_cluster.test"

	avmcWithoutTag, avmcWithTag := autonomousVMClusterResourceTestEntity.autonomousVmClusterByARN()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			autonomousVMClusterResourceTestEntity.testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             autonomousVMClusterResourceTestEntity.testAccCheckCloudAutonomousVmClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: avmcWithoutTag,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.ComposeTestCheckFunc(func(state *terraform.State) error {
						return nil
					}),
					autonomousVMClusterResourceTestEntity.checkCloudAutonomousVmClusterExists(ctx, resourceName, &avmc1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: avmcWithTag,
				Check: resource.ComposeAggregateTestCheckFunc(
					autonomousVMClusterResourceTestEntity.checkCloudAutonomousVmClusterExists(ctx, resourceName, &avmc2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.env", "dev"),
					resource.ComposeTestCheckFunc(func(state *terraform.State) error {
						if strings.Compare(*(avmc1.CloudAutonomousVmClusterId), *(avmc2.CloudAutonomousVmClusterId)) != 0 {
							return errors.New("shouldn't create a new autonomous vm cluster")
						}
						return nil
					}),
				),
			},
		},
	})
}

func TestAccODBCloudAutonomousVmCluster_withAllParams(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var cloudAVMC odbtypes.CloudAutonomousVmCluster

	resourceName := "aws_odb_cloud_autonomous_vm_cluster.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			//acctest.PreCheckPartitionHasService(t, names.ODBServiceID)
			autonomousVMClusterResourceTestEntity.testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             autonomousVMClusterResourceTestEntity.testAccCheckCloudAutonomousVmClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: autonomousVMClusterResourceTestEntity.avmcAllParamsConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					autonomousVMClusterResourceTestEntity.checkCloudAutonomousVmClusterExists(ctx, resourceName, &cloudAVMC),
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

func TestAccODBCloudAutonomousVmCluster_tagging(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var avmc1, avmc2 odbtypes.CloudAutonomousVmCluster
	resourceName := "aws_odb_cloud_autonomous_vm_cluster.test"
	withoutTag, withTag := autonomousVMClusterResourceTestEntity.avmcNoTagWithTag()
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			autonomousVMClusterResourceTestEntity.testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             autonomousVMClusterResourceTestEntity.testAccCheckCloudAutonomousVmClusterDestroy(ctx),

		Steps: []resource.TestStep{
			{
				Config: withoutTag,

				Check: resource.ComposeAggregateTestCheckFunc(
					autonomousVMClusterResourceTestEntity.checkCloudAutonomousVmClusterExists(ctx, resourceName, &avmc1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: withTag,
				Check: resource.ComposeAggregateTestCheckFunc(
					autonomousVMClusterResourceTestEntity.checkCloudAutonomousVmClusterExists(ctx, resourceName, &avmc2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.env", "dev"),
					resource.ComposeTestCheckFunc(func(state *terraform.State) error {
						if strings.Compare(*(avmc1.CloudAutonomousVmClusterId), *(avmc2.CloudAutonomousVmClusterId)) != 0 {
							return errors.New("shouldn't create a new autonomous vm cluster")
						}
						return nil
					}),
				),
			},
		},
	})
}

func TestAccODBCloudAutonomousVmCluster_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	var cloudautonomousvmcluster odbtypes.CloudAutonomousVmCluster
	resourceName := "aws_odb_cloud_autonomous_vm_cluster.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			autonomousVMClusterResourceTestEntity.testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             autonomousVMClusterResourceTestEntity.testAccCheckCloudAutonomousVmClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: autonomousVMClusterResourceTestEntity.avmcBasic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					autonomousVMClusterResourceTestEntity.checkCloudAutonomousVmClusterExists(ctx, resourceName, &cloudautonomousvmcluster),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfodb.ResourceCloudAutonomousVMCluster, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func (autonomousVMClusterResourceTest) testAccPreCheck(ctx context.Context, t *testing.T) {
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
func (autonomousVMClusterResourceTest) testAccCheckCloudAutonomousVmClusterDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_odb_cloud_autonomous_vm_cluster" {
				continue
			}

			_, err := autonomousVMClusterResourceTestEntity.findAVMC(ctx, conn, rs.Primary.ID)
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

func (autonomousVMClusterResourceTest) checkCloudAutonomousVmClusterExists(ctx context.Context, name string, cloudAutonomousVMCluster *odbtypes.CloudAutonomousVmCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.ResNameCloudAutonomousVmCluster, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.ResNameCloudAutonomousVmCluster, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)
		resp, err := autonomousVMClusterResourceTestEntity.findAVMC(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.ResNameCloudAutonomousVmCluster, rs.Primary.ID, err)
		}

		*cloudAutonomousVMCluster = *resp

		return nil
	}
}

func (autonomousVMClusterResourceTest) findAVMC(ctx context.Context, conn *odb.Client, id string) (*odbtypes.CloudAutonomousVmCluster, error) {
	input := odb.GetCloudAutonomousVmClusterInput{
		CloudAutonomousVmClusterId: aws.String(id),
	}
	out, err := conn.GetCloudAutonomousVmCluster(ctx, &input)
	if err != nil {
		if errs.IsA[*odbtypes.ResourceNotFoundException](err) {
			return nil, &sdkretry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			}
		}
		return nil, err
	}

	if out == nil || out.CloudAutonomousVmCluster == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return out.CloudAutonomousVmCluster, nil
}

func (autonomousVMClusterResourceTest) avmcBasic() string {
	exaInfraDisplayName := sdkacctest.RandomWithPrefix(autonomousVMClusterDSTestEntity.exaInfraDisplayNamePrefix)
	odbNetworkDisplayName := sdkacctest.RandomWithPrefix(autonomousVMClusterDSTestEntity.odbNetDisplayNamePrefix)
	avmcDisplayName := sdkacctest.RandomWithPrefix(autonomousVMClusterDSTestEntity.autonomousVmClusterDisplayNamePrefix)
	domain := acctest.RandomDomainName()
	emailAddress := acctest.RandomEmailAddress(domain)
	exaInfraRes := autonomousVMClusterResourceTestEntity.exaInfra(exaInfraDisplayName, emailAddress)
	odbNetRes := autonomousVMClusterResourceTestEntity.oracleDBNetwork(odbNetworkDisplayName)
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

}




`, exaInfraRes, odbNetRes, avmcDisplayName)

	return res
}

func (autonomousVMClusterResourceTest) avmcNoTagWithTag() (string, string) {
	exaInfraDisplayName := sdkacctest.RandomWithPrefix(autonomousVMClusterDSTestEntity.exaInfraDisplayNamePrefix)
	odbNetworkDisplayName := sdkacctest.RandomWithPrefix(autonomousVMClusterDSTestEntity.odbNetDisplayNamePrefix)
	avmcDisplayName := sdkacctest.RandomWithPrefix(autonomousVMClusterDSTestEntity.autonomousVmClusterDisplayNamePrefix)
	domain := acctest.RandomDomainName()
	emailAddress := acctest.RandomEmailAddress(domain)
	exaInfraRes := autonomousVMClusterResourceTestEntity.exaInfra(exaInfraDisplayName, emailAddress)
	odbNetRes := autonomousVMClusterResourceTestEntity.oracleDBNetwork(odbNetworkDisplayName)
	noTag := fmt.Sprintf(`
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

}




`, exaInfraRes, odbNetRes, avmcDisplayName)
	withTag := fmt.Sprintf(`
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




`, exaInfraRes, odbNetRes, avmcDisplayName)

	return noTag, withTag
}

func (autonomousVMClusterResourceTest) avmcAllParamsConfig() string {
	exaInfraDisplayName := sdkacctest.RandomWithPrefix(autonomousVMClusterDSTestEntity.exaInfraDisplayNamePrefix)
	odbNetworkDisplayName := sdkacctest.RandomWithPrefix(autonomousVMClusterDSTestEntity.odbNetDisplayNamePrefix)
	avmcDisplayName := sdkacctest.RandomWithPrefix(autonomousVMClusterDSTestEntity.autonomousVmClusterDisplayNamePrefix)
	domain := acctest.RandomDomainName()
	emailAddress := acctest.RandomEmailAddress(domain)
	exaInfraRes := autonomousVMClusterResourceTestEntity.exaInfra(exaInfraDisplayName, emailAddress)
	odbNetRes := autonomousVMClusterResourceTestEntity.oracleDBNetwork(odbNetworkDisplayName)
	res := fmt.Sprintf(`
%s

%s

data "aws_odb_db_servers" "test" {
  cloud_exadata_infrastructure_id = aws_odb_cloud_exadata_infrastructure.test.id
}

resource "aws_odb_cloud_autonomous_vm_cluster" "test" {
  description                           = "my first avmc"
  time_zone                             = "UTC"
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
    days_of_week       = [{ name = "MONDAY" }, { name = "TUESDAY" }]
    hours_of_day       = [4, 16]
    lead_time_in_weeks = 3
    months             = [{ name = "FEBRUARY" }, { name = "MAY" }, { name = "AUGUST" }, { name = "NOVEMBER" }]
    preference         = "CUSTOM_PREFERENCE"
    weeks_of_month     = [2, 4]
  }
  tags = {
    "env" = "dev"
  }

}




`, exaInfraRes, odbNetRes, avmcDisplayName)

	return res
}

func (autonomousVMClusterResourceTest) oracleDBNetwork(odbNetName string) string {
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

func (autonomousVMClusterResourceTest) exaInfra(exaDisplayName, emailAddress string) string {
	resource := fmt.Sprintf(`
resource "aws_odb_cloud_exadata_infrastructure" "test" {
  display_name                     = %[1]q
  shape                            = "Exadata.X9M"
  storage_count                    = 3
  compute_count                    = 2
  availability_zone_id             = "use1-az6"
  customer_contacts_to_send_to_oci = [{ email = "%[2]s" }]
  maintenance_window {
    custom_action_timeout_in_mins    = 16
    is_custom_action_timeout_enabled = true
    patching_mode                    = "ROLLING"
    preference                       = "NO_PREFERENCE"
  }
}
`, exaDisplayName, emailAddress)

	return resource
}

func autonomousVmClusterConfig_useVariables(rName string) string {
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

resource "aws_odb_cloud_autonomous_vm_cluster" "test" {
  cloud_exadata_infrastructure_id       = var.cloud_exadata_infrastructure_id
  odb_network_id                        = var.odb_network_id
  display_name                          = %[1]q
  autonomous_data_storage_size_in_tbs   = 5
  memory_per_oracle_compute_unit_in_gbs = 2
  total_container_databases             = 1
  cpu_core_count_per_node               = 40
  license_model                         = "LICENSE_INCLUDED"
  db_servers                            = ["db-server-1", "db-server-2"]
  scan_listener_port_tls                = 8561
  scan_listener_port_non_tls            = 1024
  maintenance_window {
    preference = "NO_PREFERENCE"
  }

}
`, rName)
}

func (autonomousVMClusterResourceTest) autonomousVmClusterByARN() (string, string) {
	exaInfraDisplayName := sdkacctest.RandomWithPrefix(autonomousVMClusterDSTestEntity.exaInfraDisplayNamePrefix)
	odbNetworkDisplayName := sdkacctest.RandomWithPrefix(autonomousVMClusterDSTestEntity.odbNetDisplayNamePrefix)
	avmcDisplayName := sdkacctest.RandomWithPrefix(autonomousVMClusterDSTestEntity.autonomousVmClusterDisplayNamePrefix)
	domain := acctest.RandomDomainName()
	emailAddress := acctest.RandomEmailAddress(domain)
	exaInfraRes := autonomousVMClusterResourceTestEntity.exaInfra(exaInfraDisplayName, emailAddress)
	odbNetRes := autonomousVMClusterResourceTestEntity.oracleDBNetwork(odbNetworkDisplayName)
	noTag := fmt.Sprintf(`
%s

%s

data "aws_odb_db_servers" "test" {
  cloud_exadata_infrastructure_id = aws_odb_cloud_exadata_infrastructure.test.arn
}

resource "aws_odb_cloud_autonomous_vm_cluster" "test" {
  cloud_exadata_infrastructure_arn      = aws_odb_cloud_exadata_infrastructure.test.arn
  odb_network_arn                       = aws_odb_network.test.arn
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

}
`, exaInfraRes, odbNetRes, avmcDisplayName)

	withTag := fmt.Sprintf(`
%s

%s

data "aws_odb_db_servers" "test" {
  cloud_exadata_infrastructure_id = aws_odb_cloud_exadata_infrastructure.test.id
}

resource "aws_odb_cloud_autonomous_vm_cluster" "test" {
  cloud_exadata_infrastructure_arn      = aws_odb_cloud_exadata_infrastructure.test.arn
  odb_network_arn                       = aws_odb_network.test.arn
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
`, exaInfraRes, odbNetRes, avmcDisplayName)

	return noTag, withTag
}
