// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datasync_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/datasync"
	"github.com/aws/aws-sdk-go/service/fsx"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdatasync "github.com/hashicorp/terraform-provider-aws/internal/service/datasync"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDataSyncLocationFSxONTAPFileSystem_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var v datasync.DescribeLocationFsxOntapOutput
	resourceName := "aws_datasync_location_fsx_ontap_file_system.test"
	fsResourceName := "aws_fsx_ontap_file_system.test"
	svmResourceName := "aws_fsx_ontap_storage_virtual_machine.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, fsx.EndpointsID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocationFSxONTAPDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLocationFSxONTAPFileSystemConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLocationFSxONTAPExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "datasync", regexache.MustCompile(`location/loc-.+`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationTime),
					resource.TestCheckResourceAttrPair(resourceName, "fsx_filesystem_arn", fsResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "protocol.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "protocol.0.nfs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "protocol.0.nfs.0.mount_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "protocol.0.nfs.0.mount_options.0.version", "NFS3"),
					resource.TestCheckResourceAttr(resourceName, "protocol.0.smb.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "subdirectory", "/"),
					resource.TestCheckResourceAttrPair(resourceName, "storage_virtual_machine_arn", svmResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestMatchResourceAttr(resourceName, names.AttrURI, regexache.MustCompile(`^fsxn-(nfs|smb)://.+/`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccLocationFSxONTAPImportStateID(resourceName),
			},
		},
	})
}

func TestAccDataSyncLocationFSxONTAPFileSystem_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var v datasync.DescribeLocationFsxOntapOutput
	resourceName := "aws_datasync_location_fsx_ontap_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, fsx.EndpointsID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocationFSxONTAPDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLocationFSxONTAPFileSystemConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationFSxONTAPExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfdatasync.ResourceLocationFSxONTAPFileSystem(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDataSyncLocationFSxONTAPFileSystem_smb(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	netBiosName := "tftest-" + sdkacctest.RandString(7)
	domainNetbiosName := "tftest" + sdkacctest.RandString(4)
	domainName := domainNetbiosName + ".local"
	var v datasync.DescribeLocationFsxOntapOutput
	resourceName := "aws_datasync_location_fsx_ontap_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, fsx.EndpointsID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocationFSxONTAPDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLocationFSxONTAPFileSystemConfig_smb(rName, netBiosName, domainNetbiosName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLocationFSxONTAPExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "protocol.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "protocol.0.nfs.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "protocol.0.smb.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "protocol.0.smb.0.domain", domainName),
					resource.TestCheckResourceAttr(resourceName, "protocol.0.smb.0.mount_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "protocol.0.smb.0.mount_options.0.version", "SMB3"),
					resource.TestCheckResourceAttr(resourceName, "protocol.0.smb.0.password", "MyPassw0rd1"),
					resource.TestCheckResourceAttr(resourceName, "protocol.0.smb.0.user", "Admin"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"protocol.0.smb.0.password"}, // Not returned from API.
				ImportStateIdFunc:       testAccLocationFSxONTAPImportStateID(resourceName),
			},
		},
	})
}

func TestAccDataSyncLocationFSxONTAPFileSystem_subdirectory(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var v datasync.DescribeLocationFsxOntapOutput
	resourceName := "aws_datasync_location_fsx_ontap_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, fsx.EndpointsID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocationFSxONTAPDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLocationFSxONTAPFileSystemConfig_subdirectory(rName, "/subdirectory1/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationFSxONTAPExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "subdirectory", "/subdirectory1/"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccLocationFSxONTAPImportStateID(resourceName),
			},
		},
	})
}

func TestAccDataSyncLocationFSxONTAPFileSystem_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var v datasync.DescribeLocationFsxOntapOutput
	resourceName := "aws_datasync_location_fsx_ontap_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, fsx.EndpointsID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocationFSxONTAPDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLocationFSxONTAPFileSystemConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationFSxONTAPExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccLocationFSxONTAPImportStateID(resourceName),
			},
			{
				Config: testAccLocationFSxONTAPFileSystemConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationFSxONTAPExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccLocationFSxONTAPFileSystemConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationFSxONTAPExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
		},
	})
}

func testAccCheckLocationFSxONTAPDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DataSyncClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_datasync_location_fsx_ontap_file_system" {
				continue
			}

			_, err := tfdatasync.FindLocationFSxONTAPByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("DataSync Location FSx for NetApp ONTAP File System %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckLocationFSxONTAPExists(ctx context.Context, n string, v *datasync.DescribeLocationFsxOntapOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DataSyncClient(ctx)

		output, err := tfdatasync.FindLocationFSxONTAPByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccLocationFSxONTAPImportStateID(n string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return "", fmt.Errorf("Not found: %s", n)
		}

		return fmt.Sprintf("%s#%s", rs.Primary.ID, rs.Primary.Attributes["fsx_filesystem_arn"]), nil
	}
}

func testAccFSxOntapFileSystemConfig_base(rName string, nSubnets int, deploymentType string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, nSubnets), fmt.Sprintf(`
resource "aws_security_group" "test" {
  name   = %[1]q
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
  storage_capacity    = 1024
  subnet_ids          = aws_subnet.test[*].id
  deployment_type     = %[2]q
  throughput_capacity = 512
  preferred_subnet_id = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}
`, rName, deploymentType))
}

func testAccFSxOntapFileSystemConfig_baseNFS(rName string) string {
	return acctest.ConfigCompose(testAccFSxOntapFileSystemConfig_base(rName, 1, "SINGLE_AZ_1"), fmt.Sprintf(`
resource "aws_fsx_ontap_storage_virtual_machine" "test" {
  file_system_id = aws_fsx_ontap_file_system.test.id
  name           = %[1]q
}
`, rName))
}

func testAccFSxOntapFileSystemConfig_baseSMB(rName, domainName string) string {
	return acctest.ConfigCompose(testAccFSxOntapFileSystemConfig_base(rName, 2, "MULTI_AZ_1"), fmt.Sprintf(`
resource "aws_directory_service_directory" "test" {
  edition  = "Standard"
  name     = %[1]q
  password = "MyPassw0rd1"
  type     = "MicrosoftAD"

  vpc_settings {
    subnet_ids = aws_subnet.test[*].id
    vpc_id     = aws_vpc.test.id
  }
}
`, domainName))
}

func testAccLocationFSxONTAPFileSystemConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccFSxOntapFileSystemConfig_baseNFS(rName), `
resource "aws_datasync_location_fsx_ontap_file_system" "test" {
  security_group_arns         = [aws_security_group.test.arn]
  storage_virtual_machine_arn = aws_fsx_ontap_storage_virtual_machine.test.arn

  protocol {
    nfs {
      mount_options {
        version = "NFS3"
      }
    }
  }
}
`)
}

func testAccLocationFSxONTAPFileSystemConfig_smb(rName, netBiosName, domainNetbiosName, domainName string) string {
	return acctest.ConfigCompose(testAccFSxOntapFileSystemConfig_baseSMB(rName, domainName), fmt.Sprintf(`
resource "aws_fsx_ontap_storage_virtual_machine" "test" {
  file_system_id = aws_fsx_ontap_file_system.test.id
  name           = %[1]q
  depends_on     = [aws_directory_service_directory.test]

  active_directory_configuration {
    netbios_name = %[2]q
    self_managed_active_directory_configuration {
      dns_ips                                = aws_directory_service_directory.test.dns_ip_addresses
      domain_name                            = %[3]q
      password                               = "MyPassw0rd1"
      username                               = "Admin"
      organizational_unit_distinguished_name = "OU=computers,OU=%[4]s"
      file_system_administrators_group       = "Admins"
    }
  }
}

resource "aws_datasync_location_fsx_ontap_file_system" "test" {
  security_group_arns         = [aws_security_group.test.arn]
  storage_virtual_machine_arn = aws_fsx_ontap_storage_virtual_machine.test.arn

  protocol {
    smb {
      domain = %[3]q

      mount_options {
        version = "SMB3"
      }

      password = "MyPassw0rd1"
      user     = "Admin"
    }
  }
}
`, rName, netBiosName, domainName, domainNetbiosName))
}

func testAccLocationFSxONTAPFileSystemConfig_subdirectory(rName, subdirectory string) string {
	return acctest.ConfigCompose(testAccFSxOntapFileSystemConfig_baseNFS(rName), fmt.Sprintf(`
resource "aws_datasync_location_fsx_ontap_file_system" "test" {
  security_group_arns         = [aws_security_group.test.arn]
  storage_virtual_machine_arn = aws_fsx_ontap_storage_virtual_machine.test.arn
  subdirectory                = %[1]q

  protocol {
    nfs {
      mount_options {
        version = "NFS3"
      }
    }
  }
}
`, subdirectory))
}

func testAccLocationFSxONTAPFileSystemConfig_tags1(rName, key1, value1 string) string {
	return acctest.ConfigCompose(testAccFSxOntapFileSystemConfig_baseNFS(rName), fmt.Sprintf(`
resource "aws_datasync_location_fsx_ontap_file_system" "test" {
  security_group_arns         = [aws_security_group.test.arn]
  storage_virtual_machine_arn = aws_fsx_ontap_storage_virtual_machine.test.arn

  tags = {
    %[1]q = %[2]q
  }

  protocol {
    nfs {
      mount_options {
        version = "NFS3"
      }
    }
  }
}
`, key1, value1))
}

func testAccLocationFSxONTAPFileSystemConfig_tags2(rName, key1, value1, key2, value2 string) string {
	return acctest.ConfigCompose(testAccFSxOntapFileSystemConfig_baseNFS(rName), fmt.Sprintf(`
resource "aws_datasync_location_fsx_ontap_file_system" "test" {
  security_group_arns         = [aws_security_group.test.arn]
  storage_virtual_machine_arn = aws_fsx_ontap_storage_virtual_machine.test.arn

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }

  protocol {
    nfs {
      mount_options {
        version = "NFS3"
      }
    }
  }
}
`, key1, value1, key2, value2))
}
