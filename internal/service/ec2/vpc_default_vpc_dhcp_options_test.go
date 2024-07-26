// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCDefaultVPCDHCPOptions_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:      testAccDefaultVPCDHCPOptions_basic,
		"owner":              testAccDefaultVPCDHCPOptions_owner,
		"v4.20.0_regression": testAccDefaultVPCDHCPOptions_v420Regression,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccDefaultVPCDHCPOptions_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var d awstypes.DhcpOptions
	resourceName := "aws_default_vpc_dhcp_options.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDefaultVPCDHCPOptionsConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDHCPOptionsExists(ctx, resourceName, &d),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`dhcp-options/dopt-.+`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrDomainName),
					resource.TestCheckResourceAttr(resourceName, "domain_name_servers", "AmazonProvidedDNS"),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "Default DHCP Option Set"),
				),
			},
		},
	})
}

func testAccDefaultVPCDHCPOptions_owner(t *testing.T) {
	ctx := acctest.Context(t)
	var d awstypes.DhcpOptions
	resourceName := "aws_default_vpc_dhcp_options.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDefaultVPCDHCPOptionsConfig_owner,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDHCPOptionsExists(ctx, resourceName, &d),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`dhcp-options/dopt-.+`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrDomainName),
					resource.TestCheckResourceAttr(resourceName, "domain_name_servers", "AmazonProvidedDNS"),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "Default DHCP Option Set"),
				),
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/25561.
// Don't forget to unset TF_CLI_CONFIG_FILE.
func testAccDefaultVPCDHCPOptions_v420Regression(t *testing.T) {
	ctx := acctest.Context(t)
	var d awstypes.DhcpOptions
	resourceName := "aws_default_vpc_dhcp_options.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy: acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "4.19.0",
					},
				},
				Config: testAccVPCDefaultVPCDHCPOptionsConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDHCPOptionsExists(ctx, resourceName, &d),
				),
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccVPCDefaultVPCDHCPOptionsConfig_basic,
				PlanOnly:                 true,
			},
		},
	})
}

const testAccVPCDefaultVPCDHCPOptionsConfig_basic = `
resource "aws_default_vpc_dhcp_options" "test" {
  tags = {
    Name = "Default DHCP Option Set"
  }
}
`

const testAccVPCDefaultVPCDHCPOptionsConfig_owner = `
data "aws_caller_identity" "current" {}

resource "aws_default_vpc_dhcp_options" "test" {
  owner_id = data.aws_caller_identity.current.account_id

  tags = {
    Name = "Default DHCP Option Set"
  }
}
`
