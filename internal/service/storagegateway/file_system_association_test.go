// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package storagegateway_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/service/storagegateway"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tffsx "github.com/hashicorp/terraform-provider-aws/internal/service/fsx"
	tfstoragegateway "github.com/hashicorp/terraform-provider-aws/internal/service/storagegateway"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccStorageGatewayFileSystemAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var fileSystemAssociation storagegateway.FileSystemAssociationInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_file_system_association.test"
	gatewayResourceName := "aws_storagegateway_gateway.test"
	fsxResourceName := "aws_fsx_windows_file_system.test"
	domainName := acctest.RandomDomainName()
	username := "Admin"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, storagegateway.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileSystemAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFileSystemAssociationConfig_required(rName, domainName, username),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystemAssociationExists(ctx, resourceName, &fileSystemAssociation),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "storagegateway", regexache.MustCompile(`fs-association/fsa-.+`)),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_arn", gatewayResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "location_arn", fsxResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrUsername, username),
					resource.TestCheckResourceAttr(resourceName, "cache_attributes.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrUsername, names.AttrPassword},
			},
		},
	})
}

func TestAccStorageGatewayFileSystemAssociation_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var fileSystemAssociation storagegateway.FileSystemAssociationInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_file_system_association.test"
	domainName := acctest.RandomDomainName()
	username := "Admin"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, storagegateway.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileSystemAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFileSystemAssociationConfig_tags1(rName, domainName, username, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystemAssociationExists(ctx, resourceName, &fileSystemAssociation),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "storagegateway", regexache.MustCompile(`fs-association/fsa-.+`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrUsername, names.AttrPassword},
			},
			{
				Config: testAccFileSystemAssociationConfig_tags2(rName, domainName, username, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystemAssociationExists(ctx, resourceName, &fileSystemAssociation),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "storagegateway", regexache.MustCompile(`fs-association/fsa-.+`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccFileSystemAssociationConfig_tags1(rName, domainName, username, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystemAssociationExists(ctx, resourceName, &fileSystemAssociation),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "storagegateway", regexache.MustCompile(`fs-association/fsa-.+`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccStorageGatewayFileSystemAssociation_cacheAttributes(t *testing.T) {
	ctx := acctest.Context(t)
	var fileSystemAssociation storagegateway.FileSystemAssociationInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_file_system_association.test"
	domainName := acctest.RandomDomainName()
	username := "Admin"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, storagegateway.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileSystemAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFileSystemAssociationConfig_cache(rName, domainName, username, 400),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystemAssociationExists(ctx, resourceName, &fileSystemAssociation),
					resource.TestCheckResourceAttr(resourceName, "cache_attributes.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cache_attributes.0.cache_stale_timeout_in_seconds", "400"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrUsername, names.AttrPassword},
			},
			{
				Config: testAccFileSystemAssociationConfig_cache(rName, domainName, username, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystemAssociationExists(ctx, resourceName, &fileSystemAssociation),
					resource.TestCheckResourceAttr(resourceName, "cache_attributes.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cache_attributes.0.cache_stale_timeout_in_seconds", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccStorageGatewayFileSystemAssociation_auditDestination(t *testing.T) {
	ctx := acctest.Context(t)
	var fileSystemAssociation storagegateway.FileSystemAssociationInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_file_system_association.test"
	domainName := acctest.RandomDomainName()
	username := "Admin"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, storagegateway.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileSystemAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFileSystemAssociationConfig_audit(rName, domainName, username, "aws_cloudwatch_log_group.test.arn"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystemAssociationExists(ctx, resourceName, &fileSystemAssociation),
					resource.TestCheckResourceAttrPair(resourceName, "audit_destination_arn", "aws_cloudwatch_log_group.test", names.AttrARN),
				),
			}, {
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrUsername, names.AttrPassword},
			},
			{
				Config: testAccFileSystemAssociationConfig_auditDisabled(rName, domainName, username),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystemAssociationExists(ctx, resourceName, &fileSystemAssociation),
					resource.TestCheckResourceAttr(resourceName, "audit_destination_arn", ""),
				),
			},
			{
				Config: testAccFileSystemAssociationConfig_audit(rName, domainName, username, "aws_cloudwatch_log_group.test2.arn"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystemAssociationExists(ctx, resourceName, &fileSystemAssociation),
					resource.TestCheckResourceAttrPair(resourceName, "audit_destination_arn", "aws_cloudwatch_log_group.test2", names.AttrARN),
				),
			},
		},
	})
}

func TestAccStorageGatewayFileSystemAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var fileSystemAssociation storagegateway.FileSystemAssociationInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_file_system_association.test"
	domainName := acctest.RandomDomainName()
	username := "Admin"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, storagegateway.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileSystemAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFileSystemAssociationConfig_required(rName, domainName, username),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystemAssociationExists(ctx, resourceName, &fileSystemAssociation),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfstoragegateway.ResourceFileSystemAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccStorageGatewayFileSystemAssociation_Disappears_storageGateway(t *testing.T) {
	ctx := acctest.Context(t)
	var fileSystemAssociation storagegateway.FileSystemAssociationInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_file_system_association.test"
	domainName := acctest.RandomDomainName()
	username := "Admin"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, storagegateway.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileSystemAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFileSystemAssociationConfig_required(rName, domainName, username),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystemAssociationExists(ctx, resourceName, &fileSystemAssociation),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfstoragegateway.ResourceGateway(), "aws_storagegateway_gateway.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccStorageGatewayFileSystemAssociation_Disappears_fsxFileSystem(t *testing.T) {
	t.Skip("A bug in the service API has been reported. Deleting the FSx file system before the association prevents association from being deleted.")
	ctx := acctest.Context(t)

	var fileSystemAssociation storagegateway.FileSystemAssociationInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_file_system_association.test"
	domainName := acctest.RandomDomainName()
	username := "Admin"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, storagegateway.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileSystemAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFileSystemAssociationConfig_required(rName, domainName, username),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystemAssociationExists(ctx, resourceName, &fileSystemAssociation),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tffsx.ResourceWindowsFileSystem(), "aws_fsx_windows_file_system.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckFileSystemAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).StorageGatewayConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_storagegateway_file_system_association" {
				continue
			}

			_, err := tfstoragegateway.FindFileSystemAssociationByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Storage Gateway File System Association %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckFileSystemAssociationExists(ctx context.Context, resourceName string, v *storagegateway.FileSystemAssociationInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).StorageGatewayConn(ctx)

		output, err := tfstoragegateway.FindFileSystemAssociationByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccFileSystemAssociationBase(rName, domainName, username string) string {
	return acctest.ConfigCompose(
		testAccGatewaySMBActiveDirectorySettingsBaseConfig(rName),
		testAccGatewayConfig_DirectoryServiceMicrosoftAD(rName, domainName),
		fmt.Sprintf(`
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = aws_directory_service_directory.test.id
  security_group_ids  = [aws_security_group.test.id]
  skip_final_backup   = true
  storage_capacity    = 32
  subnet_ids          = [aws_subnet.test[0].id]
  throughput_capacity = 8

  tags = {
    Name = %[1]q
  }
}

resource "aws_storagegateway_gateway" "test" {
  gateway_ip_address = aws_instance.test.public_ip
  gateway_name       = %[1]q
  gateway_timezone   = "GMT"
  gateway_type       = "FILE_FSX_SMB"

  smb_active_directory_settings {
    domain_name = aws_directory_service_directory.test.name
    password    = aws_directory_service_directory.test.password
    username    = %[2]q
  }

  tags = {
    Name = %[1]q
  }
}

`, rName, username))
}

func testAccFileSystemAssociationConfig_required(rName, domainName, username string) string {
	return testAccFileSystemAssociationBase(rName, domainName, username) + fmt.Sprintf(`
resource "aws_storagegateway_file_system_association" "test" {
  gateway_arn  = aws_storagegateway_gateway.test.arn
  location_arn = aws_fsx_windows_file_system.test.arn
  username     = %[1]q
  password     = aws_directory_service_directory.test.password
}
`, username)
}

func testAccFileSystemAssociationConfig_tags1(rName, domainName, username, tagKey1, tagValue1 string) string {
	return testAccFileSystemAssociationBase(rName, domainName, username) + fmt.Sprintf(`
resource "aws_storagegateway_file_system_association" "test" {
  gateway_arn  = aws_storagegateway_gateway.test.arn
  location_arn = aws_fsx_windows_file_system.test.arn
  username     = %[1]q
  password     = aws_directory_service_directory.test.password

  tags = {
    %[2]q = %[3]q
  }
}
`, username, tagKey1, tagValue1)
}

func testAccFileSystemAssociationConfig_tags2(rName, domainName, username, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccFileSystemAssociationBase(rName, domainName, username) + fmt.Sprintf(`
resource "aws_storagegateway_file_system_association" "test" {
  gateway_arn  = aws_storagegateway_gateway.test.arn
  location_arn = aws_fsx_windows_file_system.test.arn
  username     = %[1]q
  password     = aws_directory_service_directory.test.password

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, username, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccFileSystemAssociationConfig_audit(rName, domainName, username string, loggingDestination string) string {
	return testAccFileSystemAssociationBase(rName, domainName, username) + fmt.Sprintf(`
resource "aws_storagegateway_file_system_association" "test" {
  gateway_arn           = aws_storagegateway_gateway.test.arn
  location_arn          = aws_fsx_windows_file_system.test.arn
  username              = %[1]q
  password              = aws_directory_service_directory.test.password
  audit_destination_arn = %[2]s
}

resource "aws_cloudwatch_log_group" "test" {}
resource "aws_cloudwatch_log_group" "test2" {}
`, username, loggingDestination)
}
func testAccFileSystemAssociationConfig_auditDisabled(rName, domainName, username string) string {
	return testAccFileSystemAssociationBase(rName, domainName, username) + fmt.Sprintf(`
resource "aws_storagegateway_file_system_association" "test" {
  gateway_arn           = aws_storagegateway_gateway.test.arn
  location_arn          = aws_fsx_windows_file_system.test.arn
  username              = %[1]q
  password              = aws_directory_service_directory.test.password
  audit_destination_arn = ""
}

resource "aws_cloudwatch_log_group" "test" {}
resource "aws_cloudwatch_log_group" "test2" {}
`, username)
}

func testAccFileSystemAssociationConfig_cache(rName, domainName, username string, cache int) string {
	return testAccFileSystemAssociationBase(rName, domainName, username) + fmt.Sprintf(`
resource "aws_storagegateway_file_system_association" "test" {
  gateway_arn  = aws_storagegateway_gateway.test.arn
  location_arn = aws_fsx_windows_file_system.test.arn
  username     = %[1]q
  password     = aws_directory_service_directory.test.password

  cache_attributes {
    cache_stale_timeout_in_seconds = %[2]d
  }
}
`, username, cache)
}
