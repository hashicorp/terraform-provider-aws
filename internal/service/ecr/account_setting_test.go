// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecr_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfecr "github.com/hashicorp/terraform-provider-aws/internal/service/ecr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccECRAccountSetting_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:       testAccAccountSetting_basic,
		"blobMounting":        testAccAccountSetting_blobMounting,
		"registryPolicyScope": testAccAccountSetting_registryPolicyScope,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccAccountSetting_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ecr_account_setting.test"
	rName := "BASIC_SCAN_TYPE_VERSION"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountSettingConfig_basic(rName, "AWS_NATIVE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountSettingExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrValue, "AWS_NATIVE")),
			},
			{
				ResourceName:                         resourceName,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportStateId:                        rName,
				ImportState:                          true,
				ImportStateVerify:                    true,
			},
			{
				Config: testAccAccountSettingConfig_basic(rName, "CLAIR"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountSettingExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrValue, "CLAIR")),
			},
		},
	})
}

func testAccAccountSetting_registryPolicyScope(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ecr_account_setting.test"
	rName := "REGISTRY_POLICY_SCOPE"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountSettingConfig_basic(rName, "V1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountSettingExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrValue, "V1")),
			},
			{
				ResourceName:                         resourceName,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportStateId:                        rName,
				ImportState:                          true,
				ImportStateVerify:                    true,
			},
			{
				Config: testAccAccountSettingConfig_basic(rName, "V2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountSettingExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrValue, "V2")),
			},
		},
	})
}

func testAccAccountSetting_blobMounting(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ecr_account_setting.test"
	rName := "BLOB_MOUNTING"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountSettingConfig_basic(rName, "ENABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountSettingExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrValue, "ENABLED")),
			},
			{
				ResourceName:                         resourceName,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportStateId:                        rName,
				ImportState:                          true,
				ImportStateVerify:                    true,
			},
			{
				Config: testAccAccountSettingConfig_basic(rName, "DISABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountSettingExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrValue, "DISABLED")),
			},
		},
	})
}

func testAccCheckAccountSettingExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ECRClient(ctx)

		_, err := tfecr.FindAccountSettingByName(ctx, conn, rs.Primary.Attributes[names.AttrName])

		return err
	}
}

func testAccAccountSettingConfig_basic(name, value string) string {
	return fmt.Sprintf(`
resource "aws_ecr_account_setting" "test" {
  name  = %[1]q
  value = %[2]q
}
`, name, value)
}
