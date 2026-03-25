// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package directconnect_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directconnect/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDirectConnectConnection_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var connection awstypes.Connection
	resourceName := "aws_dx_connection.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectionExists(ctx, t, resourceName, &connection),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "directconnect", regexache.MustCompile(`dxcon/.+`)),
					resource.TestCheckResourceAttr(resourceName, "bandwidth", "1Gbps"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLocation),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrOwnerAccountID),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "partner_name", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrProviderName, ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, "vlan_id", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"request_macsec", names.AttrSkipDestroy},
			},
		},
	})
}

func TestAccDirectConnectConnection_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var connection awstypes.Connection
	resourceName := "aws_dx_connection.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(ctx, t, resourceName, &connection),
					acctest.CheckSDKResourceDisappears(ctx, t, tfdirectconnect.ResourceConnection(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDirectConnectConnection_encryptionMode(t *testing.T) {
	ctx := acctest.Context(t)
	connectionID := acctest.SkipIfEnvVarNotSet(t, "DX_CONNECTION_ID")
	connectionName := acctest.SkipIfEnvVarNotSet(t, "DX_CONNECTION_NAME")

	var connection awstypes.Connection
	resourceName := "aws_dx_connection.test"
	ckn := testAccMacSecGenerateHex()
	cak := testAccMacSecGenerateHex()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config:             testAccConnectionConfig_encryptionModeShouldEncrypt(connectionName, ckn, cak),
				ResourceName:       resourceName,
				ImportState:        true,
				ImportStateId:      connectionID,
				ImportStatePersist: true,
			},
			{
				Config: testAccConnectionConfig_encryptionModeNoEncrypt(connectionName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(ctx, t, resourceName, &connection),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "directconnect", regexache.MustCompile(`dxcon/.+`)),
					resource.TestCheckResourceAttr(resourceName, "encryption_mode", "no_encrypt"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLocation),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, connectionName),
					resource.TestCheckResourceAttr(resourceName, names.AttrSkipDestroy, acctest.CtTrue),
				),
			},
			{
				Config: testAccConnectionConfig_encryptionModeShouldEncrypt(connectionName, ckn, cak),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(ctx, t, resourceName, &connection),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "directconnect", regexache.MustCompile(`dxcon/.+`)),
					resource.TestCheckResourceAttr(resourceName, "encryption_mode", "should_encrypt"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLocation),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, connectionName),
					resource.TestCheckResourceAttr(resourceName, names.AttrSkipDestroy, acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccDirectConnectConnection_macsecRequested(t *testing.T) {
	ctx := acctest.Context(t)
	var connection awstypes.Connection
	resourceName := "aws_dx_connection.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionConfig_macsecEnabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(ctx, t, resourceName, &connection),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "directconnect", regexache.MustCompile(`dxcon/.+`)),
					resource.TestCheckResourceAttr(resourceName, "bandwidth", "100Gbps"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLocation),
					resource.TestCheckResourceAttr(resourceName, "request_macsec", acctest.CtTrue),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrOwnerAccountID),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrProviderName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"request_macsec", names.AttrSkipDestroy},
			},
		},
	})
}

func TestAccDirectConnectConnection_providerName(t *testing.T) {
	ctx := acctest.Context(t)
	var connection awstypes.Connection
	resourceName := "aws_dx_connection.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionConfig_providerName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(ctx, t, resourceName, &connection),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "directconnect", regexache.MustCompile(`dxcon/.+`)),
					resource.TestCheckResourceAttr(resourceName, "bandwidth", "1Gbps"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLocation),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrOwnerAccountID),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrProviderName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"request_macsec", names.AttrSkipDestroy},
			},
		},
	})
}

func TestAccDirectConnectConnection_skipDestroy(t *testing.T) {
	ctx := acctest.Context(t)
	var connection awstypes.Connection
	resourceName := "aws_dx_connection.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionNoDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionConfig_skipDestroy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(ctx, t, resourceName, &connection),
					resource.TestCheckResourceAttr(resourceName, names.AttrSkipDestroy, acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccDirectConnectConnection_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var connection awstypes.Connection
	resourceName := "aws_dx_connection.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(ctx, t, resourceName, &connection),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"request_macsec", names.AttrSkipDestroy},
			},
			{
				Config: testAccConnectionConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(ctx, t, resourceName, &connection),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccConnectionConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(ctx, t, resourceName, &connection),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/31732.
func TestAccDirectConnectConnection_vlanIDMigration501(t *testing.T) {
	ctx := acctest.Context(t)
	var connection awstypes.Connection
	resourceName := "aws_dx_connection.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.DirectConnectServiceID),
		CheckDestroy: testAccCheckConnectionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				// At v5.0.1 the resource's schema is v0 and vlan_id is TypeString.
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.0.1",
					},
				},
				Config: testAccConnectionConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectionExists(ctx, t, resourceName, &connection),
					resource.TestCheckResourceAttr(resourceName, "vlan_id", ""),
				),
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccConnectionConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectionExists(ctx, t, resourceName, &connection),
					resource.TestCheckResourceAttr(resourceName, "vlan_id", "0"),
				),
			},
		},
	})
}

func TestAccDirectConnectConnection_vlanIDMigration510(t *testing.T) {
	ctx := acctest.Context(t)
	var connection awstypes.Connection
	resourceName := "aws_dx_connection.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.DirectConnectServiceID),
		CheckDestroy: testAccCheckConnectionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				// At v5.1.0 the resource's schema is v0 and vlan_id is TypeInt.
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.1.0",
					},
				},
				Config: testAccConnectionConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectionExists(ctx, t, resourceName, &connection),
					resource.TestCheckResourceAttr(resourceName, "vlan_id", "0"),
				),
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccConnectionConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectionExists(ctx, t, resourceName, &connection),
					resource.TestCheckResourceAttr(resourceName, "vlan_id", "0"),
				),
			},
		},
	})
}

func testAccCheckConnectionDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DirectConnectClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dx_connection" {
				continue
			}

			_, err := tfdirectconnect.FindConnectionByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Direct Connect Connection %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckConnectionExists(ctx context.Context, t *testing.T, n string, v *awstypes.Connection) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).DirectConnectClient(ctx)

		output, err := tfdirectconnect.FindConnectionByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckConnectionNoDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DirectConnectClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dx_connection" {
				continue
			}

			_, err := tfdirectconnect.FindConnectionByID(ctx, conn, rs.Primary.ID)

			return err
		}

		return nil
	}
}

func testAccConnectionConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_dx_locations" "test" {}

locals {
  location_codes = tolist(data.aws_dx_locations.test.location_codes)
  idx            = min(2, length(local.location_codes) - 1)
}

resource "aws_dx_connection" "test" {
  name      = %[1]q
  bandwidth = "1Gbps"
  location  = local.location_codes[local.idx]
}
`, rName)
}

func testAccConnectionConfig_encryptionModeNoEncrypt(rName string) string {
	return fmt.Sprintf(`
resource "aws_dx_connection" "test" {
  name            = %[1]q
  location        = "CSOW"
  bandwidth       = "100Gbps"
  encryption_mode = "no_encrypt"
  skip_destroy    = true
}
`, rName)
}

func testAccConnectionConfig_encryptionModeShouldEncrypt(rName, ckn, cak string) string {
	return fmt.Sprintf(`
resource "aws_dx_connection" "test" {
  name            = %[1]q
  location        = "CSOW"
  bandwidth       = "100Gbps"
  encryption_mode = "should_encrypt"
  skip_destroy    = true
}

resource "aws_dx_macsec_key_association" "test" {
  connection_id = aws_dx_connection.test.id
  ckn           = %[2]q
  cak           = %[3]q
}
`, rName, ckn, cak)
}

func testAccConnectionConfig_macsecEnabled(rName string) string {
	return fmt.Sprintf(`
data "aws_dx_locations" "test" {}

locals {
  location_codes = tolist(data.aws_dx_locations.test.location_codes)
  idx            = min(2, length(local.location_codes) - 1)
}

data "aws_dx_location" "test" {
  location_code = local.location_codes[local.idx]
}

resource "aws_dx_connection" "test" {
  name           = %[1]q
  bandwidth      = "100Gbps"
  location       = data.aws_dx_location.test.location_code
  request_macsec = true

  provider_name = data.aws_dx_location.test.available_providers[0]
}
`, rName)
}

func testAccConnectionConfig_providerName(rName string) string {
	return fmt.Sprintf(`
data "aws_dx_locations" "test" {}

locals {
  location_codes = tolist(data.aws_dx_locations.test.location_codes)
  idx            = min(2, length(local.location_codes) - 1)
}

data "aws_dx_location" "test" {
  location_code = local.location_codes[local.idx]
}

resource "aws_dx_connection" "test" {
  name      = %[1]q
  bandwidth = "1Gbps"
  location  = data.aws_dx_location.test.location_code

  provider_name = data.aws_dx_location.test.available_providers[0]
}
`, rName)
}

func testAccConnectionConfig_skipDestroy(rName string) string {
	return fmt.Sprintf(`
data "aws_dx_locations" "test" {}

locals {
  location_codes = tolist(data.aws_dx_locations.test.location_codes)
  idx            = min(2, length(local.location_codes) - 1)
}

resource "aws_dx_connection" "test" {
  name         = %[1]q
  bandwidth    = "1Gbps"
  location     = local.location_codes[local.idx]
  skip_destroy = true
}
	`, rName)
}

func testAccConnectionConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
data "aws_dx_locations" "test" {}

locals {
  location_codes = tolist(data.aws_dx_locations.test.location_codes)
  idx            = min(2, length(local.location_codes) - 1)
}

resource "aws_dx_connection" "test" {
  name      = %[1]q
  bandwidth = "1Gbps"
  location  = local.location_codes[local.idx]

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccConnectionConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
data "aws_dx_locations" "test" {}

locals {
  location_codes = tolist(data.aws_dx_locations.test.location_codes)
  idx            = min(2, length(local.location_codes) - 1)
}

resource "aws_dx_connection" "test" {
  name      = %[1]q
  bandwidth = "1Gbps"
  location  = local.location_codes[local.idx]

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
