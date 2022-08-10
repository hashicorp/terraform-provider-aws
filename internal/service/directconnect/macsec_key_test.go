package directconnect_test

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

// func TestAccDirectConnectMacSecKey_withSecret(t *testing.T) {
// 	var connection directconnect.Connection
// 	resourceName := "aws_dx_macsec_key.test"
// 	dxResourceName := "aws_dx_connection.test"
// 	ckn := testAccDirecConnectMacSecGenerateHex()
// 	cak := testAccDirecConnectMacSecGenerateHex()
// 	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

// 	resource.ParallelTest(t, resource.TestCase{
// 		PreCheck:          func() { acctest.PreCheck(t) },
// 		ErrorCheck:        acctest.ErrorCheck(t, directconnect.EndpointsID),
// 		ProviderFactories: acctest.ProviderFactories,
// 		CheckDestroy:      testAccCheckConnectionDestroy,
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccDirectConnectMacSecConfig_withSecret(rName, ckn, cak),
// 				Check: resource.ComposeTestCheckFunc(
// 					testAccCheckConnectionExists(dxResourceName, &connection),
// 					// TODO: check that DX connection ID and MacSec connection ID match
// 					// resource.TestCheckResourceAttrPair(resourceName, "connection_id", dxResourceName, "id"),
// 					// TODO: check that MacSecKey exists on DX connection
// 					// resource.TestMatchResourceAttr(dxResourceName, "macsec_keys", regexp.MustCompile(``)),
// 				),
// 			},
// 			// Test import.
// 			{
// 				ResourceName:      resourceName,
// 				ImportState:       true,
// 				ImportStateVerify: true,
// 			},
// 		},
// 	})
// }

func TestAccDirectConnectMacSecKey_withCkn(t *testing.T) {
	// Requires an existing MACsec-capable DX connection set as environmental variable
	key := "DX_CONNECTION_ID"
	connectionId := os.Getenv(key)
	if connectionId == "" {
		t.Skipf("Environment variable %s is not set", key)
	}
	resourceName := "aws_dx_macsec_key.test"
	ckn := testAccDirecConnectMacSecGenerateHex()
	cak := testAccDirecConnectMacSecGenerateHex()
	// rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		// ErrorCheck:        acctest.ErrorCheck(t, directconnect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccDirectConnectMacSecConfig_withCkn(ckn, cak, connectionId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "connection_id", connectionId),
					resource.TestMatchResourceAttr(resourceName, "ckn", regexp.MustCompile(ckn)),
				),
			},
		},
	})
}

func TestAccDirectConnectMacSecKey_withSecret(t *testing.T) {
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

	resourceName := "aws_dx_macsec_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		// ErrorCheck:        acctest.ErrorCheck(t, directconnect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccDirectConnectMacSecConfig_withSecret(secretArn, connectionId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "connection_id", connectionId),
					resource.TestCheckResourceAttr(resourceName, "secret_arn", secretArn),
				),
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

func testAccDirectConnectMacSecConfig_withCkn(ckn, cak, connectionId string) string {
	conf := fmt.Sprintf(`
resource "aws_dx_macsec_key" "test" {
  connection_id = %[3]q
  ckn = %[1]q
  cak = %[2]q
}

`, ckn, cak, connectionId)
	fmt.Println(conf)
	return conf
}

// Can only be used with an EXISTING secrets - cannot create secrets from scratch
func testAccDirectConnectMacSecConfig_withSecret(secretArn, connectionId string) string {
	conf := fmt.Sprintf(`
data "aws_secretsmanager_secret" "test" {
  arn = %[1]q
}

resource "aws_dx_macsec_key" "test" {
  connection_id = %[2]q
  secret_arn = data.aws_secretsmanager_secret.test.arn
}

`, secretArn, connectionId)
	fmt.Println(conf)
	return conf
}
