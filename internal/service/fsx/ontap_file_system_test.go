// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fsx_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/fsx/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tffsx "github.com/hashicorp/terraform-provider-aws/internal/service/fsx"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccFSxONTAPFileSystem_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem awstypes.FileSystem
	resourceName := "aws_fsx_ontap_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckONTAPFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPFileSystemConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckONTAPFileSystemExists(ctx, resourceName, &filesystem),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "fsx", regexache.MustCompile(`file-system/fs-.+`)),
					resource.TestCheckResourceAttr(resourceName, "automatic_backup_retention_days", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "deployment_type", string(awstypes.OntapDeploymentTypeMultiAz1)),
					resource.TestCheckResourceAttr(resourceName, "disk_iops_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "disk_iops_configuration.0.iops", "3072"),
					resource.TestCheckResourceAttr(resourceName, "disk_iops_configuration.0.mode", "AUTOMATIC"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_ip_address_range"),
					resource.TestCheckResourceAttr(resourceName, "endpoints.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "endpoints.0.intercluster.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "endpoints.0.intercluster.0.dns_name"),
					resource.TestCheckResourceAttr(resourceName, "endpoints.0.management.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "endpoints.0.management.0.dns_name"),
					resource.TestCheckResourceAttr(resourceName, "ha_pairs", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrKMSKeyID),
					resource.TestCheckResourceAttr(resourceName, "network_interface_ids.#", acctest.Ct2),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrPair(resourceName, "preferred_subnet_id", "aws_subnet.test.0", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "route_table_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "route_table_ids.*", "aws_vpc.test", "default_route_table_id"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "storage_capacity", "1024"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageType, string(awstypes.StorageTypeSsd)),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_ids.*", "aws_subnet.test.0", names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_ids.*", "aws_subnet.test.1", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "throughput_capacity", "128"),
					resource.TestCheckResourceAttr(resourceName, "throughput_capacity_per_ha_pair", "128"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrVPCID, "aws_vpc.test", names.AttrID),
					resource.TestMatchResourceAttr(resourceName, "weekly_maintenance_start_time", regexache.MustCompile(`^\d:\d\d:\d\d$`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrSecurityGroupIDs},
			},
		},
	})
}

func TestAccFSxONTAPFileSystem_singleAZ(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem awstypes.FileSystem
	resourceName := "aws_fsx_ontap_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckONTAPFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPFileSystemConfig_singleAZ(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPFileSystemExists(ctx, resourceName, &filesystem),
					resource.TestCheckResourceAttr(resourceName, "deployment_type", string(awstypes.OntapDeploymentTypeSingleAz1)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrSecurityGroupIDs},
			},
		},
	})
}

func TestAccFSxONTAPFileSystem_multiAZ2(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1, filesystem2 awstypes.FileSystem
	throughput1 := 384
	throughput2 := 768
	throughput3 := 768
	capacity1 := 1024
	capacity2 := 1024
	capacity3 := 2048

	resourceName := "aws_fsx_ontap_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckONTAPFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPFileSystemConfig_multiAZ2(rName, throughput1, capacity1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPFileSystemExists(ctx, resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "deployment_type", string(awstypes.OntapDeploymentTypeMultiAz2)),
					resource.TestCheckResourceAttr(resourceName, "ha_pairs", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "throughput_capacity", fmt.Sprint(throughput1)),
					resource.TestCheckResourceAttr(resourceName, "throughput_capacity_per_ha_pair", fmt.Sprint(throughput1)),
					resource.TestCheckResourceAttr(resourceName, "storage_capacity", fmt.Sprint(capacity1)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrSecurityGroupIDs},
			},
			{
				Config: testAccONTAPFileSystemConfig_multiAZ2(rName, throughput2, capacity2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckONTAPFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "deployment_type", string(awstypes.OntapDeploymentTypeMultiAz2)),
					resource.TestCheckResourceAttr(resourceName, "throughput_capacity_per_ha_pair", fmt.Sprint(throughput2)),
					resource.TestCheckResourceAttr(resourceName, "storage_capacity", fmt.Sprint(capacity2)),
				),
			},
			{
				Config: testAccONTAPFileSystemConfig_multiAZ2(rName, throughput3, capacity3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckONTAPFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "deployment_type", string(awstypes.OntapDeploymentTypeMultiAz2)),
					resource.TestCheckResourceAttr(resourceName, "throughput_capacity_per_ha_pair", fmt.Sprint(throughput3)),
					resource.TestCheckResourceAttr(resourceName, "storage_capacity", fmt.Sprint(capacity3)),
				),
			},
		},
	})
}

func TestAccFSxONTAPFileSystem_haPair(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem awstypes.FileSystem
	throughput1 := 3072
	throughput2 := 256
	resourceName := "aws_fsx_ontap_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckONTAPFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPFileSystemConfig_haPair(rName, throughput1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPFileSystemExists(ctx, resourceName, &filesystem),
					resource.TestCheckResourceAttr(resourceName, "ha_pairs", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "throughput_capacity", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "throughput_capacity_per_ha_pair", fmt.Sprint(throughput1)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrSecurityGroupIDs},
			},
			{
				Config: testAccONTAPFileSystemConfig_oneHaPair(rName, throughput2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPFileSystemExists(ctx, resourceName, &filesystem),
					resource.TestCheckResourceAttr(resourceName, "ha_pairs", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "throughput_capacity", fmt.Sprint(throughput2)),
					resource.TestCheckResourceAttr(resourceName, "throughput_capacity_per_ha_pair", fmt.Sprint(throughput2)),
				),
			},
		},
	})
}

func TestAccFSxONTAPFileSystem_haPair_increase(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1, filesystem2 awstypes.FileSystem
	throughput := 3072
	capacity1 := 4096
	capacity2 := 8192
	haPair1 := 2
	haPair2 := 4
	resourceName := "aws_fsx_ontap_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckONTAPFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPFileSystemConfig_singleAZ2(rName, throughput, capacity1, haPair1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPFileSystemExists(ctx, resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "ha_pairs", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "throughput_capacity", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "throughput_capacity_per_ha_pair", fmt.Sprint(throughput)),
					resource.TestCheckResourceAttr(resourceName, "storage_capacity", fmt.Sprint(capacity1)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrSecurityGroupIDs},
			},
			{
				Config: testAccONTAPFileSystemConfig_singleAZ2(rName, throughput, capacity2, haPair2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckONTAPFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "ha_pairs", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "throughput_capacity", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "throughput_capacity_per_ha_pair", fmt.Sprint(throughput)),
					resource.TestCheckResourceAttr(resourceName, "storage_capacity", fmt.Sprint(capacity2)),
				),
			},
		},
	})
}

func TestAccFSxONTAPFileSystem_fsxAdminPassword(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1, filesystem2 awstypes.FileSystem
	resourceName := "aws_fsx_ontap_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	pass1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	pass2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckONTAPFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPFileSystemConfig_adminPassword(rName, pass1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPFileSystemExists(ctx, resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "fsx_admin_password", pass1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrSecurityGroupIDs, "fsx_admin_password"},
			},
			{
				Config: testAccONTAPFileSystemConfig_adminPassword(rName, pass2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckONTAPFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "fsx_admin_password", pass2),
				),
			},
		},
	})
}

func TestAccFSxONTAPFileSystem_endpointIPAddressRange(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem awstypes.FileSystem
	resourceName := "aws_fsx_ontap_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckONTAPFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPFileSystemConfig_endpointIPAddressRange(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPFileSystemExists(ctx, resourceName, &filesystem),
					resource.TestCheckResourceAttr(resourceName, "endpoint_ip_address_range", "198.19.255.0/24"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrSecurityGroupIDs},
			},
		},
	})
}

func TestAccFSxONTAPFileSystem_diskIOPS(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem awstypes.FileSystem
	resourceName := "aws_fsx_ontap_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckONTAPFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPFileSystemConfig_diskIOPSConfiguration(rName, 3072),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPFileSystemExists(ctx, resourceName, &filesystem),
					resource.TestCheckResourceAttr(resourceName, "disk_iops_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "disk_iops_configuration.0.mode", "USER_PROVISIONED"),
					resource.TestCheckResourceAttr(resourceName, "disk_iops_configuration.0.iops", "3072"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrSecurityGroupIDs},
			},
			{
				Config: testAccONTAPFileSystemConfig_diskIOPSConfiguration(rName, 4000),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPFileSystemExists(ctx, resourceName, &filesystem),
					resource.TestCheckResourceAttr(resourceName, "disk_iops_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "disk_iops_configuration.0.mode", "USER_PROVISIONED"),
					resource.TestCheckResourceAttr(resourceName, "disk_iops_configuration.0.iops", "4000"),
				),
			},
		},
	})
}

func TestAccFSxONTAPFileSystem_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem awstypes.FileSystem
	resourceName := "aws_fsx_ontap_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckONTAPFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPFileSystemConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPFileSystemExists(ctx, resourceName, &filesystem),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tffsx.ResourceONTAPFileSystem(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccFSxONTAPFileSystem_securityGroupIDs(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1, filesystem2 awstypes.FileSystem
	resourceName := "aws_fsx_ontap_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckONTAPFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPFileSystemConfig_securityGroupIDs1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPFileSystemExists(ctx, resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", acctest.Ct1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrSecurityGroupIDs},
			},
			{
				Config: testAccONTAPFileSystemConfig_securityGroupIDs2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckONTAPFileSystemRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccFSxONTAPFileSystem_routeTableIDs(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1 awstypes.FileSystem
	resourceName := "aws_fsx_ontap_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckONTAPFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPFileSystemConfig_routeTable(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPFileSystemExists(ctx, resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "route_table_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "route_table_ids.*", "aws_route_table.test.0", names.AttrID),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrSecurityGroupIDs},
			},
			{
				Config: testAccONTAPFileSystemConfig_routeTable(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPFileSystemExists(ctx, resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "route_table_ids.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "route_table_ids.*", "aws_route_table.test.0", names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "route_table_ids.*", "aws_route_table.test.1", names.AttrID),
				),
			},
			{
				Config: testAccONTAPFileSystemConfig_routeTable(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPFileSystemExists(ctx, resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "route_table_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "route_table_ids.*", "aws_route_table.test.0", names.AttrID),
				),
			},
		},
	})
}

func TestAccFSxONTAPFileSystem_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1, filesystem2, filesystem3 awstypes.FileSystem
	resourceName := "aws_fsx_ontap_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckONTAPFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPFileSystemConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPFileSystemExists(ctx, resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrSecurityGroupIDs},
			},
			{
				Config: testAccONTAPFileSystemConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckONTAPFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccONTAPFileSystemConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPFileSystemExists(ctx, resourceName, &filesystem3),
					testAccCheckONTAPFileSystemNotRecreated(&filesystem2, &filesystem3),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccFSxONTAPFileSystem_weeklyMaintenanceStartTime(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1, filesystem2 awstypes.FileSystem
	resourceName := "aws_fsx_ontap_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckONTAPFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPFileSystemConfig_weeklyMaintenanceStartTime(rName, "1:01:01"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPFileSystemExists(ctx, resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "weekly_maintenance_start_time", "1:01:01"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrSecurityGroupIDs},
			},
			{
				Config: testAccONTAPFileSystemConfig_weeklyMaintenanceStartTime(rName, "2:02:02"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckONTAPFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "weekly_maintenance_start_time", "2:02:02"),
				),
			},
		},
	})
}

func TestAccFSxONTAPFileSystem_automaticBackupRetentionDays(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1, filesystem2 awstypes.FileSystem
	resourceName := "aws_fsx_ontap_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckONTAPFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPFileSystemConfig_automaticBackupRetentionDays(rName, 90),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPFileSystemExists(ctx, resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "automatic_backup_retention_days", "90"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrSecurityGroupIDs},
			},
			{
				Config: testAccONTAPFileSystemConfig_automaticBackupRetentionDays(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckONTAPFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "automatic_backup_retention_days", acctest.Ct0),
				),
			},
			{
				Config: testAccONTAPFileSystemConfig_automaticBackupRetentionDays(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPFileSystemExists(ctx, resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "automatic_backup_retention_days", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccFSxONTAPFileSystem_kmsKeyID(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem awstypes.FileSystem
	resourceName := "aws_fsx_ontap_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckONTAPFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPFileSystemConfig_kmsKeyID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPFileSystemExists(ctx, resourceName, &filesystem),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyID, "aws_kms_key.test", names.AttrARN),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrSecurityGroupIDs},
			},
		},
	})
}

func TestAccFSxONTAPFileSystem_dailyAutomaticBackupStartTime(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1, filesystem2 awstypes.FileSystem
	resourceName := "aws_fsx_ontap_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLustreFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPFileSystemConfig_dailyAutomaticBackupStartTime(rName, "01:01"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPFileSystemExists(ctx, resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "daily_automatic_backup_start_time", "01:01"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrSecurityGroupIDs},
			},
			{
				Config: testAccONTAPFileSystemConfig_dailyAutomaticBackupStartTime(rName, "02:02"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckONTAPFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "daily_automatic_backup_start_time", "02:02"),
				),
			},
		},
	})
}

func TestAccFSxONTAPFileSystem_throughputCapacity(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1, filesystem2 awstypes.FileSystem
	resourceName := "aws_fsx_ontap_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckONTAPFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPFileSystemConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPFileSystemExists(ctx, resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "throughput_capacity", "128"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrSecurityGroupIDs},
			},
			{
				Config: testAccONTAPFileSystemConfig_throughputCapacity(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckONTAPFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "throughput_capacity", "256"),
				),
			},
			{
				Config: testAccONTAPFileSystemConfig_multiAZHaThrougput(rName, 256),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckONTAPFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "throughput_capacity", "256"),
					resource.TestCheckResourceAttr(resourceName, "throughput_capacity_per_ha_pair", "256"),
				),
			},
		},
	})
}

func TestAccFSxONTAPFileSystem_throughputCapacity_singleAZ1(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1, filesystem2 awstypes.FileSystem
	throughput1 := 128
	throughput2 := 256
	capacity := 1024
	resourceName := "aws_fsx_ontap_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckONTAPFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPFileSystemConfig_singleAZ1(rName, throughput1, capacity),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPFileSystemExists(ctx, resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "throughput_capacity_per_ha_pair", fmt.Sprint(throughput1)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrSecurityGroupIDs},
			},
			{
				Config: testAccONTAPFileSystemConfig_singleAZ1(rName, throughput2, capacity),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckONTAPFileSystemRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "throughput_capacity_per_ha_pair", fmt.Sprint(throughput2)),
				),
			},
		},
	})
}

func TestAccFSxONTAPFileSystem_throughputCapacity_multiAZ1(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1, filesystem2 awstypes.FileSystem
	throughput1 := 128
	throughput2 := 256
	capacity := 1024
	resourceName := "aws_fsx_ontap_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckONTAPFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPFileSystemConfig_multiAZ1(rName, throughput1, capacity),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPFileSystemExists(ctx, resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "throughput_capacity_per_ha_pair", fmt.Sprint(throughput1)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrSecurityGroupIDs},
			},
			{
				Config: testAccONTAPFileSystemConfig_multiAZ1(rName, throughput2, capacity),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckONTAPFileSystemRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "throughput_capacity_per_ha_pair", fmt.Sprint(throughput2)),
				),
			},
		},
	})
}

func TestAccFSxONTAPFileSystem_storageCapacity(t *testing.T) {
	ctx := acctest.Context(t)
	var filesystem1, filesystem2 awstypes.FileSystem
	resourceName := "aws_fsx_ontap_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckONTAPFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPFileSystemConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPFileSystemExists(ctx, resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "storage_capacity", "1024"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrSecurityGroupIDs},
			},
			{
				Config: testAccONTAPFileSystemConfig_storageCapacity(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckONTAPFileSystemExists(ctx, resourceName, &filesystem2),
					testAccCheckONTAPFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "storage_capacity", "2048"),
				),
			},
		},
	})
}

func testAccCheckONTAPFileSystemExists(ctx context.Context, n string, v *awstypes.FileSystem) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).FSxClient(ctx)

		output, err := tffsx.FindONTAPFileSystemByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckONTAPFileSystemDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).FSxClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_fsx_ontap_file_system" {
				continue
			}

			_, err := tffsx.FindONTAPFileSystemByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("FSx for NetApp ONTAP File System (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckONTAPFileSystemNotRecreated(i, j *awstypes.FileSystem) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(i.FileSystemId) != aws.ToString(j.FileSystemId) {
			return fmt.Errorf("FSx for NetApp ONTAP File System (%s) recreated", aws.ToString(i.FileSystemId))
		}

		return nil
	}
}

func testAccCheckONTAPFileSystemRecreated(i, j *awstypes.FileSystem) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(i.FileSystemId) == aws.ToString(j.FileSystemId) {
			return fmt.Errorf("FSx for NetApp ONTAP File System (%s) not recreated", aws.ToString(i.FileSystemId))
		}

		return nil
	}
}

func testAccONTAPFileSystemConfig_base(rName string) string {
	return acctest.ConfigVPCWithSubnets(rName, 2)
}

func testAccONTAPFileSystemConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccONTAPFileSystemConfig_base(rName), `
resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity    = 1024
  subnet_ids          = aws_subnet.test[*].id
  deployment_type     = "MULTI_AZ_1"
  throughput_capacity = 128
  preferred_subnet_id = aws_subnet.test[0].id
}
`)
}

func testAccONTAPFileSystemConfig_singleAZ(rName string) string {
	return acctest.ConfigCompose(testAccONTAPFileSystemConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity    = 1024
  subnet_ids          = [aws_subnet.test[0].id]
  deployment_type     = "SINGLE_AZ_1"
  throughput_capacity = 128
  preferred_subnet_id = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccONTAPFileSystemConfig_multiAZ2(rName string, throughput int, capacity int) string {
	return acctest.ConfigCompose(testAccONTAPFileSystemConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity                = %[3]d
  subnet_ids                      = aws_subnet.test[*].id
  deployment_type                 = "MULTI_AZ_2"
  ha_pairs                        = 1
  throughput_capacity_per_ha_pair = %[2]d
  preferred_subnet_id             = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}
`, rName, throughput, capacity))
}

func testAccONTAPFileSystemConfig_multiAZ1(rName string, throughput int, capacity int) string {
	return acctest.ConfigCompose(testAccONTAPFileSystemConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity                = %[3]d
  subnet_ids                      = aws_subnet.test[*].id
  deployment_type                 = "MULTI_AZ_1"
  ha_pairs                        = 1
  throughput_capacity_per_ha_pair = %[2]d
  preferred_subnet_id             = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}
`, rName, throughput, capacity))
}

func testAccONTAPFileSystemConfig_singleAZ1(rName string, throughput int, capacity int) string {
	return acctest.ConfigCompose(testAccONTAPFileSystemConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity                = %[3]d
  subnet_ids                      = [aws_subnet.test[0].id]
  deployment_type                 = "SINGLE_AZ_1"
  ha_pairs                        = 1
  throughput_capacity_per_ha_pair = %[2]d
  preferred_subnet_id             = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}
`, rName, throughput, capacity))
}

func testAccONTAPFileSystemConfig_singleAZ2(rName string, throughput int, capacity int, haPairs int) string {
	return acctest.ConfigCompose(testAccONTAPFileSystemConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity                = %[3]d
  subnet_ids                      = [aws_subnet.test[0].id]
  deployment_type                 = "SINGLE_AZ_2"
  ha_pairs                        = %[4]d
  throughput_capacity_per_ha_pair = %[2]d
  preferred_subnet_id             = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}
`, rName, throughput, capacity, haPairs))
}

func testAccONTAPFileSystemConfig_haPair(rName string, capacity int) string {
	return acctest.ConfigCompose(testAccONTAPFileSystemConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity                = 2048
  subnet_ids                      = [aws_subnet.test[0].id]
  deployment_type                 = "SINGLE_AZ_2"
  ha_pairs                        = 2
  throughput_capacity_per_ha_pair = %[2]d
  preferred_subnet_id             = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}
`, rName, capacity))
}

func testAccONTAPFileSystemConfig_oneHaPair(rName string, capacity int) string {
	return acctest.ConfigCompose(testAccONTAPFileSystemConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity                = 1024
  subnet_ids                      = [aws_subnet.test[0].id]
  deployment_type                 = "SINGLE_AZ_1"
  ha_pairs                        = 1
  throughput_capacity_per_ha_pair = %[2]d
  preferred_subnet_id             = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}
`, rName, capacity))
}

func testAccONTAPFileSystemConfig_multiAZHaThrougput(rName string, capacity int) string {
	return acctest.ConfigCompose(testAccONTAPFileSystemConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity                = 2048
  subnet_ids                      = aws_subnet.test[*].id
  deployment_type                 = "MULTI_AZ_1"
  throughput_capacity_per_ha_pair = %[2]d
  ha_pairs                        = 1
  preferred_subnet_id             = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}
`, rName, capacity))
}

func testAccONTAPFileSystemConfig_adminPassword(rName, pass string) string {
	return acctest.ConfigCompose(testAccONTAPFileSystemConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity    = 1024
  subnet_ids          = aws_subnet.test[*].id
  deployment_type     = "MULTI_AZ_1"
  throughput_capacity = 128
  preferred_subnet_id = aws_subnet.test[0].id
  fsx_admin_password  = %[2]q

  tags = {
    Name = %[1]q
  }
}
`, rName, pass))
}

func testAccONTAPFileSystemConfig_endpointIPAddressRange(rName string) string {
	return acctest.ConfigCompose(testAccONTAPFileSystemConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity          = 1024
  subnet_ids                = aws_subnet.test[*].id
  deployment_type           = "MULTI_AZ_1"
  throughput_capacity       = 128
  preferred_subnet_id       = aws_subnet.test[0].id
  endpoint_ip_address_range = "198.19.255.0/24"

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccONTAPFileSystemConfig_diskIOPSConfiguration(rName string, iops int) string {
	return acctest.ConfigCompose(testAccONTAPFileSystemConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity    = 1024
  subnet_ids          = aws_subnet.test[*].id
  deployment_type     = "MULTI_AZ_1"
  throughput_capacity = 128
  preferred_subnet_id = aws_subnet.test[0].id

  disk_iops_configuration {
    mode = "USER_PROVISIONED"
    iops = %[2]d
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, iops))
}

func testAccONTAPFileSystemConfig_routeTable(rName string, cnt int) string {
	return acctest.ConfigCompose(testAccONTAPFileSystemConfig_base(rName), fmt.Sprintf(`
resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  count = %[2]d

  vpc_id = aws_vpc.test.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity    = 1024
  subnet_ids          = aws_subnet.test[*].id
  deployment_type     = "MULTI_AZ_1"
  throughput_capacity = 128
  preferred_subnet_id = aws_subnet.test[0].id
  route_table_ids     = aws_route_table.test[*].id

  tags = {
    Name = %[1]q
  }
}
`, rName, cnt))
}

func testAccONTAPFileSystemConfig_securityGroupIDs1(rName string) string {
	return acctest.ConfigCompose(testAccONTAPFileSystemConfig_base(rName), fmt.Sprintf(`
resource "aws_security_group" "test1" {
  name   = "%[1]s-1"
  vpc_id = aws_vpc.test.id

  ingress {
    cidr_blocks = [aws_vpc.test.cidr_block]
    from_port   = 0
    protocol    = -1
    to_port     = 0
  }

  egress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 0
    protocol    = "-1"
    to_port     = 0
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_fsx_ontap_file_system" "test" {
  security_group_ids  = [aws_security_group.test1.id]
  storage_capacity    = 1024
  subnet_ids          = aws_subnet.test[*].id
  deployment_type     = "MULTI_AZ_1"
  throughput_capacity = 128
  preferred_subnet_id = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccONTAPFileSystemConfig_securityGroupIDs2(rName string) string {
	return acctest.ConfigCompose(testAccONTAPFileSystemConfig_base(rName), fmt.Sprintf(`
resource "aws_security_group" "test1" {
  name   = "%[1]s-1"
  vpc_id = aws_vpc.test.id

  ingress {
    cidr_blocks = [aws_vpc.test.cidr_block]
    from_port   = 0
    protocol    = -1
    to_port     = 0
  }

  egress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 0
    protocol    = "-1"
    to_port     = 0
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test2" {
  name   = "%[1]s-2"
  vpc_id = aws_vpc.test.id

  ingress {
    cidr_blocks = [aws_vpc.test.cidr_block]
    from_port   = 0
    protocol    = -1
    to_port     = 0
  }

  egress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 0
    protocol    = "-1"
    to_port     = 0
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_fsx_ontap_file_system" "test" {
  security_group_ids  = [aws_security_group.test1.id, aws_security_group.test2.id]
  storage_capacity    = 1024
  subnet_ids          = aws_subnet.test[*].id
  deployment_type     = "MULTI_AZ_1"
  throughput_capacity = 128
  preferred_subnet_id = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccONTAPFileSystemConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccONTAPFileSystemConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity    = 1024
  subnet_ids          = aws_subnet.test[*].id
  deployment_type     = "MULTI_AZ_1"
  throughput_capacity = 128
  preferred_subnet_id = aws_subnet.test[0].id

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccONTAPFileSystemConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccONTAPFileSystemConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity    = 1024
  subnet_ids          = aws_subnet.test[*].id
  deployment_type     = "MULTI_AZ_1"
  throughput_capacity = 128
  preferred_subnet_id = aws_subnet.test[0].id

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccONTAPFileSystemConfig_weeklyMaintenanceStartTime(rName, weeklyMaintenanceStartTime string) string {
	return acctest.ConfigCompose(testAccONTAPFileSystemConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity              = 1024
  subnet_ids                    = aws_subnet.test[*].id
  deployment_type               = "MULTI_AZ_1"
  throughput_capacity           = 128
  preferred_subnet_id           = aws_subnet.test[0].id
  weekly_maintenance_start_time = %[2]q

  tags = {
    Name = %[1]q
  }
}
`, rName, weeklyMaintenanceStartTime))
}

func testAccONTAPFileSystemConfig_dailyAutomaticBackupStartTime(rName, dailyAutomaticBackupStartTime string) string {
	return acctest.ConfigCompose(testAccONTAPFileSystemConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity                  = 1024
  subnet_ids                        = aws_subnet.test[*].id
  deployment_type                   = "MULTI_AZ_1"
  throughput_capacity               = 128
  preferred_subnet_id               = aws_subnet.test[0].id
  daily_automatic_backup_start_time = %[2]q
  automatic_backup_retention_days   = 1

  tags = {
    Name = %[1]q
  }
}
`, rName, dailyAutomaticBackupStartTime))
}

func testAccONTAPFileSystemConfig_automaticBackupRetentionDays(rName string, retention int) string {
	return acctest.ConfigCompose(testAccONTAPFileSystemConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity                = 1024
  subnet_ids                      = aws_subnet.test[*].id
  deployment_type                 = "MULTI_AZ_1"
  throughput_capacity             = 128
  preferred_subnet_id             = aws_subnet.test[0].id
  automatic_backup_retention_days = %[2]d

  tags = {
    Name = %[1]q
  }
}
`, rName, retention))
}

func testAccONTAPFileSystemConfig_kmsKeyID(rName string) string {
	return acctest.ConfigCompose(testAccONTAPFileSystemConfig_base(rName), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity    = 1024
  subnet_ids          = aws_subnet.test[*].id
  deployment_type     = "MULTI_AZ_1"
  throughput_capacity = 128
  preferred_subnet_id = aws_subnet.test[0].id
  kms_key_id          = aws_kms_key.test.arn

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccONTAPFileSystemConfig_throughputCapacity(rName string) string {
	return acctest.ConfigCompose(testAccONTAPFileSystemConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity    = 1024
  subnet_ids          = aws_subnet.test[*].id
  deployment_type     = "MULTI_AZ_1"
  throughput_capacity = 256
  preferred_subnet_id = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccONTAPFileSystemConfig_storageCapacity(rName string) string {
	return acctest.ConfigCompose(testAccONTAPFileSystemConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity    = 2048
  subnet_ids          = aws_subnet.test[*].id
  deployment_type     = "MULTI_AZ_1"
  throughput_capacity = 128
  preferred_subnet_id = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}
