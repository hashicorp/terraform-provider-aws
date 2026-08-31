// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/odb"
	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfodb "github.com/hashicorp/terraform-provider-aws/internal/service/odb"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	testAccExaDBVMClusterAvailabilityZoneID           = "use1-az6"
	testAccExaDBVMClusterAlternateGridImageIDEnvVar   = "TF_AWS_ODB_EXADB_VM_CLUSTER_ALTERNATE_REGION_GRID_IMAGE_ID"
	testAccExaDBVMClusterDisplayNamePrefix            = "ofake"
	testAccExaDBVMClusterEnabledECPUCount             = 16
	testAccExaDBVMClusterGridImageIDEnvVar            = "TF_AWS_ODB_EXADB_VM_CLUSTER_GRID_IMAGE_ID"
	testAccExaDBVMClusterNodeCount                    = 2
	testAccExaDBVMClusterShape                        = "ExaDbXS"
	testAccExaDBVMClusterTotalECPUCount               = 64
	testAccExaDBVMClusterUpdatedEnabledECPUCount      = 20
	testAccExaDBVMClusterUpdatedTotalECPUCount        = 80
	testAccExaDBVMClusterUpdatedVMFileSystemSizeInGBs = 480
	testAccExaDBVMClusterVaultStorageSizeInGBs        = 900
	testAccExaDBVMClusterVMFileSystemSizeInGBs        = 440
)

func TestAccODBExaDBVMCluster_basic(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var exaDBVMCluster odbtypes.ExadbVmCluster
	rName := testAccRandomExaDBVMClusterDisplayName(t)
	hostname := testAccRandomExaDBVMClusterHostname(t)
	gridImageID := acctest.SkipIfEnvVarNotSet(t, testAccExaDBVMClusterGridImageIDEnvVar)
	publicKey := testAccRandomExaDBVMClusterSSHPublicKey(t)
	resourceName := "aws_odb_exadb_vm_cluster.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID)
			testAccPreCheckExaDBVMCluster(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExaDBVMClusterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExaDBVMClusterConfig_basic(rName, hostname, gridImageID, publicKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExaDBVMClusterExists(ctx, t, resourceName, &exaDBVMCluster),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "odb", regexache.MustCompile(`exadb-vm-cluster/.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, rName),
					resource.TestCheckResourceAttr(resourceName, "enabled_ecpu_count", fmt.Sprintf("%d", testAccExaDBVMClusterEnabledECPUCount)),
					resource.TestCheckResourceAttrSet(resourceName, "exascale_db_storage_vault_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "exascale_db_storage_vault_id", "aws_odb_exascale_db_storage_vault.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "grid_image_id", gridImageID),
					resource.TestCheckResourceAttr(resourceName, "hostname", hostname),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "node_count", fmt.Sprintf("%d", testAccExaDBVMClusterNodeCount)),
					resource.TestCheckResourceAttrSet(resourceName, "odb_network_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "odb_network_id", "aws_odb_network.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "shape", testAccExaDBVMClusterShape),
					resource.TestCheckResourceAttr(resourceName, "ssh_public_keys.#", "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "AVAILABLE"),
					resource.TestCheckResourceAttr(resourceName, "total_ecpu_count", fmt.Sprintf("%d", testAccExaDBVMClusterTotalECPUCount)),
					resource.TestCheckResourceAttr(resourceName, "vm_file_system_storage.0.total_size_in_gbs", fmt.Sprintf("%d", testAccExaDBVMClusterVMFileSystemSizeInGBs)),
					resource.TestCheckResourceAttr(resourceName, "vm_file_system_storage_total_size_in_gbs", fmt.Sprintf("%d", testAccExaDBVMClusterVMFileSystemSizeInGBs)),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccODBExaDBVMCluster_allArguments(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var exaDBVMCluster odbtypes.ExadbVmCluster
	rName := testAccRandomExaDBVMClusterDisplayName(t)
	hostname := testAccRandomExaDBVMClusterHostname(t)
	clusterName := testAccRandomExaDBVMClusterClusterName(t)
	gridImageID := acctest.SkipIfEnvVarNotSet(t, testAccExaDBVMClusterGridImageIDEnvVar)
	publicKey := testAccRandomExaDBVMClusterSSHPublicKey(t)
	resourceName := "aws_odb_exadb_vm_cluster.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID)
			testAccPreCheckExaDBVMCluster(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExaDBVMClusterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExaDBVMClusterConfig_allArguments(rName, hostname, clusterName, gridImageID, publicKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExaDBVMClusterExists(ctx, t, resourceName, &exaDBVMCluster),
					resource.TestCheckResourceAttr(resourceName, names.AttrClusterName, clusterName),
					resource.TestCheckResourceAttr(resourceName, "data_collection_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_collection_options.0.is_diagnostics_events_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "data_collection_options.0.is_health_monitoring_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "data_collection_options.0.is_incident_logs_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "license_model", "LICENSE_INCLUDED"),
					resource.TestCheckResourceAttr(resourceName, "scan_listener_port_tcp", "1521"),
					resource.TestCheckResourceAttr(resourceName, "scan_listener_port_tcp_ssl", "2484"),
					resource.TestCheckResourceAttr(resourceName, "shape_attribute", "SMART_STORAGE"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "time_zone", "UTC"),
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

func TestAccODBExaDBVMCluster_update(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var exaDBVMCluster odbtypes.ExadbVmCluster
	rName := testAccRandomExaDBVMClusterDisplayName(t)
	hostname := testAccRandomExaDBVMClusterHostname(t)
	gridImageID := acctest.SkipIfEnvVarNotSet(t, testAccExaDBVMClusterGridImageIDEnvVar)
	publicKey1 := testAccRandomExaDBVMClusterSSHPublicKey(t)
	publicKey2 := testAccRandomExaDBVMClusterSSHPublicKey(t)
	resourceName := "aws_odb_exadb_vm_cluster.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID)
			testAccPreCheckExaDBVMCluster(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExaDBVMClusterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExaDBVMClusterConfig_updateBefore(rName, hostname, gridImageID, publicKey1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExaDBVMClusterExists(ctx, t, resourceName, &exaDBVMCluster),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, rName),
					resource.TestCheckResourceAttr(resourceName, "enabled_ecpu_count", fmt.Sprintf("%d", testAccExaDBVMClusterEnabledECPUCount)),
					resource.TestCheckResourceAttr(resourceName, "license_model", "LICENSE_INCLUDED"),
					resource.TestCheckResourceAttr(resourceName, "total_ecpu_count", fmt.Sprintf("%d", testAccExaDBVMClusterTotalECPUCount)),
					resource.TestCheckResourceAttr(resourceName, "vm_file_system_storage_total_size_in_gbs", fmt.Sprintf("%d", testAccExaDBVMClusterVMFileSystemSizeInGBs)),
				),
			},
			{
				Config: testAccExaDBVMClusterConfig_updateAfter(rName, hostname, gridImageID, publicKey2, true),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExaDBVMClusterExists(ctx, t, resourceName, &exaDBVMCluster),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, rName+"-updated"),
					resource.TestCheckResourceAttr(resourceName, "data_collection_options.0.is_diagnostics_events_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "data_collection_options.0.is_health_monitoring_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "data_collection_options.0.is_incident_logs_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "enabled_ecpu_count", fmt.Sprintf("%d", testAccExaDBVMClusterUpdatedEnabledECPUCount)),
					resource.TestCheckResourceAttr(resourceName, "license_model", "BRING_YOUR_OWN_LICENSE"),
					resource.TestCheckResourceAttr(resourceName, "ssh_public_keys.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "total_ecpu_count", fmt.Sprintf("%d", testAccExaDBVMClusterUpdatedTotalECPUCount)),
					resource.TestCheckResourceAttr(resourceName, "vm_file_system_storage_total_size_in_gbs", fmt.Sprintf("%d", testAccExaDBVMClusterUpdatedVMFileSystemSizeInGBs)),
				),
			},
			{
				Config: testAccExaDBVMClusterConfig_updateAfter(rName, hostname, gridImageID, publicKey2, false),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExaDBVMClusterExists(ctx, t, resourceName, &exaDBVMCluster),
					resource.TestCheckResourceAttr(resourceName, "license_model", "BRING_YOUR_OWN_LICENSE"),
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

func TestAccODBExaDBVMCluster_tags(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var exaDBVMCluster odbtypes.ExadbVmCluster
	rName := testAccRandomExaDBVMClusterDisplayName(t)
	hostname := testAccRandomExaDBVMClusterHostname(t)
	gridImageID := acctest.SkipIfEnvVarNotSet(t, testAccExaDBVMClusterGridImageIDEnvVar)
	publicKey := testAccRandomExaDBVMClusterSSHPublicKey(t)
	resourceName := "aws_odb_exadb_vm_cluster.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID)
			testAccPreCheckExaDBVMCluster(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExaDBVMClusterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExaDBVMClusterConfig_tags(rName, hostname, gridImageID, publicKey, fmt.Sprintf("%s = %q", acctest.CtKey1, acctest.CtValue1)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExaDBVMClusterExists(ctx, t, resourceName, &exaDBVMCluster),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				Config: testAccExaDBVMClusterConfig_tags(rName, hostname, gridImageID, publicKey, fmt.Sprintf("%s = %q\n    %s = %q", acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExaDBVMClusterExists(ctx, t, resourceName, &exaDBVMCluster),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccExaDBVMClusterConfig_tags(rName, hostname, gridImageID, publicKey, fmt.Sprintf("%s = %q", acctest.CtKey2, "")),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExaDBVMClusterExists(ctx, t, resourceName, &exaDBVMCluster),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckNoResourceAttr(resourceName, acctest.CtTagsKey1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, ""),
				),
			},
			{
				Config: testAccExaDBVMClusterConfig_basic(rName, hostname, gridImageID, publicKey),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccODBExaDBVMCluster_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := testAccRandomExaDBVMClusterDisplayName(t)
	hostname := testAccRandomExaDBVMClusterHostname(t)
	gridImageID := acctest.SkipIfEnvVarNotSet(t, testAccExaDBVMClusterGridImageIDEnvVar)
	publicKey := testAccRandomExaDBVMClusterSSHPublicKey(t)
	resourceName := "aws_odb_exadb_vm_cluster.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID)
			testAccPreCheckExaDBVMCluster(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExaDBVMClusterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExaDBVMClusterConfig_basic(rName, hostname, gridImageID, publicKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfodb.ResourceExaDBVMCluster, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccCheckExaDBVMClusterDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ODBClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_odb_exadb_vm_cluster" {
				continue
			}

			_, err := tfodb.FindExaDBVMClusterByID(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return create.Error(names.ODB, create.ErrActionCheckingDestroyed, tfodb.ResNameExaDBVMCluster, rs.Primary.ID, err)
			}

			return create.Error(names.ODB, create.ErrActionCheckingDestroyed, tfodb.ResNameExaDBVMCluster, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckExaDBVMClusterExists(ctx context.Context, t *testing.T, name string, exaDBVMCluster *odbtypes.ExadbVmCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.ResNameExaDBVMCluster, name, errors.New("not found"))
		}
		if rs.Primary.ID == "" {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.ResNameExaDBVMCluster, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).ODBClient(ctx)
		output, err := tfodb.FindExaDBVMClusterByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.ResNameExaDBVMCluster, rs.Primary.ID, err)
		}

		*exaDBVMCluster = *output

		return nil
	}
}

func testAccPreCheckExaDBVMCluster(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).ODBClient(ctx)

	_, err := conn.ListExadbVmClusters(ctx, &odb.ListExadbVmClustersInput{})
	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccRandomExaDBVMClusterDisplayName(t *testing.T) string {
	return acctest.RandomWithPrefix(t, testAccExaDBVMClusterDisplayNamePrefix)
}

func testAccRandomExaDBVMClusterHostname(t *testing.T) string {
	return "ofake" + acctest.RandStringFromCharSet(t, 6, acctest.CharSetAlphaNum)
}

func testAccRandomExaDBVMClusterClusterName(t *testing.T) string {
	return "ofake" + acctest.RandStringFromCharSet(t, 5, acctest.CharSetAlphaNum)
}

func testAccRandomExaDBVMClusterSSHPublicKey(t *testing.T) string {
	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatal(err)
	}

	return publicKey
}

func testAccExaDBVMClusterConfig_basic(rName, hostname, gridImageID, publicKey string) string {
	return testAccExaDBVMClusterConfig(rName, hostname, gridImageID, publicKey, "")
}

func testAccExaDBVMClusterConfig_allArguments(rName, hostname, clusterName, gridImageID, publicKey string) string {
	extra := fmt.Sprintf(`
  cluster_name               = %[1]q
  license_model              = "LICENSE_INCLUDED"
  scan_listener_port_tcp     = 1521
  scan_listener_port_tcp_ssl = 2484
  shape_attribute            = "SMART_STORAGE"
  time_zone                  = "UTC"

  data_collection_options {
    is_diagnostics_events_enabled = true
    is_health_monitoring_enabled  = true
    is_incident_logs_enabled      = true
  }

  tags = {
    Name = %[2]q
  }
`, clusterName, rName)

	return testAccExaDBVMClusterConfig(rName, hostname, gridImageID, publicKey, extra)
}

func testAccExaDBVMClusterConfig_updateBefore(rName, hostname, gridImageID, publicKey string) string {
	extra := `
  license_model = "LICENSE_INCLUDED"

  data_collection_options {
    is_diagnostics_events_enabled = false
    is_health_monitoring_enabled  = false
    is_incident_logs_enabled      = false
  }
`

	return testAccExaDBVMClusterConfig(rName, hostname, gridImageID, publicKey, extra)
}

func testAccExaDBVMClusterConfig_updateAfter(rName, hostname, gridImageID, publicKey string, includeLicenseModel bool) string {
	licenseModel := ""
	if includeLicenseModel {
		licenseModel = `license_model = "BRING_YOUR_OWN_LICENSE"`
	}

	extra := fmt.Sprintf(`
  %[1]s

  data_collection_options {
    is_diagnostics_events_enabled = true
    is_health_monitoring_enabled  = true
    is_incident_logs_enabled      = true
  }
`, licenseModel)

	return testAccExaDBVMClusterConfigWithSizes(rName+"-updated", hostname, gridImageID, publicKey, testAccExaDBVMClusterUpdatedEnabledECPUCount, testAccExaDBVMClusterUpdatedTotalECPUCount, testAccExaDBVMClusterUpdatedVMFileSystemSizeInGBs, extra)
}

func testAccExaDBVMClusterConfig_tags(rName, hostname, gridImageID, publicKey, tags string) string {
	extra := fmt.Sprintf(`
  tags = {
    %s
  }
`, tags)

	return testAccExaDBVMClusterConfig(rName, hostname, gridImageID, publicKey, extra)
}

func testAccExaDBVMClusterConfig(rName, hostname, gridImageID, publicKey, extra string) string {
	return testAccExaDBVMClusterConfigWithSizes(rName, hostname, gridImageID, publicKey, testAccExaDBVMClusterEnabledECPUCount, testAccExaDBVMClusterTotalECPUCount, testAccExaDBVMClusterVMFileSystemSizeInGBs, extra)
}

func testAccExaDBVMClusterConfigWithSizes(rName, hostname, gridImageID, publicKey string, enabledECPUCount, totalECPUCount, vmFileSystemSizeInGBs int, extra string) string {
	return fmt.Sprintf(`
resource "aws_odb_network" "test" {
  availability_zone_id        = %[8]q
  backup_subnet_cidr          = "10.2.1.0/24"
  client_subnet_cidr          = "10.2.0.0/24"
  delete_associated_resources = true
  display_name                = "%[1]s-network"
  s3_access                   = "DISABLED"
  zero_etl_access             = "DISABLED"
}

resource "aws_odb_exascale_db_storage_vault" "test" {
  availability_zone_id                             = %[8]q
  display_name                                     = "%[1]s-vault"
  high_capacity_database_storage_total_size_in_gbs = %[9]d
}

resource "aws_odb_exadb_vm_cluster" "test" {
  display_name                             = %[1]q
  enabled_ecpu_count                       = %[5]d
  exascale_db_storage_vault_id             = aws_odb_exascale_db_storage_vault.test.id
  grid_image_id                            = %[3]q
  hostname                                 = %[2]q
  node_count                               = %[10]d
  odb_network_id                           = aws_odb_network.test.id
  shape                                    = %[11]q
  ssh_public_keys                          = [%[4]q]
  total_ecpu_count                         = %[6]d
  vm_file_system_storage_total_size_in_gbs = %[7]d
%[12]s
}
`, rName, hostname, gridImageID, publicKey, enabledECPUCount, totalECPUCount, vmFileSystemSizeInGBs, testAccExaDBVMClusterAvailabilityZoneID, testAccExaDBVMClusterVaultStorageSizeInGBs, testAccExaDBVMClusterNodeCount, testAccExaDBVMClusterShape, extra)
}
