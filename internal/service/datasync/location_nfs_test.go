// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datasync_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/datasync"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdatasync "github.com/hashicorp/terraform-provider-aws/internal/service/datasync"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDataSyncLocationNFS_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v datasync.DescribeLocationNfsOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_location_nfs.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocationNFSDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLocationNFSConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationNFSExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "datasync", regexache.MustCompile(`location/loc-.+`)),
					resource.TestCheckResourceAttr(resourceName, "on_prem_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "on_prem_config.0.agent_arns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mount_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mount_options.0.version", "AUTOMATIC"),
					resource.TestCheckResourceAttr(resourceName, "server_hostname", "example.com"),
					resource.TestCheckResourceAttr(resourceName, "subdirectory", "/"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestMatchResourceAttr(resourceName, names.AttrURI, regexache.MustCompile(`^nfs://.+/`)),
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

func TestAccDataSyncLocationNFS_mountOptions(t *testing.T) {
	ctx := acctest.Context(t)
	var v datasync.DescribeLocationNfsOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_location_nfs.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocationNFSDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLocationNFSConfig_mountOptions(rName, "NFS4_0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationNFSExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mount_options.0.version", "NFS4_0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLocationNFSConfig_mountOptions(rName, "NFS4_1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationNFSExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mount_options.0.version", "NFS4_1"),
				),
			},
		},
	})
}

func TestAccDataSyncLocationNFS_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v datasync.DescribeLocationNfsOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_location_nfs.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocationNFSDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLocationNFSConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationNFSExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfdatasync.ResourceLocationNFS(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDataSyncLocationNFS_AgentARNs_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	var v datasync.DescribeLocationNfsOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_location_nfs.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocationNFSDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLocationNFSConfig_agentARNsMultiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationNFSExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "on_prem_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "on_prem_config.0.agent_arns.#", acctest.Ct2),
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

func TestAccDataSyncLocationNFS_subdirectory(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var v datasync.DescribeLocationNfsOutput
	resourceName := "aws_datasync_location_nfs.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocationNFSDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLocationNFSConfig_subdirectory(rName, "/subdirectory1/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationNFSExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "subdirectory", "/subdirectory1/"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLocationNFSConfig_subdirectory(rName, "/subdirectory2/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationNFSExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "subdirectory", "/subdirectory2/"),
				),
			},
		},
	})
}

func TestAccDataSyncLocationNFS_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v datasync.DescribeLocationNfsOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_location_nfs.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocationNFSDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLocationNFSConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationNFSExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLocationNFSConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationNFSExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccLocationNFSConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationNFSExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
		},
	})
}

func testAccCheckLocationNFSDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DataSyncClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_datasync_location_nfs" {
				continue
			}

			_, err := tfdatasync.FindLocationNFSByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("DataSync Location NFS %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckLocationNFSExists(ctx context.Context, n string, v *datasync.DescribeLocationNfsOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DataSyncClient(ctx)

		output, err := tfdatasync.FindLocationNFSByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccLocationNFSConfig_base(rName string) string {
	return acctest.ConfigCompose(testAccAgentAgentConfig_base(rName), fmt.Sprintf(`
resource "aws_datasync_agent" "test" {
  ip_address = aws_instance.test.public_ip
  name       = %[1]q
}
`, rName))
}

func testAccLocationNFSConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccLocationNFSConfig_base(rName), `
resource "aws_datasync_location_nfs" "test" {
  server_hostname = "example.com"
  subdirectory    = "/"

  on_prem_config {
    agent_arns = [aws_datasync_agent.test.arn]
  }
}
`)
}

func testAccLocationNFSConfig_mountOptions(rName, option string) string {
	return acctest.ConfigCompose(testAccLocationNFSConfig_base(rName), fmt.Sprintf(`
resource "aws_datasync_location_nfs" "test" {
  server_hostname = "example.com"
  subdirectory    = "/"

  on_prem_config {
    agent_arns = [aws_datasync_agent.test.arn]
  }

  mount_options {
    version = %[1]q
  }
}
`, option))
}

func testAccLocationNFSConfig_agentARNsMultiple(rName string) string {
	return acctest.ConfigCompose(testAccLocationNFSConfig_base(rName), fmt.Sprintf(`
resource "aws_instance" "test2" {
  depends_on = [aws_internet_gateway.test]

  ami                         = aws_instance.test.ami
  associate_public_ip_address = true
  instance_type               = aws_instance.test.instance_type
  vpc_security_group_ids      = [aws_security_group.test.id]
  subnet_id                   = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}

resource "aws_datasync_agent" "test2" {
  ip_address = aws_instance.test2.public_ip
  name       = "%[1]s-2"
}

resource "aws_datasync_location_nfs" "test" {
  server_hostname = "example.com"
  subdirectory    = "/"

  on_prem_config {
    agent_arns = [
      aws_datasync_agent.test.arn,
      aws_datasync_agent.test2.arn,
    ]
  }
}
`, rName))
}

func testAccLocationNFSConfig_subdirectory(rName, subdirectory string) string {
	return acctest.ConfigCompose(testAccLocationNFSConfig_base(rName), fmt.Sprintf(`
resource "aws_datasync_location_nfs" "test" {
  server_hostname = "example.com"
  subdirectory    = %[1]q

  on_prem_config {
    agent_arns = [aws_datasync_agent.test.arn]
  }
}
`, subdirectory))
}

func testAccLocationNFSConfig_tags1(rName, key1, value1 string) string {
	return acctest.ConfigCompose(testAccLocationNFSConfig_base(rName), fmt.Sprintf(`
resource "aws_datasync_location_nfs" "test" {
  server_hostname = "example.com"
  subdirectory    = "/"

  on_prem_config {
    agent_arns = [aws_datasync_agent.test.arn]
  }

  tags = {
    %[1]q = %[2]q
  }
}
`, key1, value1))
}

func testAccLocationNFSConfig_tags2(rName, key1, value1, key2, value2 string) string {
	return acctest.ConfigCompose(testAccLocationNFSConfig_base(rName), fmt.Sprintf(`
resource "aws_datasync_location_nfs" "test" {
  server_hostname = "example.com"
  subdirectory    = "/"

  on_prem_config {
    agent_arns = [aws_datasync_agent.test.arn]
  }

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, key1, value1, key2, value2))
}
