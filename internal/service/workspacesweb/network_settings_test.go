// Copyright IBM Corp. 2014, 2026
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
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfworkspacesweb "github.com/hashicorp/terraform-provider-aws/internal/service/workspacesweb"
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkSettingsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkSettingsConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkSettingsExists(ctx, t, resourceName, &networkSettings),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrVPCID, vpcResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_ids.*", subnetResourceName1, names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_ids.*", subnetResourceName2, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_group_ids.*", securityGroupResourceName1, names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_group_ids.*", securityGroupResourceName2, names.AttrID),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "network_settings_arn", "workspaces-web", regexache.MustCompile(`networkSettings/.+$`)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "network_settings_arn"),
				ImportStateVerifyIdentifierAttribute: "network_settings_arn",
			},
		},
	})
}

func TestAccWorkSpacesWebNetworkSettings_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var networkSettings awstypes.NetworkSettings
	resourceName := "aws_workspacesweb_network_settings.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkSettingsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkSettingsConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkSettingsExists(ctx, t, resourceName, &networkSettings),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfworkspacesweb.ResourceNetworkSettings, resourceName),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkSettingsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkSettingsConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkSettingsExists(ctx, t, resourceName, &networkSettings),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrVPCID, vpcResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_ids.*", subnetResourceName1, names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_ids.*", subnetResourceName2, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_group_ids.*", securityGroupResourceName1, names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_group_ids.*", securityGroupResourceName2, names.AttrID),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "network_settings_arn"),
				ImportStateVerifyIdentifierAttribute: "network_settings_arn",
			},
			{
				Config: testAccNetworkSettingsConfig_updated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkSettingsExists(ctx, t, resourceName, &networkSettings),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrVPCID, vpcResourceName2, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_ids.*", subnetResourceName3, names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_ids.*", subnetResourceName4, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_group_ids.*", securityGroupResourceName3, names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_group_ids.*", securityGroupResourceName4, names.AttrID),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "network_settings_arn"),
				ImportStateVerifyIdentifierAttribute: "network_settings_arn",
			},
		},
	})
}

func testAccCheckNetworkSettingsDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).WorkSpacesWebClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_workspacesweb_network_settings" {
				continue
			}

			_, err := tfworkspacesweb.FindNetworkSettingsByARN(ctx, conn, rs.Primary.Attributes["network_settings_arn"])

			if retry.NotFound(err) {
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

func testAccCheckNetworkSettingsExists(ctx context.Context, t *testing.T, n string, v *awstypes.NetworkSettings) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).WorkSpacesWebClient(ctx)

		output, err := tfworkspacesweb.FindNetworkSettingsByARN(ctx, conn, rs.Primary.Attributes["network_settings_arn"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccNetworkSettingsConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "test2" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 2

  vpc_id            = aws_vpc.test.id
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  availability_zone = data.aws_availability_zones.available.names[count.index]

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test2" {
  count = 2

  vpc_id            = aws_vpc.test2.id
  cidr_block        = cidrsubnet(aws_vpc.test2.cidr_block, 8, count.index)
  availability_zone = data.aws_availability_zones.available.names[count.index]

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  count = 2

  vpc_id = aws_vpc.test.id
  name   = "%[1]s-1-${count.index}"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test2" {
  count = 2

  vpc_id = aws_vpc.test2.id
  name   = "%[1]s-2-${count.index}"
}
`, rName))
}

func testAccNetworkSettingsConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccNetworkSettingsConfig_base(rName), `
resource "aws_workspacesweb_network_settings" "test" {
  vpc_id             = aws_vpc.test.id
  subnet_ids         = [aws_subnet.test[0].id, aws_subnet.test[1].id]
  security_group_ids = [aws_security_group.test[0].id, aws_security_group.test[1].id]
}
`)
}

func testAccNetworkSettingsConfig_updated(rName string) string {
	return acctest.ConfigCompose(testAccNetworkSettingsConfig_base(rName), `
resource "aws_workspacesweb_network_settings" "test" {
  vpc_id             = aws_vpc.test2.id
  subnet_ids         = [aws_subnet.test2[0].id, aws_subnet.test2[1].id]
  security_group_ids = [aws_security_group.test2[0].id, aws_security_group.test2[1].id]
}
`)
}
