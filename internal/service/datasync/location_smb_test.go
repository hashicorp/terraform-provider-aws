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

func TestAccDataSyncLocationSMB_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v datasync.DescribeLocationSmbOutput
	resourceName := "aws_datasync_location_smb.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocationSMBDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLocationSMBConfig_basic(rName, "/test/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationSMBExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "datasync", regexache.MustCompile(`location/loc-.+`)),
					resource.TestCheckResourceAttr(resourceName, "agent_arns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mount_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mount_options.0.version", "AUTOMATIC"),
					resource.TestCheckResourceAttr(resourceName, "user", "Guest"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestMatchResourceAttr(resourceName, names.AttrURI, regexache.MustCompile(`^smb://.+/`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
			{
				Config: testAccLocationSMBConfig_basic(rName, "/test2/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationSMBExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "datasync", regexache.MustCompile(`location/loc-.+`)),
					resource.TestCheckResourceAttr(resourceName, "agent_arns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mount_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mount_options.0.version", "AUTOMATIC"),
					resource.TestCheckResourceAttr(resourceName, "user", "Guest"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestMatchResourceAttr(resourceName, names.AttrURI, regexache.MustCompile(`^smb://.+/`)),
				),
			},
		},
	})
}

func TestAccDataSyncLocationSMB_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v datasync.DescribeLocationSmbOutput
	resourceName := "aws_datasync_location_smb.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocationSMBDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLocationSMBConfig_basic(rName, "/test/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationSMBExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfdatasync.ResourceLocationSMB(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDataSyncLocationSMB_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v datasync.DescribeLocationSmbOutput
	resourceName := "aws_datasync_location_smb.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocationSMBDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLocationSMBConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationSMBExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
			{
				Config: testAccLocationSMBConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationSMBExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccLocationSMBConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationSMBExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
		},
	})
}

func testAccCheckLocationSMBDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DataSyncClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_datasync_location_smb" {
				continue
			}

			_, err := tfdatasync.FindLocationSMBByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("DataSync Location SMB %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckLocationSMBExists(ctx context.Context, n string, v *datasync.DescribeLocationSmbOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DataSyncClient(ctx)

		output, err := tfdatasync.FindLocationSMBByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccLocationSMBConfig_base(rName string) string {
	return acctest.ConfigCompose(testAccAgentAgentConfig_base(rName), fmt.Sprintf(`
resource "aws_datasync_agent" "test" {
  ip_address = aws_instance.test.public_ip
  name       = %[1]q
}
`, rName))
}

func testAccLocationSMBConfig_basic(rName, dir string) string {
	return acctest.ConfigCompose(testAccLocationSMBConfig_base(rName), fmt.Sprintf(`
resource "aws_datasync_location_smb" "test" {
  agent_arns      = [aws_datasync_agent.test.arn]
  password        = "ZaphodBeeblebroxPW"
  server_hostname = aws_instance.test.public_ip
  subdirectory    = %[1]q
  user            = "Guest"
}
`, dir))
}

func testAccLocationSMBConfig_tags1(rName, key1, value1 string) string {
	return acctest.ConfigCompose(testAccLocationSMBConfig_base(rName), fmt.Sprintf(`
resource "aws_datasync_location_smb" "test" {
  agent_arns      = [aws_datasync_agent.test.arn]
  password        = "ZaphodBeeblebroxPW"
  server_hostname = aws_instance.test.public_ip
  subdirectory    = "/test/"
  user            = "Guest"

  tags = {
    %[1]q = %[2]q
  }
}
`, key1, value1))
}

func testAccLocationSMBConfig_tags2(rName, key1, value1, key2, value2 string) string {
	return acctest.ConfigCompose(testAccLocationSMBConfig_base(rName), fmt.Sprintf(`
resource "aws_datasync_location_smb" "test" {
  agent_arns      = [aws_datasync_agent.test.arn]
  password        = "ZaphodBeeblebroxPW"
  server_hostname = aws_instance.test.public_ip
  subdirectory    = "/test/"
  user            = "Guest"

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, key1, value1, key2, value2))
}
