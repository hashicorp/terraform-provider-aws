// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package directconnect_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/service/directconnect"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDirectConnectConnection_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var connection directconnect.Connection
	resourceName := "aws_dx_connection.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectionExists(ctx, resourceName, &connection),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "directconnect", regexache.MustCompile(`dxcon/.+`)),
					resource.TestCheckResourceAttr(resourceName, "bandwidth", "1Gbps"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLocation),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerAccountID),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "partner_name", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrProviderName, ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "vlan_id", acctest.Ct0),
				),
			},
			// Test import.
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
	var connection directconnect.Connection
	resourceName := "aws_dx_connection.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(ctx, resourceName, &connection),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfdirectconnect.ResourceConnection(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDirectConnectConnection_encryptionMode(t *testing.T) {
	ctx := acctest.Context(t)
	dxKey := "DX_CONNECTION_ID"
	connectionId := os.Getenv(dxKey)
	if connectionId == "" {
		t.Skipf("Environment variable %s is not set", dxKey)
	}

	dxName := "DX_CONNECTION_NAME"
	connectionName := os.Getenv(dxName)
	if connectionName == "" {
		t.Skipf("Environment variable %s is not set", dxName)
	}

	var connection directconnect.Connection
	resourceName := "aws_dx_connection.test"
	ckn := testAccDirecConnectMacSecGenerateHex()
	cak := testAccDirecConnectMacSecGenerateHex()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config:             testAccConnectionConfig_encryptionModeShouldEncrypt(connectionName, ckn, cak),
				ResourceName:       resourceName,
				ImportState:        true,
				ImportStateId:      connectionId,
				ImportStatePersist: true,
			},
			{
				Config: testAccConnectionConfig_encryptionModeNoEncrypt(connectionName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(ctx, resourceName, &connection),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "directconnect", regexache.MustCompile(`dxcon/.+`)),
					resource.TestCheckResourceAttr(resourceName, "encryption_mode", "no_encrypt"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLocation),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, connectionName),
					resource.TestCheckResourceAttr(resourceName, names.AttrSkipDestroy, acctest.CtTrue),
				),
			},
			{
				Config: testAccConnectionConfig_encryptionModeShouldEncrypt(connectionName, ckn, cak),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(ctx, resourceName, &connection),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "directconnect", regexache.MustCompile(`dxcon/.+`)),
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
	var connection directconnect.Connection
	resourceName := "aws_dx_connection.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionConfig_macsecEnabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(ctx, resourceName, &connection),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "directconnect", regexache.MustCompile(`dxcon/.+`)),
					resource.TestCheckResourceAttr(resourceName, "bandwidth", "100Gbps"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLocation),
					resource.TestCheckResourceAttr(resourceName, "request_macsec", acctest.CtTrue),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerAccountID),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrProviderName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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
	var connection directconnect.Connection
	resourceName := "aws_dx_connection.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionConfig_providerName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(ctx, resourceName, &connection),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "directconnect", regexache.MustCompile(`dxcon/.+`)),
					resource.TestCheckResourceAttr(resourceName, "bandwidth", "1Gbps"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLocation),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerAccountID),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrProviderName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			// Test import.
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
	var connection directconnect.Connection
	resourceName := "aws_dx_connection.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionNoDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionConfig_skipDestroy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(ctx, resourceName, &connection),
					resource.TestCheckResourceAttr(resourceName, names.AttrSkipDestroy, acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccDirectConnectConnection_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var connection directconnect.Connection
	resourceName := "aws_dx_connection.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(ctx, resourceName, &connection),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			// Test import.
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"request_macsec", names.AttrSkipDestroy},
			},
			{
				Config: testAccConnectionConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(ctx, resourceName, &connection),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccConnectionConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(ctx, resourceName, &connection),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/31732.
func TestAccDirectConnectConnection_vlanIDMigration501(t *testing.T) {
	ctx := acctest.Context(t)
	var connection directconnect.Connection
	resourceName := "aws_dx_connection.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.DirectConnectServiceID),
		CheckDestroy: testAccCheckConnectionDestroy(ctx),
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
					testAccCheckConnectionExists(ctx, resourceName, &connection),
					resource.TestCheckResourceAttr(resourceName, "vlan_id", ""),
				),
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccConnectionConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectionExists(ctx, resourceName, &connection),
					resource.TestCheckResourceAttr(resourceName, "vlan_id", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccDirectConnectConnection_vlanIDMigration510(t *testing.T) {
	ctx := acctest.Context(t)
	var connection directconnect.Connection
	resourceName := "aws_dx_connection.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.DirectConnectServiceID),
		CheckDestroy: testAccCheckConnectionDestroy(ctx),
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
					testAccCheckConnectionExists(ctx, resourceName, &connection),
					resource.TestCheckResourceAttr(resourceName, "vlan_id", acctest.Ct0),
				),
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccConnectionConfig_basic(rName),
				PlanOnly:                 true,
			},
		},
	})
}

func testAccCheckConnectionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DirectConnectConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dx_connection" {
				continue
			}

			_, err := tfdirectconnect.FindConnectionByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
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

func testAccCheckConnectionExists(ctx context.Context, name string, v *directconnect.Connection) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DirectConnectConn(ctx)

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Direct Connect Connection ID is set")
		}

		connection, err := tfdirectconnect.FindConnectionByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *connection

		return nil
	}
}

func testAccCheckConnectionNoDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DirectConnectConn(ctx)

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
