// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/odb"
	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/go-version"
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
	testAccExaDBVMClusterDisplayNamePrefix            = "ofake"
	testAccExaDBVMClusterEnabledECPUCount             = 16
	testAccExaDBVMClusterGIVersion                    = "26.0.0.0"
	testAccExaDBVMClusterNodeCount                    = 2
	testAccExaDBVMClusterShape                        = "ExaDbXS"
	testAccExaDBVMClusterShapeCanonical               = "EXADBXS"
	testAccExaDBVMClusterShapeFamily                  = "EXADB_XS"
	testAccExaDBVMClusterTotalECPUCount               = 64
	testAccExaDBVMClusterUpdatedEnabledECPUCount      = 20
	testAccExaDBVMClusterUpdatedTotalECPUCount        = 80
	testAccExaDBVMClusterUpdatedVMFileSystemSizeInGBs = 480
	testAccExaDBVMClusterVaultStorageSizeInGBs        = 900
	testAccExaDBVMClusterVMFileSystemSizeInGBs        = 440
)

var testAccExaDBVMClusterAvailabilityZoneIDs = map[string]string{
	endpoints.EuWest1RegionID: "euw1-az3",
	endpoints.UsEast1RegionID: "use1-az6",
}

func TestAccODBExaDBVMCluster_basic(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var exaDBVMCluster odbtypes.ExadbVmCluster
	rName := testAccRandomExaDBVMClusterDisplayName(t)
	hostname := testAccRandomExaDBVMClusterHostname(t)
	availabilityZoneID := testAccExaDBVMClusterAvailabilityZoneIDs[acctest.Region()]
	gridImageID := testAccExaDBVMClusterGridImageIDForRegion(ctx, t, acctest.Region(), availabilityZoneID)
	publicKey := testAccRandomExaDBVMClusterSSHPublicKey(t)
	resourceName := "aws_odb_exadb_vm_cluster.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID, endpoints.EuWest1RegionID)
			testAccPreCheckExaDBVMCluster(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExaDBVMClusterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExaDBVMClusterConfig_basic(rName, hostname, availabilityZoneID, gridImageID, publicKey),
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
					resource.TestCheckResourceAttr(resourceName, "shape", testAccExaDBVMClusterShapeCanonical),
					resource.TestCheckResourceAttr(resourceName, "ssh_public_keys.#", "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "AVAILABLE"),
					resource.TestCheckResourceAttrSet(resourceName, "system_version"),
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
	availabilityZoneID := testAccExaDBVMClusterAvailabilityZoneIDs[acctest.Region()]
	gridImageID := testAccExaDBVMClusterGridImageIDForRegion(ctx, t, acctest.Region(), availabilityZoneID)
	publicKey := testAccRandomExaDBVMClusterSSHPublicKey(t)
	resourceName := "aws_odb_exadb_vm_cluster.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID, endpoints.EuWest1RegionID)
			testAccPreCheckExaDBVMCluster(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExaDBVMClusterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExaDBVMClusterConfig_allArguments(rName, hostname, clusterName, availabilityZoneID, gridImageID, publicKey),
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
	availabilityZoneID := testAccExaDBVMClusterAvailabilityZoneIDs[acctest.Region()]
	gridImageID := testAccExaDBVMClusterGridImageIDForRegion(ctx, t, acctest.Region(), availabilityZoneID)
	publicKey1 := testAccRandomExaDBVMClusterSSHPublicKey(t)
	publicKey2 := testAccRandomExaDBVMClusterSSHPublicKey(t)
	resourceName := "aws_odb_exadb_vm_cluster.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID, endpoints.EuWest1RegionID)
			testAccPreCheckExaDBVMCluster(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExaDBVMClusterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExaDBVMClusterConfig_updateBefore(rName, hostname, availabilityZoneID, gridImageID, publicKey1),
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
				Config: testAccExaDBVMClusterConfig_updateAfter(rName, hostname, availabilityZoneID, gridImageID, publicKey2, true),
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
				Config: testAccExaDBVMClusterConfig_updateAfter(rName, hostname, availabilityZoneID, gridImageID, publicKey2, false),
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
	availabilityZoneID := testAccExaDBVMClusterAvailabilityZoneIDs[acctest.Region()]
	gridImageID := testAccExaDBVMClusterGridImageIDForRegion(ctx, t, acctest.Region(), availabilityZoneID)
	publicKey := testAccRandomExaDBVMClusterSSHPublicKey(t)
	resourceName := "aws_odb_exadb_vm_cluster.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID, endpoints.EuWest1RegionID)
			testAccPreCheckExaDBVMCluster(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExaDBVMClusterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExaDBVMClusterConfig_tags(rName, hostname, availabilityZoneID, gridImageID, publicKey, fmt.Sprintf("%s = %q", acctest.CtKey1, acctest.CtValue1)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExaDBVMClusterExists(ctx, t, resourceName, &exaDBVMCluster),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				Config: testAccExaDBVMClusterConfig_tags(rName, hostname, availabilityZoneID, gridImageID, publicKey, fmt.Sprintf("%s = %q\n    %s = %q", acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExaDBVMClusterExists(ctx, t, resourceName, &exaDBVMCluster),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccExaDBVMClusterConfig_tags(rName, hostname, availabilityZoneID, gridImageID, publicKey, fmt.Sprintf("%s = %q", acctest.CtKey2, "")),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExaDBVMClusterExists(ctx, t, resourceName, &exaDBVMCluster),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckNoResourceAttr(resourceName, acctest.CtTagsKey1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, ""),
				),
			},
			{
				Config: testAccExaDBVMClusterConfig_basic(rName, hostname, availabilityZoneID, gridImageID, publicKey),
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
	availabilityZoneID := testAccExaDBVMClusterAvailabilityZoneIDs[acctest.Region()]
	gridImageID := testAccExaDBVMClusterGridImageIDForRegion(ctx, t, acctest.Region(), availabilityZoneID)
	publicKey := testAccRandomExaDBVMClusterSSHPublicKey(t)
	resourceName := "aws_odb_exadb_vm_cluster.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID, endpoints.EuWest1RegionID)
			testAccPreCheckExaDBVMCluster(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExaDBVMClusterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExaDBVMClusterConfig_basic(rName, hostname, availabilityZoneID, gridImageID, publicKey),
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

func testAccExaDBVMClusterGridImageIDForRegion(ctx context.Context, t *testing.T, region, availabilityZoneID string) string {
	t.Helper()

	acctest.PreCheck(ctx, t)
	conn := acctest.ProviderMeta(ctx, t).ODBClient(ctx)
	regionOption := func(options *odb.Options) {
		options.Region = region
	}

	giVersionFound := false
	giVersionsPaginator := odb.NewListGiVersionsPaginator(conn, &odb.ListGiVersionsInput{
		Shape: aws.String(testAccExaDBVMClusterShape),
	})
	for giVersionsPaginator.HasMorePages() {
		page, err := giVersionsPaginator.NextPage(ctx, regionOption)
		if acctest.PreCheckSkipError(err) {
			t.Skipf("skipping acceptance testing: listing GI versions: %s", err)
		}
		if err != nil {
			t.Fatalf("listing GI versions: %s", err)
		}

		for _, giVersion := range page.GiVersions {
			if aws.ToString(giVersion.Version) == testAccExaDBVMClusterGIVersion {
				giVersionFound = true
				break
			}
		}
	}
	if !giVersionFound {
		t.Fatalf("GI version %q is not available for shape %q in Region %q", testAccExaDBVMClusterGIVersion, testAccExaDBVMClusterShape, region)
	}

	var latestVersion *version.Version
	var gridImageID string
	giMinorVersionsPaginator := odb.NewListGiMinorVersionsPaginator(conn, &odb.ListGiMinorVersionsInput{
		AvailabilityZoneId: aws.String(availabilityZoneID),
		GiVersion:          aws.String(testAccExaDBVMClusterGIVersion),
		ShapeFamily:        aws.String(testAccExaDBVMClusterShapeFamily),
	})
	for giMinorVersionsPaginator.HasMorePages() {
		page, err := giMinorVersionsPaginator.NextPage(ctx, regionOption)
		if acctest.PreCheckSkipError(err) {
			t.Skipf("skipping acceptance testing: listing GI minor versions: %s", err)
		}
		if err != nil {
			t.Fatalf("listing GI minor versions: %s", err)
		}

		for _, giMinorVersion := range page.GiMinorVersions {
			candidateGridImageID := aws.ToString(giMinorVersion.GridImageId)
			if candidateGridImageID == "" {
				continue
			}

			candidateVersion, err := version.NewVersion(aws.ToString(giMinorVersion.Version))
			if err != nil {
				t.Fatalf("parsing GI minor version %q: %s", aws.ToString(giMinorVersion.Version), err)
			}
			if latestVersion == nil || candidateVersion.GreaterThan(latestVersion) {
				latestVersion = candidateVersion
				gridImageID = candidateGridImageID
			}
		}
	}
	if gridImageID == "" {
		t.Fatalf("no Grid Image ID is available for GI version %q, shape family %q, and Availability Zone ID %q in Region %q", testAccExaDBVMClusterGIVersion, testAccExaDBVMClusterShapeFamily, availabilityZoneID, region)
	}

	return gridImageID
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

func testAccExaDBVMClusterConfig_basic(rName, hostname, availabilityZoneID, gridImageID, publicKey string) string {
	return testAccExaDBVMClusterConfig(rName, hostname, availabilityZoneID, gridImageID, publicKey, "")
}

func testAccExaDBVMClusterConfig_allArguments(rName, hostname, clusterName, availabilityZoneID, gridImageID, publicKey string) string {
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

	return testAccExaDBVMClusterConfig(rName, hostname, availabilityZoneID, gridImageID, publicKey, extra)
}

func testAccExaDBVMClusterConfig_updateBefore(rName, hostname, availabilityZoneID, gridImageID, publicKey string) string {
	extra := `
  license_model = "LICENSE_INCLUDED"

  data_collection_options {
    is_diagnostics_events_enabled = false
    is_health_monitoring_enabled  = false
    is_incident_logs_enabled      = false
  }
`

	return testAccExaDBVMClusterConfig(rName, hostname, availabilityZoneID, gridImageID, publicKey, extra)
}

func testAccExaDBVMClusterConfig_updateAfter(rName, hostname, availabilityZoneID, gridImageID, publicKey string, includeLicenseModel bool) string {
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

	return testAccExaDBVMClusterConfigWithSizes(rName+"-updated", hostname, availabilityZoneID, gridImageID, publicKey, testAccExaDBVMClusterUpdatedEnabledECPUCount, testAccExaDBVMClusterUpdatedTotalECPUCount, testAccExaDBVMClusterUpdatedVMFileSystemSizeInGBs, extra)
}

func testAccExaDBVMClusterConfig_tags(rName, hostname, availabilityZoneID, gridImageID, publicKey, tags string) string {
	extra := fmt.Sprintf(`
  tags = {
    %s
  }
`, tags)

	return testAccExaDBVMClusterConfig(rName, hostname, availabilityZoneID, gridImageID, publicKey, extra)
}

func testAccExaDBVMClusterConfig(rName, hostname, availabilityZoneID, gridImageID, publicKey, extra string) string {
	return testAccExaDBVMClusterConfigWithSizes(rName, hostname, availabilityZoneID, gridImageID, publicKey, testAccExaDBVMClusterEnabledECPUCount, testAccExaDBVMClusterTotalECPUCount, testAccExaDBVMClusterVMFileSystemSizeInGBs, extra)
}

func testAccExaDBVMClusterConfigWithSizes(rName, hostname, availabilityZoneID, gridImageID, publicKey string, enabledECPUCount, totalECPUCount, vmFileSystemSizeInGBs int, extra string) string {
	return fmt.Sprintf(`
resource "aws_odb_network" "test" {
  availability_zone_id        = %[3]q
  backup_subnet_cidr          = "10.2.1.0/24"
  client_subnet_cidr          = "10.2.0.0/24"
  delete_associated_resources = true
  display_name                = "%[1]s-network"
  s3_access                   = "DISABLED"
  zero_etl_access             = "DISABLED"
}

resource "aws_odb_exascale_db_storage_vault" "test" {
  availability_zone_id                             = %[3]q
  display_name                                     = "%[1]s-vault"
  high_capacity_database_storage_total_size_in_gbs = %[9]d
}

resource "aws_odb_exadb_vm_cluster" "test" {
  display_name                             = %[1]q
  enabled_ecpu_count                       = %[6]d
  exascale_db_storage_vault_id             = aws_odb_exascale_db_storage_vault.test.id
  grid_image_id                            = %[4]q
  hostname                                 = %[2]q
  node_count                               = %[10]d
  odb_network_id                           = aws_odb_network.test.id
  shape                                    = %[11]q
  ssh_public_keys                          = [%[5]q]
  total_ecpu_count                         = %[7]d
  vm_file_system_storage_total_size_in_gbs = %[8]d
%[12]s
}
`, rName, hostname, availabilityZoneID, gridImageID, publicKey, enabledECPUCount, totalECPUCount, vmFileSystemSizeInGBs, testAccExaDBVMClusterVaultStorageSizeInGBs, testAccExaDBVMClusterNodeCount, testAccExaDBVMClusterShapeCanonical, extra)
}
