// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package directconnect_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDirectConnectMacSecKeyAssociation_withCkn(t *testing.T) {
	ctx := acctest.Context(t)
	connectionID := acctest.SkipIfEnvVarNotSet(t, "DX_CONNECTION_ID")
	resourceName := "aws_dx_macsec_key_association.test"
	ckn := testAccMacSecGenerateHex()
	cak := testAccMacSecGenerateHex()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMacSecKeyAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMacSecKeyAssociationConfig_withCkn(ckn, cak, connectionID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMacSecKeyAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrConnectionID, connectionID),
					resource.TestMatchResourceAttr(resourceName, "ckn", regexache.MustCompile(ckn)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// Ignore the "cak" attribute as isn't returned by the API during read/refresh
				ImportStateVerifyIgnore: []string{"cak"},
			},
		},
	})
}

func TestAccDirectConnectMacSecKeyAssociation_withSecret(t *testing.T) {
	ctx := acctest.Context(t)
	connectionID := acctest.SkipIfEnvVarNotSet(t, "DX_CONNECTION_ID")
	secretARN := acctest.SkipIfEnvVarNotSet(t, "SECRET_ARN")
	resourceName := "aws_dx_macsec_key_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMacSecKeyAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMacSecKeyAssociationConfig_withSecret(secretARN, connectionID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMacSecKeyAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrConnectionID, connectionID),
					resource.TestCheckResourceAttr(resourceName, "secret_arn", secretARN),
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

func testAccMacSecGenerateHex() string {
	s := make([]byte, 32)
	if _, err := rand.Read(s); err != nil {
		return ""
	}
	return hex.EncodeToString(s)
}

func testAccCheckMacSecKeyAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DirectConnectClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dx_macsec_key_association" {
				continue
			}

			_, err := tfdirectconnect.FindMacSecKeyByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrConnectionID], rs.Primary.Attributes["secret_arn"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Direct Connect MACSec Key Association %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckMacSecKeyAssociationExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DirectConnectClient(ctx)

		_, err := tfdirectconnect.FindMacSecKeyByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrConnectionID], rs.Primary.Attributes["secret_arn"])

		return err
	}
}

func testAccMacSecKeyAssociationConfig_withCkn(ckn, cak, connectionID string) string {
	return fmt.Sprintf(`
resource "aws_dx_macsec_key_association" "test" {
  connection_id = %[3]q
  ckn           = %[1]q
  cak           = %[2]q
}
`, ckn, cak, connectionID)
}

// Can only be used with an EXISTING secrets created by previous association - cannot create secrets from scratch.
func testAccMacSecKeyAssociationConfig_withSecret(secretARN, connectionID string) string {
	return fmt.Sprintf(`
data "aws_secretsmanager_secret" "test" {
  arn = %[1]q
}

resource "aws_dx_macsec_key_association" "test" {
  connection_id = %[2]q
  secret_arn    = data.aws_secretsmanager_secret.test.arn
}
`, secretARN, connectionID)
}
