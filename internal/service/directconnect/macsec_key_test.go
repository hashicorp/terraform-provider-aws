// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package directconnect_test

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDirectConnectMacSecKey_withCkn(t *testing.T) {
	ctx := acctest.Context(t)
	// Requires an existing MACsec-capable DX connection set as environmental variable
	key := "DX_CONNECTION_ID"
	connectionId := os.Getenv(key)
	if connectionId == "" {
		t.Skipf("Environment variable %s is not set", key)
	}
	resourceName := "aws_dx_macsec_key_association.test"
	ckn := testAccDirecConnectMacSecGenerateHex()
	cak := testAccDirecConnectMacSecGenerateHex()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccMacSecConfig_withCkn(ckn, cak, connectionId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrConnectionID, connectionId),
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

func TestAccDirectConnectMacSecKey_withSecret(t *testing.T) {
	ctx := acctest.Context(t)
	// Requires an existing MACsec-capable DX connection set as environmental variable
	dxKey := "DX_CONNECTION_ID"
	connectionId := os.Getenv(dxKey)
	if connectionId == "" {
		t.Skipf("Environment variable %s is not set", dxKey)
	}

	secretKey := "SECRET_ARN"
	secretArn := os.Getenv(secretKey)
	if secretArn == "" {
		t.Skipf("Environment variable %s is not set", secretKey)
	}

	resourceName := "aws_dx_macsec_key_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccMacSecConfig_withSecret(secretArn, connectionId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrConnectionID, connectionId),
					resource.TestCheckResourceAttr(resourceName, "secret_arn", secretArn),
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

// testAccDirecConnectMacSecGenerateKey generates a 64-character hex string to be used as CKN or CAK
func testAccDirecConnectMacSecGenerateHex() string {
	s := make([]byte, 32)
	if _, err := rand.Read(s); err != nil {
		return ""
	}
	return hex.EncodeToString(s)
}

func testAccMacSecConfig_withCkn(ckn, cak, connectionId string) string {
	return fmt.Sprintf(`
resource "aws_dx_macsec_key_association" "test" {
  connection_id = %[3]q
  ckn           = %[1]q
  cak           = %[2]q
}


`, ckn, cak, connectionId)
}

// Can only be used with an EXISTING secrets created by previous association - cannot create secrets from scratch
func testAccMacSecConfig_withSecret(secretArn, connectionId string) string {
	return fmt.Sprintf(`
data "aws_secretsmanager_secret" "test" {
  arn = %[1]q
}

resource "aws_dx_macsec_key_association" "test" {
  connection_id = %[2]q
  secret_arn    = data.aws_secretsmanager_secret.test.arn
}


`, secretArn, connectionId)
}
