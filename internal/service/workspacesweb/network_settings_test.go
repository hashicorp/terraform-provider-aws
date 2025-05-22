// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package workspacesweb_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/workspacesweb/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfworkspacesweb "github.com/hashicorp/terraform-provider-aws/internal/service/workspacesweb"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccWorkSpacesWebNetworkSettings_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var networkSettings awstypes.NetworkSettings
	resourceName := "aws_workspacesweb_network_settings.test"
	vpcResourceName := "aws_vpc.test"
	subnetResourceName1 := "aws_subnet.test.0"
	subnetResourceName2 := "aws_subnet.test.1"
	securityGroupResourceName1 := "aws_security_group.test.0"
	securityGroupResourceName2 := "aws_security_group.test.1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkSettingsConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkSettingsExists(ctx, resourceName, &networkSettings),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrVPCID, vpcResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_ids.0", subnetResourceName1, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_ids.1", subnetResourceName2, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "2"),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_ids.0", securityGroupResourceName1, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_ids.1", securityGroupResourceName2, names.AttrID),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "network_settings_arn", "workspaces-web", regexache.MustCompile(`networkSettings/.+$`)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccNetworkSettingsImportStateIdFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: "network_settings_arn",
			},
		},
	})
}

func TestAccWorkSpacesWebNetworkSettings_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var networkSettings awstypes.NetworkSettings
	resourceName := "aws_workspacesweb_network_settings.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkSettingsConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkSettingsExists(ctx, resourceName, &networkSettings),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfworkspacesweb.ResourceNetworkSettings, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWorkSpacesWebNetworkSettings_update(t *testing.T) {
	ctx := acctest.Context(t)
	var networkSettings awstypes.NetworkSettings
	resourceName := "aws_workspacesweb_network_settings.test"
	vpcResourceName := "aws_vpc.test"
	vpcResourceName2 := "aws_vpc.test2"
	subnetResourceName1 := "aws_subnet.test.0"
	subnetResourceName2 := "aws_subnet.test.1"
	subnetResourceName3 := "aws_subnet.test2.0"
	subnetResourceName4 := "aws_subnet.test2.1"
	securityGroupResourceName1 := "aws_security_group.test.0"
	securityGroupResourceName2 := "aws_security_group.test.1"
	securityGroupResourceName3 := "aws_security_group.test2.0"
	securityGroupResourceName4 := "aws_security_group.test2.1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkSettingsConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkSettingsExists(ctx, resourceName, &networkSettings),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrVPCID, vpcResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_ids.0", subnetResourceName1, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_ids.1", subnetResourceName2, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "2"),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_ids.0", securityGroupResourceName1, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_ids.1", securityGroupResourceName2, names.AttrID),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccNetworkSettingsImportStateIdFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: "network_settings_arn",
			},
			{
				Config: testAccNetworkSettingsConfig_updated(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkSettingsExists(ctx, resourceName, &networkSettings),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrVPCID, vpcResourceName2, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_ids.0", subnetResourceName3, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_ids.1", subnetResourceName4, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "2"),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_ids.0", securityGroupResourceName3, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_ids.1", securityGroupResourceName4, names.AttrID),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccNetworkSettingsImportStateIdFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: "network_settings_arn",
			},
		},
	})
}

func testAccCheckNetworkSettingsDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).WorkSpacesWebClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_workspacesweb_network_settings" {
				continue
			}

			_, err := tfworkspacesweb.FindNetworkSettingsByARN(ctx, conn, rs.Primary.Attributes["network_settings_arn"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("WorkSpaces Web Network Settings %s still exists", rs.Primary.Attributes["network_settings_arn"])
		}

		return nil
	}
}

func testAccCheckNetworkSettingsExists(ctx context.Context, n string, v *awstypes.NetworkSettings) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WorkSpacesWebClient(ctx)

		output, err := tfworkspacesweb.FindNetworkSettingsByARN(ctx, conn, rs.Primary.Attributes["network_settings_arn"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccNetworkSettingsImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.Attributes["network_settings_arn"], nil
	}
}

func testAccNetworkSettingsConfig_base() string {
	return `
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_vpc" "test2" {
  cidr_block = "10.1.0.0/16"
}

resource "aws_subnet" "test" {
  count = 2

  vpc_id            = aws_vpc.test.id
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  availability_zone = data.aws_availability_zones.available.names[count.index]
}

resource "aws_subnet" "test2" {
  count = 2

  vpc_id            = aws_vpc.test2.id
  cidr_block        = cidrsubnet(aws_vpc.test2.cidr_block, 8, count.index)
  availability_zone = data.aws_availability_zones.available.names[count.index]
}

resource "aws_security_group" "test" {
  count = 2

  vpc_id = aws_vpc.test.id
  name   = "test-sg-${count.index}"
}

resource "aws_security_group" "test2" {
  count = 2

  vpc_id = aws_vpc.test2.id
  name   = "test-sg-${count.index}"
}

data "aws_availability_zones" "available" {
  state = "available"
  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}
`
}

func testAccNetworkSettingsConfig_basic() string {
	return acctest.ConfigCompose(testAccNetworkSettingsConfig_base(), `
resource "aws_workspacesweb_network_settings" "test" {
  vpc_id             = aws_vpc.test.id
  subnet_ids         = [aws_subnet.test[0].id, aws_subnet.test[1].id]
  security_group_ids = [aws_security_group.test[0].id, aws_security_group.test[1].id]
}
`)
}

func testAccNetworkSettingsConfig_updated() string {
	return acctest.ConfigCompose(testAccNetworkSettingsConfig_base(), `
resource "aws_workspacesweb_network_settings" "test" {
  vpc_id             = aws_vpc.test2.id
  subnet_ids         = [aws_subnet.test2[0].id, aws_subnet.test2[1].id]
  security_group_ids = [aws_security_group.test2[0].id, aws_security_group.test2[1].id]
}
`)
}
